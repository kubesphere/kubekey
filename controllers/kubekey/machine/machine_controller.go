/*
Copyright 2020 The KubeSphere Authors.

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

package machine

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/cluster-api/api/v1beta1/index"
	"sigs.k8s.io/cluster-api/controllers/external"
	"sigs.k8s.io/cluster-api/controllers/remote"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	kubekeyv1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha3"
	"github.com/kubesphere/kubekey/util"
	"github.com/kubesphere/kubekey/util/annotations"
	"github.com/kubesphere/kubekey/util/conditions"
	"github.com/kubesphere/kubekey/util/patch"
	"github.com/kubesphere/kubekey/util/predicates"
)

//+kubebuilder:rbac:groups=kubekey.kubesphere.io,resources=machines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubekey.kubesphere.io,resources=machines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubekey.kubesphere.io,resources=machines/finalizers,verbs=update

// MachineReconciler reconciles a Machine object
type MachineReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	Tracker *remote.ClusterCacheTracker

	// WatchFilterValue is the label value used to filter events prior to reconciliation.
	WatchFilterValue string

	controller      controller.Controller
	recorder        record.EventRecorder
	externalTracker external.ObjectTracker
}

// SetupWithManager sets up the controller with the Manager.
func (r *MachineReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	clusterToMachines, err := util.ClusterToObjectsMapper(mgr.GetClient(), &kubekeyv1.MachineList{}, mgr.GetScheme())
	if err != nil {
		return err
	}

	c, err := ctrl.NewControllerManagedBy(mgr).
		For(&kubekeyv1.Machine{}).
		WithOptions(options).
		WithEventFilter(predicates.ResourceNotPausedAndHasFilterLabel(ctrl.LoggerFrom(ctx), r.WatchFilterValue)).
		Build(r)
	if err != nil {
		return errors.Wrap(err, "failed setting up with a controller manager")
	}

	err = c.Watch(
		&source.Kind{Type: &kubekeyv1.Cluster{}},
		handler.EnqueueRequestsFromMapFunc(clusterToMachines),
		// TODO: should this wait for Cluster.Status.InfrastructureReady similar to Infra Machine resources?
		predicates.All(ctrl.LoggerFrom(ctx),
			predicates.Any(ctrl.LoggerFrom(ctx),
				predicates.ClusterUnpaused(ctrl.LoggerFrom(ctx)),
				predicates.ClusterControlPlaneInitialized(ctrl.LoggerFrom(ctx)),
			),
			predicates.ResourceHasFilterLabel(ctrl.LoggerFrom(ctx), r.WatchFilterValue),
		),
	)
	if err != nil {
		return errors.Wrap(err, "failed to add Watch for Clusters to controller manager")
	}

	r.controller = c

	r.recorder = mgr.GetEventRecorderFor("machine-controller")
	r.externalTracker = external.ObjectTracker{
		Controller: c,
	}

	return nil
}

func (r *MachineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, retErr error) {
	log := ctrl.LoggerFrom(ctx)

	// Fetch the Machine instance
	m := &kubekeyv1.Machine{}
	if err := r.Client.Get(ctx, req.NamespacedName, m); err != nil {
		if apierrors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return ctrl.Result{}, nil
		}

		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	cluster, err := util.GetClusterByName(ctx, r.Client, m.ObjectMeta.Namespace, m.Spec.ClusterName)
	if err != nil {
		return ctrl.Result{}, errors.Wrapf(err, "failed to get cluster %q for machine %q in namespace %q",
			m.Spec.ClusterName, m.Name, m.Namespace)
	}

	// Return early if the object or Cluster is paused.
	if annotations.IsPaused(cluster, m) {
		log.Info("Reconciliation is paused for this object")
		return ctrl.Result{}, nil
	}

	// Initialize the patch helper
	patchHelper, err := patch.NewHelper(m, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}

	defer func() {
		r.reconcilePhase(ctx, m)

		// Always attempt to patch the object and status after each reconciliation.
		// Patch ObservedGeneration only if the reconciliation completed successfully
		patchOpts := []patch.Option{}
		if retErr == nil {
			patchOpts = append(patchOpts, patch.WithStatusObservedGeneration{})
		}
		if err := patchMachine(ctx, patchHelper, m, patchOpts...); err != nil {
			retErr = kerrors.NewAggregate([]error{retErr, err})
		}
	}()

	// Reconcile labels.
	if m.Labels == nil {
		m.Labels = make(map[string]string)
	}
	m.Labels[kubekeyv1.ClusterLabelName] = m.Spec.ClusterName

	// Add finalizer first if not exist to avoid the race condition between init and delete
	if !controllerutil.ContainsFinalizer(m, kubekeyv1.MachineFinalizer) {
		controllerutil.AddFinalizer(m, kubekeyv1.MachineFinalizer)
		return ctrl.Result{}, nil
	}

	// Handle normal reconciliation loop.
	return r.reconcile(ctx, cluster, m)
}

func (r *MachineReconciler) reconcile(ctx context.Context, cluster *kubekeyv1.Cluster, m *kubekeyv1.Machine) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	log.V(4).Info("Reconcile Machine")

	if conditions.IsTrue(cluster, kubekeyv1.ControlPlaneInitializedCondition) {
		if err := r.watchClusterNodes(ctx, cluster); err != nil {
			log.Error(err, "error watching nodes on target cluster")
			return ctrl.Result{}, err
		}
	}

	// If the Machine belongs to a cluster, add an owner reference.
	if r.shouldAdopt(m) {
		m.OwnerReferences = util.EnsureOwnerRef(m.OwnerReferences, metav1.OwnerReference{
			APIVersion: kubekeyv1.GroupVersion.String(),
			Kind:       "Cluster",
			Name:       cluster.Name,
			UID:        cluster.UID,
		})
	}
	return ctrl.Result{}, nil
}

func (r *MachineReconciler) reconcileConnection(ctx context.Context, cluster *kubekeyv1.Cluster, m *kubekeyv1.Machine) (ctrl.Result, error) {
	//log := ctrl.LoggerFrom(ctx, "cluster", cluster.Name)
	return ctrl.Result{}, nil
}

func (r *MachineReconciler) watchClusterNodes(ctx context.Context, cluster *kubekeyv1.Cluster) error {
	// If there is no tracker, don't watch remote nodes
	if r.Tracker == nil {
		return nil
	}

	return r.Tracker.Watch(ctx, remote.WatchInput{
		Name:         "machine-watchNodes",
		Cluster:      util.ObjectKey(cluster),
		Watcher:      r.controller,
		Kind:         &corev1.Node{},
		EventHandler: handler.EnqueueRequestsFromMapFunc(r.nodeToMachine),
	})
}

func (r *MachineReconciler) nodeToMachine(o client.Object) []reconcile.Request {
	node, ok := o.(*corev1.Node)
	if !ok {
		panic(fmt.Sprintf("Expected a Node but got a %T", o))
	}

	var filters []client.ListOption
	// Match by clusterName when the node has the annotation.
	if clusterName, ok := node.GetAnnotations()[kubekeyv1.ClusterNameAnnotation]; ok {
		filters = append(filters, client.MatchingLabels{
			kubekeyv1.ClusterLabelName: clusterName,
		})
	}

	// Match by namespace when the node has the annotation.
	if namespace, ok := node.GetAnnotations()[kubekeyv1.ClusterNamespaceAnnotation]; ok {
		filters = append(filters, client.InNamespace(namespace))
	}

	// Match by nodeName and status.nodeRef.name.
	machineList := &kubekeyv1.MachineList{}
	if err := r.Client.List(
		context.TODO(),
		machineList,
		append(filters, client.MatchingFields{index.MachineNodeNameField: node.Name})...); err != nil {
		return nil
	}

	// There should be exactly 1 Machine for the node.
	if len(machineList.Items) == 1 {
		return []reconcile.Request{{NamespacedName: util.ObjectKey(&machineList.Items[0])}}
	}

	// There should be exactly 1 Machine for the node.
	if len(machineList.Items) == 1 {
		return []reconcile.Request{{NamespacedName: util.ObjectKey(&machineList.Items[0])}}
	}

	return nil
}

func (r *MachineReconciler) shouldAdopt(m *kubekeyv1.Machine) bool {
	return metav1.GetControllerOf(m) == nil && !util.HasOwner(m.OwnerReferences, kubekeyv1.GroupVersion.String(), []string{"Cluster"})
}

func patchMachine(ctx context.Context, patchHelper *patch.Helper, machine *kubekeyv1.Machine, options ...patch.Option) error {
	// Always update the readyCondition by summarizing the state of other conditions.
	// A step counter is added to represent progress during the provisioning process (instead we are hiding it
	// after provisioning - e.g. when a MHC condition exists - or during the deletion process).
	conditions.SetSummary(machine,
		conditions.WithConditions(
			// Infrastructure problems should take precedence over all the other conditions
			kubekeyv1.InfrastructureReadyCondition,
			// Boostrap comes after, but it is relevant only during initial machine provisioning.
			kubekeyv1.BootstrapReadyCondition,
			kubekeyv1.MachineOwnerRemediatedCondition,
		),
		conditions.WithStepCounterIf(machine.ObjectMeta.DeletionTimestamp.IsZero()),
		conditions.WithStepCounterIfOnly(
			kubekeyv1.BootstrapReadyCondition,
			kubekeyv1.InfrastructureReadyCondition,
		),
	)

	// Patch the object, ignoring conflicts on the conditions owned by this controller.
	// Also, if requested, we are adding additional options like e.g. Patch ObservedGeneration when issuing the
	// patch at the end of the reconcile loop.
	options = append(options,
		patch.WithOwnedConditions{Conditions: []kubekeyv1.ConditionType{
			kubekeyv1.ReadyCondition,
			kubekeyv1.BootstrapReadyCondition,
			kubekeyv1.InfrastructureReadyCondition,
			kubekeyv1.DrainingSucceededCondition,
			kubekeyv1.MachineOwnerRemediatedCondition,
		}},
	)

	return patchHelper.Patch(ctx, machine, options...)
}
