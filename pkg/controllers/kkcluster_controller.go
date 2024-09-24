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
	"reflect"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
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

// Defines some useful static strings.
const (
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
		r.reconcileDelete(clusterScope)

		return ctrl.Result{}, nil
	}

	// Handle non-deleted clusters
	return r.reconcileNormal(ctx, clusterScope)
}

func (r *KKClusterReconciler) reconcileNormal(ctx context.Context, s *scope.ClusterScope) (reconcile.Result, error) {
	klog.V(4).Info("Reconcile KKCluster normal")

	kkCluster := s.KKCluster

	// If the KKCluster doesn't have our finalizer, add it.
	if controllerutil.AddFinalizer(kkCluster, infrav1beta1.ClusterFinalizer) {
		// Register the finalizer immediately to avoid orphaning KK resources on delete
		if err := s.PatchObject(ctx); err != nil {
			return reconcile.Result{}, err
		}
	}

	switch kkCluster.Status.Phase {
	case "":
		// Switch kkCluster.Status.Phase to `Pending`
		excepted := kkCluster.DeepCopy()
		kkCluster.Status.Phase = infrav1beta1.KKClusterPhasePending
		if err := r.Client.Status().Patch(ctx, kkCluster, ctrlclient.MergeFrom(excepted)); err != nil {
			klog.V(5).ErrorS(err, "Update KKCluster error", "KKCluster", ctrlclient.ObjectKeyFromObject(kkCluster))

			return ctrl.Result{}, err
		}
	case infrav1beta1.KKClusterPhasePending:
		// Switch kkCluster.Status.Phase to `Pending`, also add HostReadyCondition.
		excepted := kkCluster.DeepCopy()
		kkCluster.Status.Phase = infrav1beta1.KKClusterPhaseRunning
		// Set series of conditions as `Unknown` for the next reconciles.
		conditions.MarkUnknown(s.KKCluster, infrav1beta1.HostsReadyCondition,
			infrav1beta1.WaitingCheckHostReadyReason, infrav1beta1.WaitingCheckHostReadyMessage)
		if err := r.Client.Status().Patch(ctx, kkCluster, ctrlclient.MergeFrom(excepted)); err != nil {
			klog.V(5).ErrorS(err, "Update KKCluster error", "KKCluster", ctrlclient.ObjectKeyFromObject(kkCluster))

			return ctrl.Result{}, err
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
		kkCluster.Spec.ControlPlaneEndpoint = clusterv1.APIEndpoint{
			Host: lb.Host,
			Port: s.APIServerPort(),
		}
	}

	kkCluster.Status.Ready = true

	return ctrl.Result{
		RequeueAfter: 30 * time.Second,
	}, nil
}

func (r *KKClusterReconciler) reconcileNormalRunning(ctx context.Context, s *scope.ClusterScope) error {
	var reset bool
	for {
		reset = false
		for _, condition := range s.KKCluster.Status.Conditions {
			conditionsCnt := len(s.KKCluster.Status.Conditions)
			if conditions.IsFalse(s.KKCluster, condition.Type) {
				continue
			}

			switch condition.Type {
			case infrav1beta1.HostsReadyCondition:
				if err := r.dealWithHostConnectCheck(ctx, s); err != nil {
					return err
				}
			case infrav1beta1.PreparationReadyCondition:
				if err := r.dealWithPreparation(ctx, s); err != nil {
					return err
				}
			case infrav1beta1.EtcdReadyCondition:
				if err := r.dealWithEtcdInstall(ctx, s); err != nil {
					return err
				}
			case infrav1beta1.BinaryInstallCondition:
				if err := r.dealWithBinaryInstall(ctx, s); err != nil {
					return err
				}
			case infrav1beta1.BootstrapReadyCondition:
				// kubeadm init, kubeadm join
				if err := r.dealWithBootstrapReady(ctx, s); err != nil {
					return err
				}
			case infrav1beta1.ClusterReadyCondition:
				// kubectl get node
				// master -> configmap -> kubeconfig -> Client: get node
				if err := r.dealWithClusterReadyCheck(ctx, s); err != nil {
					return err
				}
				// Switch `KKCluster.Phase` to `Succeed`
				s.KKCluster.Status.Phase = infrav1beta1.KKClusterPhaseSucceed
				if err := r.Client.Status().Update(ctx, s.KKCluster); err != nil {
					klog.V(5).ErrorS(err, "Update KKCluster error", "KKCluster",
						ctrlclient.ObjectKeyFromObject(s.KKCluster))

					return err
				}
			default:
			}

			// If add new conditions, restart loop.
			if len(s.KKCluster.Status.Conditions) > conditionsCnt {
				reset = true

				break
			}
		}

		if !reset {
			break
		}
	}

	return nil
}

func (r *KKClusterReconciler) reconcileDelete(clusterScope *scope.ClusterScope) {
	klog.V(4).Info("Reconcile KKCluster delete")

	// : pipeline delete
	switch clusterScope.KKCluster.Status.Phase {
	case infrav1beta1.KKClusterPhasePending:
		// transfer into Delete phase
	case infrav1beta1.KKClusterPhaseRunning:
		// delete running pipeline & recreate delete pipeline
	case infrav1beta1.KKClusterPhaseFailed:
		// delete
	case infrav1beta1.KKClusterPhaseSucceed:
		//
	}

	// Cluster is deleted so remove the finalizer.
	controllerutil.RemoveFinalizer(clusterScope.KKCluster, infrav1beta1.ClusterFinalizer)
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
	// Initialize node select mode
	if err := r.initNodeSelectMode(s); err != nil {
		return err
	}

	// Fetch groups and hosts of `Inventory`, replicas of `KubeadmControlPlane` and `MachineDeployment`.
	inv, err := r.getInitialedInventory(ctx, s)
	if err != nil {
		return err
	}

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
	controlPlaneGroup, err := validateInventoryGroup(s.KKCluster, inv, s.KKCluster.Spec.ControlPlaneGroupName,
		int(*kcp.Spec.Replicas), unavailableHosts, unavailableGroups, false,
	)
	if err != nil {
		return err
	}

	inv.Spec.Groups[s.KKCluster.Spec.ControlPlaneGroupName] = controlPlaneGroup

	// Validate kubernetes cluster's workerGroup.
	workerGroup, err := validateInventoryGroup(s.KKCluster, inv, s.KKCluster.Spec.WorkerGroupName,
		int(*md.Spec.Replicas), unavailableHosts, unavailableGroups, false,
	)
	if err != nil {
		return err
	}

	inv.Spec.Groups[s.KKCluster.Spec.WorkerGroupName] = workerGroup

	// Update `Inventory` resource.
	if err := r.Client.Update(ctx, inv); err != nil {
		klog.V(5).ErrorS(err, "Update Inventory error", "Inventory", ctrlclient.ObjectKeyFromObject(inv))

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

		return p, nil
	case kkcorev1.PipelinePhaseFailed:
		return p, r.dealWithExecuteFailed(p, funcWithFailed)
	default:
		return &kkcorev1.Pipeline{}, nil
	}
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

// dealWithPipelinesReconciles will reconcile all pipelines created for execute `playbookName` tasks, and belong to current cluster.
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

// initNodeSelectMode function used to initialize some necessary configuration information if yaml file not config them.
func (r *KKClusterReconciler) initNodeSelectMode(s *scope.ClusterScope) error {
	// Set default value of `KKCluster` resource.
	if s.KKCluster.Spec.NodeSelectorMode == "" {
		s.KKCluster.Spec.NodeSelectorMode = infrav1beta1.DefaultNodeSelectorMode
	}
	if s.KKCluster.Spec.ControlPlaneGroupName == "" {
		s.KKCluster.Spec.ControlPlaneGroupName = infrav1beta1.DefaultControlPlaneGroupName
	}
	if s.KKCluster.Spec.WorkerGroupName == "" {
		s.KKCluster.Spec.WorkerGroupName = infrav1beta1.DefaultWorkerGroupName
	}
	if s.KKCluster.Spec.ClusterGroupName == "" {
		s.KKCluster.Spec.ClusterGroupName = infrav1beta1.DefaultClusterGroupName
	}

	return nil
}

// getInitialedInventory function is a pre-processor function, used to process `Groups` of `Inventory`to streamline
// formal processing in `dealWithHostSelector` function.
func (r *KKClusterReconciler) getInitialedInventory(ctx context.Context, s *scope.ClusterScope) (
	*kkcorev1.Inventory, error) {
	inv, err := GetInventory(ctx, r.Client, s)
	if err != nil {
		return nil, err
	}

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
	groups[s.KKCluster.Spec.ClusterGroupName] = kkcorev1.InventoryGroup{
		Groups: []string{s.KKCluster.Spec.ControlPlaneGroupName, s.KKCluster.Spec.WorkerGroupName},
	}
	if _, exists := groups[s.KKCluster.Spec.ControlPlaneGroupName]; !exists {
		groups[s.KKCluster.Spec.ControlPlaneGroupName] = kkcorev1.InventoryGroup{}
	}
	if _, exists := groups[s.KKCluster.Spec.WorkerGroupName]; !exists {
		groups[s.KKCluster.Spec.WorkerGroupName] = kkcorev1.InventoryGroup{}
	}
	inv.Spec.Groups = groups

	if err := controllerutil.SetControllerReference(s.KKCluster, inv, r.Scheme); err != nil {
		return nil, err
	}

	if err := r.Update(ctx, inv); err != nil {
		klog.ErrorS(err, "Failed to update Inventory", "Inventory", inv)

		return nil, err
	}

	return inv, nil
}

func (r *KKClusterReconciler) updateInventoryStatus(ctx context.Context, s *scope.ClusterScope, inv *kkcorev1.Inventory) error {
	// Get HostMachineMapping, and create a new one for update.
	hostMachineMapping := inv.Status.HostMachineMapping
	newHostMachineMapping := make(map[string]kkcorev1.MachineBinding)

	// Get ControlPlaneGroup and WorkerGroup.
	controlPlaneGroup := inv.Spec.Groups[s.KKCluster.Spec.ControlPlaneGroupName]
	workerGroup := inv.Spec.Groups[s.KKCluster.Spec.WorkerGroupName]

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
		Vars:   inv.Spec.Groups[kkc.Spec.ControlPlaneGroupName].Vars,
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
	ref := s.KKCluster.Spec.PipelineRef
	if ref.Namespace == "" {
		ref.Namespace = s.Namespace()
	}

	pipelineTemplate, err := GetPipelineTemplateFromRef(ctx, r.Client, s.KKCluster.Spec.PipelineRef)
	if err != nil {
		return nil, err
	}

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
			Project:      pipelineTemplate.Spec.Project,
			Playbook:     playbook,
			InventoryRef: pipelineTemplate.Spec.InventoryRef,
			ConfigRef:    pipelineTemplate.Spec.ConfigRef,
			Tags:         pipelineTemplate.Spec.Tags,
			SkipTags:     pipelineTemplate.Spec.SkipTags,
			Debug:        pipelineTemplate.Spec.Debug,
			JobSpec:      pipelineTemplate.Spec.JobSpec,
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

// generateKKMachines function can generate KKMachines bind with both control plane nodes and worker nodes.
// func (r *KKClusterReconciler) generateKKMachines(ctx context.Context, s *scope.ClusterScope) error {
// 	// Fetch groups and hosts of `Inventory`, replicas of `KubeadmControlPlane` and `MachineDeployment`.
// 	inv, err := GetInventory(ctx, r.Client, s)
// 	if err != nil {
// 		return err
// 	}
//
// 	kcp, err := GetKubeadmControlPlane(ctx, r.Client, s)
// 	if err != nil {
// 		return err
// 	}
//
// 	md, err := GetMachineDeployment(ctx, r.Client, s)
// 	if err != nil {
// 		return err
// 	}
//
// 	controlPlaneGroup := inv.Spec.Groups[s.KKCluster.Spec.ControlPlaneGroupName]
// 	workerGroup := inv.Spec.Groups[s.KKCluster.Spec.WorkerGroupName]
// 	controlPlaneInfraRef := kcp.Spec.MachineTemplate.InfrastructureRef
// 	workerInfraRef := kcp.Spec.MachineTemplate.InfrastructureRef
//
// 	// Iterate through the control plane hosts
// 	for _, hostName := range controlPlaneGroup.Hosts {
// 		// Generate labels for control plane
// 		labels := map[string]string{
// 			clusterv1.ClusterNameLabel:         s.Name(),
// 			clusterv1.MachineControlPlaneLabel: "true",
// 		}
//
// 		// Check if the KKMachine already exists
// 		kkMachine := &infrav1beta1.KKMachine{}
// 		err := r.Client.Get(ctx, ctrlclient.ObjectKey{
// 			Name:      s.Name() + "-" + hostName, // Name convention
// 			Namespace: s.Namespace(),
// 		}, kkMachine)
//
// 		if err != nil && apierrors.IsNotFound(err) {
// 			if err := r.generateKKMachine(ctx, s, controlPlaneInfraRef, hostName, labels); err != nil {
// 				return err
// 			}
// 		} else if err == nil {
// 			// If exists, update the KKMachine if necessary
// 			if err := r.updateKKMachine(ctx, kkMachine, controlPlaneInfraRef, labels); err != nil {
// 				return err
// 			}
// 		}
// 	}
//
// 	// Iterate through the worker group hosts
// 	for _, hostName := range workerGroup.Hosts {
// 		// Generate labels for worker nodes
// 		labels := map[string]string{
// 			clusterv1.ClusterNameLabel:           s.Name(),
// 			clusterv1.MachineDeploymentNameLabel: md.Name,
// 		}
//
// 		// Check if the KKMachine already exists
// 		kkMachine := &infrav1beta1.KKMachine{}
// 		err := r.Client.Get(ctx, ctrlclient.ObjectKey{
// 			Name:      s.Name() + "-" + hostName, // Name convention
// 			Namespace: s.Namespace(),
// 		}, kkMachine)
//
// 		if err != nil && apierrors.IsNotFound(err) {
// 			// If not found, generate a new KKMachine
// 			if err := r.generateKKMachine(ctx, s, workerInfraRef, hostName, labels); err != nil {
// 				return err
// 			}
// 		} else if err == nil {
// 			// If exists, update the KKMachine if necessary
// 			if err := r.updateKKMachine(ctx, kkMachine, workerInfraRef, labels); err != nil {
// 				return err
// 			}
// 		}
// 	}
//
// 	return nil
// }

// generateKKMachine function is used for generate a `KKMachine` resource by `Ref` and `providerID` given by other CRDs.
// Param::providerID: from `Inventory` resource, Param::ref: from `KubeadmControlPlane` for `MachineDeployment` resource.
// Param::ref is used for get `KKMachineTemplate`
// Param::labels used for bind with other CRDs.
// func (r *KKClusterReconciler) generateKKMachine(ctx context.Context, s *scope.ClusterScope, ref corev1.ObjectReference,
// 	providerID string, labels map[string]string) error {
// 	kkMachineTemplate, err := GetKKMachineTemplateFromRef(ctx, r.Client, ref)
// 	if err != nil {
// 		return err
// 	}
//
// 	// Create a new KKMachine based on the template
// 	kkMachine := &infrav1beta1.KKMachine{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      s.Name() + "-" + providerID,
// 			Namespace: s.Namespace(),
// 			Labels:    kkMachineTemplate.Spec.Template.ObjectMeta.Labels,
// 		},
// 		Spec: kkMachineTemplate.Spec.Template.Spec,
// 	}
//
// 	// Add additional labels provided
// 	for k, v := range labels {
// 		kkMachine.ObjectMeta.Labels[k] = labels[v]
// 	}
//
// 	// Assign the providerID to the new KKMachine
// 	kkMachine.Spec.ProviderID = &providerID
//
// 	// Create the new KKMachine resource
// 	return r.Client.Create(ctx, kkMachine)
// }

// updateKKMachine function used for update one exist `KKMachine` resource. Usually update `labels` and `roles`.
// func (r *KKClusterReconciler) updateKKMachine(ctx context.Context, kkm *infrav1beta1.KKMachine,
// 	ref corev1.ObjectReference, labels map[string]string) error {
// 	kkMachineTemplate, err := GetKKMachineTemplateFromRef(ctx, r.Client, ref)
// 	if err != nil {
// 		return err
// 	}
//
// 	// Update labels if they don't exist
// 	for key, value := range labels {
// 		if _, exists := kkm.Labels[key]; !exists {
// 			kkm.Labels[key] = value
// 		}
// 	}
//
// 	// Append roles if they are missing
// 	// convert old role to roleSet, used for de-duplicated
// 	roleSet := make(map[string]struct{})
// 	for _, role := range kkm.Spec.Roles {
// 		roleSet[role] = struct{}{}
// 	}
// 	// Append missing roles from the template
// 	for _, role := range kkMachineTemplate.Spec.Template.Spec.Roles {
// 		if _, exists := roleSet[role]; !exists {
// 			kkm.Spec.Roles = append(kkm.Spec.Roles, role)
// 		}
// 	}
//
// 	// Update the KKMachine resource
// 	return r.Client.Update(ctx, kkm)
// }

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

// GetConfig function return cluster's `Config` resource.
func GetConfig(ctx context.Context, client ctrlclient.Client, s *scope.ClusterScope) *kkcorev1.Config {
	config := &kkcorev1.Config{}

	namespace := s.KKCluster.Spec.ConfigRef.Namespace
	if namespace == "" {
		namespace = s.Namespace()
	}

	err := client.Get(ctx,
		types.NamespacedName{
			Name:      s.KKCluster.Spec.ConfigRef.Name,
			Namespace: namespace,
		}, config)

	if err != nil {
		klog.V(5).InfoS("Cluster not found customize `Config` resource, use default configuration default",
			"Config", ctrlclient.ObjectKeyFromObject(config))

		return nil
	}

	return config
}

// GetKubeadmControlPlane function return cluster's `KubeadmControlPlane` resource.
func GetKubeadmControlPlane(ctx context.Context, client ctrlclient.Client, s *scope.ClusterScope) (*v1beta1.KubeadmControlPlane, error) {
	kcp := &v1beta1.KubeadmControlPlane{}

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

// GetPipelineTemplateFromRef function used for generate `Pipeline` resources by `PipelineTemplate`
func GetPipelineTemplateFromRef(ctx context.Context, client ctrlclient.Client, ref *corev1.ObjectReference) (*kkcorev1.PipelineTemplate, error) {
	pipelineTemplate := &kkcorev1.PipelineTemplate{}

	namespacedName := types.NamespacedName{
		Namespace: ref.Namespace,
		Name:      ref.Name,
	}

	if err := client.Get(ctx, namespacedName, pipelineTemplate); err != nil {
		return nil, err
	}

	return pipelineTemplate, nil
}

// GetKKMachineTemplateFromRef function return `KKMachineTemplate` resource based on `ObjectReference`.
// e.g. `ObjectReference` from `KubeadmControlPlane` & `MachineDeployment` resources.
// func GetKKMachineTemplateFromRef(ctx context.Context, client ctrlclient.Client, ref corev1.ObjectReference) (*infrav1beta1.KKMachineTemplate, error) {
// 	kkMachineTemplate := &infrav1beta1.KKMachineTemplate{}
//
// 	namespacedName := types.NamespacedName{
// 		Namespace: ref.Namespace,
// 		Name:      ref.Name,
// 	}
//
// 	if err := client.Get(ctx, namespacedName, kkMachineTemplate); err != nil {
// 		return nil, err
// 	}
//
// 	return kkMachineTemplate, nil
// }

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
