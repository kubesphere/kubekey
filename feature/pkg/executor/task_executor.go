package executor

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	kkcorev1 "github.com/kubesphere/kubekey/v4/pkg/apis/core/v1"
	kkcorev1alpha1 "github.com/kubesphere/kubekey/v4/pkg/apis/core/v1alpha1"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/converter/tmpl"
	"github.com/kubesphere/kubekey/v4/pkg/modules"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

type taskExecutor struct {
	*option
	task *kkcorev1alpha1.Task
}

// Exec and store Task
func (e taskExecutor) Exec(ctx context.Context) error {
	// create task
	if err := e.client.Create(ctx, e.task); err != nil {
		klog.V(5).ErrorS(err, "create task error", "task", ctrlclient.ObjectKeyFromObject(e.task), "pipeline", ctrlclient.ObjectKeyFromObject(e.pipeline))

		return err
	}
	defer func() {
		e.pipeline.Status.TaskResult.Total++
		switch e.task.Status.Phase {
		case kkcorev1alpha1.TaskPhaseSuccess:
			e.pipeline.Status.TaskResult.Success++
		case kkcorev1alpha1.TaskPhaseIgnored:
			e.pipeline.Status.TaskResult.Ignored++
		case kkcorev1alpha1.TaskPhaseFailed:
			e.pipeline.Status.TaskResult.Failed++
		}
	}()

	for !e.task.IsComplete() {
		var roleLog string
		if e.task.Annotations[kkcorev1alpha1.TaskAnnotationRole] != "" {
			roleLog = "[" + e.task.Annotations[kkcorev1alpha1.TaskAnnotationRole] + "] "
		}
		klog.V(5).InfoS("begin run task", "task", ctrlclient.ObjectKeyFromObject(e.task))
		fmt.Fprintf(e.logOutput, "%s %s%s\n", time.Now().Format(time.TimeOnly+" MST"), roleLog, e.task.Spec.Name)
		// exec task
		e.task.Status.Phase = kkcorev1alpha1.TaskPhaseRunning
		if err := e.client.Status().Update(ctx, e.task); err != nil {
			klog.V(5).ErrorS(err, "update task status error", "task", ctrlclient.ObjectKeyFromObject(e.task), "pipeline", ctrlclient.ObjectKeyFromObject(e.pipeline))
		}
		e.execTask(ctx)
		if err := e.client.Status().Update(ctx, e.task); err != nil {
			klog.V(5).ErrorS(err, "update task status error", "task", ctrlclient.ObjectKeyFromObject(e.task), "pipeline", ctrlclient.ObjectKeyFromObject(e.pipeline))

			return err
		}
	}
	// exit when task run failed
	if e.task.IsFailed() {
		var hostReason []kkcorev1.PipelineFailedDetailHost
		for _, tr := range e.task.Status.HostResults {
			hostReason = append(hostReason, kkcorev1.PipelineFailedDetailHost{
				Host:   tr.Host,
				Stdout: tr.Stdout,
				StdErr: tr.StdErr,
			})
		}
		e.pipeline.Status.FailedDetail = append(e.pipeline.Status.FailedDetail, kkcorev1.PipelineFailedDetail{
			Task:  e.task.Spec.Name,
			Hosts: hostReason,
		})
		e.pipeline.Status.Phase = kkcorev1.PipelinePhaseFailed

		return fmt.Errorf("task %s run failed", e.task.Spec.Name)
	}

	return nil
}

// execTask
func (e taskExecutor) execTask(ctx context.Context) {
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
func (e taskExecutor) execTaskHost(i int, h string) func(ctx context.Context) {
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
			stderr = fmt.Sprintf("get variable error: %v", err)

			return
		}
		// convert hostVariable to map
		had, ok := ha.(map[string]any)
		if !ok {
			stderr = fmt.Sprintf("variable is not map error: %v", err)
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
func (e taskExecutor) execTaskHostLogs(ctx context.Context, h string, stdout, stderr *string) func() {
	// placeholder format task log
	var placeholder string
	if hostNameMaxLen, err := e.variable.Get(variable.GetHostMaxLength()); err == nil {
		if hl, ok := hostNameMaxLen.(int); ok {
			placeholder = strings.Repeat(" ", hl-len(h))
		}
	}
	// progress bar for task
	var bar = progressbar.NewOptions(-1,
		progressbar.OptionSetWriter(e.logOutput),
		progressbar.OptionSpinnerCustom([]string{"            "}),
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
				return false, err
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

// execLoop parse loop to item slice and execute it. if loop contains template string. convert it.
// loop is json string. try convertor to string slice by json.
// loop is normal string. set it to empty slice and return.
func (e taskExecutor) dealLoop(ha map[string]any) []any {
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

// executeModule find register module and execute it in a single host.
func (e taskExecutor) executeModule(ctx context.Context, task *kkcorev1alpha1.Task, host string, stdout, stderr *string) {
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
		Pipeline: *e.pipeline,
	})
}

// dealWhen "when" argument in task.
func (e taskExecutor) dealWhen(had map[string]any, stdout, stderr *string) bool {
	if len(e.task.Spec.When) > 0 {
		ok, err := tmpl.ParseBool(had, e.task.Spec.When)
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
func (e taskExecutor) dealFailedWhen(had map[string]any, stdout, stderr *string) bool {
	if len(e.task.Spec.FailedWhen) > 0 {
		ok, err := tmpl.ParseBool(had, e.task.Spec.FailedWhen)
		if err != nil {
			klog.V(5).ErrorS(err, "validate failed_when condition error", "task", ctrlclient.ObjectKeyFromObject(e.task))
			*stderr = fmt.Sprintf("parse failed_when condition error: %v", err)

			return true
		}
		if ok {
			*stdout = modules.StdoutSkip
			*stderr = "reach failed_when, failed"

			return true
		}
	}

	return false
}

// dealRegister "register" argument in task.
func (e taskExecutor) dealRegister(stdout, stderr, host string) error {
	if e.task.Spec.Register != "" {
		var stdoutResult any = stdout
		var stderrResult any = stderr
		// try to convert by json
		_ = json.Unmarshal([]byte(stdout), &stdoutResult)
		// try to convert by json
		_ = json.Unmarshal([]byte(stderr), &stderrResult)
		// set variable to parent location
		if err := e.variable.Merge(variable.MergeRuntimeVariable(map[string]any{
			e.task.Spec.Register: map[string]any{
				"stdout": stdoutResult,
				"stderr": stderrResult,
			},
		}, host)); err != nil {
			return fmt.Errorf("register task result to variable error: %w", err)
		}
	}

	return nil
}
