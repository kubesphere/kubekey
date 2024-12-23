package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"

	capkkinfrav1beta1 "github.com/kubesphere/kubekey/api/capkk/infrastructure/v1beta1"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	clusterv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	kubeadmcpv1beta1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	clusterutil "sigs.k8s.io/cluster-api/util"
	clusterannotations "sigs.k8s.io/cluster-api/util/annotations"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	ctrlcontroller "sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/converter"
	"github.com/kubesphere/kubekey/v4/pkg/util"
)

const (
	kkClusterControllerName = "kkcluster"
)

// KKClusterReconciler reconciles a KKCluster object
type KKClusterReconciler struct {
	*runtime.Scheme
	ctrlclient.Client
	record.EventRecorder
}

// Name implements controllers.controller.
// Subtle: this method shadows the method (*Scheme).Name of KKClusterReconciler.Scheme.
func (r *KKClusterReconciler) Name() string {
	return kkClusterControllerName
}

// SetupWithManager implements controllers.controller.
func (r *KKClusterReconciler) SetupWithManager(mgr manager.Manager, o ctrlcontroller.Options) error {
	r.Scheme = mgr.GetScheme()
	r.Client = mgr.GetClient()
	r.EventRecorder = mgr.GetEventRecorderFor(r.Name())

	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(o).
		For(&capkkinfrav1beta1.KKCluster{}).
		// Watches inventory to sync kkMachine.
		Watches(&kkcorev1.Inventory{}, handler.EnqueueRequestsFromMapFunc(r.ownerToKKClusterMapFunc)).
		// Watches kkMachine to sync kkMachine.
		Watches(&capkkinfrav1beta1.KKMachine{}, handler.EnqueueRequestsFromMapFunc(r.ownerToKKClusterMapFunc)).
		Complete(r)
}

func (r *KKClusterReconciler) ownerToKKClusterMapFunc(ctx context.Context, obj ctrlclient.Object) []ctrl.Request {
	kkcluster := &capkkinfrav1beta1.KKCluster{}
	if err := util.GetOwnerFromObject(ctx, r.Scheme, r.Client, obj, kkcluster); err == nil {
		return []ctrl.Request{{NamespacedName: ctrlclient.ObjectKeyFromObject(kkcluster)}}
	}

	return nil
}

// +kubebuilder:rbac:groups=kubekey.kubesphere.io,resources=*,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=*,verbs=get;list;watch
// +kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=*,verbs=get;list;watch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=*,verbs=get;list;watch

func (r *KKClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, retErr error) {
	// Get KKCluster.
	kkcluster := &capkkinfrav1beta1.KKCluster{}
	err := r.Client.Get(ctx, req.NamespacedName, kkcluster)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	helper, err := patch.NewHelper(kkcluster, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}
	defer func() {
		if retErr != nil {
			if kkcluster.Status.FailureReason == "" {
				kkcluster.Status.FailureReason = capkkinfrav1beta1.KKClusterFailedUnknown
			}
			kkcluster.Status.FailureMessage = retErr.Error()
		}
		if err := r.reconcileStatus(ctx, kkcluster); err != nil {
			retErr = errors.Join(retErr, err)
		}
		if err := helper.Patch(ctx, kkcluster); err != nil {
			retErr = errors.Join(retErr, err)
		}
	}()

	// Fetch the Cluster.
	cluster, err := clusterutil.GetOwnerCluster(ctx, r.Client, kkcluster.ObjectMeta)
	if err != nil {
		return reconcile.Result{}, err
	}
	if cluster == nil {
		klog.V(5).InfoS("Cluster has not yet set OwnerRef")

		return reconcile.Result{}, nil
	}
	if kkcluster.Labels[clusterv1beta1.ClusterNameLabel] == "" {
		kkcluster.Labels[clusterv1beta1.ClusterNameLabel] = cluster.Name
	}

	// skip if cluster is paused.
	if clusterannotations.IsPaused(cluster, kkcluster) {
		klog.InfoS("cluster or kkcluster is marked as paused. Won't reconcile")

		return reconcile.Result{}, nil
	}

	// Add finalizer first if not set to avoid the race condition between init and delete.
	// Note: Finalizers in general can only be added when the deletionTimestamp is not set.
	if kkcluster.ObjectMeta.DeletionTimestamp.IsZero() && !controllerutil.ContainsFinalizer(kkcluster, capkkinfrav1beta1.KKClusterFinalizer) {
		controllerutil.AddFinalizer(kkcluster, capkkinfrav1beta1.KKClusterFinalizer)

		return ctrl.Result{}, nil
	}

	// Handle deleted clusters
	if !kkcluster.DeletionTimestamp.IsZero() {
		return reconcile.Result{}, r.reconcileDelete(ctx, kkcluster)
	}

	// Handle non-deleted clusters
	return reconcile.Result{}, r.reconcileNormal(ctx, kkcluster)
}

// reconcileDelete delete cluster
func (r *KKClusterReconciler) reconcileDelete(ctx context.Context, kkcluster *capkkinfrav1beta1.KKCluster) error {
	// waiting machine deleted
	kkMachineList := &capkkinfrav1beta1.KKMachineList{}
	if err := r.Client.List(ctx, kkMachineList, ctrlclient.MatchingLabels(map[string]string{"a": "b"})); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
	}
	if len(kkMachineList.Items) == 0 {
		controllerutil.RemoveFinalizer(kkcluster, capkkinfrav1beta1.KKClusterFinalizer)
	}

	return nil
}

// reconcileNormal
func (r *KKClusterReconciler) reconcileNormal(ctx context.Context, kkcluster *capkkinfrav1beta1.KKCluster) error {
	// sync inventoryHosts in kkcluster to inventory
	if err := r.reconcileInventory(ctx, kkcluster); err != nil {
		return fmt.Errorf("failed to sync inventory: %w", err)
	}

	// syncKKMachine
	if err := r.reconcileKKMachine(ctx, kkcluster); err != nil {
		return fmt.Errorf("failed to sync kkMachine: %w", err)
	}

	return nil
}

// reconcileInventory reconcile kkcluster's hosts to inventory's inventoryHosts.
func (r *KKClusterReconciler) reconcileInventory(ctx context.Context, kkcluster *capkkinfrav1beta1.KKCluster) error {
	inventoryHosts, err := converter.ConvertKKClusterToInventoryHost(kkcluster)
	if err != nil { // cannot convert kkcluster to inventory. may be kkcluster is not valid.
		kkcluster.Status.FailureReason = capkkinfrav1beta1.KKClusterFailedInvalidHosts

		return err
	}
	// check if inventory is exist
	inventory := &kkcorev1.Inventory{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kkcluster.Name,
			Namespace: kkcluster.Namespace,
		},
	}
	if err := r.Client.Get(ctx, ctrlclient.ObjectKey{Name: kkcluster.Name, Namespace: kkcluster.Namespace}, inventory); err != nil {
		if apierrors.IsNotFound(err) {
			// inventory is not exist. create inventory
			inventory.Name = kkcluster.Name
			inventory.Namespace = kkcluster.Namespace
			inventory.Labels = map[string]string{
				clusterv1beta1.ClusterNameLabel: kkcluster.Labels[clusterv1beta1.ClusterNameLabel],
			}
			inventory.Spec.Hosts = inventoryHosts
			if err := controllerutil.SetOwnerReference(kkcluster, inventory, r.Scheme); err != nil {
				return err
			}

			return r.Client.Create(ctx, inventory)
		}

		return err
	}

	// if inventory's host is match kkcluster.inventoryHosts. skip
	if reflect.DeepEqual(inventory.Spec.Hosts, inventoryHosts) {
		return nil
	}

	// set inventory
	inventory.Spec.Hosts = inventoryHosts
	// if host contains in group but not contains in hosts. remove it.
	for _, g := range inventory.Spec.Groups {
		newHosts := make([]string, 0)
		for _, h := range g.Hosts {
			if h == _const.VariableLocalHost {
				newHosts = append(newHosts, h)

				continue
			}
			for hn := range inventoryHosts {
				if h == hn {
					newHosts = append(newHosts, h)
				}
			}
		}
		g.Hosts = newHosts
	}
	if inventory.Annotations == nil {
		kkcluster.Annotations = make(map[string]string)
	}
	inventory.Annotations[kkcorev1.HostCheckPipelineAnnotation] = ""
	inventory.Status.Phase = kkcorev1.InventoryPhasePending
	inventory.Status.Ready = false

	return nil
}

// hostPing only check is node is online.
func (r *KKClusterReconciler) reconcileKKMachine(ctx context.Context, kkcluster *capkkinfrav1beta1.KKCluster) error {
	// the inventory and cluster has the same objectKey for kkcluster.
	inventory := &kkcorev1.Inventory{}
	if err := r.Client.Get(ctx, ctrlclient.ObjectKey{Name: kkcluster.Name, Namespace: kkcluster.Namespace}, inventory); err != nil {
		return err
	}

	if !inventory.Status.Ready {
		klog.InfoS("waiting inventory ready", "inventory", ctrlclient.ObjectKeyFromObject(inventory))

		return nil
	}

	cluster := &clusterv1beta1.Cluster{}
	if err := r.Client.Get(ctx, ctrlclient.ObjectKey{Name: kkcluster.Name, Namespace: kkcluster.Namespace}, cluster); err != nil {
		return err
	}

	// sync control_plane kk machine
	if err := r.syncControlPlaneKKMachine(ctx, kkcluster, inventory, cluster); err != nil {
		kkcluster.Status.FailureReason = capkkinfrav1beta1.KKClusterFailedSyncCPKKMachine

		return err
	}

	// sync worker kkMachine
	if err := r.syncWorkerKKMachine(ctx, kkcluster, inventory, cluster); err != nil {
		kkcluster.Status.FailureReason = capkkinfrav1beta1.KKClusterFailedSyncWorkerKKMachine

		return err
	}

	return nil
}

func (r *KKClusterReconciler) syncControlPlaneKKMachine(ctx context.Context, kkcluster *capkkinfrav1beta1.KKCluster, inventory *kkcorev1.Inventory, cluster *clusterv1beta1.Cluster) error {
	groupName := getControlPlaneGroupName()
	// sync control plane kk machine
	kkmachineList := &capkkinfrav1beta1.KKMachineList{}
	if err := util.GetObjectListFromOwner(ctx, r.Scheme, r.Client, kkcluster, kkmachineList, ctrlclient.MatchingLabels{
		capkkinfrav1beta1.KKMachineBelongGroupLabel: groupName,
	}); err != nil {
		return err
	}

	needToDelete := make([]capkkinfrav1beta1.KKMachine, 0)
	needToAdd := append(make([]string, 0), inventory.Spec.Groups[groupName].Hosts...) // Deep copy of the hosts slice

	for _, kkmachine := range kkmachineList.Items {
		// Check if the machine's ProviderID exists in the group hosts
		if kkmachine.Spec.ProviderID != nil && slices.Contains(inventory.Spec.Groups[groupName].Hosts, *kkmachine.Spec.ProviderID) {
			// Remove the ProviderID from needToAdd
			idx := slices.Index(needToAdd, *kkmachine.Spec.ProviderID)
			if idx != -1 {
				needToAdd = append(needToAdd[:idx], needToAdd[idx+1:]...)
			}
		} else {
			// If the machine's ProviderID is not in the group, add it to needToDelete
			needToDelete = append(needToDelete, kkmachine)
		}
	}
	for _, km := range needToDelete {
		if err := r.Client.Delete(ctx, &km); err != nil {
			return err
		}
	}

	kubeadmCPList := &kubeadmcpv1beta1.KubeadmControlPlaneList{}
	if err := util.GetObjectListFromOwner(ctx, r.Scheme, r.Client, cluster, kubeadmCPList); err != nil {
		return err
	}
	// should only have one ownerReferences from cluster.
	if len(kubeadmCPList.Items) != 1 {
		return fmt.Errorf("should only have one KubeadmControlPlane in cluster %s", ctrlclient.ObjectKeyFromObject(cluster))
	}

	kkMachineTemplate := &capkkinfrav1beta1.KKMachineTemplate{}
	if err := r.Client.Get(ctx, ctrlclient.ObjectKey{Name: kubeadmCPList.Items[0].Spec.MachineTemplate.InfrastructureRef.Name,
		Namespace: kubeadmCPList.Items[0].Spec.MachineTemplate.InfrastructureRef.Namespace}, kkMachineTemplate); err != nil {
		return err
	}

	for _, host := range needToAdd {
		kkmachine := &capkkinfrav1beta1.KKMachine{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: kkcluster.Name + "-",
				Namespace:    kkcluster.Namespace,
				Labels: map[string]string{
					capkkinfrav1beta1.KKMachineBelongGroupLabel: groupName,
					clusterv1beta1.ClusterNameLabel:             kkcluster.Labels[clusterv1beta1.ClusterNameLabel],
				},
			},
			Spec: capkkinfrav1beta1.KKMachineSpec{
				ProviderID: ptr.To(host),
				Roles:      kkMachineTemplate.Spec.Template.Spec.Roles,
				Config:     kkMachineTemplate.Spec.Template.Spec.Config,
			},
		}
		if err := controllerutil.SetOwnerReference(kkcluster, kkmachine, r.Scheme); err != nil {
			return err
		}
		if err := r.Client.Create(ctx, kkmachine); err != nil {
			return err
		}
	}

	return nil
}

func (r *KKClusterReconciler) syncWorkerKKMachine(ctx context.Context, kkcluster *capkkinfrav1beta1.KKCluster, inventory *kkcorev1.Inventory, cluster *clusterv1beta1.Cluster) error {
	groupName := getWorkerGroupName()
	// sync control plane kk machine
	kkmachineList := &capkkinfrav1beta1.KKMachineList{}
	if err := util.GetObjectListFromOwner(ctx, r.Scheme, r.Client, kkcluster, kkmachineList, ctrlclient.MatchingLabels{
		capkkinfrav1beta1.KKMachineBelongGroupLabel: groupName,
	}); err != nil {
		return err
	}

	needToDelete := make([]capkkinfrav1beta1.KKMachine, 0)
	needToAdd := append(make([]string, 0), inventory.Spec.Groups[groupName].Hosts...) // Deep copy of the hosts slice

	for _, kkmachine := range kkmachineList.Items {
		// Check if the machine's ProviderID exists in the group hosts
		if kkmachine.Spec.ProviderID != nil && slices.Contains(inventory.Spec.Groups[groupName].Hosts, *kkmachine.Spec.ProviderID) {
			// Remove the ProviderID from needToAdd
			idx := slices.Index(needToAdd, *kkmachine.Spec.ProviderID)
			if idx != -1 {
				needToAdd = append(needToAdd[:idx], needToAdd[idx+1:]...)
			}
		} else {
			// If the machine's ProviderID is not in the group, add it to needToDelete
			needToDelete = append(needToDelete, kkmachine)
		}
	}
	for _, km := range needToDelete {
		if err := r.Client.Delete(ctx, &km); err != nil {
			return err
		}
	}

	mdList := &clusterv1beta1.MachineDeploymentList{}
	if err := util.GetObjectListFromOwner(ctx, r.Scheme, r.Client, cluster, mdList); err != nil {
		return err
	}
	// should only have one ownerReferences from cluster.
	if len(mdList.Items) != 1 {
		return fmt.Errorf("should only have one MachineDeployment in cluster %s", ctrlclient.ObjectKeyFromObject(cluster))
	}

	kkMachineTemplate := &capkkinfrav1beta1.KKMachineTemplate{}
	if err := r.Client.Get(ctx, ctrlclient.ObjectKey{Name: mdList.Items[0].Spec.Template.Spec.InfrastructureRef.Name,
		Namespace: mdList.Items[0].Spec.Template.Spec.InfrastructureRef.Namespace}, kkMachineTemplate); err != nil {
		return err
	}

	for _, host := range needToAdd {
		kkmachine := &capkkinfrav1beta1.KKMachine{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: kkcluster.Name + "-",
				Namespace:    kkcluster.Namespace,
				Labels: map[string]string{
					capkkinfrav1beta1.KKMachineBelongGroupLabel: groupName,
					clusterv1beta1.ClusterNameLabel:             kkcluster.Labels[clusterv1beta1.ClusterNameLabel],
				},
			},
			Spec: capkkinfrav1beta1.KKMachineSpec{ProviderID: ptr.To(host),
				Roles:  kkMachineTemplate.Spec.Template.Spec.Roles,
				Config: kkMachineTemplate.Spec.Template.Spec.Config,
			},
		}
		if err := controllerutil.SetOwnerReference(kkcluster, kkmachine, r.Scheme); err != nil {
			return err
		}
		if err := r.Client.Create(ctx, kkmachine); err != nil {
			return err
		}
	}

	return nil
}

func (r *KKClusterReconciler) reconcileStatus(ctx context.Context, kkcluster *capkkinfrav1beta1.KKCluster) error {
	// sync KKClusterNodeReachedCondition.
	inventory := &kkcorev1.Inventory{}
	if err := r.Client.Get(ctx, ctrlclient.ObjectKey{Name: kkcluster.Name, Namespace: kkcluster.Namespace}, inventory); err != nil {
		if apierrors.IsNotFound(err) {
			conditions.MarkUnknown(kkcluster, capkkinfrav1beta1.KKClusterNodeReachedCondition, capkkinfrav1beta1.KKClusterNodeReachedConditionReasonWaiting, "waiting for inventory created")

			return nil
		}

		return err
	}
	switch inventory.Status.Phase {
	case kkcorev1.InventoryPhasePending:
		conditions.MarkUnknown(kkcluster, capkkinfrav1beta1.KKClusterNodeReachedCondition, capkkinfrav1beta1.KKClusterNodeReachedConditionReasonWaiting, "waiting for inventory host check pipeline.")
	case kkcorev1.InventoryPhaseSucceeded:
		conditions.MarkTrue(kkcluster, capkkinfrav1beta1.KKClusterNodeReachedCondition)
	case kkcorev1.InventoryPhaseFailed:
		conditions.MarkFalse(kkcluster, capkkinfrav1beta1.KKClusterNodeReachedCondition, capkkinfrav1beta1.KKClusterNodeReachedConditionReasonUnreached, clusterv1beta1.ConditionSeverityError,
			"inventory host check pipeline %q run failed", inventory.Annotations[kkcorev1.HostCheckPipelineAnnotation])
	}

	// after inventory is ready. continue create cluster
	// todo: when cluster node changed, Is it should be ready?
	kkcluster.Status.Ready = kkcluster.Status.Ready || inventory.Status.Ready

	// sync KKClusterKKMachineConditionReady.
	kkmachineList := &capkkinfrav1beta1.KKMachineList{}
	if err := util.GetObjectListFromOwner(ctx, r.Scheme, r.Client, kkcluster, kkmachineList, ctrlclient.InNamespace(kkcluster.Namespace)); err != nil {
		if apierrors.IsNotFound(err) {
			conditions.MarkUnknown(kkcluster, capkkinfrav1beta1.KKClusterKKMachineConditionReady, capkkinfrav1beta1.KKClusterKKMachineConditionReadyReasonWaiting, "waiting for kkmachine created")

			return nil
		}

		return err
	}

	// sync kkmachine status to kkcluster
	failedKKMachine := make([]string, 0)
	for _, kkmachine := range kkmachineList.Items {
		if kkmachine.Status.FailureReason != "" {
			failedKKMachine = append(failedKKMachine, kkmachine.Name)
		}
	}
	if len(failedKKMachine) != 0 {
		conditions.MarkFalse(kkcluster, capkkinfrav1beta1.KKClusterKKMachineConditionReady, capkkinfrav1beta1.KKMachineKKMachineConditionReasonFailed, clusterv1beta1.ConditionSeverityError,
			"failed kkmachine %s", strings.Join(failedKKMachine, ","))
		kkcluster.Status.FailureReason = capkkinfrav1beta1.KKMachineKKMachineConditionReasonFailed
		kkcluster.Status.FailureMessage = "[" + strings.Join(failedKKMachine, ",") + "]"
	}

	if kkcluster.Status.FailureReason == "" && len(kkmachineList.Items) == len(inventory.Spec.Groups[getControlPlaneGroupName()].Hosts)+len(inventory.Spec.Groups[getWorkerGroupName()].Hosts) {
		conditions.MarkTrue(kkcluster, capkkinfrav1beta1.KKClusterKKMachineConditionReady)
	}

	return nil
}
