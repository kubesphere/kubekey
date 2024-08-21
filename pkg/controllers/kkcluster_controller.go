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

	infrastructurev1alpha1 "github.com/kubesphere/kubekey/v4/pkg/apis/capkk/v1alpha1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/annotations"
	"sigs.k8s.io/cluster-api/util/predicates"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	ctrlcontroller "sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrlfinalizer "sigs.k8s.io/controller-runtime/pkg/finalizer"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/kubesphere/kubekey/v4/pkg/scope"
)

// KKClusterReconciler reconciles a KKCluster object
type KKClusterReconciler struct {
	*runtime.Scheme
	ctrlclient.Client
	record.EventRecorder

	ctrlfinalizer.Finalizers
	MaxConcurrentReconciles int
}

// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=*,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;clusters/status,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machines;machines/status,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=controlplane.cluster.x-k8s.io,resources=*,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machinedeployments;machinedeployments/status,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machinesets;machinesets/status,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=*,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups="",resources=secrets;events;configmaps,verbs=get;list;watch;create;patch

func (r *KKClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, retErr error) {
	// get kk-cluster
	kkCluster := &infrastructurev1alpha1.KKCluster{}
	err := r.Client.Get(ctx, req.NamespacedName, kkCluster)
	if err != nil {
		if apierrors.IsNotFound(err) {
			klog.V(5).InfoS("apiserver not found", "pipeline", ctrlclient.ObjectKeyFromObject(kkCluster))

			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	// Fetch the Cluster.
	cluster, err := util.GetOwnerCluster(ctx, r.Client, kkCluster.ObjectMeta)
	if err != nil {
		return reconcile.Result{}, err
	}
	if cluster == nil {
		klog.V(5).InfoS("cluster-controller has not yet set OwnerRef")

		return reconcile.Result{}, nil
	}

	klog.InfoS("", "cluster", cluster.Name)

	// Create the scope.
	clusterScope, err := scope.NewClusterScope(scope.ClusterScopeParams{
		Client:         r.Client,
		Cluster:        cluster,
		KKCluster:      kkCluster,
		ControllerName: "kk-cluster",
	})
	if err != nil {
		return reconcile.Result{}, errors.Errorf("failed to create scope: %+v", err)
	}

	// Always close the scope when exiting this function, so we can persist any KKCluster changes.
	defer func() {
		if err := clusterScope.Close(); err != nil && retErr == nil {
			klog.V(5).ErrorS(err, "failed to patch object")
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

func (r *KKClusterReconciler) reconcileNormal(ctx context.Context, clusterScope *scope.ClusterScope) (reconcile.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	log.V(4).Info("Reconcile KKCluster normal")

	kkCluster := clusterScope.KKCluster

	// If the KKCluster doesn't have our finalizer, add it.
	if controllerutil.AddFinalizer(kkCluster, infrastructurev1alpha1.ClusterFinalizer) {
		// Register the finalizer immediately to avoid orphaning KK resources on delete
		if err := clusterScope.PatchObject(); err != nil {
			return reconcile.Result{}, err
		}
	}

	switch clusterScope.KKCluster.Status.Phase {
	case "":
		// to pending
	case infrastructurev1alpha1.KKClusterPhasePending:
		// to running
	case infrastructurev1alpha1.KKClusterPhaseRunning:
		for _, condition := range clusterScope.KKCluster.Status.Conditions {
			switch condition.Type {
			case infrastructurev1alpha1.HostReadyCondition:
				// select/create group
				// 从 inventory hosts 内随机选取 node 加入到 `control plane`, `master`，检查 ssh 连通性
			case infrastructurev1alpha1.EtcdReadyCondition:
				// 安装 etcd
			case infrastructurev1alpha1.ClusterBinaryReadyCondition:
				// 安装 kubelet, kubeadm, docker, etc.
			case infrastructurev1alpha1.BootstrapReadyCondition:
				// kubeadm init, kubeadm join
			case infrastructurev1alpha1.ClusterReadyCondition:
				// kubectl get node
				// master -> configmap -> kubeconfig -> Client: get node
			}
		}

		fallthrough
	default:
	}

	// : Reconcile KKCluster, upgrade, ... By create pipeline
	// 使用 client 创建 CRD

	// pipeline := &kubekeyv1.Pipeline{}
	// err := r.Client.Create(ctx, pipeline)
	// if err != nil {
	// 	if errors.IsAlreadyExists(err) {
	// 		// CRD 已存在，可以选择更新或忽略
	// 		log.Info("CRD already exists", "name", pipeline.Name)
	// 	} else {
	// 		// 处理其他错误
	// 		return reconcile.Result{}, err
	// 	}
	// }

	// If ControlPlaneLoadBalancer is NULL? Maybe we need default configuration here
	kkCluster.Spec.ControlPlaneEndpoint = clusterv1.APIEndpoint{
		Host: clusterScope.ControlPlaneLoadBalancer().Host,
		Port: clusterScope.APIServerPort(),
	}

	kkCluster.Status.Ready = true

	return ctrl.Result{}, nil
}

func (r *KKClusterReconciler) reconcileDelete(ctx context.Context, clusterScope *scope.ClusterScope) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	log.V(4).Info("Reconcile KKCluster delete")

	// : IsPaused filtered
	if annotations.IsPaused(clusterScope.Cluster, clusterScope.KKCluster) {
		log.Info("KKCluster or linked Cluster is marked as paused. Won't reconcile")

		return reconcile.Result{}, nil
	}

	// : pipeline delete

	// Cluster is deleted so remove the finalizer.
	controllerutil.RemoveFinalizer(clusterScope.KKCluster, infrastructurev1alpha1.ClusterFinalizer)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KKClusterReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(ctrlcontroller.Options{
			MaxConcurrentReconciles: r.MaxConcurrentReconciles,
		}).
		WithEventFilter(predicates.ResourceIsNotExternallyManaged(ctrl.LoggerFrom(ctx))).
		For(&infrastructurev1alpha1.KKCluster{}).
		Complete(r)
}
