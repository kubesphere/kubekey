package executor

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	kkcorev1alpha1 "github.com/kubesphere/kubekey/api/core/v1alpha1"
	"github.com/schollz/progressbar/v3"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/converter"
	"github.com/kubesphere/kubekey/v4/pkg/converter/tmpl"
	"github.com/kubesphere/kubekey/v4/pkg/modules"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

// taskExecutor handles the execution of a single task across multiple hosts.
type taskExecutor struct {
	*option
	task *kkcorev1alpha1.Task
	// taskRunTimeout is the timeout for task executor
	taskRunTimeout time.Duration
}

// Exec creates and executes a task, updating its status and the parent playbook's status.
// It returns an error if the task creation or execution fails.
func (e *taskExecutor) Exec(ctx context.Context) error {
	if e.taskRunTimeout == time.Duration(0) {
		e.taskRunTimeout = 60 * time.Minute
	}
	// create task
	if err := e.client.Create(ctx, e.task); err != nil {
		return errors.Wrapf(err, "failed to create task %v", e.task)
	}
	defer func() {
		e.playbook.Status.Statistics.Total++
		switch e.task.Status.Phase {
		case kkcorev1alpha1.TaskPhaseSuccess:
			e.playbook.Status.Statistics.Success++
		case kkcorev1alpha1.TaskPhaseIgnored:
			e.playbook.Status.Statistics.Ignored++
		case kkcorev1alpha1.TaskPhaseFailed:
			e.playbook.Status.Statistics.Failed++
		}
	}()
	// run task
	if err := e.runTaskLoop(ctx); err != nil {
		return err
	}
	// exit when task run failed
	if e.task.IsFailed() {
		failedMsg := "\n"

		for _, result := range e.task.Status.HostResults {
			// 1. Print executor-level (host-level) error first, if exists
			if strings.TrimSpace(result.Error) != "" {
				failedMsg += fmt.Sprintf(
					"[%s][executor]: %s\n",
					result.Host,
					result.Error,
				)
			}

			// 2. Then print item-level errors (only items with error)
			for idx, r := range result.LoopResults {
				if strings.TrimSpace(r.Error) == "" {
					continue
				}

				itemInfo := "item=<nil>"
				if len(r.Item.Raw) > 0 {
					itemInfo = "item=" + string(r.Item.Raw)
				} else if r.Item.Object != nil {
					itemInfo = fmt.Sprintf("item=%#v", r.Item.Object)
				}

				failedMsg += fmt.Sprintf(
					"[%s][%s][%d]: \nstdout: %s\nstderr: %s\nerror: %s\n",
					result.Host,
					itemInfo,
					idx,
					r.Stdout,
					r.Stderr,
					r.Error,
				)
			}
		}

		return errors.Errorf(
			"task [%s](%s) run failed: %s",
			e.task.Spec.Name,
			ctrlclient.ObjectKeyFromObject(e.task),
			failedMsg,
		)
	}

	return nil
}

// runTaskLoop runs a task in a loop until it completes or times out.
// It periodically reconciles the task status and executes the task when it enters the running phase.
func (e *taskExecutor) runTaskLoop(ctx context.Context) error {
	klog.V(3).InfoS("begin run task", "task", ctrlclient.ObjectKeyFromObject(e.task))
	defer klog.V(3).InfoS("end run task", "task", ctrlclient.ObjectKeyFromObject(e.task))

	// Add role prefix to log output if role annotation exists
	var roleLog string
	if e.task.Annotations[kkcorev1alpha1.TaskAnnotationRelativePath] != "" {
		roleLog = "[" + e.task.Annotations[kkcorev1alpha1.TaskAnnotationRelativePath] + "] "
	}
	fmt.Fprintf(e.logOutput, "%s %s%s\n", time.Now().Format(time.TimeOnly+" MST"), roleLog, e.task.Spec.Name)

	for !e.task.IsComplete() {
		task := e.task.DeepCopy()
		if e.task.Status.Phase == kkcorev1alpha1.TaskPhaseFailed {
			e.task.Status.RestartCount++
		}
		e.task.Status.Phase = kkcorev1alpha1.TaskPhaseRunning
		if err := e.client.Status().Patch(ctx, e.task, ctrlclient.MergeFrom(task)); err != nil {
			return errors.Wrapf(err, "failed to patch task status of %s", ctrlclient.ObjectKeyFromObject(task))
		}
		task = e.task.DeepCopy()
		e.execTask(ctx)
		if err := e.client.Status().Patch(ctx, e.task, ctrlclient.MergeFrom(task)); err != nil {
			return errors.Wrapf(err, "failed to patch task status of %s", ctrlclient.ObjectKeyFromObject(task))
		}
	}
	return nil
}

// execTask executes the task across all specified hosts in parallel and updates the task status.
func (e *taskExecutor) execTask(ctx context.Context) {
	// check task host results
	wg := &wait.Group{}
	e.task.Status.HostResults = make([]kkcorev1alpha1.TaskHostResult, len(e.task.Spec.Hosts))
	for i, h := range e.task.Spec.Hosts {
		wg.StartWithContext(ctx, e.execTaskHost(i, h))
	}
	wg.Wait()
	// host result for task
	e.task.Status.Phase = kkcorev1alpha1.TaskPhaseSuccess
	for _, data := range e.task.Status.HostResults {
		if data.Error != "" {
			if e.task.Spec.IgnoreError != nil && *e.task.Spec.IgnoreError {
				e.task.Status.Phase = kkcorev1alpha1.TaskPhaseIgnored
			} else {
				e.task.Status.Phase = kkcorev1alpha1.TaskPhaseFailed
			}

			break
		}
	}
}

// execTaskHost handles executing a task on a single host, including variable setup,
// condition checking, and module execution. It runs in parallel for each host.
func (e *taskExecutor) execTaskHost(i int, h string) func(ctx context.Context) {
	return func(ctx context.Context) {
		var resErr error
		var loopResults []kkcorev1alpha1.LoopResult

		// task log
		deferFunc := e.execTaskHostLogs(ctx, h, &loopResults)
		defer deferFunc()

		defer func() {
			resErr = errors.Join(resErr, e.dealRegister(h, loopResults))

			var errMsg string
			if resErr != nil {
				errMsg = resErr.Error()
			}

			e.task.Status.HostResults[i] = kkcorev1alpha1.TaskHostResult{
				Host:        h,
				Error:       errMsg,
				LoopResults: loopResults,
			}
		}()

		ha, err := e.variable.Get(variable.GetAllVariable(h))
		if err != nil {
			resErr = err
			return
		}
		had, ok := ha.(map[string]any)
		if !ok {
			resErr = errors.Errorf("host: %s variable is not a map", h)
			return
		}

		for _, item := range e.dealLoop(had) {
			stdout, stderr, rendered, exeErr := e.executeModule(ctx, e.task, item, h)

			var rawItem runtime.RawExtension
			if rendered != nil {
				if bs, err := json.Marshal(rendered); err == nil {
					rawItem = runtime.RawExtension{Raw: bs}
				}
			}

			var errMsg string
			if exeErr != nil {
				errMsg = exeErr.Error()
			}

			r := kkcorev1alpha1.LoopResult{
				Item:   rawItem,
				Stdout: stdout,
				Stderr: stderr,
				Error:  errMsg,
			}

			loopResults = append(loopResults, r)
		}
	}
}

// execTaskHostLogs sets up and manages progress bar logging for task execution on a host.
// It returns a cleanup function to be called when execution completes.
func (e *taskExecutor) execTaskHostLogs(ctx context.Context, h string, loopResults *[]kkcorev1alpha1.LoopResult) func() {
	// placeholder format task log
	var placeholder string
	if hostnameMaxLen, err := e.variable.Get(variable.GetHostMaxLength()); err == nil {
		if hl, ok := hostnameMaxLen.(int); ok {
			placeholder = strings.Repeat(" ", hl-len(h))
		}
	}
	// progress bar for task
	options := []progressbar.Option{
		progressbar.OptionSetWriter(os.Stdout),
		// progressbar.OptionSpinnerCustom([]string{"            "}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetDescription(fmt.Sprintf("[\033[36m%s\033[0m]%s \033[36mrunning\033[0m", h, placeholder)),
		progressbar.OptionOnCompletion(func() {
			if _, err := os.Stdout.WriteString("\n"); err != nil {
				klog.ErrorS(err, "failed to write output", "host", h)
			}
		}),
	}
	if e.logOutput != os.Stdout {
		options = append(options, progressbar.OptionSetVisibility(false))
	}
	bar := progressbar.NewOptions(-1, options...)
	// run progress
	go func() {
		if err := wait.PollUntilContextCancel(ctx, 100*time.Millisecond, true, func(context.Context) (bool, error) {
			if bar.IsFinished() {
				return true, nil
			}
			if err := bar.Add(1); err != nil {
				return false, errors.Wrap(err, "failed to process bar")
			}

			return false, nil
		}); err != nil {
			klog.ErrorS(err, "failed to wait for task run to finish", "host", h)
		}
	}()

	return func() {
		var failed bool
		var skipped bool

		// determine overall status by scanning all register results
		skipped = true // assume skip until we find a non-skip stdout
		for _, r := range *loopResults {
			if r.Error != "" {
				failed = true
				break
			}
			if r.Stdout != modules.StdoutSkip {
				skipped = false
			}
		}

		switch {
		case failed:
			// failed or ignore
			if e.task.Spec.IgnoreError != nil && *e.task.Spec.IgnoreError {
				// ignore
				bar.Describe(fmt.Sprintf("[\033[36m%s\033[0m]%s \033[34mignore \033[0m", h, placeholder))
				if e.logOutput != os.Stdout {
					fmt.Fprintf(e.logOutput, "[%s]%s ignore \n", h, placeholder)
				}
			} else {
				// failed
				bar.Describe(fmt.Sprintf("[\033[36m%s\033[0m]%s \033[31mfailed \033[0m", h, placeholder))
				if e.logOutput != os.Stdout {
					fmt.Fprintf(e.logOutput, "[%s]%s failed \n", h, placeholder)
				}
			}
		case skipped:
			// skip
			bar.Describe(fmt.Sprintf("[\033[36m%s\033[0m]%s \033[34mskip   \033[0m", h, placeholder))
			if e.logOutput != os.Stdout {
				fmt.Fprintf(e.logOutput, "[%s]%s skip   \n", h, placeholder)
			}
		default:
			// success
			bar.Describe(fmt.Sprintf("[\033[36m%s\033[0m]%s \033[34msuccess\033[0m", h, placeholder))
			if e.logOutput != os.Stdout {
				fmt.Fprintf(e.logOutput, "[%s]%s success\n", h, placeholder)
			}
		}

		_ = bar.Finish()
	}
}

// executeModule executes a single module task on a specific host.
func (e *taskExecutor) executeModule(ctx context.Context, task *kkcorev1alpha1.Task, item any, host string) (stdout string, stderr string, rendered any, resErr error) {
	// Set loop item variable if one was provided
	if item != nil {
		// Convert item to runtime variable
		node, err := converter.ConvertMap2Node(map[string]any{_const.VariableItem: item})
		if err != nil {
			return modules.StdoutFailed, "", nil, err
		}

		// Merge item into host's runtime variables
		if err := e.variable.Merge(variable.MergeRuntimeVariable([]yaml.Node{node}, host)); err != nil {
			return modules.StdoutFailed, "", nil, err
		}
		// Clean up loop item variable after execution
		defer func() {
			if item == nil {
				return
			}
			// Reset item to null
			resetNode, err := converter.ConvertMap2Node(map[string]any{_const.VariableItem: nil})
			if err != nil {
				resErr = err
				return
			}
			if err := e.variable.Merge(variable.MergeRuntimeVariable([]yaml.Node{resetNode}, host)); err != nil {
				resErr = err
				return
			}
		}()
	}

	// Get all variables for this host, including any loop item
	ha, err := e.variable.Get(variable.GetAllVariable(host))
	if err != nil {
		return modules.StdoutFailed, "", nil, err
	}

	// Convert host variables to map type
	had, ok := ha.(map[string]any)
	if !ok {
		return modules.StdoutFailed, "", nil, err
	}
	// check when condition
	if skip, err := e.dealWhen(had); err != nil {
		return modules.StdoutFailed, "", nil, err
	} else if skip {
		return modules.StdoutSkip, "", nil, nil
	}

	// Execute the actual module with the prepared context
	stdout, stderr, resErr = modules.FindModule(task.Spec.Module.Name)(ctx, modules.ExecOptions{
		Args:      e.task.Spec.Module.Args,
		Host:      host,
		Variable:  e.variable,
		Task:      *e.task,
		Playbook:  *e.playbook,
		LogOutput: e.logOutput,
	})
	if ferr := e.dealFailedWhen(had, resErr); ferr != nil {
		return stdout, stderr, had[_const.VariableItem], ferr
	}
	return stdout, stderr, had[_const.VariableItem], nil
}

// dealLoop parses the loop specification into a slice of items to iterate over.
// If no loop is specified, returns a single nil item. Otherwise converts the loop
// specification from JSON into a slice of values.
func (e *taskExecutor) dealLoop(ha map[string]any) []any {
	var items []any
	if e.task.Spec.Loop.Raw == nil {
		// loop is not set. add one element to execute once module.
		items = []any{nil}
	} else {
		items = variable.Extension2Slice(ha, e.task.Spec.Loop)
	}

	return items
}

// dealWhen evaluates the "when" conditions for a task to determine if it should be skipped.
// Returns true if the task should be skipped, false if it should proceed.
func (e *taskExecutor) dealWhen(had map[string]any) (bool, error) {
	if len(e.task.Spec.When) > 0 {
		ok, err := tmpl.ParseBool(had, e.task.Spec.When...)
		if err != nil {
			return false, err
		}
		if !ok {
			return true, nil
		}
	}

	return false, nil
}

// dealFailedWhen evaluates the "failed_when" conditions for a task to determine if it should fail.
// Returns true if the task should be marked as failed, false if it should proceed.
func (e *taskExecutor) dealFailedWhen(had map[string]any, err error) error {
	if err != nil {
		return err
	}
	if len(e.task.Spec.FailedWhen) > 0 {
		ok, err := tmpl.ParseBool(had, e.task.Spec.FailedWhen...)
		if err != nil {
			return errors.Wrap(err, "failed to parse failed_when condition")
		}
		if ok {
			return errors.New("reach failed_when, failed")
		}
	}
	return nil
}

// dealRegister merges loopResults into global variables for the given host and processes "register" logic.
// If the task specifies a Register name, it composes the values from the loopResults slice,
// normalizes stdout data according to the RegisterType (json, yaml, or plain string),
// detects errors in any of the items, and merges the composed data into the task's runtime variables.
// It returns an error if any items are in error or if variable merge fails.
func (e *taskExecutor) dealRegister(host string, loopResults []kkcorev1alpha1.LoopResult) (resErr error) {
	// parseStdout parses stdout according to the RegisterType.
	parseStdout := func(s string) any {
		var out any = s

		switch e.task.Spec.RegisterType {
		case "json":
			// Attempt to unmarshal as JSON.
			if err := json.Unmarshal([]byte(s), &out); err != nil {
				klog.V(5).ErrorS(err, "failed to register json value")
				return s
			}
		case "yaml", "yml":
			// Attempt to unmarshal as YAML.
			if err := yaml.Unmarshal([]byte(s), &out); err != nil {
				klog.V(5).ErrorS(err, "failed to register yaml value")
				return s
			}
		default:
			// Remove trailing newline by default.
			if str, ok := out.(string); ok {
				out = strings.TrimRight(str, "\n")
			}
		}

		return out
	}

	var value any
	var hasItemError bool

	// If there is exactly one loopResults with no Item data, use the flat representation.
	if len(loopResults) == 1 && len(loopResults[0].Item.Raw) == 0 && loopResults[0].Item.Object == nil {
		r := loopResults[0]
		value = map[string]any{
			"stdout": parseStdout(r.Stdout),
			"stderr": r.Stderr,
			"error":  r.Error,
		}

		// If there is any error at the module level, set the global error flag.
		hasItemError = hasItemError || strings.TrimSpace(r.Error) != ""
	} else {
		// Otherwise, collect all items as an array of results.
		var arr []any
		for _, r := range loopResults {
			arr = append(arr, map[string]any{
				"item":   string(r.Item.Raw),
				"stdout": parseStdout(r.Stdout),
				"stderr": r.Stderr,
				"error":  r.Error,
			})

			// If any item has error, set the global error flag.
			hasItemError = hasItemError || strings.TrimSpace(r.Error) != ""
		}
		value = arr
	}

	// If any item failed, return a unified task failure error sentinel.
	if hasItemError {
		resErr = errors.Join(resErr, errors.New("module run failed"))
	}

	if e.task.Spec.Register != "" {
		// Convert the register mapping to a YAML node.
		node, err := converter.ConvertMap2Node(map[string]any{
			e.task.Spec.Register: value,
		})
		if err != nil {
			return errors.Join(resErr, err)
		}
		resErr = errors.Join(resErr, e.variable.Merge(variable.MergeRuntimeVariable([]yaml.Node{node}, host)))
	}

	return resErr
}
