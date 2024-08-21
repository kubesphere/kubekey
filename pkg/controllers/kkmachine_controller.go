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
	"github.com/pkg/errors"

	"k8s.io/klog/v2"

	"github.com/kubesphere/kubekey/v4/pkg/apis/capkk/v1alpha1"

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
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// KKMachineReconciler reconciles a KKMachine object
type KKMachineReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	record.EventRecorder

	ctrlfinalizer.Finalizers
	MaxConcurrentReconciles int
}

// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=kkmachines;kkmachines/status;kkmachines/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machines;machines/status,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=secrets;,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;update;patch

func (r *KKMachineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, retErr error) {
	// Fetch the KKMachine.
	kkMachine := &v1alpha1.KKMachine{}
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
		klog.V(5).InfoS("Machine Controller has not yet set OwnerRef")

		return ctrl.Result{}, nil
	}

	klog.InfoS("machine", machine.Name)

	// Fetch the Cluster.
	cluster, err := util.GetClusterFromMetadata(ctx, r.Client, machine.ObjectMeta)
	if err != nil {
		klog.V(5).InfoS("Machine is missing cluster label or cluster does not exist")

		return ctrl.Result{}, nil
	}

	if annotations.IsPaused(cluster, kkMachine) {
		klog.V(5).InfoS("KKMachine or linked Cluster is marked as paused. Won't reconcile")

		return ctrl.Result{}, nil
	}

	klog.InfoS("", "cluster", cluster.Name)

	klog.InfoS("", "KKCluster", cluster.Spec.InfrastructureRef.Name)
	kkClusterName := client.ObjectKey{
		Namespace: kkMachine.Namespace,
		Name:      cluster.Spec.InfrastructureRef.Name,
	}

	kkCluster := &v1alpha1.KKCluster{}
	if err := r.Client.Get(ctx, kkClusterName, kkCluster); err != nil {
		klog.V(5).InfoS("KKCluster is not ready yet")

		return ctrl.Result{}, nil
	}

	// Create the cluster scope.
	clusterScope, err := scope.NewClusterScope(scope.ClusterScopeParams{
		Client:         r.Client,
		Cluster:        cluster,
		KKCluster:      kkCluster,
		ControllerName: "kk-cluster",
	})
	if err != nil {
		return reconcile.Result{}, errors.Errorf("failed to create scope: %+v", err)
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
		if err := machineScope.Close(); err != nil && retErr == nil {
			klog.V(5).ErrorS(err, "failed to patch object")
			retErr = err
		}
	}()

	// if !kkMachine.ObjectMeta.DeletionTimestamp.IsZero() {
	// 	return r.reconcileDelete(ctx, machineScope)
	// }
	//
	// return r.reconcileNormal(ctx, machineScope, infraCluster, infraCluster)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KKMachineReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(ctrlcontroller.Options{
			MaxConcurrentReconciles: r.MaxConcurrentReconciles,
		}).
		WithEventFilter(predicates.ResourceIsNotExternallyManaged(ctrl.LoggerFrom(ctx))).
		For(&v1alpha1.KKMachine{}).
		Complete(r)
}
