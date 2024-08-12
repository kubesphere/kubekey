/*
Copyright 2024 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package executor

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	kkcorev1 "github.com/kubesphere/kubekey/v4/pkg/apis/core/v1"
	kkcorev1alpha1 "github.com/kubesphere/kubekey/v4/pkg/apis/core/v1alpha1"
	projectv1 "github.com/kubesphere/kubekey/v4/pkg/apis/project/v1"
	"github.com/kubesphere/kubekey/v4/pkg/connector"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/converter"
	"github.com/kubesphere/kubekey/v4/pkg/converter/tmpl"
	"github.com/kubesphere/kubekey/v4/pkg/modules"
	"github.com/kubesphere/kubekey/v4/pkg/project"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

// TaskExecutor all task in pipeline
type TaskExecutor interface {
	Exec(ctx context.Context) error
}

func NewTaskExecutor(client ctrlclient.Client, pipeline *kkcorev1.Pipeline, logOutput io.Writer) TaskExecutor {
	// get variable
	v, err := variable.New(client, *pipeline)
	if err != nil {
		klog.V(5).ErrorS(nil, "convert playbook error", "pipeline", ctrlclient.ObjectKeyFromObject(pipeline))
		return nil
	}

	return &executor{
		client:    client,
		pipeline:  pipeline,
		variable:  v,
		logOutput: logOutput,
	}
}

type executor struct {
	client ctrlclient.Client

	pipeline *kkcorev1.Pipeline
	variable variable.Variable

	logOutput io.Writer
}

type execBlockOptions struct {
	// playbook level config
	hosts        []string // which hosts will run playbook
	ignoreErrors *bool    // IgnoreErrors for playbook
	// blocks level config
	blocks []projectv1.Block
	role   string   // role name of blocks
	when   []string // when condition for blocks
	tags   projectv1.Taggable
}

func (e executor) Exec(ctx context.Context) error {
	klog.V(6).InfoS("deal project", "pipeline", ctrlclient.ObjectKeyFromObject(e.pipeline))
	pj, err := project.New(*e.pipeline, true)
	if err != nil {
		return fmt.Errorf("deal project error: %w", err)
	}

	// convert to transfer.Playbook struct
	pb, err := pj.MarshalPlaybook()
	if err != nil {
		return fmt.Errorf("convert playbook error: %w", err)
	}

	for _, play := range pb.Play {
		if !play.Taggable.IsEnabled(e.pipeline.Spec.Tags, e.pipeline.Spec.SkipTags) {
			// if not match the tags. skip
			continue
		}
		// hosts should contain all host's name. hosts should not be empty.
		var hosts []string
		if ahn, err := e.variable.Get(variable.GetHostnames(play.PlayHost.Hosts)); err == nil {
			hosts = ahn.([]string)
		}
		if len(hosts) == 0 { // if hosts is empty skip this playbook
			klog.V(5).Info("Hosts is empty", "pipeline", ctrlclient.ObjectKeyFromObject(e.pipeline))
			continue
		}

		// when gather_fact is set. get host's information from remote.
		if play.GatherFacts {
			for _, h := range hosts {
				gfv, err := e.getGatherFact(ctx, h, e.variable)
				if err != nil {
					return fmt.Errorf("get gather fact error: %w", err)
				}
				// merge host information to runtime variable
				if err := e.variable.Merge(variable.MergeRemoteVariable(h, gfv)); err != nil {
					klog.V(5).ErrorS(err, "Merge gather fact error", "pipeline", ctrlclient.ObjectKeyFromObject(e.pipeline), "host", h)
					return fmt.Errorf("merge gather fact error: %w", err)
				}
			}
		}

		// Batch execution, with each batch being a group of hosts run in serial.
		var batchHosts [][]string
		if play.RunOnce {
			// runOnce only run in first node
			batchHosts = [][]string{{hosts[0]}}
		} else {
			// group hosts by serial. run the playbook by serial
			batchHosts, err = converter.GroupHostBySerial(hosts, play.Serial.Data)
			if err != nil {
				return fmt.Errorf("group host by serial error: %w", err)
			}
		}

		// generate and execute task.
		for _, serials := range batchHosts {
			// each batch hosts should not be empty.
			if len(serials) == 0 {
				klog.V(5).ErrorS(nil, "Host is empty", "pipeline", ctrlclient.ObjectKeyFromObject(e.pipeline))
				return fmt.Errorf("host is empty")
			}

			if err := e.mergeVariable(ctx, e.variable, play.Vars, serials...); err != nil {
				return fmt.Errorf("merge variable error: %w", err)
			}
			// generate task from pre tasks
			if err := e.execBlock(ctx, execBlockOptions{
				hosts:        serials,
				ignoreErrors: play.IgnoreErrors,
				blocks:       play.PreTasks,
				tags:         play.Taggable,
			}); err != nil {
				return fmt.Errorf("execute pre-tasks from play error: %w", err)
			}
			// generate task from role
			for _, role := range play.Roles {
				if err := e.mergeVariable(ctx, e.variable, role.Vars, serials...); err != nil {
					return fmt.Errorf("merge variable error: %w", err)
				}
				// use the most closely configuration
				ignoreErrors := role.IgnoreErrors
				if ignoreErrors == nil {
					ignoreErrors = play.IgnoreErrors
				}

				if err := e.execBlock(ctx, execBlockOptions{
					hosts:        serials,
					ignoreErrors: ignoreErrors,
					blocks:       role.Block,
					role:         role.Role,
					when:         role.When.Data,
					tags:         projectv1.JoinTag(role.Taggable, play.Taggable),
				}); err != nil {
					return fmt.Errorf("execute role-tasks error: %w", err)
				}
			}
			// generate task from tasks
			if err := e.execBlock(ctx, execBlockOptions{
				hosts:        serials,
				ignoreErrors: play.IgnoreErrors,
				blocks:       play.Tasks,
				tags:         play.Taggable,
			}); err != nil {
				return fmt.Errorf("execute tasks error: %w", err)
			}
			// generate task from post tasks
			if err := e.execBlock(ctx, execBlockOptions{
				hosts:        serials,
				ignoreErrors: play.IgnoreErrors,
				blocks:       play.Tasks,
				tags:         play.Taggable,
			}); err != nil {
				return fmt.Errorf("execute post-tasks error: %w", err)
			}
		}
	}
	return nil
}

// getGatherFact get host info
func (e executor) getGatherFact(ctx context.Context, hostname string, vars variable.Variable) (map[string]any, error) {
	v, err := vars.Get(variable.GetParamVariable(hostname))
	if err != nil {
		klog.V(5).ErrorS(err, "Get host variable error", "hostname", hostname)
		return nil, err
	}
	connectorVars := make(map[string]any)
	if c1, ok := v.(map[string]any)[_const.VariableConnector]; ok {
		if c2, ok := c1.(map[string]any); ok {
			connectorVars = c2
		}
	}
	conn, err := connector.NewConnector(hostname, connectorVars)
	if err != nil {
		klog.V(5).ErrorS(err, "New connector error", "hostname", hostname)
		return nil, err
	}
	if err := conn.Init(ctx); err != nil {
		klog.V(5).ErrorS(err, "Init connection error", "hostname", hostname)
		return nil, err
	}
	defer conn.Close(ctx)

	if gf, ok := conn.(connector.GatherFacts); ok {
		return gf.Info(ctx)
	}
	klog.V(5).ErrorS(nil, "gather fact is not defined in this connector", "hostname", hostname)
	return nil, nil
}

// execBlock loop block and generate task.
func (e executor) execBlock(ctx context.Context, options execBlockOptions) error {
	for _, at := range options.blocks {
		if !projectv1.JoinTag(at.Taggable, options.tags).IsEnabled(e.pipeline.Spec.Tags, e.pipeline.Spec.SkipTags) {
			continue
		}
		hosts := options.hosts
		if at.RunOnce { // only run in first host
			hosts = []string{options.hosts[0]}
		}
		tags := projectv1.JoinTag(at.Taggable, options.tags)

		// use the most closely configuration
		ignoreErrors := at.IgnoreErrors
		if ignoreErrors == nil {
			ignoreErrors = options.ignoreErrors
		}
		// merge variable which defined in block
		if err := e.mergeVariable(ctx, e.variable, at.Vars, hosts...); err != nil {
			klog.V(5).ErrorS(err, "merge variable error", "pipeline", e.pipeline, "block", at.Name)
			return err
		}

		switch {
		case len(at.Block) != 0:
			var errs error
			// exec block
			if err := e.execBlock(ctx, execBlockOptions{
				hosts:        hosts,
				ignoreErrors: ignoreErrors,
				role:         options.role,
				blocks:       at.Block,
				when:         append(options.when, at.When.Data...),
				tags:         tags,
			}); err != nil {
				klog.V(5).ErrorS(err, "execute tasks from block error", "pipeline", ctrlclient.ObjectKeyFromObject(e.pipeline), "block", at.Name)
				errs = errors.Join(errs, err)
			}

			// if block exec failed exec rescue
			if e.pipeline.Status.Phase == kkcorev1.PipelinePhaseFailed && len(at.Rescue) != 0 {
				if err := e.execBlock(ctx, execBlockOptions{
					hosts:        hosts,
					ignoreErrors: ignoreErrors,
					blocks:       at.Rescue,
					role:         options.role,
					when:         append(options.when, at.When.Data...),
					tags:         tags,
				}); err != nil {
					klog.V(5).ErrorS(err, "execute tasks from rescue error", "pipeline", ctrlclient.ObjectKeyFromObject(e.pipeline), "block", at.Name)
					errs = errors.Join(errs, err)
				}
			}

			// exec always after block
			if len(at.Always) != 0 {
				if err := e.execBlock(ctx, execBlockOptions{
					hosts:        hosts,
					ignoreErrors: ignoreErrors,
					blocks:       at.Always,
					role:         options.role,
					when:         append(options.when, at.When.Data...),
					tags:         tags,
				}); err != nil {
					klog.V(5).ErrorS(err, "execute tasks from always error", "pipeline", ctrlclient.ObjectKeyFromObject(e.pipeline), "block", at.Name)
					errs = errors.Join(errs, err)
				}
			}

			// when execute error. return
			if errs != nil {
				return errs
			}

		case at.IncludeTasks != "":
			// include tasks has converted to blocks.
			// do nothing
		default:
			task := converter.MarshalBlock(ctx, options.role, hosts, append(options.when, at.When.Data...), at)
			// complete by pipeline
			task.GenerateName = e.pipeline.Name + "-"
			task.Namespace = e.pipeline.Namespace
			if err := controllerutil.SetControllerReference(e.pipeline, task, e.client.Scheme()); err != nil {
				klog.V(5).ErrorS(err, "Set controller reference error", "pipeline", ctrlclient.ObjectKeyFromObject(e.pipeline), "block", at.Name)
				return err
			}
			// complete module by unknown field
			for n, a := range at.UnknownFiled {
				data, err := json.Marshal(a)
				if err != nil {
					klog.V(5).ErrorS(err, "Marshal unknown field error", "pipeline", ctrlclient.ObjectKeyFromObject(e.pipeline), "block", at.Name, "field", n)
					return err
				}
				if m := modules.FindModule(n); m != nil {
					task.Spec.Module.Name = n
					task.Spec.Module.Args = runtime.RawExtension{Raw: data}
					break
				}
			}
			if task.Spec.Module.Name == "" { // action is necessary for a task
				klog.V(5).ErrorS(nil, "No module/action detected in task", "pipeline", ctrlclient.ObjectKeyFromObject(e.pipeline), "block", at.Name)
				return fmt.Errorf("no module/action detected in task: %s", task.Name)
			}
			// create task
			if err := e.client.Create(ctx, task); err != nil {
				klog.V(5).ErrorS(err, "create task error", "pipeline", ctrlclient.ObjectKeyFromObject(e.pipeline), "block", at.Name)
				return err
			}

			for {
				var roleLog string
				if task.Annotations[kkcorev1alpha1.TaskAnnotationRole] != "" {
					roleLog = "[" + task.Annotations[kkcorev1alpha1.TaskAnnotationRole] + "] "
				}
				klog.V(5).InfoS("begin run task", "task", ctrlclient.ObjectKeyFromObject(task))
				fmt.Fprintf(e.logOutput, "%s %s%s\n", time.Now().Format(time.TimeOnly+" MST"), roleLog, task.Spec.Name)
				// exec task
				task.Status.Phase = kkcorev1alpha1.TaskPhaseRunning
				if err := e.client.Status().Update(ctx, task); err != nil {
					klog.V(5).ErrorS(err, "update task status error", "task", ctrlclient.ObjectKeyFromObject(task))
				}
				if err := e.executeTask(ctx, task, options); err != nil {
					klog.V(5).ErrorS(err, "exec task error", "pipeline", ctrlclient.ObjectKeyFromObject(e.pipeline), "block", at.Name)
					return err
				}
				if err := e.client.Status().Update(ctx, task); err != nil {
					klog.V(5).ErrorS(err, "update task status error", "task", ctrlclient.ObjectKeyFromObject(task))
					return err
				}

				if task.IsComplete() {
					break
				}
			}
			e.pipeline.Status.TaskResult.Total++
			switch task.Status.Phase {
			case kkcorev1alpha1.TaskPhaseSuccess:
				e.pipeline.Status.TaskResult.Success++
			case kkcorev1alpha1.TaskPhaseIgnored:
				e.pipeline.Status.TaskResult.Ignored++
			case kkcorev1alpha1.TaskPhaseFailed:
				e.pipeline.Status.TaskResult.Failed++
			}

			// exit when task run failed
			if task.IsFailed() {
				var hostReason []kkcorev1.PipelineFailedDetailHost
				for _, tr := range task.Status.HostResults {
					hostReason = append(hostReason, kkcorev1.PipelineFailedDetailHost{
						Host:   tr.Host,
						Stdout: tr.Stdout,
						StdErr: tr.StdErr,
					})
				}
				e.pipeline.Status.FailedDetail = append(e.pipeline.Status.FailedDetail, kkcorev1.PipelineFailedDetail{
					Task:  task.Spec.Name,
					Hosts: hostReason,
				})
				e.pipeline.Status.Phase = kkcorev1.PipelinePhaseFailed
				return fmt.Errorf("task %s run failed", task.Spec.Name)
			}
		}
	}
	return nil
}

// executeTask parallel in each host.
func (e executor) executeTask(ctx context.Context, task *kkcorev1alpha1.Task, options execBlockOptions) error {
	// check task host results
	wg := &wait.Group{}
	task.Status.HostResults = make([]kkcorev1alpha1.TaskHostResult, len(task.Spec.Hosts))

	for i, h := range task.Spec.Hosts {
		wg.StartWithContext(ctx, func(ctx context.Context) {
			// task result
			var stdout, stderr string
			defer func() {
				if task.Spec.Register != "" {
					var stdoutResult any = stdout
					var stderrResult any = stderr
					// try to convert by json
					_ = json.Unmarshal([]byte(stdout), &stdoutResult)
					// try to convert by json
					_ = json.Unmarshal([]byte(stderr), &stderrResult)
					// set variable to parent location
					if err := e.variable.Merge(variable.MergeRuntimeVariable(h, map[string]any{
						task.Spec.Register: map[string]any{
							"stdout": stdoutResult,
							"stderr": stderrResult,
						},
					})); err != nil {
						stderr = fmt.Sprintf("register task result to variable error: %v", err)
						return
					}
				}
				if stderr != "" && task.Spec.IgnoreError != nil && *task.Spec.IgnoreError {
					klog.V(5).ErrorS(nil, "task run failed", "host", h, "stdout", stdout, "stderr", stderr, "task", ctrlclient.ObjectKeyFromObject(task))
				} else if stderr != "" {
					klog.ErrorS(nil, "task run failed", "host", h, "stdout", stdout, "stderr", stderr, "task", ctrlclient.ObjectKeyFromObject(task))
				}
				// fill result
				task.Status.HostResults[i] = kkcorev1alpha1.TaskHostResult{
					Host:   h,
					Stdout: stdout,
					StdErr: stderr,
				}
			}()
			// task log
			// placeholder format task log
			var placeholder string
			if hostNameMaxLen, err := e.variable.Get(variable.GetHostMaxLength()); err == nil {
				placeholder = strings.Repeat(" ", hostNameMaxLen.(int)-len(h))
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
			go func() {
				for !bar.IsFinished() {
					if err := bar.Add(1); err != nil {
						return
					}
					time.Sleep(100 * time.Millisecond)
				}
			}()
			defer func() {
				switch {
				case stderr != "":
					if task.Spec.IgnoreError != nil && *task.Spec.IgnoreError { // ignore
						bar.Describe(fmt.Sprintf("[\033[36m%s\033[0m]%s \033[34mignore \033[0m", h, placeholder))
					} else { // failed
						bar.Describe(fmt.Sprintf("[\033[36m%s\033[0m]%s \033[31mfailed \033[0m", h, placeholder))
					}
				case stdout == "skip": // skip
					bar.Describe(fmt.Sprintf("[\033[36m%s\033[0m]%s \033[34mskip   \033[0m", h, placeholder))
				default: //success
					bar.Describe(fmt.Sprintf("[\033[36m%s\033[0m]%s \033[34msuccess\033[0m", h, placeholder))
				}
				if err := bar.Finish(); err != nil {
					klog.ErrorS(err, "finish bar error")
				}
			}()
			// task execute
			ha, err := e.variable.Get(variable.GetAllVariable(h))
			if err != nil {
				stderr = fmt.Sprintf("get variable error: %v", err)
				return
			}
			// check when condition
			if len(task.Spec.When) > 0 {
				ok, err := tmpl.ParseBool(ha.(map[string]any), task.Spec.When)
				if err != nil {
					stderr = fmt.Sprintf("parse when condition error: %v", err)
					return
				}
				if !ok {
					stdout = "skip"
					return
				}
			}
			// execute module with loop
			// if loop is empty. execute once, and the item is null
			for _, item := range e.parseLoop(ctx, ha.(map[string]any), task) {
				// set item to runtime variable
				if err := e.variable.Merge(variable.MergeRuntimeVariable(h, map[string]any{
					_const.VariableItem: item,
				})); err != nil {
					stderr = fmt.Sprintf("set loop item to variable error: %v", err)
					return
				}
				stdout, stderr = e.executeModule(ctx, task, modules.ExecOptions{
					Args:     task.Spec.Module.Args,
					Host:     h,
					Variable: e.variable,
					Task:     *task,
					Pipeline: *e.pipeline,
				})
				// delete item
				if err := e.variable.Merge(variable.MergeRuntimeVariable(h, map[string]any{
					_const.VariableItem: nil,
				})); err != nil {
					stderr = fmt.Sprintf("clean loop item to variable error: %v", err)
					return
				}
			}
		})
	}
	wg.Wait()
	// host result for task
	task.Status.Phase = kkcorev1alpha1.TaskPhaseSuccess
	for _, data := range task.Status.HostResults {
		if data.StdErr != "" {
			if task.Spec.IgnoreError != nil && *task.Spec.IgnoreError {
				task.Status.Phase = kkcorev1alpha1.TaskPhaseIgnored
			} else {
				task.Status.Phase = kkcorev1alpha1.TaskPhaseFailed
			}
			break
		}
	}

	return nil
}

// parseLoop parse loop to slice. if loop contains template string. convert it.
// loop is json string. try convertor to string slice by json.
// loop is normal string. set it to empty slice and return.
// loop is string slice. return it.
func (e executor) parseLoop(ctx context.Context, ha map[string]any, task *kkcorev1alpha1.Task) []any {
	switch {
	case task.Spec.Loop.Raw == nil:
		// loop is not set. add one element to execute once module.
		return []any{nil}
	default:
		return variable.Extension2Slice(ha, task.Spec.Loop)
	}
}

// executeModule find register module and execute it.
func (e executor) executeModule(ctx context.Context, task *kkcorev1alpha1.Task, opts modules.ExecOptions) (string, string) {
	// get all variable. which contains item.
	lg, err := opts.Variable.Get(variable.GetAllVariable(opts.Host))
	if err != nil {
		klog.V(5).ErrorS(err, "get location variable error", "task", ctrlclient.ObjectKeyFromObject(task))
		return "", err.Error()
	}
	// check failed when condition
	if len(task.Spec.FailedWhen) > 0 {
		ok, err := tmpl.ParseBool(lg.(map[string]any), task.Spec.FailedWhen)
		if err != nil {
			klog.V(5).ErrorS(err, "validate FailedWhen condition error", "task", ctrlclient.ObjectKeyFromObject(task))
			return "", err.Error()
		}
		if ok {
			return "", "failed by failedWhen"
		}
	}

	return modules.FindModule(task.Spec.Module.Name)(ctx, opts)
}

// mergeVariable to runtime variable
func (e executor) mergeVariable(ctx context.Context, v variable.Variable, vd map[string]any, hosts ...string) error {
	if len(vd) == 0 {
		// skip
		return nil
	}
	for _, host := range hosts {
		if err := v.Merge(variable.MergeRuntimeVariable(host, vd)); err != nil {
			return err
		}
	}
	return nil
}
