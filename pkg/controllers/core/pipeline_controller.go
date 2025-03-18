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
	"errors"
	"os"

	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	ctrlcontroller "sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/kubesphere/kubekey/v4/cmd/controller-manager/app/options"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/util"
)

const (
	// pipelinePodLabel set in pod. value is which pipeline belongs to.
	podPipelineLabel     = "kubekey.kubesphere.io/pipeline"
	defaultExecutorImage = "docker.io/kubesphere/executor:latest"
	executorContainer    = "executor"
)

// PipelineReconciler reconcile pipeline
type PipelineReconciler struct {
	ctrlclient.Client
	record.EventRecorder

	MaxConcurrentReconciles int
}

var _ options.Controller = &PipelineReconciler{}
var _ reconcile.Reconciler = &PipelineReconciler{}

// Name implements controllers.controller.
func (r *PipelineReconciler) Name() string {
	return "pipeline-reconciler"
}

// SetupWithManager implements controllers.controller.
func (r *PipelineReconciler) SetupWithManager(mgr manager.Manager, o options.ControllerManagerServerOptions) error {
	r.Client = mgr.GetClient()
	r.EventRecorder = mgr.GetEventRecorderFor(r.Name())

	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(ctrlcontroller.Options{
			MaxConcurrentReconciles: o.MaxConcurrentReconciles,
		}).
		For(&kkcorev1.Pipeline{}).
		// Watches pod to sync pipeline.
		Watches(&corev1.Pod{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj ctrlclient.Object) []reconcile.Request {
			pipeline := &kkcorev1.Pipeline{}
			if err := util.GetOwnerFromObject(ctx, r.Client, obj, pipeline); err == nil {
				return []ctrl.Request{{NamespacedName: ctrlclient.ObjectKeyFromObject(pipeline)}}
			}

			return nil
		})).
		Complete(r)
}

// +kubebuilder:rbac:groups=kubekey.kubesphere.io,resources=*,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods;events,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="coordination.k8s.io",resources=leases,verbs=get;list;watch;create;update;patch;delete

func (r PipelineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, retErr error) {
	// get pipeline
	pipeline := &kkcorev1.Pipeline{}
	err := r.Client.Get(ctx, req.NamespacedName, pipeline)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	helper, err := patch.NewHelper(pipeline, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}
	defer func() {
		if retErr != nil {
			if pipeline.Status.FailureReason == "" {
				pipeline.Status.FailureReason = kkcorev1.PipelineFailedReasonUnknown
			}
			pipeline.Status.FailureMessage = retErr.Error()
		}
		if err := r.reconcileStatus(ctx, pipeline); err != nil {
			retErr = errors.Join(retErr, err)
		}
		if err := helper.Patch(ctx, pipeline); err != nil {
			retErr = errors.Join(retErr, err)
		}
	}()

	// Add finalizer first if not set to avoid the race condition between init and delete.
	// Note: Finalizers in general can only be added when the deletionTimestamp is not set.
	if pipeline.ObjectMeta.DeletionTimestamp.IsZero() && !controllerutil.ContainsFinalizer(pipeline, kkcorev1.PipelineCompletedFinalizer) {
		controllerutil.AddFinalizer(pipeline, kkcorev1.PipelineCompletedFinalizer)

		return ctrl.Result{}, nil
	}

	// Handle deleted clusters
	if !pipeline.DeletionTimestamp.IsZero() {
		return reconcile.Result{}, r.reconcileDelete(ctx, pipeline)
	}

	return ctrl.Result{}, r.reconcileNormal(ctx, pipeline)
}

func (r PipelineReconciler) reconcileStatus(ctx context.Context, pipeline *kkcorev1.Pipeline) error {
	// get pod from pipeline
	podList := &corev1.PodList{}
	if err := util.GetObjectListFromOwner(ctx, r.Client, pipeline, podList); err != nil {
		return err
	}
	// should only one pod for pipeline
	if len(podList.Items) != 1 {
		return nil
	}

	if pipeline.Status.Phase != kkcorev1.PipelinePhaseFailed && pipeline.Status.Phase != kkcorev1.PipelinePhaseSucceeded {
		switch pod := podList.Items[0]; pod.Status.Phase {
		case corev1.PodFailed:
			pipeline.Status.Phase = kkcorev1.PipelinePhaseFailed
			pipeline.Status.FailureReason = kkcorev1.PipelineFailedReasonPodFailed
			pipeline.Status.FailureMessage = pod.Status.Message
			for _, cs := range pod.Status.ContainerStatuses {
				if cs.Name == executorContainer && cs.State.Terminated != nil {
					pipeline.Status.FailureMessage = cs.State.Terminated.Reason + ": " + cs.State.Terminated.Message
				}
			}
		case corev1.PodSucceeded:
			pipeline.Status.Phase = kkcorev1.PipelinePhaseSucceeded
		default:
			// the pipeline status will set by pod
		}
	}

	return nil
}

func (r *PipelineReconciler) reconcileDelete(ctx context.Context, pipeline *kkcorev1.Pipeline) error {
	podList := &corev1.PodList{}
	if err := util.GetObjectListFromOwner(ctx, r.Client, pipeline, podList); err != nil {
		return err
	}
	if pipeline.Status.Phase == kkcorev1.PipelinePhaseFailed || pipeline.Status.Phase == kkcorev1.PipelinePhaseSucceeded {
		// pipeline has completed. delete the owner pods.
		for _, obj := range podList.Items {
			if err := r.Client.Delete(ctx, &obj); err != nil {
				return err
			}
		}
	}

	if len(podList.Items) == 0 {
		controllerutil.RemoveFinalizer(pipeline, kkcorev1.PipelineCompletedFinalizer)
	}

	return nil
}

func (r *PipelineReconciler) reconcileNormal(ctx context.Context, pipeline *kkcorev1.Pipeline) error {
	switch pipeline.Status.Phase {
	case "":
		pipeline.Status.Phase = kkcorev1.PipelinePhasePending
	case kkcorev1.PipelinePhasePending:
		pipeline.Status.Phase = kkcorev1.PipelinePhaseRunning
	case kkcorev1.PipelinePhaseRunning:
		return r.dealRunningPipeline(ctx, pipeline)
	}

	return nil
}

func (r *PipelineReconciler) dealRunningPipeline(ctx context.Context, pipeline *kkcorev1.Pipeline) error {
	// check if pod is exist
	podList := &corev1.PodList{}
	if err := r.Client.List(ctx, podList, ctrlclient.InNamespace(pipeline.Namespace), ctrlclient.MatchingLabels{
		podPipelineLabel: pipeline.Name,
	}); err != nil && !apierrors.IsNotFound(err) {
		return err
	} else if len(podList.Items) != 0 {
		// could find exist pod
		return nil
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
					Name:    executorContainer,
					Image:   defaultExecutorImage,
					Command: []string{"kk"},
					Args: []string{"pipeline",
						"--name", pipeline.Name,
						"--namespace", pipeline.Namespace},
					SecurityContext: &corev1.SecurityContext{
						RunAsUser:  ptr.To[int64](0),
						RunAsGroup: ptr.To[int64](0),
					},
					VolumeMounts: pipeline.Spec.VolumeMounts,
				},
			},
		},
	}
	// get verbose from env
	if verbose := os.Getenv(_const.ENV_EXECUTOR_VERBOSE); verbose != "" {
		pod.Spec.Containers[0].Args = append(pod.Spec.Containers[0].Args, "-v", verbose)
	}
	// get image from env
	if image := os.Getenv(_const.ENV_EXECUTOR_IMAGE); image != "" {
		pod.Spec.Containers[0].Image = image
	}
	// get image from env
	if imagePullPolicy := os.Getenv(_const.ENV_EXECUTOR_IMAGE_PULLPOLICY); imagePullPolicy != "" {
		pod.Spec.Containers[0].ImagePullPolicy = corev1.PullPolicy(imagePullPolicy)
	}
	if err := ctrl.SetControllerReference(pipeline, pod, r.Client.Scheme()); err != nil {
		klog.ErrorS(err, "set controller reference error", "pipeline", ctrlclient.ObjectKeyFromObject(pipeline))

		return err
	}

	return r.Client.Create(ctx, pod)
}
