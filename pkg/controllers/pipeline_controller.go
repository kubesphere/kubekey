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
	"os"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	ctrlcontroller "sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrlfinalizer "sigs.k8s.io/controller-runtime/pkg/finalizer"

	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
)

const (
	labelJob              = "kubekey.kubesphere.io/job"
	defaultExecutorImage  = "hub.kubesphere.com.cn/kubekey/executor:latest"
	defaultPullPolicy     = "IfNotPresent"
	defaultServiceAccount = "kk-executor"
)

type PipelineReconciler struct {
	*runtime.Scheme
	ctrlclient.Client
	record.EventRecorder

	ctrlfinalizer.Finalizers
	MaxConcurrentReconciles int
}

func (r PipelineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// get pipeline
	pipeline := &kubekeyv1.Pipeline{}
	err := r.Client.Get(ctx, req.NamespacedName, pipeline)
	if err != nil {
		if errors.IsNotFound(err) {
			klog.V(5).InfoS("pipeline not found", "pipeline", ctrlclient.ObjectKeyFromObject(pipeline))
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if pipeline.DeletionTimestamp != nil {
		klog.V(5).InfoS("pipeline is deleting", "pipeline", ctrlclient.ObjectKeyFromObject(pipeline))

		return ctrl.Result{}, nil
	}

	switch pipeline.Status.Phase {
	case "":
		excepted := pipeline.DeepCopy()
		pipeline.Status.Phase = kubekeyv1.PipelinePhasePending
		if err := r.Client.Status().Patch(ctx, pipeline, ctrlclient.MergeFrom(excepted)); err != nil {
			klog.V(5).ErrorS(err, "update pipeline error", "pipeline", ctrlclient.ObjectKeyFromObject(pipeline))
			return ctrl.Result{}, err
		}
	case kubekeyv1.PipelinePhasePending:
		excepted := pipeline.DeepCopy()
		pipeline.Status.Phase = kubekeyv1.PipelinePhaseRunning
		if err := r.Client.Status().Patch(ctx, pipeline, ctrlclient.MergeFrom(excepted)); err != nil {
			klog.V(5).ErrorS(err, "update pipeline error", "pipeline", ctrlclient.ObjectKeyFromObject(pipeline))
			return ctrl.Result{}, err
		}
	case kubekeyv1.PipelinePhaseRunning:
		return r.dealRunningPipeline(ctx, pipeline)
	case kubekeyv1.PipelinePhaseFailed:
		// do nothing
	case kubekeyv1.PipelinePhaseSucceed:
		// do nothing
	}
	return ctrl.Result{}, nil
}

func (r *PipelineReconciler) dealRunningPipeline(ctx context.Context, pipeline *kubekeyv1.Pipeline) (ctrl.Result, error) {
	cp := pipeline.DeepCopy()
	defer func() {
		// update pipeline status
		if err := r.Client.Status().Patch(ctx, pipeline, ctrlclient.MergeFrom(cp)); err != nil {
			klog.V(5).ErrorS(err, "update pipeline error", "pipeline", ctrlclient.ObjectKeyFromObject(pipeline))
		}
	}()

	// check if running executor exist
	if jobName, ok := pipeline.Labels[labelJob]; ok {
		if err := r.Client.Get(ctx, ctrlclient.ObjectKey{Namespace: pipeline.Namespace, Name: jobName}, cp); err == nil {
			// job is have create
			return ctrl.Result{}, nil
		} else if !errors.IsNotFound(err) {
			// get job failed
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// get image from env
	image, ok := os.LookupEnv("EXECUTOR_IMAGE")
	if !ok {
		image = defaultExecutorImage
	}
	// get image from env
	imagePullPolicy, ok := os.LookupEnv("EXECUTOR_IMAGE_PULLPOLICY")
	if !ok {
		imagePullPolicy = defaultPullPolicy
	}
	// get ServiceAccount name for executor pod
	saName, ok := os.LookupEnv("EXECUTOR_SERVICEACCOUNT")
	if !ok {
		saName = defaultServiceAccount
	}

	var sa = &corev1.ServiceAccount{}
	if err := r.Client.Get(ctx, ctrlclient.ObjectKey{Namespace: pipeline.Namespace, Name: saName}, sa); err != nil {
		if !errors.IsNotFound(err) {
			klog.ErrorS(err, "get service account", "pipeline", ctrlclient.ObjectKeyFromObject(pipeline))
			return ctrl.Result{}, err
		}
		// create sa
		if err := r.Client.Create(ctx, &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{Name: saName, Namespace: pipeline.Namespace},
		}); err != nil {
			klog.ErrorS(err, "create service account error", "pipeline", ctrlclient.ObjectKeyFromObject(pipeline))
			return ctrl.Result{}, err
		}
	}

	var rb = &rbacv1.ClusterRoleBinding{}
	if err := r.Client.Get(ctx, ctrlclient.ObjectKey{Namespace: pipeline.Namespace, Name: saName}, rb); err != nil {
		if !errors.IsNotFound(err) {
			klog.ErrorS(err, "create role binding error", "pipeline", ctrlclient.ObjectKeyFromObject(pipeline))
			return ctrl.Result{}, err
		}
		//create rolebinding
		if err := r.Client.Create(ctx, &rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{Namespace: pipeline.Namespace, Name: saName},
			RoleRef: rbacv1.RoleRef{
				APIGroup: rbacv1.GroupName,
				Kind:     "ClusterRole",
				Name:     saName,
			},
			Subjects: []rbacv1.Subject{
				{
					APIGroup:  corev1.GroupName,
					Kind:      "ServiceAccount",
					Name:      saName,
					Namespace: pipeline.Namespace,
				},
			},
		}); err != nil {
			klog.ErrorS(err, "create role binding error", "pipeline", ctrlclient.ObjectKeyFromObject(pipeline))
			return ctrl.Result{}, err
		}
	}

	// create a job to executor the pipeline
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: pipeline.Name + "-",
			Namespace:    pipeline.Namespace,
		},
		Spec: batchv1.JobSpec{
			Parallelism:  ptr.To[int32](1),
			Completions:  ptr.To[int32](1),
			BackoffLimit: ptr.To[int32](0),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{},
				Spec: corev1.PodSpec{
					ServiceAccountName: saName,
					RestartPolicy:      "Never",
					Containers: []corev1.Container{
						{
							Name:            "executor",
							Image:           image,
							ImagePullPolicy: corev1.PullPolicy(imagePullPolicy),
							Command:         []string{"kk"},
							Args: []string{"pipeline",
								"--name", pipeline.Name,
								"--namespace", pipeline.Namespace},
						},
					},
				},
			},
		},
	}
	if err := controllerutil.SetOwnerReference(pipeline, job, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	err := r.Create(ctx, job)
	if err != nil {
		return ctrl.Result{}, err
	}

	// add job label to pipeline
	cp = pipeline.DeepCopy()
	metav1.SetMetaDataLabel(&pipeline.ObjectMeta, labelJob, job.Name)
	// update pipeline status
	if err := r.Client.Patch(ctx, pipeline, ctrlclient.MergeFrom(cp)); err != nil {
		klog.V(5).ErrorS(err, "update pipeline error", "pipeline", ctrlclient.ObjectKeyFromObject(pipeline))
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PipelineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(ctrlcontroller.Options{
			MaxConcurrentReconciles: r.MaxConcurrentReconciles,
		}).
		For(&kubekeyv1.Pipeline{}).
		Complete(r)
}
