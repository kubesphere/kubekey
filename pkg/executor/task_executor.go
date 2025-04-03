package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	kkcorev1alpha1 "github.com/kubesphere/kubekey/api/core/v1alpha1"
	"github.com/schollz/progressbar/v3"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/converter/tmpl"
	"github.com/kubesphere/kubekey/v4/pkg/modules"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

type taskExecutor struct {
	*option
	task *kkcorev1alpha1.Task
	// runOnce only executor task once
	runOnce sync.Once
	// taskRunTimeout is the timeout for task executor
	taskRunTimeout time.Duration
}

// Exec and store Task
func (e *taskExecutor) Exec(ctx context.Context) error {
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

	// Create ticker for periodic reconciliation
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	// reconcile handles task state transitions and execution
	reconcile := func(ctx context.Context, request ctrl.Request) (_ ctrl.Result, err error) {
		task := &kkcorev1alpha1.Task{}
		if err = e.client.Get(ctx, request.NamespacedName, task); err != nil {
			return ctrl.Result{Requeue: true}, err
		}
		defer func() {
			err = e.client.Status().Patch(ctx, e.task, ctrlclient.MergeFrom(task.DeepCopy()))
		}()

		// Handle task phase transitions
		switch task.Status.Phase {
		case "", kkcorev1alpha1.TaskPhasePending:
			e.task.Status.Phase = kkcorev1alpha1.TaskPhaseRunning
		case kkcorev1alpha1.TaskPhaseRunning:
			// Execute task only once when it enters running phase
			e.runOnce.Do(func() {
				e.execTask(ctx)
			})
		case kkcorev1alpha1.TaskPhaseFailed, kkcorev1alpha1.TaskPhaseIgnored, kkcorev1alpha1.TaskPhaseSuccess:
			return ctrl.Result{}, nil
		}

		return ctrl.Result{Requeue: true}, nil
	}

	// Main loop to handle task execution and timeout
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(e.taskRunTimeout):
			return errors.Errorf("task %q execution timeout", e.task.Spec.Name)
		case <-ticker.C:
			result, err := reconcile(ctx, ctrl.Request{NamespacedName: ctrlclient.ObjectKeyFromObject(e.task)})
			if err != nil {
				klog.V(5).ErrorS(err, "failed to reconcile task", "task", ctrlclient.ObjectKeyFromObject(e.task), "playbook", ctrlclient.ObjectKeyFromObject(e.playbook))
			}
			if result.Requeue {
				continue
			}

			return nil
		}
	}
}

// execTask
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

// execTaskHost deal module in each host parallel.
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
			// set item to runtime variable
			if err := e.variable.Merge(variable.MergeRuntimeVariable(map[string]any{
				_const.VariableItem: item,
			}, h)); err != nil {
				stderr = fmt.Sprintf("set loop item to variable error: %v", err)

				return
			}
			e.executeModule(ctx, e.task, h, &stdout, &stderr)
			// delete item
			if err := e.variable.Merge(variable.MergeRuntimeVariable(map[string]any{
				_const.VariableItem: nil,
			}, h)); err != nil {
				stderr = fmt.Sprintf("clean loop item to variable error: %v", err)

				return
			}
		}
	}
}

// execTaskHostLogs logs for each host
func (e *taskExecutor) execTaskHostLogs(ctx context.Context, h string, stdout, stderr *string) func() {
	// placeholder format task log
	var placeholder string
	if hostnameMaxLen, err := e.variable.Get(variable.GetHostMaxLength()); err == nil {
		if hl, ok := hostnameMaxLen.(int); ok {
			placeholder = strings.Repeat(" ", hl-len(h))
		}
	}
	// progress bar for task
	var bar = progressbar.NewOptions(-1,
		progressbar.OptionSetWriter(e.logOutput),
		// progressbar.OptionSpinnerCustom([]string{"            "}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetDescription(fmt.Sprintf("[\033[36m%s\033[0m]%s \033[36mrunning\033[0m", h, placeholder)),
		progressbar.OptionOnCompletion(func() {
			if _, err := os.Stdout.WriteString("\n"); err != nil {
				klog.ErrorS(err, "failed to write output", "host", h)
			}
		}),
	)
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
			} else { // failed
				bar.Describe(fmt.Sprintf("[\033[36m%s\033[0m]%s \033[31mfailed \033[0m", h, placeholder))
			}
		case *stdout == modules.StdoutSkip: // skip
			bar.Describe(fmt.Sprintf("[\033[36m%s\033[0m]%s \033[34mskip   \033[0m", h, placeholder))
		default: //success
			bar.Describe(fmt.Sprintf("[\033[36m%s\033[0m]%s \033[34msuccess\033[0m", h, placeholder))
		}
		if err := bar.Finish(); err != nil {
			klog.ErrorS(err, "finish bar error")
		}
	}
}

// executeModule find register module and execute it in a single host.
func (e *taskExecutor) executeModule(ctx context.Context, task *kkcorev1alpha1.Task, host string, stdout, stderr *string) {
	// get all variable. which contains item.
	ha, err := e.variable.Get(variable.GetAllVariable(host))
	if err != nil {
		*stderr = fmt.Sprintf("failed to get host %s variable: %v", host, err)

		return
	}
	// convert hostVariable to map
	had, ok := ha.(map[string]any)
	if !ok {
		*stderr = fmt.Sprintf("host: %s variable is not a map", host)

		return
	}
	// check failed when condition
	if skip := e.dealFailedWhen(had, stdout, stderr); skip {
		return
	}
	*stdout, *stderr = modules.FindModule(task.Spec.Module.Name)(ctx, modules.ExecOptions{
		Args:     e.task.Spec.Module.Args,
		Host:     host,
		Variable: e.variable,
		Task:     *e.task,
		Playbook: *e.playbook,
	})
}

// execLoop parse loop to item slice and execute it. if loop contains template string. convert it.
// loop is json string. try convertor to string slice by json.
// loop is normal string. set it to empty slice and return.
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

// dealWhen "when" argument in task.
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

// dealFailedWhen "failed_when" argument in task.
func (e *taskExecutor) dealFailedWhen(had map[string]any, stdout, stderr *string) bool {
	if len(e.task.Spec.FailedWhen) > 0 {
		ok, err := tmpl.ParseBool(had, e.task.Spec.FailedWhen...)
		if err != nil {
			klog.V(5).ErrorS(err, "validate failed_when condition error", "task", ctrlclient.ObjectKeyFromObject(e.task))
			*stderr = fmt.Sprintf("parse failed_when condition error: %v", err)

			return true
		}
		if ok {
			*stdout = modules.StdoutFalse
			*stderr = "reach failed_when, failed"

			return true
		}
	}

	return false
}

// dealRegister "register" argument in task.
func (e *taskExecutor) dealRegister(stdout, stderr, host string) error {
	if e.task.Spec.Register != "" {
		var stdoutResult any = stdout
		var stderrResult any = stderr
		// try to convert by json
		if json.Valid([]byte(stdout)) {
			_ = json.Unmarshal([]byte(stdout), &stdoutResult)
			_ = json.Unmarshal([]byte(stderr), &stderrResult)
		}
		// set variable to parent location
		if err := e.variable.Merge(variable.MergeRuntimeVariable(map[string]any{
			e.task.Spec.Register: map[string]any{
				"stdout": stdoutResult,
				"stderr": stderrResult,
			},
		}, host)); err != nil {
			return err
		}
	}

	return nil
}
