/*
Copyright 2022.

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
	"fmt"
	"time"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/pointer"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/controllers/remote"
	expv1 "sigs.k8s.io/cluster-api/exp/api/v1beta1"
	"sigs.k8s.io/cluster-api/feature"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/annotations"
	"sigs.k8s.io/cluster-api/util/collections"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/cluster-api/util/predicates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	infrabootstrapv1 "github.com/kubesphere/kubekey/v3/bootstrap/k3s/api/v1beta1"
	infracontrolplanev1 "github.com/kubesphere/kubekey/v3/controlplane/k3s/api/v1beta1"
	k3sCluster "github.com/kubesphere/kubekey/v3/controlplane/k3s/pkg/cluster"
	"github.com/kubesphere/kubekey/v3/util/secret"
)

// K3sControlPlaneReconciler reconciles a K3sControlPlane object
type K3sControlPlaneReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	APIReader  client.Reader
	controller controller.Controller
	recorder   record.EventRecorder
	Tracker    *remote.ClusterCacheTracker

	// WatchFilterValue is the label value used to filter events prior to reconciliation.
	WatchFilterValue string

	managementCluster         k3sCluster.ManagementCluster
	managementClusterUncached k3sCluster.ManagementCluster
}

// SetupWithManager sets up the controller with the Manager.
func (r *K3sControlPlaneReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	c, err := ctrl.NewControllerManagedBy(mgr).
		For(&infracontrolplanev1.K3sControlPlane{}).
		Owns(&clusterv1.Machine{}).
		WithOptions(options).
		WithEventFilter(predicates.ResourceNotPausedAndHasFilterLabel(ctrl.LoggerFrom(ctx), r.WatchFilterValue)).
		Build(r)
	if err != nil {
		return errors.Wrap(err, "failed setting up with a controller manager")
	}

	err = c.Watch(
		&source.Kind{Type: &clusterv1.Cluster{}},
		handler.EnqueueRequestsFromMapFunc(r.ClusterToK3sControlPlane),
		predicates.All(ctrl.LoggerFrom(ctx),
			predicates.ResourceHasFilterLabel(ctrl.LoggerFrom(ctx), r.WatchFilterValue),
			predicates.ClusterUnpausedAndInfrastructureReady(ctrl.LoggerFrom(ctx)),
		),
	)
	if err != nil {
		return errors.Wrap(err, "failed adding Watch for Clusters to controller manager")
	}

	r.controller = c
	r.recorder = mgr.GetEventRecorderFor("k3s-control-plane-controller")

	if r.managementCluster == nil {
		if r.Tracker == nil {
			return errors.New("cluster cache tracker is nil, cannot create the internal management cluster resource")
		}
		r.managementCluster = &k3sCluster.Management{
			Client:  r.Client,
			Tracker: r.Tracker,
		}
	}

	if r.managementClusterUncached == nil {
		r.managementClusterUncached = &k3sCluster.Management{Client: mgr.GetAPIReader()}
	}

	return nil
}

// +kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;patch
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io;bootstrap.cluster.x-k8s.io;controlplane.cluster.x-k8s.io,resources=*,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;clusters/status,verbs=get;list;watch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machines;machines/status,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch

// Reconcile handles K3sControlPlane events.
func (r *K3sControlPlaneReconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, retErr error) {
	log := ctrl.LoggerFrom(ctx)

	// Fetch the K3sControlPlane instance.
	kcp := &infracontrolplanev1.K3sControlPlane{}
	if err := r.Client.Get(ctx, req.NamespacedName, kcp); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Fetch the Cluster.
	cluster, err := util.GetOwnerCluster(ctx, r.Client, kcp.ObjectMeta)
	if err != nil {
		log.Error(err, "Failed to retrieve owner Cluster from the API Server")
		return ctrl.Result{}, err
	}
	if cluster == nil {
		log.Info("Cluster Controller has not yet set OwnerRef")
		return ctrl.Result{}, nil
	}
	log = log.WithValues("cluster", cluster.Name)

	if annotations.IsPaused(cluster, kcp) {
		log.Info("Reconciliation is paused for this object")
		return ctrl.Result{}, nil
	}

	// Initialize the patch helper.
	patchHelper, err := patch.NewHelper(kcp, r.Client)
	if err != nil {
		log.Error(err, "Failed to configure the patch helper")
		return ctrl.Result{Requeue: true}, nil
	}

	// Add finalizer first if not exist to avoid the race condition between init and delete
	if !controllerutil.ContainsFinalizer(kcp, infracontrolplanev1.K3sControlPlaneFinalizer) {
		controllerutil.AddFinalizer(kcp, infracontrolplanev1.K3sControlPlaneFinalizer)

		// patch and return right away instead of reusing the main defer,
		// because the main defer may take too much time to get cluster status
		// Patch ObservedGeneration only if the reconciliation completed successfully
		patchOpts := []patch.Option{patch.WithStatusObservedGeneration{}}
		if err := patchHelper.Patch(ctx, kcp, patchOpts...); err != nil {
			log.Error(err, "Failed to patch K3sControlPlane to add finalizer")
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	defer func() {
		// Always attempt to update status.
		if err := r.updateStatus(ctx, kcp, cluster); err != nil {
			var connFailure *k3sCluster.RemoteClusterConnectionError
			if errors.As(err, &connFailure) {
				log.Info("Could not connect to workload cluster to fetch status", "err", err.Error())
			} else {
				log.Error(err, "Failed to update K3sControlPlane Status")
				retErr = kerrors.NewAggregate([]error{retErr, err})
			}
		}

		// Always attempt to Patch the K3sControlPlane object and status after each reconciliation.
		if err := patchK3sControlPlane(ctx, patchHelper, kcp); err != nil {
			log.Error(err, "Failed to patch K3sControlPlane")
			retErr = kerrors.NewAggregate([]error{retErr, err})
		}

		// TODO: remove this as soon as we have a proper remote cluster cache in place.
		// Make KCP to requeue in case status is not ready, so we can check for node status without waiting for a full resync (by default 10 minutes).
		// Only requeue if we are not going in exponential backoff due to error, or if we are not already re-queueing, or if the object has a deletion timestamp.
		if retErr == nil && !res.Requeue && res.RequeueAfter <= 0 && kcp.ObjectMeta.DeletionTimestamp.IsZero() {
			if !kcp.Status.Ready {
				res = ctrl.Result{RequeueAfter: 20 * time.Second}
			}
		}
	}()

	if !kcp.ObjectMeta.DeletionTimestamp.IsZero() {
		// Handle deletion reconciliation loop.
		return r.reconcileDelete(ctx, cluster, kcp)
	}

	// Handle normal reconciliation loop.
	return r.reconcile(ctx, cluster, kcp)
}

func patchK3sControlPlane(ctx context.Context, patchHelper *patch.Helper, kcp *infracontrolplanev1.K3sControlPlane) error {
	// Always update the readyCondition by summarizing the state of other conditions.
	conditions.SetSummary(kcp,
		conditions.WithConditions(
			infracontrolplanev1.MachinesCreatedCondition,
			infracontrolplanev1.MachinesSpecUpToDateCondition,
			infracontrolplanev1.ResizedCondition,
			infracontrolplanev1.MachinesReadyCondition,
			infracontrolplanev1.AvailableCondition,
			infracontrolplanev1.CertificatesAvailableCondition,
		),
	)

	// Patch the object, ignoring conflicts on the conditions owned by this controller.
	return patchHelper.Patch(
		ctx,
		kcp,
		patch.WithOwnedConditions{Conditions: []clusterv1.ConditionType{
			infracontrolplanev1.MachinesCreatedCondition,
			clusterv1.ReadyCondition,
			infracontrolplanev1.MachinesSpecUpToDateCondition,
			infracontrolplanev1.ResizedCondition,
			infracontrolplanev1.MachinesReadyCondition,
			infracontrolplanev1.AvailableCondition,
			infracontrolplanev1.CertificatesAvailableCondition,
		}},
		patch.WithStatusObservedGeneration{},
	)
}

// reconcile handles K3sControlPlane reconciliation.
func (r *K3sControlPlaneReconciler) reconcile(ctx context.Context, cluster *clusterv1.Cluster, kcp *infracontrolplanev1.K3sControlPlane) (res ctrl.Result, retErr error) {
	log := ctrl.LoggerFrom(ctx, "cluster", cluster.Name)
	log.Info("Reconcile K3sControlPlane")

	// Make sure to reconcile the external infrastructure reference.
	if err := r.reconcileExternalReference(ctx, cluster, &kcp.Spec.MachineTemplate.InfrastructureRef); err != nil {
		return ctrl.Result{}, err
	}

	// Wait for the cluster infrastructure to be ready before creating machines
	if !cluster.Status.InfrastructureReady {
		log.Info("Cluster infrastructure is not ready yet")
		return ctrl.Result{}, nil
	}

	// Generate Cluster Certificates if needed
	certificates := secret.NewCertificatesForInitialControlPlane()
	controllerRef := metav1.NewControllerRef(kcp, infracontrolplanev1.GroupVersion.WithKind("K3sControlPlane"))
	if err := certificates.LookupOrGenerate(ctx, r.Client, util.ObjectKey(cluster), *controllerRef); err != nil {
		log.Error(err, "unable to lookup or create cluster certificates")
		conditions.MarkFalse(kcp, infracontrolplanev1.CertificatesAvailableCondition, infracontrolplanev1.CertificatesGenerationFailedReason, clusterv1.ConditionSeverityWarning, err.Error())
		return ctrl.Result{}, err
	}
	conditions.MarkTrue(kcp, infracontrolplanev1.CertificatesAvailableCondition)

	// If ControlPlaneEndpoint is not set, return early
	if !cluster.Spec.ControlPlaneEndpoint.IsValid() {
		log.Info("Cluster does not yet have a ControlPlaneEndpoint defined")
		return ctrl.Result{}, nil
	}

	// Generate Cluster Kubeconfig if needed
	if result, err := r.reconcileKubeconfig(ctx, cluster, kcp); !result.IsZero() || err != nil {
		if err != nil {
			log.Error(err, "failed to reconcile Kubeconfig")
		}
		return result, err
	}

	controlPlaneMachines, err := r.managementClusterUncached.GetMachinesForCluster(ctx, cluster, collections.ControlPlaneMachines(cluster.Name))
	if err != nil {
		log.Error(err, "failed to retrieve control plane machines for cluster")
		return ctrl.Result{}, err
	}

	adoptableMachines := controlPlaneMachines.Filter(collections.AdoptableControlPlaneMachines(cluster.Name))
	if len(adoptableMachines) > 0 {
		// We adopt the Machines and then wait for the update event for the ownership reference to re-queue them so the cache is up-to-date
		err = r.adoptMachines(ctx, kcp, adoptableMachines, cluster)
		return ctrl.Result{}, err
	}

	ownedMachines := controlPlaneMachines.Filter(collections.OwnedMachines(kcp))
	if len(ownedMachines) != len(controlPlaneMachines) {
		log.Info("Not all control plane machines are owned by this K3sControlPlane, refusing to operate in mixed management mode")
		return ctrl.Result{}, nil
	}

	controlPlane, err := k3sCluster.NewControlPlane(ctx, r.Client, cluster, kcp, ownedMachines)
	if err != nil {
		log.Error(err, "failed to initialize control plane")
		return ctrl.Result{}, err
	}

	// Aggregate the operational state of all the machines; while aggregating we are adding the
	// source ref (reason@machine/name) so the problem can be easily tracked down to its source machine.
	conditions.SetAggregate(controlPlane.KCP, infracontrolplanev1.MachinesReadyCondition, ownedMachines.ConditionGetters(), conditions.AddSourceRef(), conditions.WithStepCounterIf(false))

	// Updates conditions reporting the status of static pods and the status of the etcd cluster.
	// NOTE: Conditions reporting KCP operation progress like e.g. Resized or SpecUpToDate are inlined with the rest of the execution.
	if result, err := r.reconcileControlPlaneConditions(ctx, controlPlane); err != nil || !result.IsZero() {
		return result, err
	}

	// Reconcile unhealthy machines by triggering deletion and requeue if it is considered safe to remediate,
	// otherwise continue with the other KCP operations.
	//if result, err := r.reconcileUnhealthyMachines(ctx, controlPlane); err != nil || !result.IsZero() {
	//	return result, err
	//}

	// Control plane machines rollout due to configuration changes (e.g. upgrades) takes precedence over other operations.
	needRollout := controlPlane.MachinesNeedingRollout()
	switch {
	case len(needRollout) > 0:
		log.Info("Rolling out Control Plane machines", "needRollout", needRollout.Names())
		conditions.MarkFalse(controlPlane.KCP, infracontrolplanev1.MachinesSpecUpToDateCondition, infracontrolplanev1.RollingUpdateInProgressReason, clusterv1.ConditionSeverityWarning, "Rolling %d replicas with outdated spec (%d replicas up to date)", len(needRollout), len(controlPlane.Machines)-len(needRollout))
		return r.upgradeControlPlane(ctx, cluster, kcp, controlPlane, needRollout)
	default:
		// make sure last upgrade operation is marked as completed.
		// NOTE: we are checking the condition already exists in order to avoid to set this condition at the first
		// reconciliation/before a rolling upgrade actually starts.
		if conditions.Has(controlPlane.KCP, infracontrolplanev1.MachinesSpecUpToDateCondition) {
			conditions.MarkTrue(controlPlane.KCP, infracontrolplanev1.MachinesSpecUpToDateCondition)
		}
	}

	// If we've made it this far, we can assume that all ownedMachines are up to date
	numMachines := len(ownedMachines)
	desiredReplicas := int(*kcp.Spec.Replicas)

	switch {
	// We are creating the first replica
	case numMachines < desiredReplicas && numMachines == 0:
		// Create new Machine w/ init
		log.Info("Initializing control plane", "Desired", desiredReplicas, "Existing", numMachines)
		conditions.MarkFalse(controlPlane.KCP, infracontrolplanev1.AvailableCondition, infracontrolplanev1.WaitingForKubeadmInitReason, clusterv1.ConditionSeverityInfo, "")
		return r.initializeControlPlane(ctx, cluster, kcp, controlPlane)
	// We are scaling up
	case numMachines < desiredReplicas && numMachines > 0:
		// Create a new Machine w/ join
		log.Info("Scaling up control plane", "Desired", desiredReplicas, "Existing", numMachines)
		return r.scaleUpControlPlane(ctx, cluster, kcp, controlPlane)
	// We are scaling down
	case numMachines > desiredReplicas:
		log.Info("Scaling down control plane", "Desired", desiredReplicas, "Existing", numMachines)
		// The last parameter (i.e. machines needing to be rolled out) should always be empty here.
		return r.scaleDownControlPlane(ctx, cluster, kcp, controlPlane, collections.Machines{})
	}

	return ctrl.Result{}, nil
}

// reconcileDelete handles K3sControlPlane deletion.
// The implementation does not take non-control plane workloads into consideration. This may or may not change in the future.
// Please see https://github.com/kubernetes-sigs/cluster-api/issues/2064.
func (r *K3sControlPlaneReconciler) reconcileDelete(ctx context.Context, cluster *clusterv1.Cluster, kcp *infracontrolplanev1.K3sControlPlane) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx, "cluster", cluster.Name)
	log.Info("Reconcile K3sControlPlane deletion")

	// Gets all machines, not just control plane machines.
	allMachines, err := r.managementCluster.GetMachinesForCluster(ctx, cluster)
	if err != nil {
		return ctrl.Result{}, err
	}
	ownedMachines := allMachines.Filter(collections.OwnedMachines(kcp))

	// If no control plane machines remain, remove the finalizer
	if len(ownedMachines) == 0 {
		controllerutil.RemoveFinalizer(kcp, infracontrolplanev1.K3sControlPlaneFinalizer)
		return ctrl.Result{}, nil
	}

	controlPlane, err := k3sCluster.NewControlPlane(ctx, r.Client, cluster, kcp, ownedMachines)
	if err != nil {
		log.Error(err, "failed to initialize control plane")
		return ctrl.Result{}, err
	}

	// Updates conditions reporting the status of static pods and the status of the etcd cluster.
	// NOTE: Ignoring failures given that we are deleting
	if _, err := r.reconcileControlPlaneConditions(ctx, controlPlane); err != nil {
		log.Info("failed to reconcile conditions", "error", err.Error())
	}

	// Aggregate the operational state of all the machines; while aggregating we are adding the
	// source ref (reason@machine/name) so the problem can be easily tracked down to its source machine.
	// However, during delete we are hiding the counter (1 of x) because it does not make sense given that
	// all the machines are deleted in parallel.
	conditions.SetAggregate(kcp, infracontrolplanev1.MachinesReadyCondition, ownedMachines.ConditionGetters(), conditions.AddSourceRef(), conditions.WithStepCounterIf(false))

	allMachinePools := &expv1.MachinePoolList{}
	// Get all machine pools.
	if feature.Gates.Enabled(feature.MachinePool) {
		allMachinePools, err = r.managementCluster.GetMachinePoolsForCluster(ctx, cluster)
		if err != nil {
			return ctrl.Result{}, err
		}
	}
	// Verify that only control plane machines remain
	if len(allMachines) != len(ownedMachines) || len(allMachinePools.Items) != 0 {
		log.Info("Waiting for worker nodes to be deleted first")
		conditions.MarkFalse(kcp, infracontrolplanev1.ResizedCondition, clusterv1.DeletingReason, clusterv1.ConditionSeverityInfo, "Waiting for worker nodes to be deleted first")
		return ctrl.Result{RequeueAfter: deleteRequeueAfter}, nil
	}

	// Delete control plane machines in parallel
	machinesToDelete := ownedMachines.Filter(collections.Not(collections.HasDeletionTimestamp))
	var errs []error
	for i := range machinesToDelete {
		m := machinesToDelete[i]
		logger := log.WithValues("machine", m)
		if err := r.Client.Delete(ctx, machinesToDelete[i]); err != nil && !apierrors.IsNotFound(err) {
			logger.Error(err, "Failed to cleanup owned machine")
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		err := kerrors.NewAggregate(errs)
		r.recorder.Eventf(kcp, corev1.EventTypeWarning, "FailedDelete",
			"Failed to delete control plane Machines for cluster %s/%s control plane: %v", cluster.Namespace, cluster.Name, err)
		return ctrl.Result{}, err
	}
	conditions.MarkFalse(kcp, infracontrolplanev1.ResizedCondition, clusterv1.DeletingReason, clusterv1.ConditionSeverityInfo, "")
	return ctrl.Result{RequeueAfter: deleteRequeueAfter}, nil
}

func (r *K3sControlPlaneReconciler) adoptMachines(ctx context.Context, kcp *infracontrolplanev1.K3sControlPlane, machines collections.Machines, cluster *clusterv1.Cluster) error {
	// We do an uncached full quorum read against the KCP to avoid re-adopting Machines the garbage collector just intentionally orphaned
	// See https://github.com/kubernetes/kubernetes/issues/42639
	uncached := infracontrolplanev1.K3sControlPlane{}
	err := r.managementClusterUncached.Get(ctx, client.ObjectKey{Namespace: kcp.Namespace, Name: kcp.Name}, &uncached)
	if err != nil {
		return errors.Wrapf(err, "failed to check whether %v/%v was deleted before adoption", kcp.GetNamespace(), kcp.GetName())
	}
	if !uncached.DeletionTimestamp.IsZero() {
		return errors.Errorf("%v/%v has just been deleted at %v", kcp.GetNamespace(), kcp.GetName(), kcp.GetDeletionTimestamp())
	}

	kcpVersion, err := semver.ParseTolerant(kcp.Spec.Version)
	if err != nil {
		return errors.Wrapf(err, "failed to parse kubernetes version %q", kcp.Spec.Version)
	}

	for _, m := range machines {
		ref := m.Spec.Bootstrap.ConfigRef

		// TODO instead of returning error here, we should instead Event and add a watch on potentially adoptable Machines
		if ref == nil || ref.Kind != "K3sConfig" {
			return errors.Errorf("unable to adopt Machine %v/%v: expected a ConfigRef of kind K3sConfig but instead found %v", m.Namespace, m.Name, ref)
		}

		// TODO instead of returning error here, we should instead Event and add a watch on potentially adoptable Machines
		if ref.Namespace != "" && ref.Namespace != kcp.Namespace {
			return errors.Errorf("could not adopt resources from K3sConfig %v/%v: cannot adopt across namespaces", ref.Namespace, ref.Name)
		}

		if m.Spec.Version == nil {
			// if the machine's version is not immediately apparent, assume the operator knows what they're doing
			continue
		}

		machineVersion, err := semver.ParseTolerant(*m.Spec.Version)
		if err != nil {
			return errors.Wrapf(err, "failed to parse kubernetes version %q", *m.Spec.Version)
		}

		if !util.IsSupportedVersionSkew(kcpVersion, machineVersion) {
			r.recorder.Eventf(kcp, corev1.EventTypeWarning, "AdoptionFailed", "Could not adopt Machine %s/%s: its version (%q) is outside supported +/- one minor version skew from KCP's (%q)", m.Namespace, m.Name, *m.Spec.Version, kcp.Spec.Version)
			// avoid returning an error here so we don't cause the KCP controller to spin until the operator clarifies their intent
			return nil
		}
	}

	for _, m := range machines {
		ref := m.Spec.Bootstrap.ConfigRef
		cfg := &infrabootstrapv1.K3sConfig{}

		if err := r.Client.Get(ctx, client.ObjectKey{Name: ref.Name, Namespace: kcp.Namespace}, cfg); err != nil {
			return err
		}

		if err := r.adoptOwnedSecrets(ctx, kcp, cfg, cluster.Name); err != nil {
			return err
		}

		patchHelper, err := patch.NewHelper(m, r.Client)
		if err != nil {
			return err
		}

		if err := controllerutil.SetControllerReference(kcp, m, r.Client.Scheme()); err != nil {
			return err
		}

		// Note that ValidateOwnerReferences() will reject this patch if another
		// OwnerReference exists with controller=true.
		if err := patchHelper.Patch(ctx, m); err != nil {
			return err
		}
	}
	return nil
}

func (r *K3sControlPlaneReconciler) adoptOwnedSecrets(ctx context.Context, kcp *infracontrolplanev1.K3sControlPlane, currentOwner *infrabootstrapv1.K3sConfig, clusterName string) error {
	secrets := corev1.SecretList{}
	if err := r.Client.List(ctx, &secrets, client.InNamespace(kcp.Namespace), client.MatchingLabels{clusterv1.ClusterLabelName: clusterName}); err != nil {
		return errors.Wrap(err, "error finding secrets for adoption")
	}

	for i := range secrets.Items {
		s := secrets.Items[i]
		if !util.IsOwnedByObject(&s, currentOwner) {
			continue
		}
		// avoid taking ownership of the bootstrap data secret
		if currentOwner.Status.DataSecretName != nil && s.Name == *currentOwner.Status.DataSecretName {
			continue
		}

		ss := s.DeepCopy()

		ss.SetOwnerReferences(util.ReplaceOwnerRef(ss.GetOwnerReferences(), currentOwner, metav1.OwnerReference{
			APIVersion:         infracontrolplanev1.GroupVersion.String(),
			Kind:               "K3sControlPlane",
			Name:               kcp.Name,
			UID:                kcp.UID,
			Controller:         pointer.Bool(true),
			BlockOwnerDeletion: pointer.Bool(true),
		}))

		if err := r.Client.Update(ctx, ss); err != nil {
			return errors.Wrapf(err, "error changing secret %v ownership from K3sConfig/%v to K3sControlPlane/%v", s.Name, currentOwner.GetName(), kcp.Name)
		}
	}

	return nil
}

// reconcileControlPlaneConditions is responsible of reconciling conditions reporting the status of static pods and
// the status of the etcd cluster.
func (r *K3sControlPlaneReconciler) reconcileControlPlaneConditions(ctx context.Context, controlPlane *k3sCluster.ControlPlane) (ctrl.Result, error) {
	// If the cluster is not yet initialized, there is no way to connect to the workload cluster and fetch information
	// for updating conditions. Return early.
	if !controlPlane.KCP.Status.Initialized {
		return ctrl.Result{}, nil
	}

	workloadCluster, err := r.managementCluster.GetWorkloadCluster(ctx, util.ObjectKey(controlPlane.Cluster))
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "cannot get remote client to workload cluster")
	}

	// Update conditions status
	workloadCluster.UpdateAgentConditions(ctx, controlPlane)
	workloadCluster.UpdateEtcdConditions(ctx, controlPlane)

	// Patch machines with the updated conditions.
	if err := controlPlane.PatchMachines(ctx); err != nil {
		return ctrl.Result{}, err
	}

	// KCP will be patched at the end of Reconcile to reflect updated conditions, so we can return now.
	return ctrl.Result{}, nil
}

// ClusterToK3sControlPlane is a handler.ToRequestsFunc to be used to enqueue requests for reconciliation
// for K3sControlPlane based on updates to a Cluster.
func (r *K3sControlPlaneReconciler) ClusterToK3sControlPlane(o client.Object) []ctrl.Request {
	c, ok := o.(*clusterv1.Cluster)
	if !ok {
		panic(fmt.Sprintf("Expected a Cluster but got a %T", o))
	}

	controlPlaneRef := c.Spec.ControlPlaneRef
	if controlPlaneRef != nil && controlPlaneRef.Kind == "K3sControlPlane" {
		return []ctrl.Request{{NamespacedName: client.ObjectKey{Namespace: controlPlaneRef.Namespace, Name: controlPlaneRef.Name}}}
	}

	return nil
}
