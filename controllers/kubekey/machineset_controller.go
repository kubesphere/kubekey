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

package kubekey

import (
	"context"
	"fmt"
	"time"

	"github.com/kubesphere/kubekey/util"
	"github.com/kubesphere/kubekey/util/annotations"
	"github.com/kubesphere/kubekey/util/collections"
	"github.com/kubesphere/kubekey/util/conditions"
	"github.com/kubesphere/kubekey/util/predicates"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/cluster-api/controllers/remote"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	kubekeyv1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha3"
)

var (
	// machineSetKind contains the schema.GroupVersionKind for the MachineSet type.
	machineSetKind = kubekeyv1.GroupVersion.WithKind("MachineSet")

	// stateConfirmationTimeout is the amount of time allowed to wait for desired state.
	stateConfirmationTimeout = 10 * time.Second

	// stateConfirmationInterval is the amount of time between polling for the desired state.
	// The polling is against a local memory cache.
	stateConfirmationInterval = 100 * time.Millisecond
)

//+kubebuilder:rbac:groups=kubekey.kubesphere.io,resources=machinesets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubekey.kubesphere.io,resources=machinesets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubekey.kubesphere.io,resources=machinesets/finalizers,verbs=update

// MachineSetReconciler reconciles a MachineSet object
type MachineSetReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	Tracker *remote.ClusterCacheTracker

	// WatchFilterValue is the label value used to filter events prior to reconciliation.
	WatchFilterValue string

	recorder record.EventRecorder
}

// SetupWithManager sets up the controller with the Manager.
func (r *MachineSetReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	_, err := ctrl.NewControllerManagedBy(mgr).
		For(&kubekeyv1.MachineSet{}).
		Owns(&kubekeyv1.Machine{}).
		WithOptions(options).
		WithEventFilter(predicates.ResourceNotPausedAndHasFilterLabel(ctrl.LoggerFrom(ctx), r.WatchFilterValue)).
		Build(r)
	if err != nil {
		return errors.Wrap(err, "failed setting up with a controller manager")
	}

	r.recorder = mgr.GetEventRecorderFor("machineset-controller")
	return nil
}

func (r *MachineSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	machineSet := &kubekeyv1.MachineSet{}
	if err := r.Client.Get(ctx, req.NamespacedName, machineSet); err != nil {
		if apierrors.IsNotFound(err) {
			// Object not found, return. Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	cluster, err := util.GetClusterByName(ctx, r.Client, machineSet.ObjectMeta.Namespace, machineSet.Spec.ClusterName)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Return early if the object or Cluster is paused.
	if annotations.IsPaused(cluster, machineSet) {
		log.Info("Reconciliation is paused for this object")
		return ctrl.Result{}, nil
	}

	// Ignore deleted MachineSets, this can happen when foregroundDeletion
	// is enabled
	if !machineSet.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	result, err := r.reconcile(ctx, cluster, machineSet)
	if err != nil {
		log.Error(err, "Failed to reconcile MachineSet")
		r.recorder.Eventf(machineSet, corev1.EventTypeWarning, "ReconcileError", "%v", err)
	}
	return result, err
}

func (r *MachineSetReconciler) reconcile(ctx context.Context, cluster *kubekeyv1.Cluster, machineSet *kubekeyv1.MachineSet) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	log.V(4).Info("Reconcile MachineSet")

	// Reconcile and retrieve the Cluster object.
	if machineSet.Labels == nil {
		machineSet.Labels = make(map[string]string)
	}
	machineSet.Labels[kubekeyv1.ClusterLabelName] = machineSet.Spec.ClusterName

	if r.shouldAdopt(machineSet) {
		machineSet.OwnerReferences = util.EnsureOwnerRef(machineSet.OwnerReferences, metav1.OwnerReference{
			APIVersion: kubekeyv1.GroupVersion.String(),
			Kind:       "Cluster",
			Name:       cluster.Name,
			UID:        cluster.UID,
		})
	}

	// Make sure selector and template to be in the same cluster.
	if machineSet.Spec.Selector.MatchLabels == nil {
		machineSet.Spec.Selector.MatchLabels = make(map[string]string)
	}

	if machineSet.Spec.Template.Labels == nil {
		machineSet.Spec.Template.Labels = make(map[string]string)
	}

	machineSet.Spec.Selector.MatchLabels[kubekeyv1.ClusterLabelName] = machineSet.Spec.ClusterName
	machineSet.Spec.Template.Labels[kubekeyv1.ClusterLabelName] = machineSet.Spec.ClusterName

	selectorMap, err := metav1.LabelSelectorAsMap(&machineSet.Spec.Selector)
	if err != nil {
		return ctrl.Result{}, errors.Wrapf(err, "failed to convert MachineSet %q label selector to a map", machineSet.Name)
	}

	// Get all Machines linked to this MachineSet.
	allMachines := &kubekeyv1.MachineList{}
	err = r.Client.List(ctx,
		allMachines,
		client.InNamespace(machineSet.Namespace),
		client.MatchingLabels(selectorMap),
	)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "failed to list machines")
	}

	// Filter out irrelevant machines (deleting/mismatch labels) and claim orphaned machines.
	filteredMachines := make([]*kubekeyv1.Machine, 0, len(allMachines.Items))
	for idx := range allMachines.Items {
		machine := &allMachines.Items[idx]
		if shouldExcludeMachine(machineSet, machine) {
			continue
		}

		filteredMachines = append(filteredMachines, machine)
	}

	//var errs []error
	for _, machine := range filteredMachines {
		// filteredMachines contains machines in deleting status to calculate correct status.
		// skip remediation for those in deleting status.
		if !machine.DeletionTimestamp.IsZero() {
			continue
		}
		//if conditions.IsFalse(machine, kubekeyv1.MachineOwnerRemediatedCondition) {
		//	log.Info("Deleting unhealthy machine", "machine", machine.GetName())
		//	patch := client.MergeFrom(machine.DeepCopy())
		//	if err := r.Client.Delete(ctx, machine); err != nil {
		//		errs = append(errs, errors.Wrap(err, "failed to delete"))
		//		continue
		//	}
		//	conditions.MarkTrue(machine, kubekeyv1.MachineOwnerRemediatedCondition)
		//	if err := r.Client.Status().Patch(ctx, machine, patch); err != nil && !apierrors.IsNotFound(err) {
		//		errs = append(errs, errors.Wrap(err, "failed to update status"))
		//	}
		//}
	}

	//err = kerrors.NewAggregate(errs)
	//if err != nil {
	//	log.Info("Failed while deleting unhealthy machines", "err", err)
	//	return ctrl.Result{}, errors.Wrap(err, "failed to remediate machines")
	//}

	syncErr := r.syncReplicas(ctx, machineSet, filteredMachines)

	// Always updates status as machines come up or die.
	if err := r.updateStatus(ctx, cluster, machineSet, filteredMachines); err != nil {
		return ctrl.Result{}, errors.Wrapf(kerrors.NewAggregate([]error{err, syncErr}), "failed to update MachineSet's Status")
	}

	if syncErr != nil {
		return ctrl.Result{}, errors.Wrapf(syncErr, "failed to sync MachineSet replicas")
	}

	var replicas int32
	if len(machineSet.Spec.Template.Machines) > 0 {
		replicas = int32(len(machineSet.Spec.Template.Machines))
	}

	// Quickly reconcile until the nodes become Ready.
	if machineSet.Status.ReadyReplicas != replicas {
		log.V(4).Info("Some nodes are not ready yet, requeuing until they are ready")
		return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
	}

	return ctrl.Result{}, nil
}

// syncReplicas scales Machine resources up or down.
func (r *MachineSetReconciler) syncReplicas(ctx context.Context, ms *kubekeyv1.MachineSet, filteredMachines []*kubekeyv1.Machine) error {
	log := ctrl.LoggerFrom(ctx)

	specMachines, _, err := r.specMachines(ms)
	if err != nil {
		return err
	}

	specMachinesSet := collections.FromMachines(specMachines...)
	filteredMachinesSet := collections.FromMachines(filteredMachines...)

	//
	var (
		createMachineList []*kubekeyv1.Machine
		errs              []error
	)
	diffSpec := specMachinesSet.Difference(filteredMachinesSet)
	for _, machine := range diffSpec {
		log.Info(fmt.Sprintf("Creating machine %s", machine.GetName()))

		if ms.Annotations != nil {
			if _, ok := ms.Annotations[kubekeyv1.DisableMachineCreate]; ok {
				log.V(2).Info("Automatic creation of new machines disabled for machine set")
				return nil
			}
		}

		if err := r.Client.Create(ctx, machine); err != nil {
			log.Error(err, "Unable to create Machine", "machine", machine.Name)
			r.recorder.Eventf(ms, corev1.EventTypeWarning, "FailedCreate", "Failed to create machine %q: %v", machine.Name, err)
			errs = append(errs, err)
			conditions.MarkFalse(ms, kubekeyv1.MachinesCreatedCondition, kubekeyv1.MachineCreationFailedReason,
				kubekeyv1.ConditionSeverityError, err.Error())
			continue
		}

		log.Info("Created machine", "machine", machine.Name)
		r.recorder.Eventf(ms, corev1.EventTypeNormal, "SuccessfulCreate", "Created machine %q", machine.Name)
		createMachineList = append(createMachineList, machine)
	}

	var deleteMachineList []*kubekeyv1.Machine
	diffFiltered := filteredMachinesSet.Difference(specMachinesSet)
	for _, machine := range diffFiltered {
		log.Info(fmt.Sprintf("Deleting machine %s", machine.GetName()))

		if err := r.Client.Delete(ctx, machine); err != nil {
			log.Error(err, "Unable to delete Machine", "machine", machine.Name)
			r.recorder.Eventf(ms, corev1.EventTypeWarning, "FailedDelete", "Failed to delete machine %q: %v", machine.Name, err)
			errs = append(errs, err)
			continue
		}
		log.Info("Deleted machine", "machine", machine.Name)
		r.recorder.Eventf(ms, corev1.EventTypeNormal, "SuccessfulDelete", "Deleted machine %q", machine.Name)
		deleteMachineList = append(deleteMachineList, machine)
	}

	if len(errs) > 0 {
		return kerrors.NewAggregate(errs)
	}

	if err := r.waitForMachineCreation(ctx, createMachineList); err != nil {
		return err
	}

	if err := r.waitForMachineDeletion(ctx, deleteMachineList); err != nil {
		return err
	}

	return nil
}

// updateStatus updates the Status field for the MachineSet
// It checks for the current state of the replicas and updates the Status of the MachineSet.
func (r *MachineSetReconciler) updateStatus(ctx context.Context, cluster *kubekeyv1.Cluster, ms *kubekeyv1.MachineSet, filteredMachines []*kubekeyv1.Machine) error {
	log := ctrl.LoggerFrom(ctx)
	newStatus := ms.Status.DeepCopy()

	// Copy label selector to its status counterpart in string format.
	// This is necessary for CRDs including scale subresources.
	selector, err := metav1.LabelSelectorAsSelector(&ms.Spec.Selector)
	if err != nil {
		return errors.Wrapf(err, "failed to update status for MachineSet %s/%s", ms.Namespace, ms.Name)
	}
	newStatus.Selector = selector.String()

	fullyLabeledReplicasCount := 0
	readyReplicasCount := 0
	desiredReplicas := int32(len(ms.Spec.Template.Machines))
	templateLabel := labels.Set(ms.Spec.Template.Labels).AsSelectorPreValidated()

	for _, machine := range filteredMachines {
		if templateLabel.Matches(labels.Set(machine.Labels)) {
			fullyLabeledReplicasCount++
		}

		if machine.Status.NodeRef == nil {
			log.V(2).Info("Unable to retrieve Node status, missing NodeRef", "machine", machine.Name)
			continue
		}
	}

	newStatus.Replicas = int32(len(filteredMachines))
	newStatus.FullyLabeledReplicas = int32(fullyLabeledReplicasCount)
	newStatus.ReadyReplicas = int32(readyReplicasCount)

	// Copy the newly calculated status into the machineset
	if ms.Status.Replicas != newStatus.Replicas ||
		ms.Status.FullyLabeledReplicas != newStatus.FullyLabeledReplicas ||
		ms.Status.ReadyReplicas != newStatus.ReadyReplicas ||
		ms.Generation != ms.Status.ObservedGeneration {
		// Save the generation number we acted on, otherwise we might wrongfully indicate
		// that we've seen a spec update when we retry.
		newStatus.ObservedGeneration = ms.Generation
		newStatus.DeepCopyInto(&ms.Status)

		log.V(4).Info(fmt.Sprintf("Updating status for %v: %s/%s, ", ms.Kind, ms.Namespace, ms.Name) +
			fmt.Sprintf("replicas %d->%d (need %d), ", ms.Status.Replicas, newStatus.Replicas, desiredReplicas) +
			fmt.Sprintf("fullyLabeledReplicas %d->%d, ", ms.Status.FullyLabeledReplicas, newStatus.FullyLabeledReplicas) +
			fmt.Sprintf("readyReplicas %d->%d, ", ms.Status.ReadyReplicas, newStatus.ReadyReplicas) +
			fmt.Sprintf("sequence No: %v->%v", ms.Status.ObservedGeneration, newStatus.ObservedGeneration))
	}

	switch {
	// We are scaling up
	case newStatus.Replicas < desiredReplicas:
		conditions.MarkFalse(ms, kubekeyv1.ResizedCondition, kubekeyv1.ScalingUpReason, kubekeyv1.ConditionSeverityWarning, "Scaling up MachineSet to %d replicas (actual %d)", desiredReplicas, newStatus.Replicas)
	// We are scaling down
	case newStatus.Replicas > desiredReplicas:
		conditions.MarkFalse(ms, kubekeyv1.ResizedCondition, kubekeyv1.ScalingDownReason, kubekeyv1.ConditionSeverityWarning, "Scaling down MachineSet to %d replicas (actual %d)", desiredReplicas, newStatus.Replicas)
		// This means that there was no error in generating the desired number of machine objects
		conditions.MarkTrue(ms, kubekeyv1.MachinesCreatedCondition)
	default:
		// Make sure last resize operation is marked as completed.
		// NOTE: we are checking the number of machines ready so we report resize completed only when the machines
		// are actually provisioned (vs reporting completed immediately after the last machine object is created). This convention is also used by KCP.
		if newStatus.ReadyReplicas == newStatus.Replicas {
			conditions.MarkTrue(ms, kubekeyv1.ResizedCondition)
		}
		// This means that there was no error in generating the desired number of machine objects
		conditions.MarkTrue(ms, kubekeyv1.MachinesCreatedCondition)
	}

	// Aggregate the operational state of all the machines; while aggregating we are adding the
	// source ref (reason@machine/name) so the problem can be easily tracked down to its source machine.
	conditions.SetAggregate(
		ms,
		kubekeyv1.MachinesReadyCondition,
		collections.FromMachines(filteredMachines...).ConditionGetters(),
		conditions.AddSourceRef(),
		conditions.WithStepCounterIf(false))

	return nil
}

func (r *MachineSetReconciler) shouldAdopt(ms *kubekeyv1.MachineSet) bool {
	//todo: there is only a cluster owner, may be need to add other role?
	return !util.HasOwner(ms.OwnerReferences, kubekeyv1.GroupVersion.String(), []string{"Cluster"})
}

func (r *MachineSetReconciler) waitForMachineCreation(ctx context.Context, machineList []*kubekeyv1.Machine) error {
	log := ctrl.LoggerFrom(ctx)

	for i := 0; i < len(machineList); i++ {
		machine := machineList[i]
		pollErr := util.PollImmediate(stateConfirmationInterval, stateConfirmationTimeout, func() (bool, error) {
			key := client.ObjectKey{Namespace: machine.Namespace, Name: machine.Name}
			if err := r.Client.Get(ctx, key, &kubekeyv1.Machine{}); err != nil {
				if apierrors.IsNotFound(err) {
					return false, nil
				}
				return false, err
			}

			return true, nil
		})

		if pollErr != nil {
			log.Error(pollErr, "Failed waiting for machine object to be created")
			return errors.Wrap(pollErr, "failed waiting for machine object to be created")
		}
	}

	return nil
}

func (r *MachineSetReconciler) waitForMachineDeletion(ctx context.Context, machineList []*kubekeyv1.Machine) error {
	log := ctrl.LoggerFrom(ctx)

	for i := 0; i < len(machineList); i++ {
		machine := machineList[i]
		pollErr := util.PollImmediate(stateConfirmationInterval, stateConfirmationTimeout, func() (bool, error) {
			m := &kubekeyv1.Machine{}
			key := client.ObjectKey{Namespace: machine.Namespace, Name: machine.Name}
			err := r.Client.Get(ctx, key, m)
			if apierrors.IsNotFound(err) || !m.DeletionTimestamp.IsZero() {
				return true, nil
			}
			return false, err
		})

		if pollErr != nil {
			log.Error(pollErr, "Failed waiting for machine object to be deleted")
			return errors.Wrap(pollErr, "failed waiting for machine object to be deleted")
		}
	}
	return nil
}

func (r *MachineSetReconciler) specMachines(machineSet *kubekeyv1.MachineSet) ([]*kubekeyv1.Machine, map[string]*kubekeyv1.Machine, error) {
	// todo: need encrypt password
	machines := make([]*kubekeyv1.Machine, len(machineSet.Spec.Template.Machines))
	machinesMap := make(map[string]*kubekeyv1.Machine)
	for i := range machineSet.Spec.Template.Machines {
		specMachine := machineSet.Spec.Template.Machines[i]

		if _, ok := machinesMap[specMachine.Name]; ok {
			return machines, machinesMap, errors.Errorf("the machine name %s in Spec for machineset %v can not be duplicated", specMachine.Name, machineSet.Name)
		}

		if err := specMachine.FillAuth(&machineSet.Spec.Template.Auth); err != nil {
			return machines, machinesMap, errors.Wrapf(err, "the machine %s fill auth failed", specMachine.Name)
		}

		if err := specMachine.FillContainerManager(&machineSet.Spec.Template.ContainerManager); err != nil {
			return machines, machinesMap, errors.Wrapf(err, "the machine %s fill container manager failed", specMachine.Name)
		}

		gv := kubekeyv1.GroupVersion
		machine := &kubekeyv1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Name:            specMachine.Name,
				OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(machineSet, machineSetKind)},
				Namespace:       machineSet.Namespace,
				Labels:          machineSet.Spec.Template.Labels,
				Annotations:     machineSet.Spec.Template.Annotations,
			},
			TypeMeta: metav1.TypeMeta{
				Kind:       gv.WithKind("Machine").Kind,
				APIVersion: gv.String(),
			},
			Spec: specMachine,
		}

		machine.Spec.ClusterName = machineSet.Spec.ClusterName
		if machine.Labels == nil {
			machine.Labels = make(map[string]string)
		}

		machinesMap[machine.Name] = machine
		machines = append(machines, machine)
	}

	return machines, machinesMap, nil
}

func (r *MachineSetReconciler) getMachineNode(ctx context.Context, cluster *kubekeyv1.Cluster, machine *kubekeyv1.Machine) (*corev1.Node, error) {
	remoteClient, err := r.Tracker.GetClient(ctx, util.ObjectKey(cluster))
	if err != nil {
		return nil, err
	}
	node := &corev1.Node{}
	if err := remoteClient.Get(ctx, client.ObjectKey{Name: machine.Status.NodeRef.Name}, node); err != nil {
		return nil, errors.Wrapf(err, "error retrieving node %s for machine %s/%s", machine.Status.NodeRef.Name, machine.Namespace, machine.Name)
	}
	return node, nil
}

// shouldExcludeMachine returns true if the machine should be filtered out, false otherwise.
func shouldExcludeMachine(machineSet *kubekeyv1.MachineSet, machine *kubekeyv1.Machine) bool {
	if metav1.GetControllerOf(machine) != nil && !metav1.IsControlledBy(machine, machineSet) {
		return true
	}

	return false
}
