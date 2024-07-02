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
	"fmt"
	"os"

	"github.com/schollz/progressbar/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	kkcorev1 "github.com/kubesphere/kubekey/v4/pkg/apis/core/v1"
	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
	kubekeyv1alpha1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/v4/pkg/connector"
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

func NewTaskExecutor(client ctrlclient.Client, pipeline *kubekeyv1.Pipeline) TaskExecutor {
	// get variable
	v, err := variable.New(client, *pipeline)
	if err != nil {
		klog.V(4).ErrorS(nil, "convert playbook error", "pipeline", ctrlclient.ObjectKeyFromObject(pipeline))
		return nil
	}

	return &executor{
		client:   client,
		pipeline: pipeline,
		variable: v,
	}
}

type executor struct {
	client ctrlclient.Client

	pipeline *kubekeyv1.Pipeline
	variable variable.Variable
}

type execBlockOptions struct {
	// playbook level config
	hosts        []string // which hosts will run playbook
	ignoreErrors *bool    // IgnoreErrors for playbook
	// blocks level config
	blocks []kkcorev1.Block
	role   string   // role name of blocks
	when   []string // when condition for blocks
	tags   kkcorev1.Taggable
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
			klog.V(4).Info("Hosts is empty", "pipeline", ctrlclient.ObjectKeyFromObject(e.pipeline))
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
					klog.V(4).ErrorS(err, "Merge gather fact error", "pipeline", ctrlclient.ObjectKeyFromObject(e.pipeline), "host", h)
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

		// generate task by each batch.
		for _, serials := range batchHosts {
			// each batch hosts should not be empty.
			if len(serials) == 0 {
				klog.V(4).ErrorS(nil, "Host is empty", "pipeline", ctrlclient.ObjectKeyFromObject(e.pipeline))
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
					tags:         kkcorev1.JoinTag(role.Taggable, play.Taggable),
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
		klog.V(4).ErrorS(err, "Get host variable error", "hostname", hostname)
		return nil, err
	}

	conn, err := connector.NewConnector(hostname, v.(map[string]any))
	if err != nil {
		klog.V(4).ErrorS(err, "New connector error", "hostname", hostname)
		return nil, err
	}
	if err := conn.Init(ctx); err != nil {
		klog.V(4).ErrorS(err, "Init connection error", "hostname", hostname)
		return nil, err
	}
	defer conn.Close(ctx)

	if gf, ok := conn.(connector.GatherFacts); ok {
		return gf.Info(ctx)
	}
	klog.V(4).ErrorS(nil, "gather fact is not defined in this connector", "hostname", hostname)
	return nil, nil
}

func (e executor) execBlock(ctx context.Context, options execBlockOptions) error {
	for _, at := range options.blocks {
		if !kkcorev1.JoinTag(at.Taggable, options.tags).IsEnabled(e.pipeline.Spec.Tags, e.pipeline.Spec.SkipTags) {
			continue
		}
		hosts := options.hosts
		if at.RunOnce { // only run in first host
			hosts = []string{options.hosts[0]}
		}

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
			// exec block
			if err := e.execBlock(ctx, execBlockOptions{
				hosts:        hosts,
				ignoreErrors: ignoreErrors,
				role:         options.role,
				blocks:       at.Block,
				when:         append(options.when, at.When.Data...),
				tags:         kkcorev1.JoinTag(at.Taggable, options.tags),
			}); err != nil {
				klog.V(4).ErrorS(err, "execute tasks from block error", "pipeline", ctrlclient.ObjectKeyFromObject(e.pipeline), "block", at.Name)
				return err
			}

			// if block exec failed exec rescue
			if e.pipeline.Status.Phase == kubekeyv1.PipelinePhaseFailed && len(at.Rescue) != 0 {
				if err := e.execBlock(ctx, execBlockOptions{
					hosts:        hosts,
					ignoreErrors: ignoreErrors,
					blocks:       at.Rescue,
					role:         options.role,
					when:         append(options.when, at.When.Data...),
					tags:         kkcorev1.JoinTag(at.Taggable, options.tags),
				}); err != nil {
					klog.V(4).ErrorS(err, "execute tasks from rescue error", "pipeline", ctrlclient.ObjectKeyFromObject(e.pipeline), "block", at.Name)
					return err
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
					tags:         kkcorev1.JoinTag(at.Taggable, options.tags),
				}); err != nil {
					klog.V(4).ErrorS(err, "execute tasks from always error", "pipeline", ctrlclient.ObjectKeyFromObject(e.pipeline), "block", at.Name)
					return err
				}
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
				klog.V(4).ErrorS(err, "Set controller reference error", "pipeline", ctrlclient.ObjectKeyFromObject(e.pipeline), "block", at.Name)
				return err
			}
			// complete module by unknown field
			for n, a := range at.UnknownFiled {
				data, err := json.Marshal(a)
				if err != nil {
					klog.V(4).ErrorS(err, "Marshal unknown field error", "pipeline", ctrlclient.ObjectKeyFromObject(e.pipeline), "block", at.Name, "field", n)
					return err
				}
				if m := modules.FindModule(n); m != nil {
					task.Spec.Module.Name = n
					task.Spec.Module.Args = runtime.RawExtension{Raw: data}
					break
				}
			}
			if task.Spec.Module.Name == "" { // action is necessary for a task
				klog.V(4).ErrorS(nil, "No module/action detected in task", "pipeline", ctrlclient.ObjectKeyFromObject(e.pipeline), "block", at.Name)
				return fmt.Errorf("no module/action detected in task: %s", task.Name)
			}
			// create task
			if err := e.client.Create(ctx, task); err != nil {
				klog.V(4).ErrorS(err, "create task error", "pipeline", ctrlclient.ObjectKeyFromObject(e.pipeline), "block", at.Name)
				return err
			}

			for {
				klog.Infof("[Task %s] task exec \"%s\" begin for %v times", ctrlclient.ObjectKeyFromObject(task), task.Spec.Name, task.Status.RestartCount+1)
				// exec task
				task.Status.Phase = kubekeyv1alpha1.TaskPhaseRunning
				if err := e.client.Status().Update(ctx, task); err != nil {
					klog.V(5).ErrorS(err, "update task status error", "task", ctrlclient.ObjectKeyFromObject(task))
				}
				if err := e.executeTask(ctx, task, options); err != nil {
					klog.V(4).ErrorS(err, "exec task error", "pipeline", ctrlclient.ObjectKeyFromObject(e.pipeline), "block", at.Name)
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
			klog.Infof("[Task %s] task exec \"%s\" end status is %s", ctrlclient.ObjectKeyFromObject(task), task.Spec.Name, task.Status.Phase)
			e.pipeline.Status.TaskResult.Total++
			switch task.Status.Phase {
			case kubekeyv1alpha1.TaskPhaseSuccess:
				e.pipeline.Status.TaskResult.Success++
			case kubekeyv1alpha1.TaskPhaseIgnored:
				e.pipeline.Status.TaskResult.Ignored++
			case kubekeyv1alpha1.TaskPhaseFailed:
				e.pipeline.Status.TaskResult.Failed++
			}

			// exit when task run failed
			if task.IsFailed() {
				var hostReason []kubekeyv1.PipelineFailedDetailHost
				for _, tr := range task.Status.HostResults {
					hostReason = append(hostReason, kubekeyv1.PipelineFailedDetailHost{
						Host:   tr.Host,
						Stdout: tr.Stdout,
						StdErr: tr.StdErr,
					})
				}
				e.pipeline.Status.FailedDetail = append(e.pipeline.Status.FailedDetail, kubekeyv1.PipelineFailedDetail{
					Task:  task.Spec.Name,
					Hosts: hostReason,
				})
				e.pipeline.Status.Phase = kubekeyv1.PipelinePhaseFailed
				return fmt.Errorf("task %s run failed", task.Spec.Name)
			}
		}
	}
	return nil
}

func (e executor) executeTask(ctx context.Context, task *kubekeyv1alpha1.Task, options execBlockOptions) error {
	cd := kubekeyv1alpha1.TaskCondition{
		StartTimestamp: metav1.Now(),
	}
	defer func() {
		cd.EndTimestamp = metav1.Now()
		task.Status.Conditions = append(task.Status.Conditions, cd)
	}()

	// check task host results
	wg := &wait.Group{}
	dataChan := make(chan kubekeyv1alpha1.TaskHostResult, len(task.Spec.Hosts))
	for _, h := range task.Spec.Hosts {
		host := h
		wg.StartWithContext(ctx, func(ctx context.Context) {
			var stdout, stderr string

			// progress bar for task
			var bar = progressbar.NewOptions(1,
				progressbar.OptionEnableColorCodes(true),
				progressbar.OptionSetDescription(fmt.Sprintf("[%s] running...", h)),
				progressbar.OptionOnCompletion(func() {
					if _, err := os.Stdout.WriteString("\n"); err != nil {
						klog.ErrorS(err, "failed to write output", "host", h)
					}
				}),
				progressbar.OptionShowElapsedTimeOnFinish(),
				progressbar.OptionSetPredictTime(false),
			)
			defer func() {
				if task.Spec.Register != "" {
					var stdoutResult any = stdout
					var stderrResult any = stderr
					// try to convert by json
					_ = json.Unmarshal([]byte(stdout), &stdoutResult)
					// try to convert by json
					_ = json.Unmarshal([]byte(stderr), &stderrResult)
					// set variable to parent location
					if err := e.variable.Merge(variable.MergeRuntimeVariable(host, map[string]any{
						task.Spec.Register: map[string]any{
							"stdout": stdoutResult,
							"stderr": stderrResult,
						},
					})); err != nil {
						stderr = fmt.Sprintf("register task result to variable error: %v", err)
						return
					}
				}

				switch {
				case stderr != "": // failed
					bar.Describe(fmt.Sprintf("[%s] failed", h))
					if err := bar.Finish(); err != nil {
						klog.ErrorS(err, "fail to finish bar")
					}
					klog.Errorf("[Task %s] run failed: %s", ctrlclient.ObjectKeyFromObject(task), stderr)
				case stdout == "skip": // skip
					bar.Describe(fmt.Sprintf("[%s] skip", h))
					if err := bar.Finish(); err != nil {
						klog.ErrorS(err, "fail to finish bar")
					}
				default: //success
					bar.Describe(fmt.Sprintf("[%s] success", h))
					if err := bar.Finish(); err != nil {
						klog.ErrorS(err, "fail to finish bar")
					}
				}

				// fill result
				dataChan <- kubekeyv1alpha1.TaskHostResult{
					Host:   host,
					Stdout: stdout,
					StdErr: stderr,
				}
			}()

			ha, err := e.variable.Get(variable.GetAllVariable(host))
			if err != nil {
				stderr = fmt.Sprintf("get variable error: %v", err)
				return
			}
			// execute module with loop
			loop, err := e.execLoop(ctx, ha.(map[string]any), task)
			if err != nil {
				stderr = fmt.Sprintf("parse loop vars error: %v", err)
				return
			}
			bar.ChangeMax(len(loop)*3 + 1)

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

			for _, item := range loop {
				// set item to runtime variable
				if err := e.variable.Merge(variable.MergeRuntimeVariable(host, map[string]any{
					"item": item,
				})); err != nil {
					stderr = fmt.Sprintf("set loop item to variable error: %v", err)
					return
				}
				if err := bar.Add(1); err != nil {
					klog.ErrorS(err, "fail to add bar")
				}
				stdout, stderr = e.executeModule(ctx, task, modules.ExecOptions{
					Args:     task.Spec.Module.Args,
					Host:     host,
					Variable: e.variable,
					Task:     *task,
					Pipeline: *e.pipeline,
				})
				if err := bar.Add(1); err != nil {
					klog.ErrorS(err, "fail to add bar")
				}
				// delete item
				if err := e.variable.Merge(variable.MergeRuntimeVariable(host, map[string]any{
					"item": nil,
				})); err != nil {
					stderr = fmt.Sprintf("clean loop item to variable error: %v", err)
					return
				}
				if err := bar.Add(1); err != nil {
					klog.ErrorS(err, "fail to add bar")
				}
			}
		})
	}
	go func() {
		wg.Wait()
		close(dataChan)
	}()

	task.Status.Phase = kubekeyv1alpha1.TaskPhaseSuccess
	for data := range dataChan {
		if data.StdErr != "" {
			if task.Spec.IgnoreError != nil && *task.Spec.IgnoreError {
				task.Status.Phase = kubekeyv1alpha1.TaskPhaseIgnored
			} else {
				task.Status.Phase = kubekeyv1alpha1.TaskPhaseFailed
				task.Status.HostResults = append(task.Status.HostResults, kubekeyv1alpha1.TaskHostResult{
					Host:   data.Host,
					Stdout: data.Stdout,
					StdErr: data.StdErr,
				})
			}
		}
		cd.HostResults = append(cd.HostResults, data)
	}

	return nil
}

func (e executor) execLoop(ctx context.Context, ha map[string]any, task *kubekeyv1alpha1.Task) ([]any, error) {
	switch {
	case task.Spec.Loop.Raw == nil:
		// loop is not set. add one element to execute once module.
		return []any{nil}, nil
	default:
		return variable.Extension2Slice(ha, task.Spec.Loop), nil
	}
}

func (e executor) executeModule(ctx context.Context, task *kubekeyv1alpha1.Task, opts modules.ExecOptions) (string, string) {
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

// merge defined variable to host variable
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
