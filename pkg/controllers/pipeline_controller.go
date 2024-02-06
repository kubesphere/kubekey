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
	"os"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrlfinalizer "sigs.k8s.io/controller-runtime/pkg/finalizer"

	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
	kubekeyv1alpha1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1alpha1"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/task"
)

const (
	pipelineFinalizer = "kubekey.kubesphere.io/pipeline"
)

type PipelineReconciler struct {
	ctrlclient.Client
	record.EventRecorder

	TaskController task.Controller

	ctrlfinalizer.Finalizers
}

func (r PipelineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.V(5).InfoS("start pipeline reconcile", "pipeline", req.String())
	defer klog.V(5).InfoS("finish pipeline reconcile", "pipeline", req.String())
	// get pipeline
	pipeline := &kubekeyv1.Pipeline{}
	err := r.Client.Get(ctx, req.NamespacedName, pipeline)
	if err != nil {
		if errors.IsNotFound(err) {
			klog.V(5).InfoS("pipeline not found", "pipeline", req.String())
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if pipeline.DeletionTimestamp != nil {
		klog.V(5).InfoS("pipeline is deleting", "pipeline", req.String())
		if controllerutil.ContainsFinalizer(pipeline, pipelineFinalizer) {
			r.clean(ctx, pipeline)
			// remove finalizer

		}

		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(pipeline, pipelineFinalizer) {
		excepted := pipeline.DeepCopy()
		controllerutil.AddFinalizer(pipeline, pipelineFinalizer)
		if err := r.Client.Patch(ctx, pipeline, ctrlclient.MergeFrom(excepted)); err != nil {
			klog.V(5).ErrorS(err, "update pipeline error", "pipeline", req.String())
			return ctrl.Result{}, err
		}
	}

	switch pipeline.Status.Phase {
	case "":
		excepted := pipeline.DeepCopy()
		pipeline.Status.Phase = kubekeyv1.PipelinePhasePending
		if err := r.Client.Status().Patch(ctx, pipeline, ctrlclient.MergeFrom(excepted)); err != nil {
			klog.V(5).ErrorS(err, "update pipeline error", "pipeline", req.String())
			return ctrl.Result{}, err
		}
	case kubekeyv1.PipelinePhasePending:
		excepted := pipeline.DeepCopy()
		pipeline.Status.Phase = kubekeyv1.PipelinePhaseRunning
		if err := r.Client.Status().Patch(ctx, pipeline, ctrlclient.MergeFrom(excepted)); err != nil {
			klog.V(5).ErrorS(err, "update pipeline error", "pipeline", req.String())
			return ctrl.Result{}, err
		}
	case kubekeyv1.PipelinePhaseRunning:
		return r.dealRunningPipeline(ctx, pipeline)
	case kubekeyv1.PipelinePhaseFailed:
		// do nothing
	case kubekeyv1.PipelinePhaseSucceed:
		if !pipeline.Spec.Debug {
			r.clean(ctx, pipeline)
		}
	}
	return ctrl.Result{}, nil
}

func (r *PipelineReconciler) dealRunningPipeline(ctx context.Context, pipeline *kubekeyv1.Pipeline) (ctrl.Result, error) {
	if _, ok := pipeline.Annotations[kubekeyv1.PauseAnnotation]; ok {
		// if pipeline is paused, do nothing
		klog.V(5).InfoS("pipeline is paused", "pipeline", ctrlclient.ObjectKeyFromObject(pipeline))
		return ctrl.Result{}, nil
	}

	cp := pipeline.DeepCopy()
	defer func() {
		// update pipeline status
		if err := r.Client.Status().Patch(ctx, pipeline, ctrlclient.MergeFrom(cp)); err != nil {
			klog.V(5).ErrorS(err, "update pipeline error", "pipeline", ctrlclient.ObjectKeyFromObject(pipeline))
		}
	}()

	if err := r.TaskController.AddTasks(ctx, task.AddTaskOptions{
		Pipeline: pipeline,
	}); err != nil {
		klog.V(5).ErrorS(err, "add task error", "pipeline", ctrlclient.ObjectKeyFromObject(pipeline))
		pipeline.Status.Phase = kubekeyv1.PipelinePhaseFailed
		pipeline.Status.Reason = fmt.Sprintf("add task to controller failed: %v", err)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// clean runtime directory
func (r *PipelineReconciler) clean(ctx context.Context, pipeline *kubekeyv1.Pipeline) {
	klog.V(5).InfoS("clean runtimeDir", "pipeline", ctrlclient.ObjectKeyFromObject(pipeline))
	// delete reference task
	taskList := &kubekeyv1alpha1.TaskList{}
	if err := r.Client.List(ctx, taskList, ctrlclient.MatchingFields{}); err != nil {
		klog.V(5).ErrorS(err, "list task error", "pipeline", ctrlclient.ObjectKeyFromObject(pipeline))
		return
	}
	if err := os.RemoveAll(_const.GetRuntimeDir()); err != nil {
		klog.V(5).ErrorS(err, "clean runtime directory error", "runtime dir", _const.GetRuntimeDir(), "pipeline", ctrlclient.ObjectKeyFromObject(pipeline))
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *PipelineReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options Options) error {
	if !options.IsControllerEnabled("pipeline") {
		klog.V(5).InfoS("controller is disabled", "controller", "pipeline")
		return nil
	}

	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(options.Options).
		For(&kubekeyv1.Pipeline{}).
		Complete(r)
}
