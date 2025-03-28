package infrastructure

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"
	capkkinfrav1beta1 "github.com/kubesphere/kubekey/api/capkk/infrastructure/v1beta1"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	clusterv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
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

// InventoryReconciler reconciles a Inventory object
type InventoryReconciler struct {
	ctrlclient.Client
	record.EventRecorder
}

var _ options.Controller = &InventoryReconciler{}
var _ reconcile.Reconciler = &InventoryReconciler{}

// HostSelectorFunc is a type alias for a function that selects and synchronizes hosts from the given inventory.
//
// This function is responsible for ensuring that the number of hosts in a specified group
// within the inventory matches the desired count provided by the groupHosts parameter.
//
// Parameters:
// - ctx: The context for managing deadlines, cancelation signals, and other request-scoped values.
// - groupName: The name of the host group within the inventory to be synchronized.
// - groupHostNum: The number of hosts in the specified group.
// - inventory: A pointer to the Inventory object (kkcorev1.Inventory) from which hosts will be selected or synchronized.
type HostSelectorFunc = func(ctx context.Context, groupName string, groupHostNum int, inventory *kkcorev1.Inventory) []string

// Name implements controllers.typeController.
// Subtle: this method shadows the method (*Scheme).Name of InventoryReconciler.Scheme.
func (r *InventoryReconciler) Name() string {
	return "inventory-reconciler"
}

// SetupWithManager implements controllers.typeController.
func (r *InventoryReconciler) SetupWithManager(mgr manager.Manager, o options.ControllerManagerServerOptions) error {
	r.Client = mgr.GetClient()
	r.EventRecorder = mgr.GetEventRecorderFor(r.Name())

	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(ctrlcontroller.Options{
			MaxConcurrentReconciles: o.MaxConcurrentReconciles,
		}).
		For(&kkcorev1.Inventory{}).
		// Watches kkmachine to sync group.
		Watches(&capkkinfrav1beta1.KKMachine{}, handler.EnqueueRequestsFromMapFunc(r.objectToInventoryMapFunc)).
		// Watch Playbook to sync inventory status.
		Watches(&kkcorev1.Playbook{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj ctrlclient.Object) []ctrl.Request {
			// only need host check playbook.
			inventory := &kkcorev1.Inventory{}
			if err := util.GetOwnerFromObject(ctx, r.Client, obj, inventory); err == nil {
				return []ctrl.Request{{NamespacedName: ctrlclient.ObjectKeyFromObject(inventory)}}
			}

			return nil
		})).
		Complete(r)
}

// ownerToInventoryMapFunc returns a function that maps the owner of an object to its corresponding Inventory.
// the owner usally is a KKCluster.
func (r *InventoryReconciler) objectToInventoryMapFunc(ctx context.Context, obj ctrlclient.Object) []ctrl.Request {
	// get clusterName from object label.
	clusterName := obj.GetLabels()[clusterv1beta1.ClusterNameLabel]
	if clusterName == "" {
		return nil
	}

	// inventory
	invlist := &kkcorev1.InventoryList{}
	if err := r.Client.List(ctx, invlist, ctrlclient.MatchingLabels{
		clusterv1beta1.ClusterNameLabel: clusterName,
	}); err != nil {
		return nil
	}
	reqs := make([]ctrl.Request, 0)
	for _, inventory := range invlist.Items {
		reqs = append(reqs, ctrl.Request{NamespacedName: ctrlclient.ObjectKeyFromObject(&inventory)})
	}

	return reqs
}

// Reconcile implements controllers.typeController.
func (r *InventoryReconciler) Reconcile(ctx context.Context, req reconcile.Request) (_ reconcile.Result, retErr error) {
	inventory := &kkcorev1.Inventory{}
	if err := r.Client.Get(ctx, req.NamespacedName, inventory); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, errors.Wrapf(err, "failed to get inventory %q", req.String())
	}
	clusterName := inventory.Labels[clusterv1beta1.ClusterNameLabel]
	if clusterName == "" {
		klog.V(5).InfoS("inventory is not belong cluster. skip", "inventory", req.String())

		return ctrl.Result{}, nil
	}
	scope, err := newClusterScope(ctx, r.Client, reconcile.Request{NamespacedName: types.NamespacedName{
		Namespace: req.Namespace,
		Name:      clusterName,
	}})
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}
	if err := scope.newPatchHelper(scope.Inventory); err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}
	defer func() {
		if err := scope.PatchHelper.Patch(ctx, scope.Inventory); err != nil {
			retErr = errors.Join(retErr, errors.WithStack(err))
		}
	}()

	// skip if cluster is paused.
	if scope.isPaused() {
		klog.InfoS("cluster or kkcluster is marked as paused. Won't reconcile")

		return reconcile.Result{}, nil
	}

	// Add finalizer first if not set to avoid the race condition between init and delete.
	// Note: Finalizers in general can only be added when the deletionTimestamp is not set.
	if scope.Inventory.ObjectMeta.DeletionTimestamp.IsZero() && !controllerutil.ContainsFinalizer(scope.Inventory, kkcorev1.InventoryCAPKKFinalizer) {
		controllerutil.AddFinalizer(scope.Inventory, kkcorev1.InventoryCAPKKFinalizer)

		return ctrl.Result{}, nil
	}

	// Handle deleted inventory
	if !scope.Inventory.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, errors.WithStack(r.reconcileDelete(ctx, scope))
	}

	return ctrl.Result{}, errors.WithStack(r.reconcileNormal(ctx, scope))
}

func (r *InventoryReconciler) reconcileDelete(ctx context.Context, scope *clusterScope) error {
	// waiting playbook delete
	playbookList := &kkcorev1.PlaybookList{}
	if err := util.GetObjectListFromOwner(ctx, r.Client, scope.Inventory, playbookList); err != nil {
		return errors.Wrapf(err, "failed to get playbook list from inventory %q", ctrlclient.ObjectKeyFromObject(scope.Inventory))
	}
	for _, obj := range playbookList.Items {
		if err := r.Client.Delete(ctx, &obj); err != nil {
			return errors.Wrapf(err, "failed to delete playbook %q", ctrlclient.ObjectKeyFromObject(&obj))
		}
	}
	// delete kkmachine for machine deployment
	mdList := &capkkinfrav1beta1.KKMachineList{}
	if err := r.Client.List(ctx, mdList, ctrlclient.MatchingLabels{
		clusterv1beta1.ClusterNameLabel: scope.Name,
	}, ctrlclient.HasLabels{clusterv1beta1.MachineDeploymentNameLabel}); err != nil {
		return errors.Wrapf(err, "failed to list machineDeployment with label %s=%s", clusterv1beta1.ClusterNameLabel, scope.Name)
	}
	for _, obj := range mdList.Items {
		if err := r.Client.Delete(ctx, &obj); err != nil {
			return errors.Wrapf(err, "failed to delete machineDeployment %q", ctrlclient.ObjectKeyFromObject(&obj))
		}
	}
	if len(mdList.Items) != 0 {
		// waiting kkmachine for worker delete first
		return nil
	}
	// delete kkmachine for control-plane
	cpList := &capkkinfrav1beta1.KKMachineList{}
	if err := r.Client.List(ctx, cpList, ctrlclient.MatchingLabels{
		clusterv1beta1.ClusterNameLabel: scope.Name,
	}, ctrlclient.HasLabels{clusterv1beta1.MachineControlPlaneNameLabel}); err != nil {
		return errors.Wrapf(err, "failed to list machineControlPlane with label %q", clusterv1beta1.ClusterNameLabel, scope.Name)
	}
	for _, obj := range cpList.Items {
		if err := r.Client.Delete(ctx, &obj); err != nil {
			return errors.Wrapf(err, "failed to delete machineControlPlane %q", ctrlclient.ObjectKeyFromObject(&obj))
		}
	}

	if len(playbookList.Items) == 0 && len(mdList.Items) == 0 && len(cpList.Items) == 0 {
		// Delete finalizer.
		controllerutil.RemoveFinalizer(scope.Inventory, kkcorev1.InventoryCAPKKFinalizer)
	}

	return nil
}

// Cluster API creates separate and independent Machine resources for ControlPlane and MachineDeployment. Specifically:
// • Machines created by ControlPlane always have the Kubernetes control-plane role and may also have the worker role.
// • Machines created by MachineDeployment always have the worker role but do not necessarily have the control-plane role.
// As a result, the hosts in the control-plane group and the worker group within the inventory should be mutually exclusive.
func (r *InventoryReconciler) reconcileNormal(ctx context.Context, scope *clusterScope) error {
	switch scope.Inventory.Status.Phase {
	case "", kkcorev1.InventoryPhasePending:
		// when it's empty: inventory is first created.
		// when it's pending: inventory's host haved changed.
		scope.Inventory.Status.Ready = false
		if err := r.createHostCheckPlaybook(ctx, scope); err != nil {
			return errors.Wrapf(err, "failed to create host check playbook in inventory %q", ctrlclient.ObjectKeyFromObject(scope.Inventory))
		}
		scope.Inventory.Status.Phase = kkcorev1.InventoryPhaseRunning
	case kkcorev1.InventoryPhaseRunning:
		// sync inventory's status from playbook
		if err := r.reconcileInventoryPlaybook(ctx, scope); err != nil {
			return errors.Wrapf(err, "failed to reconcile running inventory %q", ctrlclient.ObjectKeyFromObject(scope.Inventory))
		}
	case kkcorev1.InventoryPhaseSucceeded:
		// sync inventory's control_plane groups from ControlPlane
		scope.Inventory.Status.Ready = true
	case kkcorev1.InventoryPhaseFailed:
		if scope.KKCluster.Spec.Tolerate {
			scope.Inventory.Status.Ready = true
		}
		if scope.Inventory.Annotations[kkcorev1.HostCheckPlaybookAnnotation] == "" {
			// change to pending
			scope.Inventory.Status.Phase = kkcorev1.InventoryPhasePending
		}
	}

	if scope.Inventory.Status.Ready {
		if scope.Inventory.Spec.Groups == nil {
			scope.Inventory.Spec.Groups = make(map[string]kkcorev1.InventoryGroup)
		}
		if err := r.syncInventoryControlPlaneGroups(ctx, scope); err != nil {
			return errors.Wrapf(err, "failed to sync control-plane groups in inventory %q", ctrlclient.ObjectKeyFromObject(scope.Inventory))
		}
		// sync inventory's worker groups from machinedeployment
		if err := r.syncInventoryWorkerGroups(ctx, scope); err != nil {
			return errors.Wrapf(err, "failed to sync worker groups in inventory %q", ctrlclient.ObjectKeyFromObject(scope.Inventory))
		}
		scope.Inventory.Spec.Groups[defaultClusterGroup] = kkcorev1.InventoryGroup{
			Groups: []string{getControlPlaneGroupName(), getWorkerGroupName()},
		}
	}

	return nil
}

func (r *InventoryReconciler) reconcileInventoryPlaybook(ctx context.Context, scope *clusterScope) error {
	// get playbook from inventory
	if scope.Inventory.Annotations[kkcorev1.HostCheckPlaybookAnnotation] == "" {
		// cannot find playbook. should create it
		return r.createHostCheckPlaybook(ctx, scope)
	}
	playbook := &kkcorev1.Playbook{}
	if err := r.Client.Get(ctx, ctrlclient.ObjectKey{Name: scope.Inventory.Annotations[kkcorev1.HostCheckPlaybookAnnotation], Namespace: scope.Namespace}, playbook); err != nil {
		if apierrors.IsNotFound(err) {
			return r.createHostCheckPlaybook(ctx, scope)
		}

		return errors.Wrapf(err, "failed to get playbook with inventory %q annotation %q", ctrlclient.ObjectKeyFromObject(scope.Inventory), kkcorev1.HostCheckPlaybookAnnotation)
	}
	switch playbook.Status.Phase {
	case kkcorev1.PlaybookPhaseSucceeded:
		scope.Inventory.Status.Phase = kkcorev1.InventoryPhaseSucceeded
	case kkcorev1.PlaybookPhaseFailed:
		scope.Inventory.Status.Phase = kkcorev1.InventoryPhaseFailed
	}

	return nil
}

// createHostCheckPlaybook if inventory hosts is reachable.
func (r *InventoryReconciler) createHostCheckPlaybook(ctx context.Context, scope *clusterScope) error {
	if ok, _ := scope.ifPlaybookCompleted(ctx, scope.Inventory); !ok {
		return nil
	}
	// todo when install offline. should mount workdir to playbook.
	volumes, volumeMounts := scope.getVolumeMounts(ctx)
	playbook := &kkcorev1.Playbook{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: scope.Inventory.Name + "-",
			Namespace:    scope.Namespace,
			Labels: map[string]string{
				clusterv1beta1.ClusterNameLabel: scope.Name,
			},
		},
		Spec: kkcorev1.PlaybookSpec{
			Project: kkcorev1.PlaybookProject{
				Addr: _const.CAPKKProjectdir,
			},
			Playbook:     _const.CAPKKPlaybookHostCheck,
			InventoryRef: util.ObjectRef(r.Client, scope.Inventory),
			Config: kkcorev1.Config{
				Spec: runtime.RawExtension{
					Raw: fmt.Appendf(nil, `{"workdir":"%s","check_group":"%s"}`, _const.CAPKKWorkdir, scope.KKCluster.Spec.HostCheckGroup),
				},
			},
			VolumeMounts: volumeMounts,
			Volumes:      volumes,
		},
	}
	if err := ctrl.SetControllerReference(scope.Inventory, playbook, r.Client.Scheme()); err != nil {
		return errors.Wrapf(err, "failed to set ownerReference of inventory %q to playbook", ctrlclient.ObjectKeyFromObject(scope.Inventory))
	}
	if err := r.Create(ctx, playbook); err != nil {
		return errors.Wrapf(err, "failed to create playbook use inventory %q", ctrlclient.ObjectKeyFromObject(scope.Inventory))
	}

	if scope.Inventory.Annotations == nil {
		scope.Inventory.Annotations = make(map[string]string)
	}
	scope.Inventory.Annotations[kkcorev1.HostCheckPlaybookAnnotation] = playbook.Name

	return nil
}

// syncInventoryControlPlaneGroups syncs the control plane group in the inventory based on the ControlPlane.
func (r *InventoryReconciler) syncInventoryControlPlaneGroups(ctx context.Context, scope *clusterScope) error {
	groupNum, _, err := unstructured.NestedInt64(scope.ControlPlane.Object, "spec", "replicas")
	if err != nil {
		return errors.Wrapf(err, "failed to get replicas from controlPlane %q in cluster %q", ctrlclient.ObjectKeyFromObject(scope.ControlPlane), scope.String())
	}
	// Ensure the control plane group's replica count is singular. because etcd is deploy in controlPlane.
	// todo: now we only support internal etcd groups.
	if groupNum%2 != 1 {
		return errors.Errorf("controlPlane %q replicas must be singular in cluster %q", ctrlclient.ObjectKeyFromObject(scope.ControlPlane), scope.String())
	}

	// get machineList from controlPlane
	machineList := &clusterv1beta1.MachineList{}
	if err := util.GetObjectListFromOwner(ctx, r.Client, scope.ControlPlane, machineList); err != nil {
		return errors.Wrapf(err, "failed to get machineList from controlPlane %q", ctrlclient.ObjectKeyFromObject(scope.ControlPlane))
	}
	if len(machineList.Items) != int(groupNum) {
		klog.Info("waiting machine synced.")

		return nil
	}
	// get exist controlPlane hosts form kkmachine
	existControlPlaneHosts := make([]string, 0)
	kkmachineList := &capkkinfrav1beta1.KKMachineList{}
	if err := r.Client.List(ctx, kkmachineList, ctrlclient.MatchingLabels{
		clusterv1beta1.MachineControlPlaneNameLabel: scope.ControlPlane.GetName(),
	}); err != nil {
		return errors.Wrapf(err, "failed to get kkMachineList with label %s=%s", clusterv1beta1.MachineControlPlaneNameLabel, scope.ControlPlane.GetName())
	}
	for _, kkmachine := range kkmachineList.Items {
		if kkmachine.Spec.ProviderID != nil {
			existControlPlaneHosts = append(existControlPlaneHosts, _const.ProviderID2Host(scope.Name, kkmachine.Spec.ProviderID))
		}
	}
	if len(existControlPlaneHosts) > len(machineList.Items) {
		klog.Info("waiting kkmachine synced.")

		return nil
	}
	group := scope.Inventory.Spec.Groups[getControlPlaneGroupName()]
	group.Hosts = existControlPlaneHosts
	scope.Inventory.Spec.Groups[getControlPlaneGroupName()] = group
	// sync inventory's control_plane group.

	return r.setProviderID(ctx, scope.Name, kkmachineList,
		RandomSelector(ctx, getControlPlaneGroupName(), len(machineList.Items)-len(existControlPlaneHosts), scope.Inventory))
}

// syncInventoryWorkerGroups sync inventory's worker groups from machinedeployment.
func (r *InventoryReconciler) syncInventoryWorkerGroups(ctx context.Context, scope *clusterScope) error {
	groupNum := ptr.Deref(scope.MachineDeployment.Spec.Replicas, 0)
	// get machineList from machinedeployment
	machineList := &clusterv1beta1.MachineList{}
	if err := r.Client.List(ctx, machineList, ctrlclient.MatchingLabels{
		clusterv1beta1.MachineDeploymentNameLabel: scope.MachineDeployment.Name,
	}); err != nil {
		return errors.Wrapf(err, "failed to get machineList with label %s=%s", clusterv1beta1.MachineDeploymentNameLabel, scope.MachineDeployment.Name)
	}
	if len(machineList.Items) != int(groupNum) {
		klog.Info("waiting machine synced.")

		return nil
	}
	// get exist worker hosts form kkmachine
	existWorkerHosts := make([]string, 0)
	kkmachineList := &capkkinfrav1beta1.KKMachineList{}
	if err := r.Client.List(ctx, kkmachineList, ctrlclient.MatchingLabels{
		clusterv1beta1.MachineDeploymentNameLabel: scope.MachineDeployment.Name,
	}); err != nil {
		return errors.Wrapf(err, "failed to get kkmachineList with label %s=%s", clusterv1beta1.MachineDeploymentNameLabel, scope.MachineDeployment.Name)
	}
	for _, kkmachine := range kkmachineList.Items {
		if kkmachine.Spec.ProviderID != nil {
			existWorkerHosts = append(existWorkerHosts, _const.ProviderID2Host(scope.Name, kkmachine.Spec.ProviderID))
		}
	}
	if len(existWorkerHosts) > len(machineList.Items) {
		klog.Info("waiting kkmachine synced.")

		return nil
	}

	group := scope.Inventory.Spec.Groups[getWorkerGroupName()]
	group.Hosts = existWorkerHosts
	scope.Inventory.Spec.Groups[getWorkerGroupName()] = group

	return r.setProviderID(ctx, scope.Name, kkmachineList,
		RandomSelector(ctx, getWorkerGroupName(), len(machineList.Items)-len(existWorkerHosts), scope.Inventory))
}

// setProviderID set providerID to kkmachine from inventory.groups[groupName].
// if machine already has providerID, skip.
func (r *InventoryReconciler) setProviderID(ctx context.Context, clusterName string, kkmachineList *capkkinfrav1beta1.KKMachineList, availableHosts []string) error {
	// kkmachine belong to different inventory's group
	for _, host := range availableHosts {
		for _, kkmachine := range kkmachineList.Items {
			if kkmachine.Spec.ProviderID != nil {
				continue
			}
			kkmachine.Spec.ProviderID = _const.Host2ProviderID(clusterName, host)
			if err := r.Client.Update(ctx, &kkmachine); err != nil {
				return errors.Wrapf(err, "failed to set provider to kkmachine %q", ctrlclient.ObjectKeyFromObject(&kkmachine))
			}
		}
	}

	return nil
}
