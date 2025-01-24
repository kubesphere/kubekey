package infrastructure

import (
	"context"
	"errors"
	"fmt"

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
	"sigs.k8s.io/cluster-api/util/annotations"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	ctrlcontroller "sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/util"
)

const (
	inventoryControllerName = "inventory"
)

// InventoryReconciler reconciles a Inventory object
type InventoryReconciler struct {
	*runtime.Scheme
	ctrlclient.Client
	record.EventRecorder
}

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
type HostSelectorFunc = func(ctx context.Context, groupName string, groupHostNum int, inventory *kkcorev1.Inventory)

// Name implements controllers.typeController.
// Subtle: this method shadows the method (*Scheme).Name of InventoryReconciler.Scheme.
func (r *InventoryReconciler) Name() string {
	return inventoryControllerName
}

// SetupWithManager implements controllers.typeController.
func (r *InventoryReconciler) SetupWithManager(mgr manager.Manager, o ctrlcontroller.TypedOptions[reconcile.Request]) error {
	r.Scheme = mgr.GetScheme()
	r.Client = mgr.GetClient()
	r.EventRecorder = mgr.GetEventRecorderFor(r.Name())

	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(o).
		For(&kkcorev1.Inventory{}).
		// Watch KubeadmControlPlane to sync control_plane group.
		Watches(&kubeadmcpv1beta1.KubeadmControlPlane{}, handler.EnqueueRequestsFromMapFunc(r.objectToInventoryMapFunc)).
		// Watch MachineDeployment to sync worker group.
		Watches(&clusterv1beta1.MachineDeployment{}, handler.EnqueueRequestsFromMapFunc(r.objectToInventoryMapFunc)).
		// Watch Pipeline to sync inventory status.
		Watches(&kkcorev1.Pipeline{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, o ctrlclient.Object) []ctrl.Request {
			// only need host check pipeline.
			pipeline, ok := o.(*kkcorev1.Pipeline)
			if !ok {
				return nil
			}
			inventory := &kkcorev1.Inventory{}
			if err := util.GetOwnerFromObject(ctx, r.Scheme, r.Client, pipeline, inventory); err == nil {
				if inventory.Annotations[kkcorev1.HostCheckPipelineAnnotation] == pipeline.Name {
					return []ctrl.Request{{NamespacedName: ctrlclient.ObjectKeyFromObject(inventory)}}
				}
			}

			return nil
		})).
		Complete(r)
}

// ownerToInventoryMapFunc returns a function that maps the owner of an object to its corresponding Inventory.
// the owner usally is a KKCluster.
func (r *InventoryReconciler) objectToInventoryMapFunc(ctx context.Context, o ctrlclient.Object) []ctrl.Request {
	// get cluster from object.
	cluster := &clusterv1beta1.Cluster{}
	if err := util.GetOwnerFromObject(ctx, r.Scheme, r.Client, o, cluster); err != nil {
		return nil
	}
	kkclusterList := &capkkinfrav1beta1.KKClusterList{}
	if err := util.GetObjectListFromOwner(ctx, r.Scheme, r.Client, cluster, kkclusterList); err != nil {
		return nil
	}
	reqs := make([]ctrl.Request, 0)
	for _, kkcluster := range kkclusterList.Items {
		// get cluster in current namespace.
		inventoryList := &kkcorev1.InventoryList{}
		if err := util.GetObjectListFromOwner(ctx, r.Scheme, r.Client, &kkcluster, inventoryList); err != nil {
			continue
		}
		for _, inventory := range inventoryList.Items {
			reqs = append(reqs, ctrl.Request{NamespacedName: ctrlclient.ObjectKeyFromObject(&inventory)})
		}
	}

	return reqs
}

// Reconcile implements controllers.typeController.
func (r *InventoryReconciler) Reconcile(ctx context.Context, req reconcile.Request) (_ reconcile.Result, retErr error) {
	// Get inventory.
	inventory := &kkcorev1.Inventory{}
	if err := r.Client.Get(ctx, req.NamespacedName, inventory); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}
	// patch helper will patch inventory and kkcluster after reconcile.
	helper, err := util.NewPatchHelper(r.Scheme, r.Client, inventory)
	if err != nil {
		return ctrl.Result{}, err
	}
	defer func() {
		if err := helper.Patch(ctx, inventory); err != nil {
			retErr = errors.Join(retErr, err)
		}
	}()

	// Fetch kkcluster from inventory
	kkcluster := &capkkinfrav1beta1.KKCluster{}
	if err := util.GetOwnerFromObject(ctx, r.Scheme, r.Client, inventory, kkcluster); err != nil {
		return ctrl.Result{}, err
	}

	// Fetch the cluster.
	cluster := &clusterv1beta1.Cluster{}
	if err := util.GetOwnerFromObject(ctx, r.Scheme, r.Client, kkcluster, cluster); err != nil {
		return ctrl.Result{}, err
	}

	// skip if cluster is paused.
	if annotations.IsPaused(cluster, inventory) {
		klog.InfoS("cluster or inventory is marked as paused. Won't reconcile")

		return reconcile.Result{}, nil
	}

	// Add finalizer first if not set to avoid the race condition between init and delete.
	// Note: Finalizers in general can only be added when the deletionTimestamp is not set.
	if inventory.ObjectMeta.DeletionTimestamp.IsZero() && !controllerutil.ContainsFinalizer(inventory, kkcorev1.InventoryCAPKKFinalizer) {
		controllerutil.AddFinalizer(inventory, kkcorev1.InventoryCAPKKFinalizer)

		return ctrl.Result{}, nil
	}

	// Handle deleted clusters
	if !kkcluster.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, r.reconcileDelete(ctx, inventory)
	}

	return ctrl.Result{}, r.reconcileNormal(ctx, inventory, kkcluster, cluster)
}

func (r *InventoryReconciler) reconcileDelete(ctx context.Context, inventory *kkcorev1.Inventory) error {
	// waiting pipeline delete
	pipelineList := &kkcorev1.PipelineList{}
	if err := util.GetObjectListFromOwner(ctx, r.Scheme, r.Client, inventory, pipelineList); err != nil {
		return err
	}
	if len(pipelineList.Items) == 0 {
		// Delete finalizer. Also remove the owner labels.
		controllerutil.RemoveFinalizer(inventory, kkcorev1.InventoryCAPKKFinalizer)
	}

	return nil
}

// Cluster API creates separate and independent Machine resources for KubeadmControlPlane and MachineDeployment. Specifically:
// • Machines created by KubeadmControlPlane always have the Kubernetes control-plane role and may also have the worker role.
// • Machines created by MachineDeployment always have the worker role but do not necessarily have the control-plane role.
// As a result, the hosts in the control-plane group and the worker group within the inventory should be mutually exclusive.
func (r *InventoryReconciler) reconcileNormal(ctx context.Context, inventory *kkcorev1.Inventory, kkcluster *capkkinfrav1beta1.KKCluster, cluster *clusterv1beta1.Cluster) error {
	switch inventory.Status.Phase {
	case "", kkcorev1.InventoryPhasePending:
		// when it's empty: inventory is first created.
		// when it's pending: inventory's host haved changed.
		inventory.Status.Ready = false
		if err := r.createHostCheckPipeline(ctx, inventory); err != nil {
			return err
		}
	case kkcorev1.InventoryPhaseRunning:
		// sync inventory's status from pipeline
		if err := r.reconcileInventoryPipeline(ctx, inventory); err != nil {
			return err
		}
	case kkcorev1.InventoryPhaseSucceeded:
		// sync inventory's control_plane groups from kubeadmcontrolplane
		if err := r.syncInventoryControlPlaneGroups(ctx, inventory, kkcluster, cluster); err != nil {
			return err
		}
		// sync inventory's worker groups from machinedeployment
		if err := r.syncInventoryWorkerGroups(ctx, inventory, kkcluster, cluster); err != nil {
			return err
		}
		inventory.Spec.Groups[defaultClusterGroup] = kkcorev1.InventoryGroup{
			Groups: []string{getControlPlaneGroupName(), getWorkerGroupName()},
		}
		inventory.Status.Ready = true
	case kkcorev1.InventoryPhaseFailed:
		if kkcluster.Spec.Tolerate {
			// sync inventory's control_plane groups from kubeadmcontrolplane
			if err := r.syncInventoryControlPlaneGroups(ctx, inventory, kkcluster, cluster); err != nil {
				return err
			}
			// sync inventory's worker groups from machinedeployment
			if err := r.syncInventoryWorkerGroups(ctx, inventory, kkcluster, cluster); err != nil {
				return err
			}
			inventory.Spec.Groups[defaultClusterGroup] = kkcorev1.InventoryGroup{
				Groups: []string{getControlPlaneGroupName(), getWorkerGroupName()},
			}
			inventory.Status.Ready = true
		}
	}

	return nil
}

func (r *InventoryReconciler) reconcileInventoryPipeline(ctx context.Context, inventory *kkcorev1.Inventory) error {
	// get pipeline from inventory
	if inventory.Annotations[kkcorev1.HostCheckPipelineAnnotation] == "" {
		return nil
	}
	pipeline := &kkcorev1.Pipeline{}
	if err := r.Client.Get(ctx, ctrlclient.ObjectKey{Name: inventory.Annotations[kkcorev1.HostCheckPipelineAnnotation], Namespace: inventory.Namespace}, pipeline); err != nil {
		return err
	}
	switch pipeline.Status.Phase {
	case kkcorev1.PipelinePhaseSucceeded:
		inventory.Status.Phase = kkcorev1.InventoryPhaseSucceeded

	case kkcorev1.PipelinePhaseFailed:
		inventory.Status.Phase = kkcorev1.InventoryPhaseFailed
	}

	return nil
}

// syncInventoryControlPlaneGroups syncs the control plane group in the inventory based on the KubeadmControlPlane.
func (r *InventoryReconciler) syncInventoryControlPlaneGroups(ctx context.Context, inventory *kkcorev1.Inventory, kkcluster *capkkinfrav1beta1.KKCluster, cluster *clusterv1beta1.Cluster) error {
	kubeadmCPList := &kubeadmcpv1beta1.KubeadmControlPlaneList{}
	if err := util.GetObjectListFromOwner(ctx, r.Scheme, r.Client, cluster, kubeadmCPList); err != nil {
		return err
	}
	// should only have one ownerReferences from cluster.
	if len(kubeadmCPList.Items) != 1 {
		return fmt.Errorf("should only have one KubeadmControlPlane in cluster %s", ctrlclient.ObjectKeyFromObject(cluster))
	}
	var groupNum int
	if ptr.To(kubeadmCPList.Items[0]).Spec.Replicas != nil {
		groupNum = int(*ptr.To(kubeadmCPList.Items[0]).Spec.Replicas)
	}
	// Ensure the control plane group's replica count is singular.
	// todo: now we only support internal etcd groups.
	if groupNum%2 != 1 && ptr.To(kubeadmCPList.Items[0]).Spec.Replicas != nil {
		return fmt.Errorf("kubeadmControlPlane %s replicas must be singular", ctrlclient.ObjectKeyFromObject(ptr.To(kubeadmCPList.Items[0])))
	}

	groupName := getControlPlaneGroupName()
	// sync inventory's control_plane group.
	getHostSelectorFunc(kkcluster.Spec.HostSelectorPolicy)(ctx, groupName, groupNum, inventory)

	return nil
}

// syncInventoryWorkerGroups sync inventory's worker groups from machinedeployment.
func (r *InventoryReconciler) syncInventoryWorkerGroups(ctx context.Context, inventory *kkcorev1.Inventory, kkcluster *capkkinfrav1beta1.KKCluster, cluster *clusterv1beta1.Cluster) error {
	mdList := &clusterv1beta1.MachineDeploymentList{}
	if err := util.GetObjectListFromOwner(ctx, r.Scheme, r.Client, cluster, mdList); err != nil {
		return err
	}
	// should only have one ownerReferences from cluster.
	if len(mdList.Items) != 1 {
		return fmt.Errorf("should only have one MachineDeployment in cluster %s", ctrlclient.ObjectKeyFromObject(cluster))
	}
	var groupNum int
	if mdList.Items[0].Spec.Replicas != nil {
		groupNum = int(*mdList.Items[0].Spec.Replicas)
	}
	groupName := getWorkerGroupName()
	// sync inventory's worker group.
	getHostSelectorFunc(kkcluster.Spec.HostSelectorPolicy)(ctx, groupName, groupNum, inventory)

	return nil
}

// createHostCheckPipeline if inventory hosts is reachable.
func (r *InventoryReconciler) createHostCheckPipeline(ctx context.Context, inventory *kkcorev1.Inventory) error {
	if ok, _ := checkIfPipelineCompleted(ctx, r.Scheme, r.Client, inventory); !ok {
		return nil
	}
	// todo when install offline. should mount workdir to pipeline.
	volumes, volumeMounts := getVolumeMountsFromEnv()
	pipeline := &kkcorev1.Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: inventory.Name + "-",
			Namespace:    inventory.Namespace,
			Annotations: map[string]string{
				kkcorev1.HostCheckPipelineAnnotation: inventory.Name,
			},
		},
		Spec: kkcorev1.PipelineSpec{
			Project: kkcorev1.PipelineProject{
				Addr: _const.CAPKKProjectdir,
			},
			Playbook:     _const.CAPKKPlaybookHostCheck,
			InventoryRef: util.ObjectRef(r.Scheme, inventory),
			Config: kkcorev1.Config{
				Spec: runtime.RawExtension{
					Raw: fmt.Appendf(nil, `{"workdir":"%s"}`, _const.CAPKKWorkdir),
				},
			},
			VolumeMounts: volumeMounts,
			Volumes:      volumes,
		},
	}
	if err := controllerutil.SetOwnerReference(inventory, pipeline, r.Scheme); err != nil {
		return err
	}
	if err := r.Create(ctx, pipeline); err != nil {
		return err
	}

	inventory.Status.Phase = kkcorev1.InventoryPhaseRunning
	inventory.Annotations[kkcorev1.HostCheckPipelineAnnotation] = pipeline.Name

	return nil
}
