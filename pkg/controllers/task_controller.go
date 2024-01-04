/*
Copyright 2023 The KubeSphere Authors.

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

package controllers

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"k8s.io/utils/strings/slices"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
	kubekeyv1alpha1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/v4/pkg/cache"
	"github.com/kubesphere/kubekey/v4/pkg/converter"
	"github.com/kubesphere/kubekey/v4/pkg/converter/tmpl"
	"github.com/kubesphere/kubekey/v4/pkg/modules"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

type TaskReconciler struct {
	// Client to resources
	ctrlclient.Client
	// VariableCache to store variable
	VariableCache cache.Cache
}

type taskReconcileOptions struct {
	*kubekeyv1.Pipeline
	*kubekeyv1alpha1.Task
	variable.Variable
}

func (r *TaskReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	klog.V(5).Infof("[Task %s] start reconcile", request.String())
	defer klog.V(5).Infof("[Task %s] finish reconcile", request.String())
	// get task
	var task = &kubekeyv1alpha1.Task{}
	if err := r.Client.Get(ctx, request.NamespacedName, task); err != nil {
		klog.Errorf("get task %s error %v", request, err)
		return ctrl.Result{}, nil
	}

	// if task is deleted, skip
	if task.DeletionTimestamp != nil {
		klog.V(5).Infof("[Task %s] task is deleted, skip", request.String())
		return ctrl.Result{}, nil
	}

	// get pipeline
	var pipeline = &kubekeyv1.Pipeline{}
	for _, ref := range task.OwnerReferences {
		if ref.Kind == "Pipeline" {
			if err := r.Client.Get(ctx, types.NamespacedName{Namespace: task.Namespace, Name: ref.Name}, pipeline); err != nil {
				klog.Errorf("[Task %s] get pipeline %s error %v", request.String(), types.NamespacedName{Namespace: task.Namespace, Name: ref.Name}.String(), err)
				if errors.IsNotFound(err) {
					klog.V(4).Infof("[Task %s] pipeline is deleted, skip", request.String())
					return ctrl.Result{}, nil
				}
				return ctrl.Result{}, err
			}
			break
		}
	}

	if _, ok := pipeline.Annotations[kubekeyv1.PauseAnnotation]; ok {
		klog.V(5).Infof("[Task %s] pipeline is paused, skip", request.String())
		return ctrl.Result{}, nil
	}

	// get variable
	var v variable.Variable
	if vc, ok := r.VariableCache.Get(string(pipeline.UID)); !ok {
		// create new variable
		nv, err := variable.New(variable.Options{
			Ctx:      ctx,
			Client:   r.Client,
			Pipeline: *pipeline,
		})
		if err != nil {
			return ctrl.Result{}, err
		}
		r.VariableCache.Put(string(pipeline.UID), nv)
		v = nv
	} else {
		v = vc.(variable.Variable)
	}

	defer func() {
		var nsTasks = &kubekeyv1alpha1.TaskList{}
		klog.V(5).Infof("[Task %s] update pipeline %s status", ctrlclient.ObjectKeyFromObject(task).String(), ctrlclient.ObjectKeyFromObject(pipeline).String())
		if err := r.Client.List(ctx, nsTasks, ctrlclient.InNamespace(task.Namespace)); err != nil {
			klog.Errorf("[Task %s] list task error %v", ctrlclient.ObjectKeyFromObject(task).String(), err)
			return
		}
		// filter by ownerReference
		for i := len(nsTasks.Items) - 1; i >= 0; i-- {
			var hasOwner bool
			for _, ref := range nsTasks.Items[i].OwnerReferences {
				if ref.UID == pipeline.UID && ref.Kind == "Pipeline" {
					hasOwner = true
				}
			}

			if !hasOwner {
				nsTasks.Items = append(nsTasks.Items[:i], nsTasks.Items[i+1:]...)
			}
		}
		cp := pipeline.DeepCopy()
		converter.CalculatePipelineStatus(nsTasks, pipeline)
		if err := r.Client.Status().Patch(ctx, pipeline, ctrlclient.MergeFrom(cp)); err != nil {
			klog.Errorf("[Task %s] update pipeline %s status error %v", ctrlclient.ObjectKeyFromObject(task).String(), pipeline.Name, err)
		}
	}()

	switch task.Status.Phase {
	case kubekeyv1alpha1.TaskPhaseFailed:
		if task.Spec.Retries > task.Status.RestartCount {
			task.Status.Phase = kubekeyv1alpha1.TaskPhasePending
			task.Status.RestartCount++
			if err := r.Client.Update(ctx, task); err != nil {
				klog.Errorf("update task %s error %v", task.Name, err)
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	case kubekeyv1alpha1.TaskPhasePending:
		// deal pending task
		return r.dealPendingTask(ctx, taskReconcileOptions{
			Pipeline: pipeline,
			Task:     task,
			Variable: v,
		})
	case kubekeyv1alpha1.TaskPhaseRunning:
		// deal running task
		return r.dealRunningTask(ctx, taskReconcileOptions{
			Pipeline: pipeline,
			Task:     task,
			Variable: v,
		})
	default:
		return ctrl.Result{}, nil
	}
}

func (r *TaskReconciler) dealPendingTask(ctx context.Context, options taskReconcileOptions) (ctrl.Result, error) {
	// find dependency tasks
	dl, err := options.Variable.Get(variable.DependencyTasks{
		LocationUID: string(options.Task.UID),
	})
	if err != nil {
		klog.Errorf("[Task %s] find dependency error %v", ctrlclient.ObjectKeyFromObject(options.Task).String(), err)
		return ctrl.Result{}, err
	}
	dt, ok := dl.(variable.DependencyTask)
	if !ok {
		klog.Errorf("[Task %s] failed to convert dependency", ctrlclient.ObjectKeyFromObject(options.Task).String())
		return ctrl.Result{}, fmt.Errorf("[Task %s] failed to convert dependency", ctrlclient.ObjectKeyFromObject(options.Task).String())
	}

	var nsTasks = &kubekeyv1alpha1.TaskList{}
	if err := r.Client.List(ctx, nsTasks, ctrlclient.InNamespace(options.Task.Namespace)); err != nil {
		klog.Errorf("[Task %s] list task error %v", ctrlclient.ObjectKeyFromObject(options.Task).String(), err)
		return ctrl.Result{}, err
	}
	// filter by ownerReference
	for i := len(nsTasks.Items) - 1; i >= 0; i-- {
		var hasOwner bool
		for _, ref := range nsTasks.Items[i].OwnerReferences {
			if ref.UID == options.Pipeline.UID && ref.Kind == "Pipeline" {
				hasOwner = true
			}
		}

		if !hasOwner {
			nsTasks.Items = append(nsTasks.Items[:i], nsTasks.Items[i+1:]...)
		}
	}
	var dts []kubekeyv1alpha1.Task
	for _, t := range nsTasks.Items {
		if slices.Contains(dt.Tasks, string(t.UID)) {
			dts = append(dts, t)
		}
	}
	// Based on the results of the executed tasks dependent on, infer the next phase of the current task.
	switch dt.Strategy(dts) {
	case kubekeyv1alpha1.TaskPhasePending:
		return ctrl.Result{Requeue: true}, nil
	case kubekeyv1alpha1.TaskPhaseRunning:
		// update task phase to running
		options.Task.Status.Phase = kubekeyv1alpha1.TaskPhaseRunning
		if err := r.Client.Update(ctx, options.Task); err != nil {
			klog.Errorf("[Task %s] update task to Running error %v", ctrlclient.ObjectKeyFromObject(options.Task), err)
		}
		return ctrl.Result{Requeue: true}, nil
	case kubekeyv1alpha1.TaskPhaseSkipped:
		options.Task.Status.Phase = kubekeyv1alpha1.TaskPhaseSkipped
		if err := r.Client.Update(ctx, options.Task); err != nil {
			klog.Errorf("[Task %s] update task to Skipped error %v", ctrlclient.ObjectKeyFromObject(options.Task), err)
		}
		return ctrl.Result{}, nil
	default:
		return ctrl.Result{}, fmt.Errorf("unknown TependencyTask.Strategy result. only support: Pending, Running, Skipped")
	}
}

func (r *TaskReconciler) dealRunningTask(ctx context.Context, options taskReconcileOptions) (ctrl.Result, error) {
	// find task in location
	klog.Infof("[Task %s] dealRunningTask begin", ctrlclient.ObjectKeyFromObject(options.Task))
	defer func() {
		klog.Infof("[Task %s] dealRunningTask end, task phase: %s", ctrlclient.ObjectKeyFromObject(options.Task), options.Task.Status.Phase)
	}()

	if err := r.executeTask(ctx, options); err != nil {
		klog.Errorf("[Task %s] execute task error %v", ctrlclient.ObjectKeyFromObject(options.Task), err)
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, nil
}

func (r *TaskReconciler) executeTask(ctx context.Context, options taskReconcileOptions) error {
	cd := kubekeyv1alpha1.TaskCondition{
		StartTimestamp: metav1.Now(),
	}
	defer func() {
		cd.EndTimestamp = metav1.Now()
		options.Task.Status.Conditions = append(options.Task.Status.Conditions, cd)
		if err := r.Client.Update(ctx, options.Task); err != nil {
			klog.Errorf("[Task %s] update task status error %v", ctrlclient.ObjectKeyFromObject(options.Task), err)
		}
	}()

	// check task host results
	wg := &wait.Group{}
	dataChan := make(chan kubekeyv1alpha1.TaskHostResult, len(options.Task.Spec.Hosts))
	for _, h := range options.Task.Spec.Hosts {
		host := h
		wg.StartWithContext(ctx, func(ctx context.Context) {
			var stdout, stderr string
			defer func() {
				if stderr != "" {
					klog.Errorf("[Task %s] run failed: %s", ctrlclient.ObjectKeyFromObject(options.Task), stderr)
				}

				dataChan <- kubekeyv1alpha1.TaskHostResult{
					Host:   host,
					Stdout: stdout,
					StdErr: stderr,
				}
				if options.Task.Spec.Register != "" {
					puid, err := options.Variable.Get(variable.ParentLocation{LocationUID: string(options.Task.UID)})
					if err != nil {
						klog.Errorf("[Task %s] get location error %v", ctrlclient.ObjectKeyFromObject(options.Task), err)
						return
					}
					// set variable to parent location
					if err := options.Variable.Merge(variable.HostMerge{
						HostNames:   []string{h},
						LocationUID: puid.(string),
						Data: variable.VariableData{
							options.Task.Spec.Register: map[string]string{
								"stdout": stdout,
								"stderr": stderr,
							},
						},
					}); err != nil {
						klog.Errorf("[Task %s] register error %v", ctrlclient.ObjectKeyFromObject(options.Task), err)
						return
					}
				}
			}()

			lg, err := options.Variable.Get(variable.LocationVars{
				HostName:    host,
				LocationUID: string(options.Task.UID),
			})
			if err != nil {
				stderr = err.Error()
				return
			}
			// check when condition
			if len(options.Task.Spec.When) > 0 {
				ok, err := tmpl.ParseBool(lg.(variable.VariableData), options.Task.Spec.When)
				if err != nil {
					stderr = err.Error()
					return
				}
				if !ok {
					stdout = "skip by when"
					return
				}
			}

			data := variable.Extension2Slice(options.Task.Spec.Loop)
			if len(data) == 0 {
				stdout, stderr = r.executeModule(ctx, options.Task, modules.ExecOptions{
					Args:     options.Task.Spec.Module.Args,
					Host:     host,
					Variable: options.Variable,
					Task:     *options.Task,
					Pipeline: *options.Pipeline,
				})
			} else {
				for _, item := range data {
					switch item.(type) {
					case string:
						item, err = tmpl.ParseString(lg.(variable.VariableData), item.(string))
						if err != nil {
							stderr = err.Error()
							return
						}
					case variable.VariableData:
						for k, v := range item.(variable.VariableData) {
							sv, err := tmpl.ParseString(lg.(variable.VariableData), v.(string))
							if err != nil {
								stderr = err.Error()
								return
							}
							item.(map[string]any)[k] = sv
						}
					default:
						stderr = "unknown loop vars, only support string or map[string]string"
						return
					}
					// set item to runtime variable
					options.Variable.Merge(variable.HostMerge{
						HostNames:   []string{h},
						LocationUID: string(options.Task.UID),
						Data: variable.VariableData{
							"item": item,
						},
					})
					stdout, stderr = r.executeModule(ctx, options.Task, modules.ExecOptions{
						Args:     options.Task.Spec.Module.Args,
						Host:     host,
						Variable: options.Variable,
						Task:     *options.Task,
						Pipeline: *options.Pipeline,
					})
				}
			}
		})
	}
	go func() {
		wg.Wait()
		close(dataChan)
	}()

	options.Task.Status.Phase = kubekeyv1alpha1.TaskPhaseSuccess
	for data := range dataChan {
		if data.StdErr != "" {
			if options.Task.Spec.IgnoreError {
				options.Task.Status.Phase = kubekeyv1alpha1.TaskPhaseIgnored
			} else {
				options.Task.Status.Phase = kubekeyv1alpha1.TaskPhaseFailed
				options.Task.Status.FailedDetail = append(options.Task.Status.FailedDetail, kubekeyv1alpha1.TaskFailedDetail{
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

func (r *TaskReconciler) executeModule(ctx context.Context, task *kubekeyv1alpha1.Task, opts modules.ExecOptions) (string, string) {
	lg, err := opts.Variable.Get(variable.LocationVars{
		HostName:    opts.Host,
		LocationUID: string(task.UID),
	})
	if err != nil {
		klog.Errorf("[Task %s] get location variable error %v", ctrlclient.ObjectKeyFromObject(task), err)
		return "", err.Error()
	}

	// check failed when condition
	if len(task.Spec.FailedWhen) > 0 {
		ok, err := tmpl.ParseBool(lg.(variable.VariableData), task.Spec.FailedWhen)
		if err != nil {
			klog.Errorf("[Task %s] validate FailedWhen condition error %v", ctrlclient.ObjectKeyFromObject(task), err)
			return "", err.Error()
		}
		if ok {
			return "", "failed by failedWhen"
		}
	}

	return modules.FindModule(task.Spec.Module.Name)(ctx, opts)
}
