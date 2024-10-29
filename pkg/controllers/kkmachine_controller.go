/*
Copyright 2024 The KubeSphere Authors.

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
	"errors"
	"fmt"
	"reflect"
	"time"

	"sigs.k8s.io/cluster-api/util/conditions"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"k8s.io/klog/v2"

	infrav1beta1 "github.com/kubesphere/kubekey/v4/pkg/apis/capkk/v1beta1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/annotations"
	"sigs.k8s.io/cluster-api/util/predicates"
	ctrlcontroller "sigs.k8s.io/controller-runtime/pkg/controller"
	ctrlfinalizer "sigs.k8s.io/controller-runtime/pkg/finalizer"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/kubesphere/kubekey/v4/pkg/scope"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// KKMachineReconciler reconciles a KKMachine object
type KKMachineReconciler struct {
	ctrlclient.Client
	Scheme *runtime.Scheme
	record.EventRecorder

	ctrlfinalizer.Finalizers
	MaxConcurrentReconciles int
}

// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=*,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machines;machines/status,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=secrets;,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;update;patch

func (r *KKMachineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, retErr error) {
	// Fetch the KKMachine.
	kkMachine := &infrav1beta1.KKMachine{}
	err := r.Get(ctx, req.NamespacedName, kkMachine)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	// Fetch the Machine.
	machine, err := util.GetOwnerMachine(ctx, r.Client, kkMachine.ObjectMeta)
	if err != nil {
		return ctrl.Result{}, err
	}
	if machine == nil {
		klog.V(5).InfoS("Machine has not yet set OwnerRef",
			"ProviderID", kkMachine.Spec.ProviderID)

		return ctrl.Result{}, nil
	}

	klog.V(4).InfoS("Fetched machine", "machine", machine.Name)

	// Fetch the Cluster & KKCluster.
	cluster, err := util.GetClusterFromMetadata(ctx, r.Client, machine.ObjectMeta)
	if err != nil {
		klog.V(5).InfoS("Machine is missing cluster label or cluster does not exist")

		return ctrl.Result{}, nil
	}
	klog.V(4).InfoS("Fetched cluster", "cluster", cluster.Name)

	if annotations.IsPaused(cluster, kkMachine) {
		klog.V(5).InfoS("KKMachine or linked Cluster is marked as paused. Won't reconcile")

		return ctrl.Result{}, nil
	}

	kkCluster := &infrav1beta1.KKCluster{}
	kkClusterName := ctrlclient.ObjectKey{
		Namespace: kkMachine.Namespace,
		Name:      cluster.Spec.InfrastructureRef.Name,
	}
	if err := r.Client.Get(ctx, kkClusterName, kkCluster); err != nil {
		klog.V(5).InfoS("KKCluster is not ready yet")

		return ctrl.Result{}, nil
	}
	klog.V(4).InfoS("Fetched kk-cluster", "kk-cluster", kkCluster.Name)

	// Create the cluster scope.
	clusterScope, err := scope.NewClusterScope(scope.ClusterScopeParams{
		Client:         r.Client,
		Cluster:        cluster,
		KKCluster:      kkCluster,
		ControllerName: "kk-cluster",
	})
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("failed to create scope: %w", err)
	}

	// Create the machine scope
	machineScope, err := scope.NewMachineScope(scope.MachineScopeParams{
		Client:       r.Client,
		ClusterScope: clusterScope,
		Machine:      machine,
		KKMachine:    kkMachine,
	})
	if err != nil {
		klog.V(5).ErrorS(err, "failed to create scope")

		return ctrl.Result{}, err
	}

	// Always close the scope when exiting this function, so we can persist any KKMachine changes.
	defer func() {
		if err := machineScope.Close(ctx); err != nil && retErr == nil {
			klog.V(5).ErrorS(err, "failed to patch object")
			retErr = err
		}
	}()

	// Handle Deleted machines
	if !kkMachine.ObjectMeta.DeletionTimestamp.IsZero() {
		r.reconcileDelete(kkMachine)

		return ctrl.Result{}, nil
	}

	// Handle normal machines
	return r.reconcileNormal(ctx, machineScope)
}

func (r *KKMachineReconciler) reconcileNormal(ctx context.Context, s *scope.MachineScope) (reconcile.Result, error) {
	klog.V(4).Info("Reconcile KKMachine normal")

	// If the KKMachine doesn't have our finalizer, add it.
	if controllerutil.AddFinalizer(s.KKMachine, infrav1beta1.MachineFinalizer) {
		// Register the finalizer immediately to avoid orphaning KK resources on delete
		if err := s.PatchObject(ctx); err != nil {
			return reconcile.Result{}, err
		}
	}

	if !s.ClusterScope.Cluster.Status.InfrastructureReady {
		return reconcile.Result{}, nil
	}

	if s.IsRole(infrav1beta1.ControlPlaneRole) {
		s.KKMachine.Labels[clusterv1.MachineControlPlaneLabel] = "true"
	}

	if !conditions.Has(s.ClusterScope.KKCluster, infrav1beta1.ClusterReadyCondition) {
		return reconcile.Result{}, nil
	}

	if err := refreshProviderID(ctx, r.Client, s); err != nil {
		return reconcile.Result{}, err
	}

	s.KKMachine.Status.Ready = true

	return ctrl.Result{
		RequeueAfter: 30 * time.Second,
	}, nil
}

func (r *KKMachineReconciler) reconcileDelete(kkm *infrav1beta1.KKMachine) {
	klog.V(4).Info("Reconcile KKMachine delete")

	// Machine is deleted so remove the finalizer.
	controllerutil.RemoveFinalizer(kkm, infrav1beta1.MachineFinalizer)
}

func refreshProviderID(ctx context.Context, client ctrlclient.Client, s *scope.MachineScope) error {
	inv, err := GetInventory(ctx, client, s.ClusterScope)
	if err != nil {
		return err
	}

	hostMachineMap := inv.Status.HostMachineMapping
	if hostMachineMap == nil {
		err := errors.New("failed to get host machine mapping")
		klog.V(5).ErrorS(err, "", "Inventory", inv)

		return err
	}

	// Remove ProviderID if it's not exist in map.
	if s.KKMachine.Spec.ProviderID != nil {
		exist := false
		for _, bindInfo := range hostMachineMap {
			if bindInfo.Machine == s.Name() {
				exist = true
				if !slicesEqualUnordered(bindInfo.Roles, s.KKMachine.Spec.Roles) {
					s.KKMachine.Spec.ProviderID = nil
				}

				break
			}
		}
		if !exist {
			s.KKMachine.Spec.ProviderID = nil
		}
	}

	// Add ProviderID if there is an unbound host.
	if s.KKMachine.Spec.ProviderID == nil {
		for hn, bindInfo := range hostMachineMap {
			if bindInfo.Machine != "" || !slicesEqualUnordered(bindInfo.Roles, s.KKMachine.Spec.Roles) {
				continue
			}

			// Bind ProviderID
			s.SetProviderID(hn)

			// Update Inventory
			expected := inv.DeepCopy()
			bindInfo.Machine = s.Name()
			inv.Status.HostMachineMapping[hn] = bindInfo
			if err := client.Status().Patch(ctx, inv, ctrlclient.MergeFrom(expected)); err != nil {
				return err
			}

			break
		}
	}

	return nil
}

func slicesEqualUnordered(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	countA := make(map[string]int)
	countB := make(map[string]int)

	for _, item := range a {
		countA[item]++
	}

	for _, item := range b {
		countB[item]++
	}

	if len(countA) != len(countB) {
		return false
	}

	for key, count := range countA {
		if countB[key] != count {
			return false
		}
	}

	return true
}

// SetupWithManager sets up the controller with the Manager.
func (r *KKMachineReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	// Avoid reconciling if the event triggering the reconciliation is related to incremental status updates
	// for KKMachine resources only
	kkMachineFilter := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldKKMachine, okOld := e.ObjectOld.(*infrav1beta1.KKMachine)
			newKKMachine, okNew := e.ObjectNew.(*infrav1beta1.KKMachine)

			if !okOld || !okNew {
				return false
			}

			oldCluster := oldKKMachine.DeepCopy()
			newCluster := newKKMachine.DeepCopy()

			oldCluster.Status = infrav1beta1.KKMachineStatus{}
			newCluster.Status = infrav1beta1.KKMachineStatus{}

			oldCluster.ObjectMeta.ResourceVersion = ""
			newCluster.ObjectMeta.ResourceVersion = ""

			return !reflect.DeepEqual(oldCluster, newCluster)
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(ctrlcontroller.Options{
			MaxConcurrentReconciles: r.MaxConcurrentReconciles,
		}).
		WithEventFilter(predicates.ResourceIsNotExternallyManaged(ctrl.LoggerFrom(ctx))).
		WithEventFilter(kkMachineFilter).
		For(&infrav1beta1.KKMachine{}).
		Complete(r)
}
