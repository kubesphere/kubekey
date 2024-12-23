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

package core

import (
	"context"
	"os"

	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	ctrlcontroller "sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
)

const (
	pipelineControllerName = "pipeline"
	// pipelinePodLabel set in pod. value is which pipeline belongs to.
	podPipelineLabel     = "kubekey.kubesphere.io/pipeline"
	defaultExecutorImage = "hub.kubesphere.com.cn/kubekey/executor:latest"
	defaultPullPolicy    = "IfNotPresent"
)

// PipelineReconciler reconcile pipeline
type PipelineReconciler struct {
	*runtime.Scheme
	ctrlclient.Client
	record.EventRecorder

	MaxConcurrentReconciles int
}

// Name implements controllers.controller.
// Subtle: this method shadows the method (*Scheme).Name of PipelineReconciler.Scheme.
func (r *PipelineReconciler) Name() string {
	return pipelineControllerName
}

// SetupWithManager implements controllers.controller.
func (r *PipelineReconciler) SetupWithManager(mgr manager.Manager, o ctrlcontroller.Options) error {
	r.Scheme = mgr.GetScheme()
	r.Client = mgr.GetClient()
	r.EventRecorder = mgr.GetEventRecorderFor(r.Name())

	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(o).
		For(&kkcorev1.Pipeline{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=kubekey.kubesphere.io,resources=*,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods;events,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="coordination.k8s.io",resources=leases,verbs=get;list;watch;create;update;patch;delete

func (r PipelineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// get pipeline
	pipeline := &kkcorev1.Pipeline{}
	err := r.Client.Get(ctx, req.NamespacedName, pipeline)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	if pipeline.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	switch pipeline.Status.Phase {
	case "":
		excepted := pipeline.DeepCopy()
		pipeline.Status.Phase = kkcorev1.PipelinePhasePending
		if err := r.Client.Status().Patch(ctx, pipeline, ctrlclient.MergeFrom(excepted)); err != nil {
			return ctrl.Result{}, err
		}
	case kkcorev1.PipelinePhasePending:
		excepted := pipeline.DeepCopy()
		pipeline.Status.Phase = kkcorev1.PipelinePhaseRunning
		if err := r.Client.Status().Patch(ctx, pipeline, ctrlclient.MergeFrom(excepted)); err != nil {
			return ctrl.Result{}, err
		}
	case kkcorev1.PipelinePhaseRunning:
		return r.dealRunningPipeline(ctx, pipeline)
	}

	return ctrl.Result{}, nil
}

func (r *PipelineReconciler) dealRunningPipeline(ctx context.Context, pipeline *kkcorev1.Pipeline) (ctrl.Result, error) {
	// check if pod is exist
	pods := &corev1.PodList{}
	if err := r.Client.List(ctx, pods, ctrlclient.InNamespace(pipeline.Namespace), ctrlclient.MatchingLabels{
		podPipelineLabel: pipeline.Name,
	}); err != nil && !apierrors.IsNotFound(err) {
		return ctrl.Result{}, err
	} else if len(pods.Items) != 0 {
		// could find exist pod
		return ctrl.Result{}, nil
	}
	// get image from env
	image := os.Getenv(_const.ENV_EXECUTOR_IMAGE)
	if image == "" {
		image = defaultExecutorImage
	}
	// get image from env
	imagePullPolicy := os.Getenv(_const.ENV_EXECUTOR_IMAGE_PULLPOLICY)
	if imagePullPolicy == "" {
		imagePullPolicy = defaultPullPolicy
	}

	// create pod
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: pipeline.Name + "-",
			Namespace:    pipeline.Namespace,
			Labels: map[string]string{
				podPipelineLabel: pipeline.Name,
			},
		},
		Spec: corev1.PodSpec{
			ServiceAccountName: pipeline.Spec.ServiceAccountName,
			RestartPolicy:      "Never",
			Volumes:            pipeline.Spec.Volumes,
			Containers: []corev1.Container{
				{
					Name:            "executor",
					Image:           image,
					ImagePullPolicy: corev1.PullPolicy(imagePullPolicy),
					Command:         []string{"kk"},
					Args: []string{"pipeline",
						"-v", "6",
						"--name", pipeline.Name,
						"--namespace", pipeline.Namespace},
					VolumeMounts: pipeline.Spec.VolumeMounts,
				},
			},
		},
	}
	if err := ctrl.SetControllerReference(pipeline, pod, r.Scheme); err != nil {
		klog.ErrorS(err, "set controller reference error", "pipeline", ctrlclient.ObjectKeyFromObject(pipeline))

		return ctrl.Result{}, err
	}

	return ctrl.Result{}, r.Client.Create(ctx, pod)
}
