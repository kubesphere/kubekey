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
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	kubedrain "k8s.io/kubectl/pkg/drain"
	"sigs.k8s.io/cluster-api/api/v1beta1/index"
	"sigs.k8s.io/cluster-api/controllers/external"
	"sigs.k8s.io/cluster-api/controllers/noderefutil"
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
	"github.com/kubesphere/kubekey/util/collections"
	"github.com/kubesphere/kubekey/util/conditions"
	"github.com/kubesphere/kubekey/util/patch"
	"github.com/kubesphere/kubekey/util/predicates"
)

const (
	// controllerName defines the controller used when creating clients.
	controllerName = "machine-controller"
)

var (
	errNilNodeRef                 = errors.New("noderef is nil")
	errLastControlPlaneNode       = errors.New("last control plane member")
	errNoControlPlaneNodes        = errors.New("no control plane members")
	errClusterIsBeingDeleted      = errors.New("cluster is being deleted")
	errControlPlaneIsBeingDeleted = errors.New("control plane is being deleted")
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

	// nodeDeletionRetryTimeout determines how long the controller will retry deleting a node
	// during a single reconciliation.
	nodeDeletionRetryTimeout time.Duration
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

	// Handle deletion reconciliation loop.
	if !m.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, cluster, m)
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

	phases := []func(context.Context, *kubekeyv1.Cluster, *kubekeyv1.Machine) (ctrl.Result, error){
		r.reconcilePing,
		r.reconcileNode,
	}

	res := ctrl.Result{}
	errs := []error{}
	for _, phase := range phases {
		// Call the inner reconciliation methods.
		phaseResult, err := phase(ctx, cluster, m)
		if err != nil {
			errs = append(errs, err)
		}
		if len(errs) > 0 {
			continue
		}
		res = util.LowestNonZeroResult(res, phaseResult)
	}
	return res, kerrors.NewAggregate(errs)
}

func (r *MachineReconciler) reconcileDelete(ctx context.Context, cluster *kubekeyv1.Cluster, m *kubekeyv1.Machine) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx, "cluster", cluster.Name)

	err := r.isDeleteNodeAllowed(ctx, cluster, m)
	isDeleteNodeAllowed := err == nil
	if err != nil {
		switch err {
		case errNoControlPlaneNodes, errLastControlPlaneNode, errNilNodeRef, errClusterIsBeingDeleted, errControlPlaneIsBeingDeleted:
			log.Info("Deleting Kubernetes Node associated with Machine is not allowed", "node", m.Status.NodeRef, "cause", err.Error())
		default:
			return ctrl.Result{}, errors.Wrapf(err, "failed to check if Kubernetes Node deletion is allowed")
		}
	}

	// Drain node before deletion and issue a patch in order to make this operation visible to the users.
	if r.isNodeDrainAllowed(m) {
		patchHelper, err := patch.NewHelper(m, r.Client)
		if err != nil {
			return ctrl.Result{}, err
		}

		log.Info("Draining node", "node", m.Status.NodeRef.Name)
		// The DrainingSucceededCondition never exists before the node is drained for the first time,
		// so its transition time can be used to record the first time draining.
		// This `if` condition prevents the transition time to be changed more than once.
		if conditions.Get(m, kubekeyv1.DrainingSucceededCondition) == nil {
			conditions.MarkFalse(m, kubekeyv1.DrainingSucceededCondition, kubekeyv1.DrainingReason, kubekeyv1.ConditionSeverityInfo, "Draining the node before deletion")
		}

		if err := patchMachine(ctx, patchHelper, m); err != nil {
			return ctrl.Result{}, errors.Wrap(err, "failed to patch Machine")
		}

		if result, err := r.drainNode(ctx, cluster, m.Status.NodeRef.Name); !result.IsZero() || err != nil {
			if err != nil {
				conditions.MarkFalse(m, kubekeyv1.DrainingSucceededCondition, kubekeyv1.DrainingFailedReason, kubekeyv1.ConditionSeverityWarning, err.Error())
				r.recorder.Eventf(m, corev1.EventTypeWarning, "FailedDrainNode", "error draining Machine's node %q: %v", m.Status.NodeRef.Name, err)
			}
			return result, err
		}

		conditions.MarkTrue(m, kubekeyv1.DrainingSucceededCondition)
		r.recorder.Eventf(m, corev1.EventTypeNormal, "SuccessfulDrainNode", "success draining Machine's node %q", m.Status.NodeRef.Name)
	}

	// Return early and don't remove the finalizer if we got an error or
	// the external reconciliation deletion isn't ready.

	patchHelper, err := patch.NewHelper(m, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}
	conditions.MarkFalse(m, kubekeyv1.MachineNodeHealthyCondition, kubekeyv1.DeletingReason, kubekeyv1.ConditionSeverityInfo, "")
	if err := patchMachine(ctx, patchHelper, m); err != nil {
		conditions.MarkFalse(m, kubekeyv1.MachineNodeHealthyCondition, kubekeyv1.DeletionFailedReason, kubekeyv1.ConditionSeverityInfo, "")
		return ctrl.Result{}, errors.Wrap(err, "failed to patch Machine")
	}

	// todo: if need to delete bootstrap
	// We only delete the node after the underlying infrastructure is gone.
	// https://github.com/kubernetes-sigs/cluster-api/issues/2565
	if isDeleteNodeAllowed {
		log.Info("Deleting node", "node", m.Status.NodeRef.Name)

		var deleteNodeErr error
		waitErr := wait.PollImmediate(2*time.Second, r.nodeDeletionRetryTimeout, func() (bool, error) {
			if deleteNodeErr = r.deleteNode(ctx, cluster, m.Status.NodeRef.Name); deleteNodeErr != nil && !apierrors.IsNotFound(errors.Cause(deleteNodeErr)) {
				return false, nil
			}
			return true, nil
		})
		if waitErr != nil {
			log.Error(deleteNodeErr, "Timed out deleting node", "node", m.Status.NodeRef.Name)
			conditions.MarkFalse(m, kubekeyv1.MachineNodeHealthyCondition, kubekeyv1.DeletionFailedReason, kubekeyv1.ConditionSeverityWarning, "")
			r.recorder.Eventf(m, corev1.EventTypeWarning, "FailedDeleteNode", "error deleting Machine's node: %v", deleteNodeErr)

			// If the node deletion timeout is not expired yet, requeue the Machine for reconciliation.
			if m.Spec.NodeDeletionTimeout == nil || m.Spec.NodeDeletionTimeout.Nanoseconds() == 0 || m.DeletionTimestamp.Add(m.Spec.NodeDeletionTimeout.Duration).After(time.Now()) {
				return ctrl.Result{}, deleteNodeErr
			}
			log.Info("Node deletion timeout expired, continuing without Node deletion.")
		}
	}

	controllerutil.RemoveFinalizer(m, kubekeyv1.MachineFinalizer)
	return ctrl.Result{}, nil
}

// isDeleteNodeAllowed returns nil only if the Machine's NodeRef is not nil
// and if the Machine is not the last control plane node in the cluster.
func (r *MachineReconciler) isDeleteNodeAllowed(ctx context.Context, cluster *kubekeyv1.Cluster, machine *kubekeyv1.Machine) error {
	log := ctrl.LoggerFrom(ctx, "cluster", cluster.Name)
	// Return early if the cluster is being deleted.
	if !cluster.DeletionTimestamp.IsZero() {
		return errClusterIsBeingDeleted
	}

	// Cannot delete something that doesn't exist.
	if machine.Status.NodeRef == nil {
		return errNilNodeRef
	}

	// controlPlaneRef is an optional field in the Cluster so skip the external
	// managed control plane check if it is nil
	if cluster.Spec.ControlPlaneRef != nil {
		controlPlane, err := external.Get(ctx, r.Client, cluster.Spec.ControlPlaneRef, cluster.Spec.ControlPlaneRef.Namespace)
		if apierrors.IsNotFound(err) {
			// If control plane object in the reference does not exist, log and skip check for
			// external managed control plane
			log.Error(err, "control plane object specified in cluster spec.controlPlaneRef does not exist", "kind", cluster.Spec.ControlPlaneRef.Kind, "name", cluster.Spec.ControlPlaneRef.Name)
		} else {
			if err != nil {
				// If any other error occurs when trying to get the control plane object,
				// return the error so we can retry
				return err
			}

			// Return early if the object referenced by controlPlaneRef is being deleted.
			if !controlPlane.GetDeletionTimestamp().IsZero() {
				return errControlPlaneIsBeingDeleted
			}
		}
	}

	// Get all of the active machines that belong to this cluster.
	machines, err := collections.GetFilteredMachinesForCluster(ctx, r.Client, cluster, collections.ActiveMachines)
	if err != nil {
		return err
	}

	// Whether or not it is okay to delete the NodeRef depends on the
	// number of remaining control plane members and whether or not this
	// machine is one of them.
	numControlPlaneMachines := len(machines.Filter(collections.ControlPlaneMachines(cluster.Name)))
	if numControlPlaneMachines == 0 {
		// Do not delete the NodeRef if there are no remaining members of
		// the control plane.
		return errNoControlPlaneNodes
	}
	// Otherwise it is okay to delete the NodeRef.
	return nil
}

func (r *MachineReconciler) isNodeDrainAllowed(m *kubekeyv1.Machine) bool {
	if _, exists := m.ObjectMeta.Annotations[kubekeyv1.ExcludeNodeDrainingAnnotation]; exists {
		return false
	}

	if r.nodeDrainTimeoutExceeded(m) {
		return false
	}

	return true
}

func (r *MachineReconciler) nodeDrainTimeoutExceeded(machine *kubekeyv1.Machine) bool {
	// if the NodeDrainTimeout type is not set by user
	if machine.Spec.NodeDrainTimeout == nil || machine.Spec.NodeDrainTimeout.Seconds() <= 0 {
		return false
	}

	// if the draining succeeded condition does not exist
	if conditions.Get(machine, kubekeyv1.DrainingSucceededCondition) == nil {
		return false
	}

	now := time.Now()
	firstTimeDrain := conditions.GetLastTransitionTime(machine, kubekeyv1.DrainingSucceededCondition)
	diff := now.Sub(firstTimeDrain.Time)
	return diff.Seconds() >= machine.Spec.NodeDrainTimeout.Seconds()
}

func (r *MachineReconciler) drainNode(ctx context.Context, cluster *kubekeyv1.Cluster, nodeName string) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx, "cluster", cluster.Name, "node", nodeName)

	restConfig, err := remote.RESTConfig(ctx, controllerName, r.Client, util.ObjectKey(cluster))
	if err != nil {
		log.Error(err, "Error creating a remote client while deleting Machine, won't retry")
		return ctrl.Result{}, nil
	}
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		log.Error(err, "Error creating a remote client while deleting Machine, won't retry")
		return ctrl.Result{}, nil
	}

	node, err := kubeClient.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			// If an admin deletes the node directly, we'll end up here.
			log.Error(err, "Could not find node from noderef, it may have already been deleted")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.Errorf("unable to get node %q: %v", nodeName, err)
	}

	drainer := &kubedrain.Helper{
		Client:              kubeClient,
		Ctx:                 ctx,
		Force:               true,
		IgnoreAllDaemonSets: true,
		DeleteEmptyDirData:  true,
		GracePeriodSeconds:  -1,
		// If a pod is not evicted in 20 seconds, retry the eviction next time the
		// machine gets reconciled again (to allow other machines to be reconciled).
		Timeout: 20 * time.Second,
		OnPodDeletedOrEvicted: func(pod *corev1.Pod, usingEviction bool) {
			verbStr := "Deleted"
			if usingEviction {
				verbStr = "Evicted"
			}
			log.Info(fmt.Sprintf("%s pod from Node", verbStr),
				"pod", fmt.Sprintf("%s/%s", pod.Name, pod.Namespace))
		},
		Out: writer{log.Info},
		ErrOut: writer{func(msg string, keysAndValues ...interface{}) {
			log.Error(nil, msg, keysAndValues...)
		}},
	}

	if noderefutil.IsNodeUnreachable(node) {
		// When the node is unreachable and some pods are not evicted for as long as this timeout, we ignore them.
		drainer.SkipWaitForDeleteTimeoutSeconds = 60 * 5 // 5 minutes
	}

	if err := kubedrain.RunCordonOrUncordon(drainer, node, true); err != nil {
		// Machine will be re-reconciled after a cordon failure.
		log.Error(err, "Cordon failed")
		return ctrl.Result{}, errors.Errorf("unable to cordon node %s: %v", node.Name, err)
	}

	if err := kubedrain.RunNodeDrain(drainer, node.Name); err != nil {
		// Machine will be re-reconciled after a drain failure.
		log.Error(err, "Drain failed, retry in 20s")
		return ctrl.Result{RequeueAfter: 20 * time.Second}, nil
	}

	log.Info("Drain successful")
	return ctrl.Result{}, nil
}

func (r *MachineReconciler) deleteNode(ctx context.Context, cluster *kubekeyv1.Cluster, name string) error {
	log := ctrl.LoggerFrom(ctx, "cluster", cluster.Name)

	remoteClient, err := r.Tracker.GetClient(ctx, util.ObjectKey(cluster))
	if err != nil {
		log.Error(err, "Error creating a remote client for cluster while deleting Machine, won't retry")
		return nil
	}

	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	if err := remoteClient.Delete(ctx, node); err != nil {
		return errors.Wrapf(err, "error deleting node %s", name)
	}
	return nil
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

// writer implements io.Writer interface as a pass-through for klog.
type writer struct {
	logFunc func(msg string, keysAndValues ...interface{})
}

// Write passes string(p) into writer's logFunc and always returns len(p).
func (w writer) Write(p []byte) (n int, err error) {
	w.logFunc(string(p))
	return len(p), nil
}
