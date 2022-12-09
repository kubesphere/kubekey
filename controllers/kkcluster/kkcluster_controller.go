/*
 Copyright 2022 The KubeSphere Authors.

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

package kkcluster

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/annotations"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/cluster-api/util/predicates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	infrav1 "github.com/kubesphere/kubekey/api/v1beta1"
	"github.com/kubesphere/kubekey/pkg/scope"
	"github.com/kubesphere/kubekey/util/collections"
)

const (
	// upgradeCheckFailedRequeueAfter is how long to wait before requeuing a cluster for which the upgrade check failed.
	upgradeCheckFailedRequeueAfter = 30 * time.Second
)

// Reconciler reconciles a KKCluster object
type Reconciler struct {
	client.Client
	Recorder         record.EventRecorder
	Scheme           *runtime.Scheme
	WatchFilterValue string
	DataDir          string
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	log := ctrl.LoggerFrom(ctx)
	c, err := ctrl.NewControllerManagedBy(mgr).
		WithOptions(options).
		For(&infrav1.KKCluster{}).
		WithEventFilter(predicates.ResourceHasFilterLabel(log, r.WatchFilterValue)).
		WithEventFilter(
			predicate.Funcs{
				// Avoid reconciling if the event triggering the reconciliation is related to incremental status updates
				// for KKCluster resources only
				UpdateFunc: func(e event.UpdateEvent) bool {
					if _, ok := e.ObjectOld.(*infrav1.KKCluster); !ok {
						return true
					}

					oldCluster := e.ObjectOld.(*infrav1.KKCluster).DeepCopy()
					newCluster := e.ObjectNew.(*infrav1.KKCluster).DeepCopy()

					oldCluster.Status = infrav1.KKClusterStatus{}
					newCluster.Status = infrav1.KKClusterStatus{}

					oldCluster.ObjectMeta.ResourceVersion = ""
					newCluster.ObjectMeta.ResourceVersion = ""

					return !cmp.Equal(oldCluster, newCluster)
				},
			},
		).
		WithEventFilter(predicates.ResourceIsNotExternallyManaged(log)).
		Build(r)
	if err != nil {
		return errors.Wrap(err, "error creating controller")
	}

	return c.Watch(
		&source.Kind{Type: &clusterv1.Cluster{}},
		handler.EnqueueRequestsFromMapFunc(r.requeueKKClusterForUnpausedCluster(ctx, log)),
		predicates.ClusterUnpaused(log),
	)
}

// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=*,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;clusters/status,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machines;machines/status,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=controlplane.cluster.x-k8s.io,resources=*,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machinedeployments;machinedeployments/status,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machinesets;machinesets/status,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=*,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups="",resources=secrets;events;configmaps,verbs=get;list;watch;create;patch

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, retErr error) {
	log := ctrl.LoggerFrom(ctx)

	kkCluster := &infrav1.KKCluster{}
	err := r.Get(ctx, req.NamespacedName, kkCluster)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// Fetch the Cluster.
	cluster, err := util.GetOwnerCluster(ctx, r.Client, kkCluster.ObjectMeta)
	if err != nil {
		return reconcile.Result{}, err
	}

	if cluster == nil {
		log.Info("Cluster Controller has not yet set OwnerRef")
		return reconcile.Result{}, nil
	}

	log = log.WithValues("cluster", cluster.Name)
	helper, err := patch.NewHelper(kkCluster, r.Client)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to init patch helper")
	}

	defer func() {
		e := helper.Patch(
			context.TODO(),
			kkCluster,
			patch.WithOwnedConditions{Conditions: []clusterv1.ConditionType{
				infrav1.PrincipalPreparedCondition,
			}})
		if e != nil {
			fmt.Println(e.Error())
		}
	}()

	// Create the scope.
	clusterScope, err := scope.NewClusterScope(scope.ClusterScopeParams{
		Client:         r.Client,
		Logger:         &log,
		Cluster:        cluster,
		KKCluster:      kkCluster,
		ControllerName: "kkcluster",
		RootFsBasePath: r.DataDir,
	})
	if err != nil {
		return reconcile.Result{}, errors.Errorf("failed to create scope: %+v", err)
	}

	// Always close the scope when exiting this function, so we can persist any KKCluster changes.
	defer func() {
		if err := clusterScope.Close(); err != nil && retErr == nil {
			log.Error(err, "failed to patch object")
			retErr = err
		}
	}()

	// Handle deleted clusters
	if !kkCluster.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, clusterScope)
	}

	// Handle non-deleted clusters
	return r.reconcileNormal(ctx, clusterScope)
}

func (r *Reconciler) reconcileDelete(ctx context.Context, clusterScope *scope.ClusterScope) (ctrl.Result, error) { //nolint:unparam
	log := ctrl.LoggerFrom(ctx)
	log.V(4).Info("Reconcile KKCluster delete")

	if annotations.IsPaused(clusterScope.Cluster, clusterScope.KKCluster) {
		log.Info("KKCluster or linked Cluster is marked as paused. Won't reconcile")
		return reconcile.Result{}, nil
	}

	// Cluster is deleted so remove the finalizer.
	controllerutil.RemoveFinalizer(clusterScope.KKCluster, infrav1.ClusterFinalizer)
	return ctrl.Result{}, nil
}

func (r *Reconciler) reconcileNormal(ctx context.Context, clusterScope *scope.ClusterScope) (reconcile.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	log.V(4).Info("Reconcile KKCluster normal")

	kkCluster := clusterScope.KKCluster

	// If the KKCluster doesn't have our finalizer, add it.
	controllerutil.AddFinalizer(kkCluster, infrav1.ClusterFinalizer)
	// Register the finalizer immediately to avoid orphaning KK resources on delete
	if err := clusterScope.PatchObject(); err != nil {
		return reconcile.Result{}, err
	}

	if _, err := net.LookupIP(kkCluster.Spec.ControlPlaneLoadBalancer.Host); err != nil {
		conditions.MarkFalse(kkCluster, infrav1.ExternalLoadBalancerReadyCondition, infrav1.WaitForDNSNameResolveReason, clusterv1.ConditionSeverityInfo, "")
		clusterScope.Info("Waiting on API server DNS name to resolve")
		return reconcile.Result{RequeueAfter: 15 * time.Second}, nil //nolint:nilerr
	}
	conditions.MarkTrue(kkCluster, infrav1.ExternalLoadBalancerReadyCondition)

	kkCluster.Spec.ControlPlaneEndpoint = clusterv1.APIEndpoint{
		Host: clusterScope.ControlPlaneLoadBalancer().Host,
		Port: clusterScope.APIServerPort(),
	}

	kkCluster.Status.Ready = true

	if res, err := r.reconcileInPlaceUpgrade(ctx, clusterScope); !res.IsZero() || err != nil {
		return res, err
	}

	if res, err := r.reconcilePatchAnnotations(ctx, clusterScope); !res.IsZero() || err != nil {
		return res, err
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) reconcileInPlaceUpgrade(ctx context.Context, clusterScope *scope.ClusterScope) (ctrl.Result, error) {
	kkCluster := clusterScope.KKCluster
	if _, ok := kkCluster.GetAnnotations()[infrav1.InPlaceUpgradeVersionAnnotation]; !ok {
		return ctrl.Result{}, nil
	}

	clusterScope.Info("Reconcile KKCluster in-place upgrade")

	if !kkCluster.Status.Ready {
		return ctrl.Result{}, nil
	}

	cluster := clusterScope.Cluster
	phases := []func(context.Context, *scope.ClusterScope) (ctrl.Result, error){
		// pause the cluster
		func(ctx context.Context, clusterScope *scope.ClusterScope) (ctrl.Result, error) {
			return r.reconcilePausedCluster(ctx, clusterScope, true)
		},
		// set up the KKInstance annotations
		func(ctx context.Context, clusterScope *scope.ClusterScope) (ctrl.Result, error) {
			return r.reconcileKKInstanceInPlaceUpgrade(ctx, clusterScope, collections.ActiveKKInstances,
				collections.OwnedKKInstances(kkCluster),
				collections.ControlPlaneKKInstances(cluster.Name))
		},
		func(ctx context.Context, clusterScope *scope.ClusterScope) (ctrl.Result, error) {
			return r.reconcileKKInstanceInPlaceUpgrade(ctx, clusterScope, collections.ActiveKKInstances,
				collections.OwnedKKInstances(kkCluster),
				collections.Not(collections.ControlPlaneKKInstances(cluster.Name)))
		},
		r.reconcileKKInstanceUpgradeCheck,
		// patch the cluster-api resource .spec.version
		r.reconcilePatchResourceSpecVersion,
		// if upgrade is done, unpause the cluster
		func(ctx context.Context, clusterScope *scope.ClusterScope) (ctrl.Result, error) {
			return r.reconcilePausedCluster(ctx, clusterScope, false)
		},
	}

	for _, phase := range phases {
		// Call the inner reconciliation methods.
		if phaseResult, err := phase(ctx, clusterScope); !phaseResult.IsZero() || err != nil {
			return phaseResult, err
		}
	}
	return ctrl.Result{}, nil
}

func (r *Reconciler) requeueKKClusterForUnpausedCluster(ctx context.Context, log logr.Logger) handler.MapFunc {
	return func(o client.Object) []ctrl.Request {
		c, ok := o.(*clusterv1.Cluster)
		if !ok {
			panic(fmt.Sprintf("Expected a Cluster but got a %T", o))
		}

		log := log.WithValues("objectMapper", "clusterToKKCluster", "namespace", c.Namespace, "cluster", c.Name)

		// Don't handle deleted clusters
		if !c.ObjectMeta.DeletionTimestamp.IsZero() {
			log.V(4).Info("Cluster has a deletion timestamp, skipping mapping.")
			return nil
		}

		// Make sure the ref is set
		if c.Spec.InfrastructureRef == nil {
			log.V(4).Info("Cluster does not have an InfrastructureRef, skipping mapping.")
			return nil
		}

		if c.Spec.InfrastructureRef.GroupVersionKind().Kind != "KKCluster" {
			log.V(4).Info("Cluster has an InfrastructureRef for a different type, skipping mapping.")
			return nil
		}

		kkCluster := &infrav1.KKCluster{}
		key := types.NamespacedName{Namespace: c.Spec.InfrastructureRef.Namespace, Name: c.Spec.InfrastructureRef.Name}

		if err := r.Get(ctx, key, kkCluster); err != nil {
			log.V(4).Error(err, "Failed to get KubeKey cluster")
			return nil
		}

		if annotations.IsExternallyManaged(kkCluster) {
			log.V(4).Info("KKCluster is externally managed, skipping mapping.")
			return nil
		}

		log.V(4).Info("Adding request.", "kkCluster", c.Spec.InfrastructureRef.Name)
		return []ctrl.Request{
			{
				NamespacedName: client.ObjectKey{Namespace: c.Namespace, Name: c.Spec.InfrastructureRef.Name},
			},
		}
	}
}
