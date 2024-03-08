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
	"reflect"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
	kubekeyv1alpha1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/v4/pkg/converter"
	"github.com/kubesphere/kubekey/v4/pkg/converter/tmpl"
	"github.com/kubesphere/kubekey/v4/pkg/modules"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

type TaskReconciler struct {
	// Client to resources
	ctrlclient.Client
}

type taskReconcileOptions struct {
	*kubekeyv1.Pipeline
	*kubekeyv1alpha1.Task
	variable.Variable
}

func (r *TaskReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	klog.V(5).InfoS("start task reconcile", "task", request.String())
	defer klog.V(5).InfoS("finish task reconcile", "task", request.String())
	// get task
	var task = &kubekeyv1alpha1.Task{}
	if err := r.Client.Get(ctx, request.NamespacedName, task); err != nil {
		klog.V(5).ErrorS(err, "get task error", "task", request.String())
		return ctrl.Result{}, nil
	}

	// if task is deleted, skip
	if task.DeletionTimestamp != nil {
		klog.V(5).InfoS("task is deleted, skip", "task", request.String())
		return ctrl.Result{}, nil
	}

	// get pipeline
	var pipeline = &kubekeyv1.Pipeline{}
	for _, ref := range task.OwnerReferences {
		if ref.Kind == "Pipeline" {
			if err := r.Client.Get(ctx, types.NamespacedName{Namespace: task.Namespace, Name: ref.Name}, pipeline); err != nil {
				if errors.IsNotFound(err) {
					klog.V(5).InfoS("pipeline is deleted, skip", "task", request.String())
					return ctrl.Result{}, nil
				}
				klog.V(5).ErrorS(err, "get pipeline error", "task", request.String(), "pipeline", types.NamespacedName{Namespace: task.Namespace, Name: ref.Name}.String())
				return ctrl.Result{}, err
			}
			break
		}
	}

	if _, ok := pipeline.Annotations[kubekeyv1.PauseAnnotation]; ok {
		klog.V(5).InfoS("pipeline is paused, skip", "task", request.String())
		return ctrl.Result{}, nil
	}

	// get variable
	v, err := variable.GetVariable(variable.Options{
		Ctx:      ctx,
		Client:   r.Client,
		Pipeline: *pipeline,
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	defer func() {
		if task.IsComplete() {
			klog.Infof("[Task %s] \"%s\" is complete.Result is: %s", request.String(), task.Spec.Name, task.Status.Phase)
		}
		var nsTasks = &kubekeyv1alpha1.TaskList{}
		klog.V(5).InfoS("update pipeline status", "task", request.String(), "pipeline", ctrlclient.ObjectKeyFromObject(pipeline).String())
		if err := r.Client.List(ctx, nsTasks, ctrlclient.InNamespace(pipeline.Namespace), ctrlclient.MatchingFields{
			kubekeyv1alpha1.TaskOwnerField: ctrlclient.ObjectKeyFromObject(pipeline).String(),
		}); err != nil {
			klog.V(5).ErrorS(err, "list task error", "task", request.String())
			return
		}
		cp := pipeline.DeepCopy()
		converter.CalculatePipelineStatus(nsTasks, pipeline)
		if err := r.Client.Status().Patch(ctx, pipeline, ctrlclient.MergeFrom(cp)); err != nil {
			klog.V(5).ErrorS(err, "update pipeline status error", "task", request.String(), "pipeline", ctrlclient.ObjectKeyFromObject(pipeline).String())
		}
	}()

	switch task.Status.Phase {
	case kubekeyv1alpha1.TaskPhaseFailed:
		if task.Spec.Retries > task.Status.RestartCount {
			task.Status.Phase = kubekeyv1alpha1.TaskPhasePending
			task.Status.RestartCount++
			if err := r.Client.Status().Update(ctx, task); err != nil {
				klog.V(5).ErrorS(err, "update task error", "task", request.String())
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
	var nsTasks = &kubekeyv1alpha1.TaskList{}
	if err := r.Client.List(ctx, nsTasks, ctrlclient.InNamespace(options.Pipeline.Namespace), ctrlclient.MatchingFields{
		kubekeyv1alpha1.TaskOwnerField: ctrlclient.ObjectKeyFromObject(options.Pipeline).String(),
	}); err != nil {
		klog.V(5).ErrorS(err, "list task error", "task", ctrlclient.ObjectKeyFromObject(options.Task).String(), err)
		return ctrl.Result{}, err
	}

	// Infer the current task's phase from its dependent tasks.
	dl, err := options.Variable.Get(variable.InferPhase{
		LocationUID: string(options.Task.UID),
		Tasks:       nsTasks.Items,
	})
	klog.InfoS("infer phase", "phase", dl, "task-name", options.Task.Spec.Name)
	if err != nil {
		klog.V(5).ErrorS(err, "find dependency error", "task", ctrlclient.ObjectKeyFromObject(options.Task).String())
		return ctrl.Result{}, err
	}

	// Based on the results of the executed tasks dependent on, infer the next phase of the current task.
	switch dl.(kubekeyv1alpha1.TaskPhase) {
	case kubekeyv1alpha1.TaskPhasePending:
		return ctrl.Result{Requeue: true}, nil
	case kubekeyv1alpha1.TaskPhaseRunning:
		// update task phase to running
		options.Task.Status.Phase = kubekeyv1alpha1.TaskPhaseRunning
		if err := r.Client.Status().Update(ctx, options.Task); err != nil {
			klog.V(5).ErrorS(err, "update task to Running error", "task", ctrlclient.ObjectKeyFromObject(options.Task))
		}
		return ctrl.Result{Requeue: true}, nil
	case kubekeyv1alpha1.TaskPhaseSkipped:
		options.Task.Status.Phase = kubekeyv1alpha1.TaskPhaseSkipped
		if err := r.Client.Status().Update(ctx, options.Task); err != nil {
			klog.V(5).ErrorS(err, "update task to Skipped error", "task", ctrlclient.ObjectKeyFromObject(options.Task))
		}
		return ctrl.Result{}, nil
	default:
		return ctrl.Result{}, fmt.Errorf("unknown TependencyTask.Strategy result. only support: Pending, Running, Skipped")
	}
}

func (r *TaskReconciler) dealRunningTask(ctx context.Context, options taskReconcileOptions) (ctrl.Result, error) {
	if err := r.prepareTask(ctx, options); err != nil {
		klog.V(5).ErrorS(err, "prepare task error", "task", ctrlclient.ObjectKeyFromObject(options.Task))
		return ctrl.Result{}, nil
	}
	// find task in location
	if err := r.executeTask(ctx, options); err != nil {
		klog.V(5).ErrorS(err, "execute task error", "task", ctrlclient.ObjectKeyFromObject(options.Task))
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, nil
}

func (r *TaskReconciler) prepareTask(ctx context.Context, options taskReconcileOptions) error {
	// trans variable to location
	// if variable contains template syntax. parse it and store in host.
	for _, h := range options.Task.Spec.Hosts {
		host := h
		lg, err := options.Variable.Get(variable.LocationVars{
			HostName:    host,
			LocationUID: string(options.Task.UID),
		})
		if err != nil {
			klog.V(5).ErrorS(err, "get location variable error", "task", ctrlclient.ObjectKeyFromObject(options.Task))
			options.Task.Status.Phase = kubekeyv1alpha1.TaskPhaseFailed
			options.Task.Status.FailedDetail = append(options.Task.Status.FailedDetail, kubekeyv1alpha1.TaskFailedDetail{
				Host:   host,
				StdErr: "parse variable error",
			})
			return err
		}

		var curVariable = lg.(variable.VariableData).DeepCopy()
		if pt := variable.BoolVar(curVariable, "prepareTask"); pt != nil && *pt {
			klog.InfoS("prepareTask is true, skip", "task", ctrlclient.ObjectKeyFromObject(options.Task), "host", host)
			continue
		}

		var parseTmpl = func(tmplStr string) (string, error) {
			return tmpl.ParseString(curVariable, tmplStr)
		}
		// parse variable with three time. ( support 3 level reference.)
		for i := 0; i < 3; i++ {
			if err := r.parseVariable(ctx, curVariable, parseTmpl); err != nil {
				klog.V(5).ErrorS(err, "parse variable error", "task", ctrlclient.ObjectKeyFromObject(options.Task))
				options.Task.Status.Phase = kubekeyv1alpha1.TaskPhaseFailed
				options.Task.Status.FailedDetail = append(options.Task.Status.FailedDetail, kubekeyv1alpha1.TaskFailedDetail{
					Host:   host,
					StdErr: fmt.Sprintf("parse variable error: %s", err.Error()),
				})
				return err
			}
		}

		// set prepareTask to true
		curVariable["prepareTask"] = true
		if err := options.Variable.Merge(variable.HostMerge{
			HostNames:   []string{h},
			LocationUID: string(options.Task.UID),
			Data:        curVariable,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (r *TaskReconciler) parseVariable(ctx context.Context, in variable.VariableData, parseTmplFunc func(string) (string, error)) error {
	for k, v := range in {
		switch reflect.TypeOf(v).Kind() {
		case reflect.String:
			if r.isTmplSyntax(v.(string)) {
				newValue, err := parseTmplFunc(v.(string))
				if err != nil {
					return err
				}
				in[k] = newValue
			}
		case reflect.Map:
			// variable.VariableData has one more String() method than map[string]any,
			// so variable.VariableData and map[string]any cannot be converted to each other.
			if vv, ok := v.(map[string]interface{}); ok {
				if err := r.parseVariable(ctx, vv, parseTmplFunc); err != nil {
					return err
				}
			}
			if vv, ok := v.(variable.VariableData); ok {
				if err := r.parseVariable(ctx, vv, parseTmplFunc); err != nil {
					return err
				}
			}
		case reflect.Slice:
			for i := 0; i < reflect.ValueOf(v).Len(); i++ {
				elem := reflect.ValueOf(v).Index(i)
				switch elem.Kind() {
				case reflect.String:
					if r.isTmplSyntax(elem.Interface().(string)) {
						newValue, err := parseTmplFunc(elem.Interface().(string))
						if err != nil {
							return err
						}
						reflect.ValueOf(v).Index(i).SetString(newValue)
					}
				case reflect.Map:
					// variable.VariableData has one more String() method than map[string]any,
					// so variable.VariableData and map[string]any cannot be converted to each other.
					if vv, ok := v.(map[string]interface{}); ok {
						if err := r.parseVariable(ctx, vv, parseTmplFunc); err != nil {
							return err
						}
					}
					if vv, ok := v.(variable.VariableData); ok {
						if err := r.parseVariable(ctx, vv, parseTmplFunc); err != nil {
							return err
						}
					}
				}
			}
		}
	}
	return nil
}

func (r *TaskReconciler) isTmplSyntax(s string) bool {
	return (strings.Contains(s, "{{") && strings.Contains(s, "}}")) ||
		(strings.Contains(s, "{%") && strings.Contains(s, "%}"))
}

func (r *TaskReconciler) executeTask(ctx context.Context, options taskReconcileOptions) error {
	cd := kubekeyv1alpha1.TaskCondition{
		StartTimestamp: metav1.Now(),
	}
	defer func() {
		cd.EndTimestamp = metav1.Now()
		options.Task.Status.Conditions = append(options.Task.Status.Conditions, cd)
		if err := r.Client.Status().Update(ctx, options.Task); err != nil {
			klog.V(5).ErrorS(err, "update task status error", "task", ctrlclient.ObjectKeyFromObject(options.Task))
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
						klog.V(5).ErrorS(err, "get location error", "task", ctrlclient.ObjectKeyFromObject(options.Task))
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
						klog.V(5).ErrorS(err, "register task result to variable error", "task", ctrlclient.ObjectKeyFromObject(options.Task))
						return
					}
				}
			}()

			lg, err := options.Variable.Get(variable.LocationVars{
				HostName:    host,
				LocationUID: string(options.Task.UID),
			})
			if err != nil {
				klog.V(5).ErrorS(err, "get location variable error", "task", ctrlclient.ObjectKeyFromObject(options.Task))
				stderr = err.Error()
				return
			}
			// check when condition
			if len(options.Task.Spec.When) > 0 {
				ok, err := tmpl.ParseBool(lg.(variable.VariableData), options.Task.Spec.When)
				if err != nil {
					klog.V(5).ErrorS(err, "parse when condition error", "task", ctrlclient.ObjectKeyFromObject(options.Task))
					stderr = err.Error()
					return
				}
				if !ok {
					stdout = "skip by when"
					return
				}
			}

			// execute module with loop
			loop, err := r.execLoop(ctx, host, options)
			if err != nil {
				klog.V(5).ErrorS(err, "parse loop vars error", "task", ctrlclient.ObjectKeyFromObject(options.Task))
				stderr = err.Error()
				return
			}

			for _, item := range loop {
				switch item.(type) {
				case nil:
					// do nothing
				case string:
					item, err = tmpl.ParseString(lg.(variable.VariableData), item.(string))
					if err != nil {
						klog.V(5).ErrorS(err, "parse loop vars error", "task", ctrlclient.ObjectKeyFromObject(options.Task))
						stderr = err.Error()
						return
					}
				case variable.VariableData:
					for k, v := range item.(variable.VariableData) {
						sv, err := tmpl.ParseString(lg.(variable.VariableData), v.(string))
						if err != nil {
							klog.V(5).ErrorS(err, "parse loop vars error", "task", ctrlclient.ObjectKeyFromObject(options.Task))
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
				if err := options.Variable.Merge(variable.HostMerge{
					HostNames:   []string{h},
					LocationUID: string(options.Task.UID),
					Data: variable.VariableData{
						"item": item,
					},
				}); err != nil {
					stderr = "set loop item to variable error"
					return
				}
				stdout, stderr = r.executeModule(ctx, options.Task, modules.ExecOptions{
					Args:     options.Task.Spec.Module.Args,
					Host:     host,
					Variable: options.Variable,
					Task:     *options.Task,
					Pipeline: *options.Pipeline,
				})
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

func (r *TaskReconciler) execLoop(ctx context.Context, host string, options taskReconcileOptions) ([]any, error) {
	switch {
	case options.Task.Spec.Loop.Raw == nil:
		// loop is not set. add one element to execute once module.
		return []any{nil}, nil
	case variable.Extension2Slice(options.Task.Spec.Loop) != nil:
		return variable.Extension2Slice(options.Task.Spec.Loop), nil
	case variable.Extension2String(options.Task.Spec.Loop) != "":
		value := variable.Extension2String(options.Task.Spec.Loop)
		// parse value by pongo2. if
		data, err := options.Variable.Get(variable.LocationVars{
			HostName:    host,
			LocationUID: string(options.Task.UID),
		})
		if err != nil {
			return nil, err
		}
		sv, err := tmpl.ParseString(data.(variable.VariableData), value)
		if err != nil {
			return nil, err
		}
		switch {
		case regexp.MustCompile(`^<\[\](.*?) Value>$`).MatchString(sv):
			// in pongo2 we cannot get slice value. add extension filter value.
			vdata, err := options.Variable.Get(variable.KeyPath{
				HostName:    host,
				LocationUID: string(options.Task.UID),
				Path: strings.Split(strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(value, "{{"), "}}")),
					"."),
			})
			if err != nil {
				return nil, err
			}
			if _, ok := vdata.([]any); ok {
				return vdata.([]any), nil
			}
		default:
			// value is simple string
			return []any{sv}, nil
		}
	}
	return nil, fmt.Errorf("unsupport loop value")
}

func (r *TaskReconciler) executeModule(ctx context.Context, task *kubekeyv1alpha1.Task, opts modules.ExecOptions) (string, string) {
	lg, err := opts.Variable.Get(variable.LocationVars{
		HostName:    opts.Host,
		LocationUID: string(task.UID),
	})
	if err != nil {
		klog.V(5).ErrorS(err, "get location variable error", "task", ctrlclient.ObjectKeyFromObject(task))
		return "", err.Error()
	}

	// check failed when condition
	if len(task.Spec.FailedWhen) > 0 {
		ok, err := tmpl.ParseBool(lg.(variable.VariableData), task.Spec.FailedWhen)
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
