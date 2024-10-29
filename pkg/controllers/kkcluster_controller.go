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
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"k8s.io/utils/ptr"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	kcv1beta1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	kcpv1beta1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/annotations"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/cluster-api/util/predicates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	ctrlcontroller "sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	ctrlfinalizer "sigs.k8s.io/controller-runtime/pkg/finalizer"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	infrav1beta1 "github.com/kubesphere/kubekey/v4/pkg/apis/capkk/v1beta1"
	kkcorev1 "github.com/kubesphere/kubekey/v4/pkg/apis/core/v1"
	"github.com/kubesphere/kubekey/v4/pkg/scope"
)

// This const defines some useful static strings.
const (
	TrueString               string = "true"
	MaxPipelineCounts        int    = 3
	PipelineUpperLimitReason string = "PipelineUpperLimit"

	pipelineNameLabel = "kubekey.kubesphere.capkk.io/pipeline"

	CheckConnectPlaybookName string = "check-connect"
	CheckConnectPlaybook     string = "capkk/playbooks/capkk_check_connect.yaml"

	PreparationPlaybookName string = "preparation"
	PreparationPlaybook     string = "capkk/playbooks/capkk_preparation.yaml"

	EtcdInstallPlaybookName string = "etcd-install"
	EtcdInstallPlaybook     string = "capkk/playbooks/capkk_etcd_binary_install.yaml"

	BinaryInstallPlaybookName string = "binary-install"
	BinaryInstallPlaybook     string = "capkk/playbooks/capkk_binary_install.yaml"

	BootstrapPlaybookName string = "bootstrap-ready"
	BootstrapPlaybook     string = "capkk/playbooks/capkk_bootstrap_ready.yaml"

	ClusterDeletingPlaybookName string = "delete-cluster"
	ClusterDeletingPlaybook     string = "capkk/playbooks/capkk_delete_cluster.yaml"

	CloudConfigValueKey string = "value"

	KubernetesDir string = "/etc/kubernetes/pki/"
	// KCPCertificateAuthoritySecretInfix string = "ca"
	// KCPCertificateAuthorityMountPath   string = "/etc/kubernetes/pki/ca"
	KCPKubeadmConfigSecretInfix string = "control-plane"
	KCPKubeadmConfigMountPath   string = "/etc/kubernetes/pki/kubeadmconfig"
	// KCPEtcdSecretInfix                 string = "etcd"
	// KCPEtcdMountPath                   string = "/etc/kubernetes/pki/etcd"
	KCPKubeConfigSecretInfix string = "kubeconfig"
	KCPKubeConfigMountPath   string = "/etc/kubernetes/pki/kubeconfig"
	// KCPProxySecretInfix                string = "proxy"
	// KCPProxyMountPath                  string = "/etc/kubernetes/pki/proxy"
	// KCPServiceAccountInfix             string = "sa"
	// KCPServiceAccountMountPath         string = "/etc/kubernetes/pki/sa"
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
	// Get KKCluster.
	kkCluster := &infrav1beta1.KKCluster{}
	err := r.Client.Get(ctx, req.NamespacedName, kkCluster)
	if err != nil {
		if apierrors.IsNotFound(err) {
			klog.V(5).InfoS("`api-server` not found",
				"KKCluster", ctrlclient.ObjectKeyFromObject(kkCluster))

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
		klog.V(5).InfoS("Cluster has not yet set OwnerRef")

		return reconcile.Result{}, nil
	}

	klog.V(4).InfoS("Fetched cluster", "cluster", cluster.Name)

	// Create the scope.
	clusterScope, err := scope.NewClusterScope(scope.ClusterScopeParams{
		Client:         r.Client,
		Cluster:        cluster,
		KKCluster:      kkCluster,
		ControllerName: cluster.Name,
	})
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("[%s]: failed to create scope: %w", cluster.Name, err)
	}

	// Always close the scope when exiting this function, so we can persist any KKCluster changes.
	defer func() {
		if err := clusterScope.Close(ctx); err != nil && retErr == nil {
			klog.V(5).ErrorS(err, "Failed to patch object")
			retErr = err
		}
	}()

	// : IsPaused filtered
	if annotations.IsPaused(clusterScope.Cluster, clusterScope.KKCluster) {
		klog.InfoS("KKCluster or linked Cluster is marked as paused. Won't reconcile")

		return reconcile.Result{}, nil
	}

	// Handle deleted clusters
	if !kkCluster.DeletionTimestamp.IsZero() {
		err := r.reconcileDelete(ctx, clusterScope)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	// Handle non-deleted clusters
	return r.reconcileNormal(ctx, clusterScope)
}

func (r *KKClusterReconciler) reconcileNormal(ctx context.Context, s *scope.ClusterScope) (reconcile.Result, error) {
	klog.V(4).Info("Reconcile KKCluster normal")

	// If the KKCluster doesn't have our finalizer, add it.
	if controllerutil.AddFinalizer(s.KKCluster, infrav1beta1.ClusterFinalizer) {
		// Register the finalizer immediately to avoid orphaning KK resources on delete
		if err := s.PatchObject(ctx); err != nil {
			return reconcile.Result{}, err
		}
	}

	//nolint:exhaustive
	switch s.KKCluster.Status.Phase {
	case "":
		// Switch kkCluster.Status.Phase to `Pending`
		err := s.PatchClusterPhase(ctx, infrav1beta1.KKClusterPhasePending)
		if err != nil {
			return reconcile.Result{}, err
		}
	case infrav1beta1.KKClusterPhasePending:
		err := s.PatchClusterWithFunc(ctx, func(kkc *infrav1beta1.KKCluster) {
			// Switch kkCluster.Status.Phase to `Pending`
			kkc.Status.Phase = infrav1beta1.KKClusterPhaseRunning
			// Set series of conditions as `Unknown` for the next reconciles.
			conditions.MarkUnknown(s.KKCluster, infrav1beta1.HostsReadyCondition,
				infrav1beta1.WaitingCheckHostReadyReason, infrav1beta1.WaitingCheckHostReadyMessage)
		})
		if err != nil {
			return reconcile.Result{}, err
		}
	case infrav1beta1.KKClusterPhaseRunning:
		if err := r.reconcileNormalRunning(ctx, s); err != nil {
			return ctrl.Result{}, err
		}
	case infrav1beta1.KKClusterPhaseSucceed:
		return ctrl.Result{}, nil
	case infrav1beta1.KKClusterPhaseFailed:
		return ctrl.Result{}, nil
	default:
		return ctrl.Result{}, nil
	}

	if lb := s.ControlPlaneLoadBalancer(); lb != nil {
		s.KKCluster.Spec.ControlPlaneEndpoint = clusterv1.APIEndpoint{
			Host: lb.Host,
			Port: s.APIServerPort(),
		}
	}

	// Initialize node select mode
	if !s.KKCluster.Status.Ready {
		if err := r.initKKCluster(ctx, s); err != nil {
			return ctrl.Result{}, err
		}
	}

	s.KKCluster.Status.Ready = true

	return ctrl.Result{
		RequeueAfter: 30 * time.Second,
	}, nil
}

func (r *KKClusterReconciler) reconcileNormalRunning(ctx context.Context, s *scope.ClusterScope) error {
	for {
		conditionsIsChanged, err := r.dealWithKKClusterConditions(ctx, s)
		if err != nil {
			return err
		}
		if !conditionsIsChanged {
			break
		}
	}

	return nil
}

//nolint:gocognit,cyclop
func (r *KKClusterReconciler) dealWithKKClusterConditions(ctx context.Context, s *scope.ClusterScope) (bool, error) {
	for _, condition := range s.KKCluster.Status.Conditions {
		conditionsCnt := len(s.KKCluster.Status.Conditions)
		if conditions.IsFalse(s.KKCluster, condition.Type) {
			continue
		}

		//nolint:exhaustive
		switch condition.Type {
		case infrav1beta1.HostsReadyCondition:
			if err := r.dealWithHostConnectCheck(ctx, s); err != nil {
				return false, err
			}
		case infrav1beta1.PreparationReadyCondition:
			// Refresh KCP secrets if annotation is true.
			if val, ok := s.KKCluster.Annotations[infrav1beta1.KCPSecretsRefreshAnnotation]; ok && val == TrueString {
				if err := dealWithSecrets(ctx, r.Client, s); err != nil {
					return false, err
				}
			}
			if err := r.dealWithPreparation(ctx, s); err != nil {
				return false, err
			}
		case infrav1beta1.EtcdReadyCondition:
			if err := r.dealWithEtcdInstall(ctx, s); err != nil {
				return false, err
			}
		case infrav1beta1.BinaryInstallCondition:
			if err := r.dealWithBinaryInstall(ctx, s); err != nil {
				return false, err
			}
		case infrav1beta1.BootstrapReadyCondition:
			// kubeadm init, kubeadm join
			if err := r.dealWithBootstrapReady(ctx, s); err != nil {
				return false, err
			}
		case infrav1beta1.ClusterReadyCondition:
			// kubectl get node
			// master -> configmap -> kubeconfig -> Client: get node
			if err := r.dealWithClusterReadyCheck(ctx, s); err != nil {
				return false, err
			}
			// Switch `KKCluster.Phase` to `Succeed`
			s.KKCluster.Status.Phase = infrav1beta1.KKClusterPhaseSucceed
			if err := r.Client.Status().Update(ctx, s.KKCluster); err != nil {
				klog.V(5).ErrorS(err, "Update KKCluster error", "KKCluster",
					ctrlclient.ObjectKeyFromObject(s.KKCluster))

				return false, err
			}
		default:
		}

		// If add new conditions, restart loop.
		if len(s.KKCluster.Status.Conditions) > conditionsCnt {
			return true, nil
		}
	}

	return false, nil
}

func (r *KKClusterReconciler) reconcileDelete(ctx context.Context, s *scope.ClusterScope) error {
	klog.V(4).Info("Reconcile KKCluster delete")

	// : pipeline delete
	switch s.KKCluster.Status.Phase {
	case infrav1beta1.KKClusterPhasePending:
		// Switch kkCluster.Status.Phase to `Deleting`
		err := s.PatchClusterPhase(ctx, infrav1beta1.KKClusterPhaseDeleting)
		if err != nil {
			return err
		}
	case infrav1beta1.KKClusterPhaseRunning:
		// delete running pipeline
		if err := r.dealWithDeletePipelines(ctx, s); err != nil {
			return err
		}

		err := s.PatchClusterPhase(ctx, infrav1beta1.KKClusterPhaseDeleting)
		if err != nil {
			return err
		}
	case infrav1beta1.KKClusterPhaseFailed:
		// Switch kkCluster.Status.Phase to `Deleting`
		err := s.PatchClusterPhase(ctx, infrav1beta1.KKClusterPhaseDeleting)
		if err != nil {
			return err
		}
	case infrav1beta1.KKClusterPhaseSucceed:
		// Switch kkCluster.Status.Phase to `Deleting`
		err := s.PatchClusterPhase(ctx, infrav1beta1.KKClusterPhaseDeleting)
		if err != nil {
			return err
		}
	case infrav1beta1.KKClusterPhaseDeleting:
		if err := r.dealWithClusterDeleting(ctx, s); err != nil {
			return err
		}
	}

	// Cluster is deleted so remove the finalizer.
	if conditions.IsFalse(s.KKCluster, infrav1beta1.ClusterDeletingCondition) {
		controllerutil.RemoveFinalizer(s.KKCluster, infrav1beta1.ClusterFinalizer)
	}

	return nil
}

// dealWithHostConnectCheck and dealWithHostSelector function used to pre-check inventory configuration, especially
// hosts and groups. In CAPKK, we defined three default groups to describe a complete kubernetes cluster. Firstly,
// dealWithHostConnectCheck function will check hosts connectivity by one simple pipeline. Secondly, dealWithHostSelector
// function will automatically initialize `Groups` defined by `Inventory` from connected hosts.
// Note: The second step always be executed although all hosts are disconnected.
func (r *KKClusterReconciler) dealWithHostConnectCheck(ctx context.Context, s *scope.ClusterScope) error {
	var p *kkcorev1.Pipeline
	var err error
	if p, err = r.dealWithExecutePlaybookReconcile(
		ctx, s, CheckConnectPlaybook, CheckConnectPlaybookName,
		func(_ *kkcorev1.Pipeline) {
			conditions.MarkTrueWithNegativePolarity(s.KKCluster, infrav1beta1.HostsReadyCondition,
				infrav1beta1.WaitingHostsSelectReason, clusterv1.ConditionSeverityInfo, infrav1beta1.WaitingHostsSelectMessage)
		},
		func(p *kkcorev1.Pipeline) {
			r.EventRecorder.Eventf(s.KKCluster, corev1.EventTypeWarning, infrav1beta1.HostsNotReadyReason, p.Status.Reason)
			conditions.MarkTrueWithNegativePolarity(s.KKCluster, infrav1beta1.HostsReadyCondition,
				infrav1beta1.HostsNotReadyReason, clusterv1.ConditionSeverityError, p.Status.Reason,
			)
		}); err != nil {
		return err
	}

	if p.Status.Phase == kkcorev1.PipelinePhaseSucceed {
		return r.dealWithHostSelector(ctx, s, *p)
	}

	return nil
}

// dealWithHostSelector function will be executed by dealWithHostConnectCheck function, if relevant pipeline run complete.
func (r *KKClusterReconciler) dealWithHostSelector(ctx context.Context, s *scope.ClusterScope, _ kkcorev1.Pipeline) error {
	// Fetch groups and hosts of `Inventory`, replicas of `KubeadmControlPlane` and `MachineDeployment`.
	inv, err := r.getInitialedInventory(ctx, s)
	if err != nil {
		return err
	}

	originalInventory := inv.DeepCopy()

	kcp, err := GetKubeadmControlPlane(ctx, r.Client, s)
	if err != nil {
		return err
	}

	md, err := GetMachineDeployment(ctx, r.Client, s)
	if err != nil {
		return err
	}

	// Initialize unavailable map to de-duplicate.
	unavailableHosts, unavailableGroups := make(map[string]struct{}), make(map[string]struct{})

	// Validate kubernetes cluster's controlPlaneGroup.
	controlPlaneGroup, err := validateInventoryGroup(s.KKCluster, inv,
		s.KKCluster.Annotations[infrav1beta1.ControlPlaneGroupNameAnnotation],
		int(*kcp.Spec.Replicas), unavailableHosts, unavailableGroups, false,
	)
	if err != nil {
		return err
	}

	inv.Spec.Groups[s.KKCluster.Annotations[infrav1beta1.ControlPlaneGroupNameAnnotation]] = controlPlaneGroup

	// Validate kubernetes cluster's workerGroup.
	workerGroup, err := validateInventoryGroup(s.KKCluster, inv,
		s.KKCluster.Annotations[infrav1beta1.WorkerGroupNameAnnotation],
		int(*md.Spec.Replicas), unavailableHosts, unavailableGroups, false,
	)
	if err != nil {
		return err
	}

	inv.Spec.Groups[s.KKCluster.Annotations[infrav1beta1.WorkerGroupNameAnnotation]] = workerGroup

	// Update `Inventory` resource.
	if err := r.Client.Patch(ctx, inv, ctrlclient.MergeFrom(originalInventory)); err != nil {
		klog.V(5).ErrorS(err, "Failed to patch Inventory", "Inventory", ctrlclient.ObjectKeyFromObject(inv))

		return err
	}

	// Update Conditions of `KKCluster`.
	conditions.MarkUnknown(s.KKCluster, infrav1beta1.PreparationReadyCondition,
		infrav1beta1.WaitingPreparationReason, infrav1beta1.WaitingPreparationMessage)
	if conditions.GetReason(s.KKCluster, infrav1beta1.HostsReadyCondition) == infrav1beta1.WaitingHostsSelectReason {
		conditions.MarkFalse(s.KKCluster, infrav1beta1.HostsReadyCondition, infrav1beta1.HostsReadyReason,
			clusterv1.ConditionSeverityInfo, infrav1beta1.HostsReadyMessage)
	} else {
		condition := conditions.Get(s.KKCluster, infrav1beta1.HostsReadyCondition)
		conditions.MarkTrueWithNegativePolarity(s.KKCluster, infrav1beta1.HostsReadyCondition, condition.Reason,
			clusterv1.ConditionSeverityWarning, condition.Message)
	}

	return nil
}

// dealWithPreparation function will pre-check & pre-install artifacts and initialize os system.
func (r *KKClusterReconciler) dealWithPreparation(ctx context.Context, s *scope.ClusterScope) error {
	if _, err := r.dealWithExecutePlaybookReconcile(
		ctx, s, PreparationPlaybook, PreparationPlaybookName,
		func(_ *kkcorev1.Pipeline) {
			conditions.MarkUnknown(s.KKCluster, infrav1beta1.EtcdReadyCondition,
				infrav1beta1.WaitingInstallEtcdReason, infrav1beta1.WaitingInstallEtcdMessage)
			conditions.MarkFalse(s.KKCluster, infrav1beta1.PreparationReadyCondition, infrav1beta1.PreparationReadyReason,
				clusterv1.ConditionSeverityInfo, infrav1beta1.PreparationReadyMessage)
		},
		func(p *kkcorev1.Pipeline) {
			r.EventRecorder.Eventf(s.KKCluster, corev1.EventTypeWarning, infrav1beta1.PreparationNotReadyReason, p.Status.Reason)
			conditions.MarkTrueWithNegativePolarity(s.KKCluster, infrav1beta1.PreparationReadyCondition,
				infrav1beta1.PreparationNotReadyReason, clusterv1.ConditionSeverityError, p.Status.Reason,
			)
		}); err != nil {
		return err
	}

	return nil
}

// dealWithEtcdInstall function will install binary Etcd.
func (r *KKClusterReconciler) dealWithEtcdInstall(ctx context.Context, s *scope.ClusterScope) error {
	if _, err := r.dealWithExecutePlaybookReconcile(
		ctx, s, EtcdInstallPlaybook, EtcdInstallPlaybookName,
		func(_ *kkcorev1.Pipeline) {
			conditions.MarkUnknown(s.KKCluster, infrav1beta1.BinaryInstallCondition,
				infrav1beta1.WaitingInstallClusterBinaryReason, infrav1beta1.WaitingInstallClusterBinaryMessage)
			conditions.MarkFalse(s.KKCluster, infrav1beta1.EtcdReadyCondition, infrav1beta1.EtcdReadyReason,
				clusterv1.ConditionSeverityInfo, infrav1beta1.EtcdReadyMessage)
		},
		func(p *kkcorev1.Pipeline) {
			r.EventRecorder.Eventf(s.KKCluster, corev1.EventTypeWarning, infrav1beta1.EtcdNotReadyReason, p.Status.Reason)
			conditions.MarkTrueWithNegativePolarity(s.KKCluster, infrav1beta1.EtcdReadyCondition,
				infrav1beta1.EtcdNotReadyReason, clusterv1.ConditionSeverityError, p.Status.Reason,
			)
		}); err != nil {
		return err
	}

	return nil
}

// dealWithBinaryInstall function will install cluster binary tools.
func (r *KKClusterReconciler) dealWithBinaryInstall(ctx context.Context, s *scope.ClusterScope) error {
	if _, err := r.dealWithExecutePlaybookReconcile(
		ctx, s, BinaryInstallPlaybook, BinaryInstallPlaybookName,
		func(_ *kkcorev1.Pipeline) {
			conditions.MarkUnknown(s.KKCluster, infrav1beta1.BootstrapReadyCondition,
				infrav1beta1.WaitingCheckBootstrapReadyReason, infrav1beta1.WaitingCheckBootstrapReadyMessage)
			conditions.MarkFalse(s.KKCluster, infrav1beta1.BinaryInstallCondition, infrav1beta1.BinaryReadyReason,
				clusterv1.ConditionSeverityInfo, infrav1beta1.BinaryReadyMessage)
		},
		func(p *kkcorev1.Pipeline) {
			r.EventRecorder.Eventf(s.KKCluster, corev1.EventTypeWarning, infrav1beta1.BinaryNotReadyReason, p.Status.Reason)
			conditions.MarkTrueWithNegativePolarity(s.KKCluster, infrav1beta1.BinaryInstallCondition,
				infrav1beta1.BinaryNotReadyReason, clusterv1.ConditionSeverityError, p.Status.Reason,
			)
		}); err != nil {
		return err
	}

	return nil
}

// dealWithBootstrapReady function will initialize cluster, and finally switch `KKCluster.Phase` to `Succeed`.
func (r *KKClusterReconciler) dealWithBootstrapReady(ctx context.Context, s *scope.ClusterScope) error {
	if _, err := r.dealWithExecutePlaybookReconcile(
		ctx, s, BootstrapPlaybook, BootstrapPlaybookName,
		func(_ *kkcorev1.Pipeline) {
			conditions.MarkUnknown(s.KKCluster, infrav1beta1.ClusterReadyCondition,
				infrav1beta1.WaitingCheckClusterReadyReason, infrav1beta1.WaitingCheckClusterReadyMessage)
			conditions.MarkFalse(s.KKCluster, infrav1beta1.BootstrapReadyCondition, infrav1beta1.BootstrapReadyReason,
				clusterv1.ConditionSeverityInfo, infrav1beta1.BootstrapReadyMessage)
		},
		func(p *kkcorev1.Pipeline) {
			r.EventRecorder.Eventf(s.KKCluster, corev1.EventTypeWarning, infrav1beta1.BootstrapNotReadyReason, p.Status.Reason)
			conditions.MarkTrueWithNegativePolarity(s.KKCluster, infrav1beta1.BootstrapReadyCondition,
				infrav1beta1.BootstrapNotReadyReason, clusterv1.ConditionSeverityError, p.Status.Reason,
			)
		}); err != nil {
		return err
	}

	return nil
}

// dealWithBootstrapReady function will initialize cluster, and finally switch `KKCluster.Phase` to `Succeed`.
func (r *KKClusterReconciler) dealWithClusterReadyCheck(ctx context.Context, s *scope.ClusterScope) error {
	inv, err := GetInventory(ctx, r.Client, s)
	if err != nil {
		return err
	}

	return r.updateInventoryStatus(ctx, s, inv)
}

// dealWithClusterDeleting function will delete the cluster.
func (r *KKClusterReconciler) dealWithClusterDeleting(ctx context.Context, s *scope.ClusterScope) error {
	var err error
	if _, err = r.dealWithExecutePlaybookReconcile(
		ctx, s, ClusterDeletingPlaybook, ClusterDeletingPlaybookName,
		func(_ *kkcorev1.Pipeline) {
			conditions.MarkFalse(s.KKCluster, infrav1beta1.ClusterDeletingCondition, infrav1beta1.ClusterDeletingSucceedReason,
				clusterv1.ConditionSeverityInfo, infrav1beta1.ClusterDeletingSucceedMessage)
		},
		func(p *kkcorev1.Pipeline) {
			r.EventRecorder.Eventf(s.KKCluster, corev1.EventTypeWarning, infrav1beta1.ClusterDeletingFailedReason, p.Status.Reason)
			conditions.MarkTrueWithNegativePolarity(s.KKCluster, infrav1beta1.ClusterDeletingCondition,
				infrav1beta1.ClusterDeletingFailedReason, clusterv1.ConditionSeverityError, p.Status.Reason,
			)
		}); err != nil {
		return err
	}

	return nil
}

// dealWithExecutePlaybookReconcile will judge the closest pipeline's `.Status.Phase` to the latest state of the cluster,
// and execute exactly stage to adjustment the cluster conditions. It will return one pipeline if it's useful, used for
// the other judgements.
func (r *KKClusterReconciler) dealWithExecutePlaybookReconcile(ctx context.Context, s *scope.ClusterScope,
	playbook, playbookName string, funcWithSucceed, funcWithFailed func(p *kkcorev1.Pipeline)) (*kkcorev1.Pipeline, error) {
	p, err := r.dealWithPipelinesReconcile(ctx, s, playbook, playbookName)
	if err != nil {
		return &kkcorev1.Pipeline{}, err
	}

	switch p.Status.Phase {
	case kkcorev1.PipelinePhasePending:
		return &kkcorev1.Pipeline{}, nil
	case kkcorev1.PipelinePhaseRunning:
		return &kkcorev1.Pipeline{}, nil
	case kkcorev1.PipelinePhaseSucceed:
		r.dealWithExecuteSucceed(p, funcWithSucceed)
		if err := r.Client.Status().Update(ctx, s.KKCluster); err != nil {
			return p, err
		}
	case kkcorev1.PipelinePhaseFailed:
		if err := r.dealWithExecuteFailed(p, funcWithFailed); err != nil {
			return p, err
		}
		if err := r.Client.Status().Update(ctx, s.KKCluster); err != nil {
			return p, err
		}
	default:
		return &kkcorev1.Pipeline{}, nil
	}

	return p, nil
}

// dealWithExecuteSucceed function used by dealWithExecutePlaybookReconcile, mark current condition as false and mark the
// next condition as true (if exist).
func (r *KKClusterReconciler) dealWithExecuteSucceed(p *kkcorev1.Pipeline, function func(p *kkcorev1.Pipeline)) {
	function(p)
	klog.V(5).InfoS("Pipeline execute succeed", "pipeline", p.Name)
}

// dealWithExecuteSucceed function used by dealWithExecutePlaybookReconcile, throw one warning event and mark the current
// condition as negative polarity.
func (r *KKClusterReconciler) dealWithExecuteFailed(p *kkcorev1.Pipeline, function func(p *kkcorev1.Pipeline)) error {
	function(p)
	err := fmt.Errorf("pipeline %s execute failed", p.Name)
	klog.V(5).ErrorS(err, "")

	return err
}

// dealWithSecrets function fetches secrets, and uses them to create a cluster.
func dealWithSecrets(ctx context.Context, client ctrlclient.Client, s *scope.ClusterScope) error {
	// Fetch all secrets.
	secrets := &corev1.SecretList{}
	if err := client.List(ctx, secrets, ctrlclient.MatchingLabels{
		clusterv1.ClusterNameLabel: s.Name(),
	}); err != nil {
		return err
	}

	// Deal with secrets
	for _, secret := range secrets.Items {
		if err := dealWithKCSecrets(ctx, client, s, &secret); err != nil {
			return err
		}
	}

	return s.PatchClusterWithFunc(ctx, func(kkc *infrav1beta1.KKCluster) {
		delete(kkc.Annotations, infrav1beta1.KCPSecretsRefreshAnnotation)
	})
}

// dealWithKCSecrets function fetches secrets created by KubeadmConfig.
func dealWithKCSecrets(ctx context.Context, client ctrlclient.Client, s *scope.ClusterScope, secret *corev1.Secret) error {
	// Fetch control plane's KubeadmConfig.
	var kcOwnRef metav1.OwnerReference
	if kc, err := GetControlPlaneKubeadmConfig(ctx, client, s); err != nil {
		return err
	} else if kc != nil {
		kcOwnRef = metav1.OwnerReference{
			APIVersion: kc.APIVersion,
			Kind:       kc.Kind,
			Name:       kc.Name,
			UID:        kc.UID,
			Controller: ptr.To(true),
		}
	}

	if !util.HasOwnerRef(secret.OwnerReferences, kcOwnRef) {
		return nil
	}

	// if secret format is cloud-config, parse and generate relevant secrets bind with `.Spec.PipelineTemplate`
	if strings.HasPrefix(secret.Name, s.KKCluster.Name+"-"+KCPKubeadmConfigSecretInfix) {
		return GenerateAndBindSecretsFromCloudConfig(ctx, client, s, secret, s.Name())
	}

	return nil
}

// mountSecretOnPipelineTemplate function handle one secret created by kcp, and bind with `PipelineTemplate` resource.
func mountSecretOnPipelineTemplate(ctx context.Context, client ctrlclient.Client, s *scope.ClusterScope,
	secret *corev1.Secret, mountPath string) error {
	// Define `Volume` and `VolumeMount`
	volume := corev1.Volume{
		Name: secret.Name + "-volume",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: secret.Name,
			},
		},
	}
	// Mount `Volume` on `VolumeMount`
	volumeMount := corev1.VolumeMount{
		Name:      secret.Name + "-volume",
		MountPath: mountPath,
	}

	// DeepCopy `KKCluster` for patch update.
	originalKKCluster := s.KKCluster.DeepCopy()

	// Fetch `.Spec.PipelineTemplate`.
	pipelineTemplate := &s.KKCluster.Spec.PipelineTemplate

	// Append or Update `Volume` for `.Spec.PipelineTemplate`.
	volumeExists := false
	for i, v := range pipelineTemplate.JobSpec.Volumes {
		if v.Name == volume.Name {
			pipelineTemplate.JobSpec.Volumes[i] = volume
			volumeExists = true

			break
		}
	}
	if !volumeExists {
		pipelineTemplate.JobSpec.Volumes = append(pipelineTemplate.JobSpec.Volumes, volume)
	}

	// Append or Update `VolumeMount` for `.Spec.PipelineTemplate`.
	volumeMountExists := false
	for i, vm := range pipelineTemplate.JobSpec.VolumeMounts {
		if vm.Name == volumeMount.Name {
			pipelineTemplate.JobSpec.VolumeMounts[i] = volumeMount
			volumeMountExists = true

			break
		}
	}
	if !volumeMountExists {
		pipelineTemplate.JobSpec.VolumeMounts = append(pipelineTemplate.JobSpec.VolumeMounts, volumeMount)
	}

	// Patch `KKCluster`.
	return client.Patch(ctx, s.KKCluster, ctrlclient.MergeFrom(originalKKCluster))
}

// dealWithPipelinesReconcile will reconcile all pipelines created for execute `playbookName` test1, and belong to current cluster.
// It will create one
func (r *KKClusterReconciler) dealWithPipelinesReconcile(ctx context.Context, s *scope.ClusterScope,
	playbook, playbookName string) (*kkcorev1.Pipeline, error) {
	pipelines := &kkcorev1.PipelineList{}

	// Check if pipeline existed, or an unexpected error happened.
	if err := r.Client.List(ctx, pipelines, ctrlclient.InNamespace(s.Namespace()), ctrlclient.MatchingLabels{
		clusterv1.ClusterNameLabel: s.Name(),
		pipelineNameLabel:          playbookName,
	}); err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}

	// Fetch the latest pipeline
	var latestPipeline *kkcorev1.Pipeline
	allPipelinesFailed := true

	for _, pipeline := range pipelines.Items {
		if pipeline.Status.Phase == kkcorev1.PipelinePhaseSucceed {
			return &pipeline, nil
		}
		if pipeline.Status.Phase != kkcorev1.PipelinePhaseFailed {
			allPipelinesFailed = false
		}
		if latestPipeline == nil || pipeline.CreationTimestamp.After(latestPipeline.CreationTimestamp.Time) {
			pipelineCopy := pipeline.DeepCopy()
			latestPipeline = pipelineCopy
		}
	}

	// If pipeline count less than upper limit and all pipeline are failed, create new one.
	if allPipelinesFailed && len(pipelines.Items) < MaxPipelineCounts {
		return r.generatePipelineByTemplate(ctx, s, playbookName, playbook)
	} else if len(pipelines.Items) >= MaxPipelineCounts {
		r.EventRecorder.Eventf(s.KKCluster, corev1.EventTypeWarning, PipelineUpperLimitReason,
			fmt.Sprintf("Can't create more %s pipeline, the upperlimit is %d", playbookName, MaxPipelineCounts))
	}

	return latestPipeline, nil
}

// dealWithDeletePipelines delete all existed pipeline created by cluster.
func (r *KKClusterReconciler) dealWithDeletePipelines(ctx context.Context, s *scope.ClusterScope) error {
	pipelines := &kkcorev1.PipelineList{}

	// Check if pipelines exist, or an unexpected error occurred.
	if err := r.Client.List(ctx, pipelines, ctrlclient.InNamespace(s.Namespace()), ctrlclient.MatchingLabels{
		clusterv1.ClusterNameLabel: s.Name(),
	}); err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	// Iterate through all pipelines and delete them.
	for _, pipeline := range pipelines.Items {
		if err := r.Client.Delete(ctx, &pipeline); err != nil && !apierrors.IsNotFound(err) {
			return err
		}
	}

	return nil
}

// initKKCluster function used to initialize some necessary configuration information if yaml file not config them.
func (r *KKClusterReconciler) initKKCluster(ctx context.Context, s *scope.ClusterScope) error {
	originalKKCluster := s.KKCluster.DeepCopy()

	if s.KKCluster.Spec.NodeSelectorMode == "" {
		s.KKCluster.Spec.NodeSelectorMode = infrav1beta1.DefaultNodeSelectorMode
	}

	// Fetch annotations of `KKCluster`.
	kkcAnnotations := s.KKCluster.Annotations
	if kkcAnnotations == nil {
		kkcAnnotations = make(map[string]string)
	}

	// Init annotations of `KKCluster`.
	if _, exists := kkcAnnotations[infrav1beta1.ControlPlaneGroupNameAnnotation]; !exists {
		kkcAnnotations[infrav1beta1.ControlPlaneGroupNameAnnotation] = infrav1beta1.DefaultControlPlaneGroupName
	}

	if _, exists := kkcAnnotations[infrav1beta1.WorkerGroupNameAnnotation]; !exists {
		kkcAnnotations[infrav1beta1.WorkerGroupNameAnnotation] = infrav1beta1.DefaultWorkerGroupName
	}

	if _, exists := kkcAnnotations[infrav1beta1.ClusterGroupNameAnnotation]; !exists {
		kkcAnnotations[infrav1beta1.ClusterGroupNameAnnotation] = infrav1beta1.DefaultClusterGroupName
	}

	if _, exists := kkcAnnotations[infrav1beta1.EtcdGroupNameAnnotation]; !exists {
		kkcAnnotations[infrav1beta1.EtcdGroupNameAnnotation] = infrav1beta1.DefaultEtcdGroupName
	}

	if _, exists := kkcAnnotations[infrav1beta1.RegistryGroupNameAnnotation]; !exists {
		kkcAnnotations[infrav1beta1.RegistryGroupNameAnnotation] = infrav1beta1.DefaultRegistryGroupName
	}

	if _, exists := kkcAnnotations[infrav1beta1.KCPSecretsRefreshAnnotation]; !exists {
		kkcAnnotations[infrav1beta1.KCPSecretsRefreshAnnotation] = "true"
	}

	// Patch annotations of KKCluster.
	s.KKCluster.Annotations = kkcAnnotations

	return r.Patch(ctx, s.KKCluster, ctrlclient.MergeFrom(originalKKCluster))
}

// getInitialedInventory function is a pre-processor function, used to process `Groups` of `Inventory`to streamline
// formal processing in `dealWithHostSelector` function.
func (r *KKClusterReconciler) getInitialedInventory(ctx context.Context, s *scope.ClusterScope) (
	*kkcorev1.Inventory, error) {
	inv, err := GetInventory(ctx, r.Client, s)
	if err != nil {
		return nil, err
	}

	originalInventory := inv.DeepCopy()

	hosts := inv.Spec.Hosts
	groups := inv.Spec.Groups
	if groups == nil {
		groups = make(map[string]kkcorev1.InventoryGroup)
	}

	// Assert hosts must be available (not empty).
	if len(hosts) == 0 {
		err := errors.New("unavailable hosts")
		klog.V(5).InfoS("Unavailable hosts, please check `Inventory` resource")

		return nil, err
	}

	// Initialize kubernetes necessary groups.
	groups[s.KKCluster.Annotations[infrav1beta1.ClusterGroupNameAnnotation]] = kkcorev1.InventoryGroup{
		Groups: []string{
			s.KKCluster.Annotations[infrav1beta1.ControlPlaneGroupNameAnnotation],
			s.KKCluster.Annotations[infrav1beta1.WorkerGroupNameAnnotation],
		},
	}
	if _, exists := groups[s.KKCluster.Annotations[infrav1beta1.ControlPlaneGroupNameAnnotation]]; !exists {
		groups[s.KKCluster.Annotations[infrav1beta1.ControlPlaneGroupNameAnnotation]] = kkcorev1.InventoryGroup{}
	}
	if _, exists := groups[s.KKCluster.Annotations[infrav1beta1.WorkerGroupNameAnnotation]]; !exists {
		groups[s.KKCluster.Annotations[infrav1beta1.WorkerGroupNameAnnotation]] = kkcorev1.InventoryGroup{}
	}
	inv.Spec.Groups = groups

	if err := controllerutil.SetControllerReference(s.KKCluster, inv, r.Scheme); err != nil {
		return nil, err
	}

	if err := r.Patch(ctx, inv, ctrlclient.MergeFrom(originalInventory)); err != nil {
		klog.ErrorS(err, "Failed to patch Inventory", "Inventory", inv)
	}

	return inv, nil
}

func (r *KKClusterReconciler) updateInventoryStatus(ctx context.Context, s *scope.ClusterScope, inv *kkcorev1.Inventory) error {
	// Get HostMachineMapping, and create a new one for update.
	hostMachineMapping := inv.Status.HostMachineMapping
	newHostMachineMapping := make(map[string]kkcorev1.MachineBinding)

	// Get ControlPlaneGroup and WorkerGroup.
	controlPlaneGroup := inv.Spec.Groups[s.KKCluster.Annotations[infrav1beta1.ControlPlaneGroupNameAnnotation]]
	workerGroup := inv.Spec.Groups[s.KKCluster.Annotations[infrav1beta1.WorkerGroupNameAnnotation]]

	// Update control-plane nodes.
	for _, h := range controlPlaneGroup.Hosts {
		if binding, exists := hostMachineMapping[h]; exists {
			newHostMachineMapping[h] = binding
		} else {
			newHostMachineMapping[h] = kkcorev1.MachineBinding{
				Machine: "",
				Roles:   []string{infrav1beta1.ControlPlaneRole},
			}
		}
	}

	// Update worker nodes.
	for _, h := range workerGroup.Hosts {
		if binding, exists := hostMachineMapping[h]; exists {
			newHostMachineMapping[h] = binding
		} else {
			newHostMachineMapping[h] = kkcorev1.MachineBinding{
				Machine: "",
				Roles:   []string{infrav1beta1.WorkerRole},
			}
		}
	}

	// Update HostMachineMapping.
	inv.Status.HostMachineMapping = newHostMachineMapping

	return r.Client.Status().Update(ctx, inv)
}

// validateInventoryGroup function validates an invalidated group defined in `Inventory` and returns.
// Param::ghosts is the regional group hosts list.
// Param::hosts is all usable hosts of the `Inventory`.
// Param::gName is the name of the `InventoryGroup`.
// Param::cnt is the target hosts count of the `InventoryGroup`.
// Param::unavailableHosts & Param::unavailableGroups are used to remove duplicates, ensuring non-repeatability of
// nodes and avoiding cyclic group dependencies.
// Param::isRepeatable defines if the group hosts are repeatable. If true, all hosts will not be added into
// unavailableHosts.
func validateInventoryGroup(
	kkc *infrav1beta1.KKCluster, inv *kkcorev1.Inventory, gName string, cnt int,
	unavailableHosts, unavailableGroups map[string]struct{}, isRepeatable bool,
) (kkcorev1.InventoryGroup, error) {
	// Get the hosts already assigned to this group
	ghosts := kkcorev1.GetHostsFromGroup(inv, gName, unavailableHosts, unavailableGroups)
	hosts := inv.Spec.Hosts

	// Check if we have fewer hosts than needed
	if len(ghosts) < cnt {
		var availableHosts []string
		for host := range hosts {
			if _, exists := unavailableHosts[host]; !exists {
				availableHosts = append(availableHosts, host)
			}
		}

		remainingHostsCount := cnt - len(ghosts)
		// If not enough hosts are available, return an error
		if len(availableHosts) < remainingHostsCount {
			conditions.MarkTrueWithNegativePolarity(kkc, infrav1beta1.HostsReadyCondition, infrav1beta1.HostsNotReadyReason,
				clusterv1.ConditionSeverityError, infrav1beta1.HostsSelectFailedMessage,
			)

			return kkcorev1.InventoryGroup{}, fmt.Errorf("not enough available hosts for group %s", gName)
		}

		// Select the remaining hosts based on the selector mode
		hs, err := groupHostsSelector(availableHosts, remainingHostsCount, kkc.Spec.NodeSelectorMode)
		if err != nil {
			return kkcorev1.InventoryGroup{}, err
		}

		// Append the selected hosts to ghosts
		ghosts = append(ghosts, hs...)
	} else {
		var err error
		ghosts, err = groupHostsSelector(ghosts, cnt, kkc.Spec.NodeSelectorMode)
		if err != nil {
			return kkcorev1.InventoryGroup{}, err
		}
	}

	if !isRepeatable {
		for _, host := range ghosts {
			if _, exists := unavailableHosts[host]; !exists {
				unavailableHosts[host] = struct{}{}
			}
		}
	}

	return kkcorev1.InventoryGroup{
		Groups: make([]string, 0),
		Hosts:  ghosts,
		Vars:   inv.Spec.Groups[gName].Vars,
	}, nil
}

// groupHostsSelector selects nodes based on the NodeSelectorMode
func groupHostsSelector(availableHosts []string, cnt int, nodeSelectMode infrav1beta1.NodeSelectorMode) ([]string, error) {
	if cnt >= len(availableHosts) {
		return availableHosts, nil
	}

	selectedHosts := append([]string(nil), availableHosts...)

	switch nodeSelectMode {
	case infrav1beta1.RandomNodeSelectorMode:
		shuffledHosts, err := secureShuffle(selectedHosts)
		if err != nil {
			return nil, err
		}

		return shuffledHosts[:cnt], nil

	case infrav1beta1.SequenceNodeSelectorMode:
		return selectedHosts[:cnt], nil
	}

	return selectedHosts[:cnt], nil
}

// Secure shuffle function using crypto/rand
func secureShuffle(hosts []string) ([]string, error) {
	shuffledHosts := append([]string(nil), hosts...)
	n := len(shuffledHosts)

	for i := n - 1; i > 0; i-- {
		j, err := secureRandomInt(i + 1)
		if err != nil {
			return nil, err
		}
		shuffledHosts[i], shuffledHosts[j] = shuffledHosts[j], shuffledHosts[i]
	}

	return shuffledHosts, nil
}

// Generate secure random integer using crypto/rand
func secureRandomInt(upperLimit int) (int, error) {
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(upperLimit)))
	if err != nil {
		return 0, err
	}

	return int(nBig.Int64()), nil
}

// generatePipelineByTemplate function can generate a generic pipeline by `PipelineTemplate`.
func (r *KKClusterReconciler) generatePipelineByTemplate(ctx context.Context, s *scope.ClusterScope, name string, playbook string,
) (*kkcorev1.Pipeline, error) {
	pipelineTemplate := s.KKCluster.Spec.PipelineTemplate

	pipeline := &kkcorev1.Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: name + "-",
			Namespace:    s.Namespace(),
			Labels: map[string]string{
				clusterv1.ClusterNameLabel: s.Name(),
				pipelineNameLabel:          name,
			},
		},
		Spec: kkcorev1.PipelineSpec{
			Project:      pipelineTemplate.Project,
			Playbook:     playbook,
			InventoryRef: pipelineTemplate.InventoryRef,
			ConfigRef:    pipelineTemplate.ConfigRef,
			Tags:         pipelineTemplate.Tags,
			SkipTags:     pipelineTemplate.SkipTags,
			Debug:        pipelineTemplate.Debug,
			JobSpec:      pipelineTemplate.JobSpec,
		},
	}

	// Check if `.Spec.InventoryRef` is nil
	if pipeline.Spec.InventoryRef == nil {
		err := errors.New("pipeline must be generated with `.Spec.inventoryRef`, but given nil")
		klog.V(5).ErrorS(err, "", "pipeline", pipeline.Name)

		return nil, err
	}

	// Create pipeline.
	if err := controllerutil.SetControllerReference(s.KKCluster, pipeline, r.Scheme); err != nil {
		return nil, err
	}

	if err := r.Client.Create(ctx, pipeline); err != nil {
		return nil, err
	}
	klog.V(5).InfoS("Create pipeline successfully", "pipeline", pipeline.Name)

	return pipeline, nil
}

// GetInventory function return cluster's `Inventory` resource.
func GetInventory(ctx context.Context, client ctrlclient.Client, s *scope.ClusterScope) (*kkcorev1.Inventory, error) {
	inventory := &kkcorev1.Inventory{}

	namespace := s.KKCluster.Spec.InventoryRef.Namespace
	if namespace == "" {
		namespace = s.Namespace()
	}

	err := client.Get(ctx,
		types.NamespacedName{
			Name:      s.KKCluster.Spec.InventoryRef.Name,
			Namespace: namespace,
		}, inventory)
	if err != nil {
		klog.V(5).InfoS("`kk-cluster` must set `InventoryRef`, but not found",
			"Inventory", ctrlclient.ObjectKeyFromObject(inventory))

		return nil, fmt.Errorf("[%s]: kk-cluster must set `InventoryRef`, but not found",
			s.Cluster.Name)
	}

	return inventory, nil
}

// GetKubeadmControlPlane function return cluster's `KubeadmControlPlane` resource.
func GetKubeadmControlPlane(ctx context.Context, client ctrlclient.Client, s *scope.ClusterScope) (*kcpv1beta1.KubeadmControlPlane, error) {
	kcp := &kcpv1beta1.KubeadmControlPlane{}

	namespace := s.Cluster.Spec.ControlPlaneRef.Namespace
	if namespace == "" {
		namespace = s.Namespace()
	}

	if err := client.Get(ctx,
		types.NamespacedName{
			Name:      s.Cluster.Spec.ControlPlaneRef.Name,
			Namespace: namespace,
		}, kcp,
	); err != nil {
		return nil, err
	}

	return kcp, nil
}

// GetControlPlaneKubeadmConfig function return cluster's `KubeadmConfig` resource belonged to control plane `Machine`.
func GetControlPlaneKubeadmConfig(ctx context.Context, client ctrlclient.Client, s *scope.ClusterScope) (*kcv1beta1.KubeadmConfig, error) {
	kcList := &kcv1beta1.KubeadmConfigList{}

	namespace := s.Cluster.Spec.ControlPlaneRef.Namespace
	if namespace == "" {
		namespace = s.Namespace()
	}

	err := client.List(ctx, kcList,
		ctrlclient.InNamespace(namespace),
		ctrlclient.MatchingLabels{
			clusterv1.ClusterNameLabel: s.Name(),
		}, ctrlclient.HasLabels{
			clusterv1.MachineControlPlaneNameLabel,
		})
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, fmt.Errorf("error listing MachineDeployments: %w", err)
	}

	if len(kcList.Items) == 0 {
		return nil, errors.New("no control plane's KubeadmConfig found for cluster " + s.Name())
	}

	if len(kcList.Items) > 1 {
		return nil, errors.New("multiple control plane's KubeadmConfig found for cluster " + s.Name())
	}

	return &kcList.Items[0], nil
}

// GetMachineDeployment function return cluster's `MachineDeployment` resource.
func GetMachineDeployment(ctx context.Context, client ctrlclient.Client, s *scope.ClusterScope) (*clusterv1.MachineDeployment, error) {
	mdList := &clusterv1.MachineDeploymentList{}

	namespace := s.KKCluster.Spec.InventoryRef.Namespace
	if namespace == "" {
		namespace = s.Namespace()
	}

	err := client.List(ctx, mdList, ctrlclient.InNamespace(namespace), ctrlclient.MatchingLabels{
		clusterv1.ClusterNameLabel: s.Name(),
	})
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, fmt.Errorf("error listing MachineDeployments: %w", err)
	}

	if len(mdList.Items) == 0 {
		return nil, errors.New("no MachineDeployment found for cluster " + s.Name())
	}

	if len(mdList.Items) > 1 {
		return nil, errors.New("multiple MachineDeployments found for cluster " + s.Name())
	}

	return &mdList.Items[0], nil
}

// GenerateAndBindSecretsFromCloudConfig fetches the cloud-config format secret, then parse it and returns secrets.
func GenerateAndBindSecretsFromCloudConfig(ctx context.Context, client ctrlclient.Client, s *scope.ClusterScope,
	secret *corev1.Secret, secretNamePrefix string) error {
	// WriteFile represents the structure of each write_files entry in cloud-init.
	type WriteFile struct {
		Path        string `yaml:"path"`
		Owner       string `yaml:"owner"`
		Permissions string `yaml:"permissions"`
		Content     string `yaml:"content"`
	}

	// CloudConfig represents the structure of the cloud-init data.
	type CloudConfig struct {
		WriteFiles []WriteFile `yaml:"write_files"`
	}

	// Step 1: Get the YAML data from the secret.
	data, exists := secret.Data[CloudConfigValueKey]
	if !exists {
		return fmt.Errorf("key %s not found in secret", CloudConfigValueKey)
	}

	// Step 2: Parse the cloud-init content into CloudConfig struct.
	var cloudConfig CloudConfig
	if err := yaml.Unmarshal(data, &cloudConfig); err != nil {
		return fmt.Errorf("failed to unmarshal cloud-init YAML: %w", err)
	}

	for _, file := range cloudConfig.WriteFiles {
		var secretName string
		if strings.HasPrefix(file.Path, KubernetesDir) {
			// Create a secret name based on the file path (replace slashes with hyphens).
			secretName = fmt.Sprintf("%s-%s", secretNamePrefix,
				strings.NewReplacer(".", "-", "/", "-").Replace(strings.TrimPrefix(file.Path, KubernetesDir)))
		} else if file.Path == "/run/kubeadm/kubeadm.yaml" {
			secretName = fmt.Sprintf("%s-%s", secretNamePrefix, "kubeadm-config")
		} else {
			continue
		}

		ownerRef := metav1.OwnerReference{
			APIVersion:         s.KKCluster.APIVersion,
			Kind:               s.KKCluster.Kind,
			Name:               s.KKCluster.Name,
			UID:                s.KKCluster.UID,
			Controller:         ptr.To(true),
			BlockOwnerDeletion: ptr.To(true),
		}

		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:            secretName,
				Namespace:       s.Namespace(),
				OwnerReferences: []metav1.OwnerReference{ownerRef},
			},
			Data: map[string][]byte{
				filepath.Base(file.Path): []byte(file.Content),
			},
			Type: corev1.SecretTypeBootstrapToken,
		}

		// Create the Secret in the cluster
		if err := client.Create(ctx, secret); err != nil {
			return err
		}

		if err := mountSecretOnPipelineTemplate(ctx, client, s, secret, file.Path); err != nil {
			return err
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KKClusterReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	pipelinePhaseFilter := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			newPipeline, ok := e.ObjectNew.(*kkcorev1.Pipeline)
			if !ok {
				return false
			}

			return newPipeline.Status.Phase == kkcorev1.PipelinePhaseSucceed || newPipeline.Status.Phase == kkcorev1.PipelinePhaseFailed
		},
		CreateFunc: func(e event.CreateEvent) bool {
			pipeline, ok := e.Object.(*kkcorev1.Pipeline)
			if !ok {
				return false
			}

			return pipeline.Status.Phase == kkcorev1.PipelinePhaseSucceed || pipeline.Status.Phase == kkcorev1.PipelinePhaseFailed
		},
	}

	// Avoid reconciling if the event triggering the reconciliation is related to incremental status updates
	// for KKCluster resources only
	kkClusterFilter := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldCluster, okOld := e.ObjectOld.(*infrav1beta1.KKCluster)
			newCluster, okNew := e.ObjectNew.(*infrav1beta1.KKCluster)

			if !okOld || !okNew {
				return true
			}

			oldClusterCopy := oldCluster.DeepCopy()
			newClusterCopy := newCluster.DeepCopy()

			oldClusterCopy.Status = infrav1beta1.KKClusterStatus{}
			newClusterCopy.Status = infrav1beta1.KKClusterStatus{}

			oldClusterCopy.ObjectMeta.ResourceVersion = ""
			newClusterCopy.ObjectMeta.ResourceVersion = ""

			return !reflect.DeepEqual(oldClusterCopy, newClusterCopy)
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(ctrlcontroller.Options{
			MaxConcurrentReconciles: r.MaxConcurrentReconciles,
		}).
		WithEventFilter(predicates.ResourceIsNotExternallyManaged(ctrl.LoggerFrom(ctx))).
		WithEventFilter(kkClusterFilter).
		For(&infrav1beta1.KKCluster{}).
		Owns(&kkcorev1.Pipeline{}, builder.WithPredicates(pipelinePhaseFilter)).
		Owns(&kkcorev1.Inventory{}).
		Complete(r)
}
