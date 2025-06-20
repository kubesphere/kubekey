package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	kkcorev1alpha1 "github.com/kubesphere/kubekey/api/core/v1alpha1"
	"github.com/schollz/progressbar/v3"
	"gopkg.in/yaml.v3"
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
		return errors.Wrapf(err, "failed to create task %q", e.task.Spec.Name)
	}
	defer func() {
		e.playbook.Status.TaskResult.Total++
		switch e.task.Status.Phase {
		case kkcorev1alpha1.TaskPhaseSuccess:
			e.playbook.Status.TaskResult.Success++
		case kkcorev1alpha1.TaskPhaseIgnored:
			e.playbook.Status.TaskResult.Ignored++
		case kkcorev1alpha1.TaskPhaseFailed:
			e.playbook.Status.TaskResult.Failed++
		}
	}()
	// run task
	if err := e.runTaskLoop(ctx); err != nil {
		return err
	}
	// exit when task run failed
	if e.task.IsFailed() {
		var hostReason []kkcorev1.PlaybookFailedDetailHost
		for _, tr := range e.task.Status.HostResults {
			hostReason = append(hostReason, kkcorev1.PlaybookFailedDetailHost{
				Host:   tr.Host,
				Stdout: tr.Stdout,
				StdErr: tr.StdErr,
			})
		}
		e.playbook.Status.FailedDetail = append(e.playbook.Status.FailedDetail, kkcorev1.PlaybookFailedDetail{
			Task:  e.task.Spec.Name,
			Hosts: hostReason,
		})
		e.playbook.Status.Phase = kkcorev1.PlaybookPhaseFailed

		return errors.Errorf("task %q run failed", e.task.Spec.Name)
	}

	return nil
}

// runTaskLoop runs a task in a loop until it completes or times out.
// It periodically reconciles the task status and executes the task when it enters the running phase.
func (e *taskExecutor) runTaskLoop(ctx context.Context) error {
	klog.V(5).InfoS("begin run task", "task", ctrlclient.ObjectKeyFromObject(e.task))
	defer klog.V(5).InfoS("end run task", "task", ctrlclient.ObjectKeyFromObject(e.task))

	// Add role prefix to log output if role annotation exists
	var roleLog string
	if e.task.Annotations[kkcorev1alpha1.TaskAnnotationRelativePath] != "" {
		roleLog = "[" + e.task.Annotations[kkcorev1alpha1.TaskAnnotationRelativePath] + "] "
	}
	fmt.Fprintf(e.logOutput, "%s %s%s\n", time.Now().Format(time.TimeOnly+" MST"), roleLog, e.task.Spec.Name)

	for {
		if e.task.IsComplete() {
			break
		}
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
		if data.StdErr != "" {
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
		// task result
		var stdout, stderr string
		defer func() {
			if err := e.dealRegister(stdout, stderr, h); err != nil {
				stderr = err.Error()
			}
			if stderr != "" && e.task.Spec.IgnoreError != nil && *e.task.Spec.IgnoreError {
				klog.V(5).ErrorS(nil, "task run failed", "host", h, "stdout", stdout, "stderr", stderr, "task", ctrlclient.ObjectKeyFromObject(e.task))
			} else if stderr != "" {
				klog.ErrorS(nil, "task run failed", "host", h, "stdout", stdout, "stderr", stderr, "task", ctrlclient.ObjectKeyFromObject(e.task))
			}
			// fill result
			e.task.Status.HostResults[i] = kkcorev1alpha1.TaskHostResult{
				Host:   h,
				Stdout: stdout,
				StdErr: stderr,
			}
		}()
		// task log
		deferFunc := e.execTaskHostLogs(ctx, h, &stdout, &stderr)
		defer deferFunc()
		// task execute
		ha, err := e.variable.Get(variable.GetAllVariable(h))
		if err != nil {
			stderr = fmt.Sprintf("failed to get host %s variable: %v", h, err)

			return
		}
		// convert hostVariable to map
		had, ok := ha.(map[string]any)
		if !ok {
			stderr = fmt.Sprintf("host: %s variable is not a map", h)
		}
		// check when condition
		if skip := e.dealWhen(had, &stdout, &stderr); skip {
			return
		}
		// execute module in loop with loop item.
		// if loop is empty. execute once, and the item is null
		for _, item := range e.dealLoop(had) {
			e.executeModule(ctx, e.task, item, h, &stdout, &stderr)
		}
	}
}

// execTaskHostLogs sets up and manages progress bar logging for task execution on a host.
// It returns a cleanup function to be called when execution completes.
func (e *taskExecutor) execTaskHostLogs(ctx context.Context, h string, stdout, stderr *string) func() {
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
		err := wait.PollUntilContextCancel(ctx, 100*time.Millisecond, true, func(context.Context) (bool, error) {
			if bar.IsFinished() {
				return true, nil
			}
			if err := bar.Add(1); err != nil {
				return false, errors.Wrap(err, "failed to process bar")
			}

			return false, nil
		})
		if err != nil {
			klog.ErrorS(err, "failed to wait for task run to finish", "host", h)
		}
	}()

	return func() {
		switch {
		case *stderr != "":
			if e.task.Spec.IgnoreError != nil && *e.task.Spec.IgnoreError { // ignore
				bar.Describe(fmt.Sprintf("[\033[36m%s\033[0m]%s \033[34mignore \033[0m", h, placeholder))
				if e.logOutput != os.Stdout {
					fmt.Fprintf(e.logOutput, "[%s]%s ignore \n", h, placeholder)
				}
			} else { // failed
				bar.Describe(fmt.Sprintf("[\033[36m%s\033[0m]%s \033[31mfailed \033[0m", h, placeholder))
				if e.logOutput != os.Stdout {
					fmt.Fprintf(e.logOutput, "[%s]%s failed \n", h, placeholder)
				}
			}
		case *stdout == modules.StdoutSkip: // skip
			bar.Describe(fmt.Sprintf("[\033[36m%s\033[0m]%s \033[34mskip   \033[0m", h, placeholder))
			if e.logOutput != os.Stdout {
				fmt.Fprintf(e.logOutput, "[%s]%s skip   \n", h, placeholder)
			}
		default: //success
			bar.Describe(fmt.Sprintf("[\033[36m%s\033[0m]%s \033[34msuccess\033[0m", h, placeholder))
			if e.logOutput != os.Stdout {
				fmt.Fprintf(e.logOutput, "[%s]%s success\n", h, placeholder)
			}
		}
		if err := bar.Finish(); err != nil {
			klog.ErrorS(err, "finish bar error")
		}
	}
}

// executeModule executes a single module task on a specific host.
func (e *taskExecutor) executeModule(ctx context.Context, task *kkcorev1alpha1.Task, item any, host string, stdout, stderr *string) {
	// Set loop item variable if one was provided
	if item != nil {
		// Convert item to runtime variable
		node, err := converter.ConvertMap2Node(map[string]any{_const.VariableItem: item})
		if err != nil {
			*stderr = fmt.Sprintf("convert loop item error: %v", err)
			return
		}

		// Merge item into host's runtime variables
		if err := e.variable.Merge(variable.MergeRuntimeVariable(node, host)); err != nil {
			*stderr = fmt.Sprintf("set loop item to variable error: %v", err)
			return
		}

		// Clean up loop item variable after execution
		defer func() {
			if item == nil {
				return
			}
			// Reset item to null
			resetNode, err := converter.ConvertMap2Node(map[string]any{_const.VariableItem: nil})
			if err != nil {
				*stderr = fmt.Sprintf("convert loop item error: %v", err)
				return
			}
			if err := e.variable.Merge(variable.MergeRuntimeVariable(resetNode, host)); err != nil {
				*stderr = fmt.Sprintf("clean loop item to variable error: %v", err)
				return
			}
		}()
	}

	// Get all variables for this host, including any loop item
	ha, err := e.variable.Get(variable.GetAllVariable(host))
	if err != nil {
		*stderr = fmt.Sprintf("failed to get host %s variable: %v", host, err)
		return
	}

	// Convert host variables to map type
	had, ok := ha.(map[string]any)
	if !ok {
		*stderr = fmt.Sprintf("host: %s variable is not a map", host)
		return
	}

	// Execute the actual module with the prepared context
	*stdout, *stderr = modules.FindModule(task.Spec.Module.Name)(ctx, modules.ExecOptions{
		Args:      e.task.Spec.Module.Args,
		Host:      host,
		Variable:  e.variable,
		Task:      *e.task,
		Playbook:  *e.playbook,
		LogOutput: e.logOutput,
	})

	e.dealFailedWhen(had, stderr)
}

// dealLoop parses the loop specification into a slice of items to iterate over.
// If no loop is specified, returns a single nil item. Otherwise converts the loop
// specification from JSON into a slice of values.
func (e *taskExecutor) dealLoop(ha map[string]any) []any {
	var items []any
	switch {
	case e.task.Spec.Loop.Raw == nil:
		// loop is not set. add one element to execute once module.
		items = []any{nil}
	default:
		items = variable.Extension2Slice(ha, e.task.Spec.Loop)
	}

	return items
}

// dealWhen evaluates the "when" conditions for a task to determine if it should be skipped.
// Returns true if the task should be skipped, false if it should proceed.
func (e *taskExecutor) dealWhen(had map[string]any, stdout, stderr *string) bool {
	if len(e.task.Spec.When) > 0 {
		ok, err := tmpl.ParseBool(had, e.task.Spec.When...)
		if err != nil {
			klog.V(5).ErrorS(err, "validate when condition error", "task", ctrlclient.ObjectKeyFromObject(e.task))
			*stderr = fmt.Sprintf("parse when condition error: %v", err)

			return true
		}
		if !ok {
			*stdout = modules.StdoutSkip

			return true
		}
	}

	return false
}

// dealFailedWhen evaluates the "failed_when" conditions for a task to determine if it should fail.
// Returns true if the task should be marked as failed, false if it should proceed.
func (e *taskExecutor) dealFailedWhen(had map[string]any, stderr *string) {
	if len(e.task.Spec.FailedWhen) > 0 {
		ok, err := tmpl.ParseBool(had, e.task.Spec.FailedWhen...)
		if err != nil {
			klog.V(5).ErrorS(err, "validate failed_when condition error", "task", ctrlclient.ObjectKeyFromObject(e.task))
			*stderr = fmt.Sprintf("parse failed_when condition error: %v", err)
		}
		if ok {
			*stderr = "reach failed_when, failed"
		}
	}
}

// dealRegister handles storing task output in a registered variable if specified.
// The output can be stored as raw string, JSON, or YAML based on the register type.
func (e *taskExecutor) dealRegister(stdout, stderr, host string) error {
	if e.task.Spec.Register != "" {
		var stdoutResult any = stdout
		var stderrResult any = stderr
		switch e.task.Spec.RegisterType {
		case "json":
			_ = json.Unmarshal([]byte(stdout), &stdoutResult)
		case "yaml", "yml":
			_ = yaml.Unmarshal([]byte(stdout), &stdoutResult)
		default:
			// store by string
		}
		// set variable to parent location
		node, err := converter.ConvertMap2Node(map[string]any{
			e.task.Spec.Register: map[string]any{
				"stdout": stdoutResult,
				"stderr": stderrResult,
			},
		})
		if err != nil {
			return err
		}
		if err := e.variable.Merge(variable.MergeRuntimeVariable(node, host)); err != nil {
			return err
		}
	}

	return nil
}
