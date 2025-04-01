package infrastructure

import (
	"context"
	"reflect"
	"strings"

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
	"sigs.k8s.io/cluster-api/util/conditions"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	ctrlcontroller "sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/kubesphere/kubekey/v4/cmd/controller-manager/app/options"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/converter"
	"github.com/kubesphere/kubekey/v4/pkg/util"
)

// KKClusterReconciler reconciles a KKCluster object
type KKClusterReconciler struct {
	*runtime.Scheme
	ctrlclient.Client
	record.EventRecorder
}

var _ options.Controller = &KKClusterReconciler{}
var _ reconcile.Reconciler = &KKClusterReconciler{}

// Name implements controllers.controller.
// Subtle: this method shadows the method (*Scheme).Name of KKClusterReconciler.Scheme.
func (r *KKClusterReconciler) Name() string {
	return "kkcluster-reconciler"
}

// SetupWithManager implements controllers.controller.
func (r *KKClusterReconciler) SetupWithManager(mgr manager.Manager, o options.ControllerManagerServerOptions) error {
	r.Scheme = mgr.GetScheme()
	r.Client = mgr.GetClient()
	r.EventRecorder = mgr.GetEventRecorderFor(r.Name())

	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(ctrlcontroller.Options{
			MaxConcurrentReconciles: o.MaxConcurrentReconciles,
		}).
		For(&capkkinfrav1beta1.KKCluster{}).
		// Watches inventory to sync kkmachine.
		Watches(&kkcorev1.Inventory{}, handler.EnqueueRequestsFromMapFunc(r.ownerToKKClusterMapFunc)).
		// Watches kkmachine to sync kkmachine.
		Watches(&capkkinfrav1beta1.KKMachine{}, handler.EnqueueRequestsFromMapFunc(r.ownerToKKClusterMapFunc)).
		Complete(r)
}

func (r *KKClusterReconciler) ownerToKKClusterMapFunc(ctx context.Context, obj ctrlclient.Object) []ctrl.Request {
	kkcluster := &capkkinfrav1beta1.KKCluster{}
	if err := util.GetOwnerFromObject(ctx, r.Client, obj, kkcluster); err == nil {
		return []ctrl.Request{{NamespacedName: ctrlclient.ObjectKeyFromObject(kkcluster)}}
	}

	return nil
}

// +kubebuilder:rbac:groups=kubekey.kubesphere.io,resources=*,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=*,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=*,verbs=get;list;watch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=*,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

func (r *KKClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, retErr error) {
	// Get KKCluster.
	kkcluster := &capkkinfrav1beta1.KKCluster{}
	err := r.Client.Get(ctx, req.NamespacedName, kkcluster)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return ctrl.Result{}, errors.Wrapf(err, "failed to get kkcluster %q", req.String())
		}

		return ctrl.Result{}, nil
	}
	clusterName := kkcluster.Labels[clusterv1beta1.ClusterNameLabel]
	if clusterName == "" {
		klog.V(5).InfoS("kkcluster is not belong cluster. skip", "inventory", req.String())

		return ctrl.Result{}, nil
	}
	scope, err := newClusterScope(ctx, r.Client, reconcile.Request{NamespacedName: types.NamespacedName{
		Namespace: req.Namespace,
		Name:      clusterName,
	}})
	if err != nil {
		return ctrl.Result{}, err
	}
	if err := scope.newPatchHelper(scope.KKCluster, scope.Inventory); err != nil {
		return ctrl.Result{}, err
	}
	defer func() {
		if retErr != nil {
			if scope.KKCluster.Status.FailureReason == "" {
				scope.KKCluster.Status.FailureReason = capkkinfrav1beta1.KKClusterFailedUnknown
			}
			scope.KKCluster.Status.FailureMessage = retErr.Error()
		}
		if err := r.reconcileStatus(ctx, scope); err != nil {
			retErr = errors.Join(retErr, err)
		}
		if err := scope.PatchHelper.Patch(ctx, scope.KKCluster, scope.Inventory); err != nil {
			retErr = errors.Join(retErr, err)
		}
	}()

	// skip if cluster is paused.
	if scope.isPaused() {
		klog.InfoS("cluster or kkcluster is marked as paused. Won't reconcile")

		return reconcile.Result{}, nil
	}

	// Add finalizer first if not set to avoid the race condition between init and delete.
	// Note: Finalizers in general can only be added when the deletionTimestamp is not set.
	if scope.KKCluster.ObjectMeta.DeletionTimestamp.IsZero() && !controllerutil.ContainsFinalizer(scope.KKCluster, capkkinfrav1beta1.KKClusterFinalizer) {
		controllerutil.AddFinalizer(scope.KKCluster, capkkinfrav1beta1.KKClusterFinalizer)

		return ctrl.Result{}, nil
	}

	// Handle deleted clusters
	if !scope.KKCluster.DeletionTimestamp.IsZero() {
		return reconcile.Result{}, r.reconcileDelete(ctx, scope)
	}

	// Handle non-deleted clusters
	return reconcile.Result{}, r.reconcileNormal(ctx, scope)
}

// reconcileDelete delete cluster
func (r *KKClusterReconciler) reconcileDelete(ctx context.Context, scope *clusterScope) error {
	// waiting inventory deleted
	inventoryList := &kkcorev1.InventoryList{}
	if err := util.GetObjectListFromOwner(ctx, r.Client, scope.KKCluster, inventoryList); err != nil {
		return err
	}
	for _, obj := range inventoryList.Items {
		if err := r.Client.Delete(ctx, &obj); err != nil {
			return errors.Wrapf(err, "failed to delete inventory %q", ctrlclient.ObjectKeyFromObject(&obj))
		}
	}

	if len(inventoryList.Items) == 0 {
		// Delete finalizer.
		controllerutil.RemoveFinalizer(scope.KKCluster, capkkinfrav1beta1.KKClusterFinalizer)
	}

	return nil
}

// reconcileInventory reconcile kkcluster's hosts to inventory's inventoryHosts.
func (r *KKClusterReconciler) reconcileNormal(ctx context.Context, scope *clusterScope) error {
	inventoryHosts, err := converter.ConvertKKClusterToInventoryHost(scope.KKCluster)
	if err != nil { // cannot convert kkcluster to inventory. may be kkcluster is not valid.
		scope.KKCluster.Status.FailureReason = capkkinfrav1beta1.KKClusterFailedInvalidHosts

		return err
	}
	// if inventory is not exist. create it
	if scope.Inventory.Name == "" {
		scope.Inventory = &kkcorev1.Inventory{
			ObjectMeta: metav1.ObjectMeta{
				Name:      scope.Name,
				Namespace: scope.Namespace,
				Labels: map[string]string{
					clusterv1beta1.ClusterNameLabel: scope.Name,
				},
			},
			Spec: kkcorev1.InventorySpec{
				Hosts: inventoryHosts,
			},
		}
		if err := ctrl.SetControllerReference(scope.KKCluster, scope.Inventory, r.Scheme); err != nil {
			return errors.Wrapf(err, "failed to set ownerReference from kkcluster %q to inventory", ctrlclient.ObjectKeyFromObject(scope.KKCluster))
		}

		return errors.Wrapf(r.Client.Create(ctx, scope.Inventory), "failed to create inventory for kkcluster %q", ctrlclient.ObjectKeyFromObject(scope.KKCluster))
	}

	// if inventory's host is match kkcluster.inventoryHosts. skip
	if reflect.DeepEqual(scope.Inventory.Spec.Hosts, inventoryHosts) {
		return nil
	}

	// set inventory
	scope.Inventory.Spec.Hosts = inventoryHosts
	// if host contains in group but not contains in hosts. remove it.
	for _, g := range scope.Inventory.Spec.Groups {
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
	if scope.Inventory.Annotations == nil {
		scope.Inventory.Annotations = make(map[string]string)
	}
	scope.Inventory.Annotations[kkcorev1.HostCheckPlaybookAnnotation] = ""
	scope.Inventory.Status.Phase = kkcorev1.InventoryPhasePending
	scope.Inventory.Status.Ready = false

	return nil
}

func (r *KKClusterReconciler) reconcileStatus(ctx context.Context, scope *clusterScope) error {
	// sync KKClusterNodeReachedCondition.
	switch scope.Inventory.Status.Phase {
	case kkcorev1.InventoryPhasePending:
		conditions.MarkUnknown(scope.KKCluster, capkkinfrav1beta1.KKClusterNodeReachedCondition, capkkinfrav1beta1.KKClusterNodeReachedConditionReasonWaiting, "waiting for inventory host check playbook.")
	case kkcorev1.InventoryPhaseSucceeded:
		conditions.MarkTrue(scope.KKCluster, capkkinfrav1beta1.KKClusterNodeReachedCondition)
	case kkcorev1.InventoryPhaseFailed:
		conditions.MarkFalse(scope.KKCluster, capkkinfrav1beta1.KKClusterNodeReachedCondition, capkkinfrav1beta1.KKClusterNodeReachedConditionReasonUnreached, clusterv1beta1.ConditionSeverityError,
			"inventory host check playbook %q run failed", scope.Inventory.Annotations[kkcorev1.HostCheckPlaybookAnnotation])
	}

	// after inventory is ready. continue create cluster
	// todo: when cluster node changed, Is it should be ready?
	scope.KKCluster.Status.Ready = scope.KKCluster.Status.Ready || scope.Inventory.Status.Ready

	// sync KKClusterKKMachineConditionReady.
	kkmachineList := &capkkinfrav1beta1.KKMachineList{}
	if err := r.Client.List(ctx, kkmachineList, ctrlclient.MatchingLabels{
		clusterv1beta1.ClusterNameLabel: scope.Name,
	}); err != nil {
		return errors.Wrapf(err, "failed to get kkMachineList with label %s=%s", clusterv1beta1.ClusterNameLabel, scope.Name)
	}

	// sync kkmachine status to kkcluster
	failedKKMachine := make([]string, 0)
	for _, kkmachine := range kkmachineList.Items {
		if kkmachine.Status.FailureReason != "" {
			failedKKMachine = append(failedKKMachine, kkmachine.Name)
		}
	}
	if len(failedKKMachine) != 0 {
		conditions.MarkFalse(scope.KKCluster, capkkinfrav1beta1.KKClusterKKMachineConditionReady, capkkinfrav1beta1.KKMachineKKMachineConditionReasonFailed, clusterv1beta1.ConditionSeverityError,
			"failed kkmachine %s", strings.Join(failedKKMachine, ","))
		scope.KKCluster.Status.FailureReason = capkkinfrav1beta1.KKMachineKKMachineConditionReasonFailed
		scope.KKCluster.Status.FailureMessage = "[" + strings.Join(failedKKMachine, ",") + "]"
	}

	cpn, _, err := unstructured.NestedInt64(scope.ControlPlane.Object, "spec", "replicas")
	if err != nil {
		return errors.Wrapf(err, "failed to get replicas from machineControlPlane %s", ctrlclient.ObjectKeyFromObject(scope.ControlPlane))
	}
	mdn := int(ptr.Deref(scope.MachineDeployment.Spec.Replicas, 0))
	if scope.KKCluster.Status.Ready && scope.KKCluster.Status.FailureReason == "" &&
		len(kkmachineList.Items) == int(cpn)+mdn {
		conditions.MarkTrue(scope.KKCluster, capkkinfrav1beta1.KKClusterKKMachineConditionReady)
	}

	return nil
}
