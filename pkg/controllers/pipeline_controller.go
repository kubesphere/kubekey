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
	"path/filepath"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/task"
)

type PipelineReconciler struct {
	ctrlclient.Client
	record.EventRecorder

	TaskController task.Controller
}

func (r PipelineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.Infof("[Pipeline %s] begin reconcile", req.NamespacedName.String())
	defer func() {
		klog.Infof("[Pipeline %s] end reconcile", req.NamespacedName.String())
	}()

	pipeline := &kubekeyv1.Pipeline{}
	err := r.Client.Get(ctx, req.NamespacedName, pipeline)
	if err != nil {
		if errors.IsNotFound(err) {
			klog.V(5).Infof("[Pipeline %s] pipeline not found", req.NamespacedName.String())
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if pipeline.DeletionTimestamp != nil {
		klog.V(5).Infof("[Pipeline %s] pipeline is deleting", req.NamespacedName.String())
		return ctrl.Result{}, nil
	}

	switch pipeline.Status.Phase {
	case "":
		excepted := pipeline.DeepCopy()
		pipeline.Status.Phase = kubekeyv1.PipelinePhasePending
		if err := r.Client.Status().Patch(ctx, pipeline, ctrlclient.MergeFrom(excepted)); err != nil {
			klog.Errorf("[Pipeline %s] update pipeline error: %v", ctrlclient.ObjectKeyFromObject(pipeline), err)
			return ctrl.Result{}, err
		}
	case kubekeyv1.PipelinePhasePending:
		excepted := pipeline.DeepCopy()
		pipeline.Status.Phase = kubekeyv1.PipelinePhaseRunning
		if err := r.Client.Status().Patch(ctx, pipeline, ctrlclient.MergeFrom(excepted)); err != nil {
			klog.Errorf("[Pipeline %s] update pipeline error: %v", ctrlclient.ObjectKeyFromObject(pipeline), err)
			return ctrl.Result{}, err
		}
	case kubekeyv1.PipelinePhaseRunning:
		return r.dealRunningPipeline(ctx, pipeline)
	case kubekeyv1.PipelinePhaseFailed:
		r.clean(ctx, pipeline)
	case kubekeyv1.PipelinePhaseSucceed:
		r.clean(ctx, pipeline)
	}
	return ctrl.Result{}, nil
}

func (r *PipelineReconciler) dealRunningPipeline(ctx context.Context, pipeline *kubekeyv1.Pipeline) (ctrl.Result, error) {
	if _, ok := pipeline.Annotations[kubekeyv1.PauseAnnotation]; ok {
		// if pipeline is paused, do nothing
		klog.V(5).Infof("[Pipeline %s] pipeline is paused", ctrlclient.ObjectKeyFromObject(pipeline))
		return ctrl.Result{}, nil
	}

	cp := pipeline.DeepCopy()
	defer func() {
		// update pipeline status
		if err := r.Client.Status().Patch(ctx, pipeline, ctrlclient.MergeFrom(cp)); err != nil {
			klog.Errorf("[Pipeline %s] update pipeline error: %v", ctrlclient.ObjectKeyFromObject(pipeline), err)
		}
	}()

	if err := r.TaskController.AddTasks(ctx, task.AddTaskOptions{
		Pipeline: pipeline,
	}); err != nil {
		klog.Errorf("[Pipeline %s] add task error: %v", ctrlclient.ObjectKeyFromObject(pipeline), err)
		pipeline.Status.Phase = kubekeyv1.PipelinePhaseFailed
		pipeline.Status.Reason = fmt.Sprintf("add task to controller failed: %v", err)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// clean runtime directory
func (r *PipelineReconciler) clean(ctx context.Context, pipeline *kubekeyv1.Pipeline) {
	if !pipeline.Spec.Debug && pipeline.Status.Phase == kubekeyv1.PipelinePhaseSucceed {
		klog.Infof("[Pipeline %s] clean runtimeDir", ctrlclient.ObjectKeyFromObject(pipeline))
		// clean runtime directory
		if err := os.RemoveAll(filepath.Join(_const.GetWorkDir(), _const.RuntimeDir)); err != nil {
			klog.Errorf("clean runtime directory %s error: %v", filepath.Join(_const.GetWorkDir(), _const.RuntimeDir), err)
		}
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *PipelineReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options Options) error {
	if !options.IsControllerEnabled("pipeline") {
		klog.V(5).Infof("pipeline controller is disabled")
		return nil
	}

	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(options.Options).
		For(&kubekeyv1.Pipeline{}).
		Complete(r)
}
