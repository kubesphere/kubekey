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

	"github.com/cockroachdb/errors"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
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
	// playbookPodLabel set in pod. value is which playbook belongs to.
	podPlaybookLabel     = "kubekey.kubesphere.io/playbook"
	defaultExecutorImage = "docker.io/kubesphere/executor:latest"
	executorContainer    = "executor"
)

// PlaybookReconciler reconcile playbook
type PlaybookReconciler struct {
	ctrlclient.Client
	record.EventRecorder

	MaxConcurrentReconciles int
}

var _ options.Controller = &PlaybookReconciler{}
var _ reconcile.Reconciler = &PlaybookReconciler{}

// Name implements controllers.controller.
func (r *PlaybookReconciler) Name() string {
	return "playbook-reconciler"
}

// SetupWithManager implements controllers.controller.
func (r *PlaybookReconciler) SetupWithManager(mgr manager.Manager, o options.ControllerManagerServerOptions) error {
	r.Client = mgr.GetClient()
	r.EventRecorder = mgr.GetEventRecorderFor(r.Name())

	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(ctrlcontroller.Options{
			MaxConcurrentReconciles: o.MaxConcurrentReconciles,
		}).
		For(&kkcorev1.Playbook{}).
		// Watches pod to sync playbook.
		Watches(&corev1.Pod{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj ctrlclient.Object) []reconcile.Request {
			playbook := &kkcorev1.Playbook{}
			if err := util.GetOwnerFromObject(ctx, r.Client, obj, playbook); err == nil {
				return []ctrl.Request{{NamespacedName: ctrlclient.ObjectKeyFromObject(playbook)}}
			}

			return nil
		})).
		Complete(r)
}

// +kubebuilder:rbac:groups=kubekey.kubesphere.io,resources=*,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods;events,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="coordination.k8s.io",resources=leases,verbs=get;list;watch;create;update;patch;delete

func (r PlaybookReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, retErr error) {
	// get playbook
	playbook := &kkcorev1.Playbook{}
	if err := r.Client.Get(ctx, req.NamespacedName, playbook); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, errors.Wrapf(err, "failed to get playbook %q", req.String())
	}

	helper, err := patch.NewHelper(playbook, r.Client)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}
	defer func() {
		if retErr != nil {
			if playbook.Status.FailureReason == "" {
				playbook.Status.FailureReason = kkcorev1.PlaybookFailedReasonUnknown
			}
			playbook.Status.FailureMessage = retErr.Error()
		}
		if err := r.reconcileStatus(ctx, playbook); err != nil {
			retErr = errors.Join(retErr, errors.WithStack(err))
		}
		if err := helper.Patch(ctx, playbook); err != nil {
			retErr = errors.Join(retErr, errors.WithStack(err))
		}
	}()

	// Add finalizer first if not set to avoid the race condition between init and delete.
	// Note: Finalizers in general can only be added when the deletionTimestamp is not set.
	if playbook.ObjectMeta.DeletionTimestamp.IsZero() && !controllerutil.ContainsFinalizer(playbook, kkcorev1.PlaybookCompletedFinalizer) {
		controllerutil.AddFinalizer(playbook, kkcorev1.PlaybookCompletedFinalizer)

		return ctrl.Result{}, nil
	}

	// Handle deleted clusters
	if !playbook.DeletionTimestamp.IsZero() {
		return reconcile.Result{}, errors.WithStack(r.reconcileDelete(ctx, playbook))
	}

	return ctrl.Result{}, errors.WithStack(r.reconcileNormal(ctx, playbook))
}

func (r PlaybookReconciler) reconcileStatus(ctx context.Context, playbook *kkcorev1.Playbook) error {
	// get pod from playbook
	podList := &corev1.PodList{}
	if err := util.GetObjectListFromOwner(ctx, r.Client, playbook, podList); err != nil {
		return errors.Wrapf(err, "failed to get pod list from playbook %q", ctrlclient.ObjectKeyFromObject(playbook))
	}
	// should only one pod for playbook
	if len(podList.Items) != 1 {
		return nil
	}

	if playbook.Status.Phase != kkcorev1.PlaybookPhaseFailed && playbook.Status.Phase != kkcorev1.PlaybookPhaseSucceeded {
		switch pod := podList.Items[0]; pod.Status.Phase {
		case corev1.PodFailed:
			playbook.Status.Phase = kkcorev1.PlaybookPhaseFailed
			playbook.Status.FailureReason = kkcorev1.PlaybookFailedReasonPodFailed
			playbook.Status.FailureMessage = pod.Status.Message
			for _, cs := range pod.Status.ContainerStatuses {
				if cs.Name == executorContainer && cs.State.Terminated != nil {
					playbook.Status.FailureMessage = cs.State.Terminated.Reason + ": " + cs.State.Terminated.Message
				}
			}
		case corev1.PodSucceeded:
			playbook.Status.Phase = kkcorev1.PlaybookPhaseSucceeded
		default:
			// the playbook status will set by pod
		}
	}

	return nil
}

func (r *PlaybookReconciler) reconcileDelete(ctx context.Context, playbook *kkcorev1.Playbook) error {
	podList := &corev1.PodList{}
	if err := util.GetObjectListFromOwner(ctx, r.Client, playbook, podList); err != nil {
		return errors.Wrapf(err, "failed to get pod list from playbook %q", ctrlclient.ObjectKeyFromObject(playbook))
	}
	if playbook.Status.Phase == kkcorev1.PlaybookPhaseFailed || playbook.Status.Phase == kkcorev1.PlaybookPhaseSucceeded {
		// playbook has completed. delete the owner pods.
		for _, obj := range podList.Items {
			if err := r.Client.Delete(ctx, &obj); err != nil {
				return errors.WithStack(err)
			}
		}
	}

	if len(podList.Items) == 0 {
		controllerutil.RemoveFinalizer(playbook, kkcorev1.PlaybookCompletedFinalizer)
	}

	return nil
}

func (r *PlaybookReconciler) reconcileNormal(ctx context.Context, playbook *kkcorev1.Playbook) error {
	switch playbook.Status.Phase {
	case "":
		playbook.Status.Phase = kkcorev1.PlaybookPhasePending
	case kkcorev1.PlaybookPhasePending:
		playbook.Status.Phase = kkcorev1.PlaybookPhaseRunning
	case kkcorev1.PlaybookPhaseRunning:
		return r.dealRunningPlaybook(ctx, playbook)
	}

	return nil
}

func (r *PlaybookReconciler) dealRunningPlaybook(ctx context.Context, playbook *kkcorev1.Playbook) error {
	// check if pod is exist
	podList := &corev1.PodList{}
	if err := r.Client.List(ctx, podList, ctrlclient.InNamespace(playbook.Namespace), ctrlclient.MatchingLabels{
		podPlaybookLabel: playbook.Name,
	}); err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrapf(err, "failed to list pod with label %s=%s", podPlaybookLabel, playbook.Name)
	} else if len(podList.Items) != 0 {
		// could find exist pod
		return nil
	}
	// create pod
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: playbook.Name + "-",
			Namespace:    playbook.Namespace,
			Labels: map[string]string{
				podPlaybookLabel: playbook.Name,
			},
		},
		Spec: corev1.PodSpec{
			ServiceAccountName: playbook.Spec.ServiceAccountName,
			RestartPolicy:      "Never",
			Volumes:            playbook.Spec.Volumes,
			Containers: []corev1.Container{
				{
					Name:    executorContainer,
					Image:   defaultExecutorImage,
					Command: []string{"kk"},
					Args: []string{"playbook",
						"--name", playbook.Name,
						"--namespace", playbook.Namespace},
					SecurityContext: &corev1.SecurityContext{
						RunAsUser:  ptr.To[int64](0),
						RunAsGroup: ptr.To[int64](0),
					},
					VolumeMounts: playbook.Spec.VolumeMounts,
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
	if err := ctrl.SetControllerReference(playbook, pod, r.Client.Scheme()); err != nil {
		return errors.Wrapf(err, "failed to set ownerReference to playbook pod %q", pod.GenerateName)
	}

	return errors.WithStack(r.Client.Create(ctx, pod))
}
