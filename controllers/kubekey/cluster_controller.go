/*
Copyright 2020 The KubeSphere Authors.

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

package kubekey

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	kubekeyv1alpha2 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha2"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=kubekey.kubesphere.io,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kubekey.kubesphere.io,resources=clusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=*,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=storage.k8s.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=*
// +kubebuilder:rbac:groups=installer.kubesphere.io,resources=clusterconfigurations,verbs=*
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=apiregistration.k8s.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=auditing.kubesphere.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=autoscaling,resources=*,verbs=*
// +kubebuilder:rbac:groups=certificates.k8s.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=config.istio.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=core.kubefed.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=devops.kubesphere.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=events.kubesphere.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=batch,resources=*,verbs=*
// +kubebuilder:rbac:groups=extensions,resources=*,verbs=*
// +kubebuilder:rbac:groups=iam.kubesphere.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=jaegertracing.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=logging.kubesphere.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=*,verbs=*
// +kubebuilder:rbac:groups=networking.istio.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=notification.kubesphere.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=policy,resources=*,verbs=*
// +kubebuilder:rbac:groups=storage.kubesphere.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=tenant.kubesphere.io,resources=*,verbs=*
// +kubebuilder:rbac:groups=cluster.kubesphere.io,resources=*,verbs=*

func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	return ctrl.Result{RequeueAfter: 2 * time.Minute}, nil
}

func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubekeyv1alpha2.Cluster{}).
		WithEventFilter(ignoreDeletionPredicate()).
		Complete(r)
}

func ignoreDeletionPredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore updates to CR status in which case metadata.Generation does not change
			return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			// Evaluates to false if the object has been confirmed deleted.
			return !e.DeleteStateUnknown
		},
	}
}
