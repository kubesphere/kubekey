/*
 Copyright 2022 The KubeSphere Authors.

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

package kkmachine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/pointer"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/controllers/noderefutil"
	"sigs.k8s.io/cluster-api/controllers/remote"
	capierrors "sigs.k8s.io/cluster-api/errors"
	cutil "sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/annotations"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/cluster-api/util/predicates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	infrav1 "github.com/kubesphere/kubekey/v3/api/v1beta1"
	"github.com/kubesphere/kubekey/v3/pkg"
	"github.com/kubesphere/kubekey/v3/pkg/scope"
	"github.com/kubesphere/kubekey/v3/util"
)

var (
	// kkMachineKind contains the schema.GroupVersionKind for the KKMachine type.
	kkMachineKind = infrav1.GroupVersion.WithKind("KKMachine")
)

// InstanceIDIndex defines the kk machine controller's instance ID index.
const (
	InstanceIDIndex = ".spec.instanceID"
)

// Reconciler reconciles a KKMachine object
type Reconciler struct {
	client.Client
	mutex            sync.Mutex
	Scheme           *runtime.Scheme
	Recorder         record.EventRecorder
	Tracker          *remote.ClusterCacheTracker
	WatchFilterValue string
	DataDir          string
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	log := ctrl.LoggerFrom(ctx)

	c, err := ctrl.NewControllerManagedBy(mgr).
		WithOptions(options).
		For(&infrav1.KKMachine{}).
		Watches(
			&source.Kind{Type: &clusterv1.Machine{}},
			handler.EnqueueRequestsFromMapFunc(cutil.MachineToInfrastructureMapFunc(infrav1.GroupVersion.WithKind("KKMachine"))),
		).
		Watches(
			&source.Kind{Type: &infrav1.KKCluster{}},
			handler.EnqueueRequestsFromMapFunc(r.KKClusterToKKMachines(log)),
		).
		WithEventFilter(predicates.ResourceNotPausedAndHasFilterLabel(log, r.WatchFilterValue)).
		Build(r)
	if err != nil {
		return err
	}

	// Add index to KKMachine to find by providerID
	if err := mgr.GetFieldIndexer().IndexField(ctx, &infrav1.KKMachine{},
		InstanceIDIndex,
		r.indexKKMachineByInstanceID,
	); err != nil {
		return errors.Wrap(err, "error setting index fields")
	}

	requeueKKMachinesForUnpausedCluster := r.requeueKKMachinesForUnpausedCluster(log)
	return c.Watch(
		&source.Kind{Type: &clusterv1.Cluster{}},
		handler.EnqueueRequestsFromMapFunc(requeueKKMachinesForUnpausedCluster),
		predicates.ClusterUnpausedAndInfrastructureReady(log),
	)
}

//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=kkmachines;kkmachines/status;kkmachines/finalizers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machines;machines/status,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=secrets;,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;update;patch

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, retErr error) {
	log := ctrl.LoggerFrom(ctx)

	// Fetch the KKMachine.
	kkMachine := &infrav1.KKMachine{}
	err := r.Get(ctx, req.NamespacedName, kkMachine)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Fetch the Machine.
	machine, err := cutil.GetOwnerMachine(ctx, r.Client, kkMachine.ObjectMeta)
	if err != nil {
		return ctrl.Result{}, err
	}
	if machine == nil {
		log.Info("Machine Controller has not yet set OwnerRef")
		return ctrl.Result{}, nil
	}

	log = log.WithValues("machine", machine.Name)

	// Fetch the Cluster.
	cluster, err := cutil.GetClusterFromMetadata(ctx, r.Client, machine.ObjectMeta)
	if err != nil {
		log.Info("Machine is missing cluster label or cluster does not exist")
		return ctrl.Result{}, nil
	}

	if annotations.IsPaused(cluster, kkMachine) {
		log.Info("KKMachine or linked Cluster is marked as paused. Won't reconcile")
		return ctrl.Result{}, nil
	}

	log = log.WithValues("cluster", cluster.Name)

	infraCluster, err := util.GetInfraCluster(ctx, r.Client, log, cluster, "kkmachine", r.DataDir)
	if err != nil {
		return ctrl.Result{}, errors.Wrapf(err, "error getting infra provider cluster object")
	}
	if infraCluster == nil {
		log.Info("KKCluster is not ready yet")
		return ctrl.Result{}, nil
	}

	// Create the machine scope
	machineScope, err := scope.NewMachineScope(scope.MachineScopeParams{
		Client:       r.Client,
		Logger:       &log,
		Cluster:      cluster,
		Machine:      machine,
		InfraCluster: infraCluster,
		KKMachine:    kkMachine,
	})
	if err != nil {
		log.Error(err, "failed to create scope")
		return ctrl.Result{}, err
	}

	// Always close the scope when exiting this function, so we can persist any KKMachine changes.
	defer func() {
		if err := machineScope.Close(); err != nil && retErr == nil {
			log.Error(err, "failed to patch object")
			retErr = err
		}
	}()

	if !kkMachine.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, machineScope)
	}

	return r.reconcileNormal(ctx, machineScope, infraCluster, infraCluster)
}

func (r *Reconciler) reconcileDelete(ctx context.Context, machineScope *scope.MachineScope) (ctrl.Result, error) {
	machineScope.Info("Reconcile KKMachine delete")

	// Set the InstanceReadyCondition and patch the object before the blocking operation
	conditions.MarkFalse(machineScope.KKMachine, infrav1.InstanceReadyCondition, clusterv1.DeletingReason, clusterv1.ConditionSeverityInfo, "")
	if err := machineScope.PatchObject(); err != nil {
		machineScope.Error(err, "failed to patch object")
		return ctrl.Result{}, err
	}

	instance, err := r.reconcileDeleteKKInstance(ctx, machineScope)
	if instance != nil || err != nil {
		return ctrl.Result{RequeueAfter: 5 * time.Second}, err
	}

	machineScope.V(4).Info("Unable to locate KubeKey instance by ID")

	conditions.MarkFalse(machineScope.KKMachine, infrav1.InstanceReadyCondition, clusterv1.DeletedReason, clusterv1.ConditionSeverityInfo, "")
	controllerutil.RemoveFinalizer(machineScope.KKMachine, infrav1.MachineFinalizer)
	return ctrl.Result{}, nil
}

func (r *Reconciler) reconcileDeleteKKInstance(ctx context.Context, machineScope *scope.MachineScope) (*infrav1.KKInstance, error) {
	// Find existing instance
	instance, err := r.findInstance(ctx, machineScope)
	if err != nil && !apierrors.IsNotFound(err) {
		machineScope.Error(err, "unable to find instance")
		return nil, err
	}

	if instance != nil {
		machineScope.V(3).Info("KubeKey instance found matching deleted KKMachine", "instance", instance.Name)
		if conditions.IsTrue(instance, infrav1.KKInstanceDeletingBootstrapCondition) {
			machineScope.V(5).Info("instance has not been cleaned up yet", "instance", instance.Name)
			return instance, errors.Errorf("kkinstance %s has not been cleaned up yet", instance.Name)
		}

		if err := r.deleteInstance(ctx, instance); err != nil {
			conditions.MarkFalse(machineScope.KKMachine, infrav1.InstanceReadyCondition, "DeletingFailed", clusterv1.ConditionSeverityWarning, err.Error())
			r.Recorder.Eventf(machineScope.KKMachine, corev1.EventTypeWarning, "FailedDelete", "Failed to delete instance %q: %v", instance.Name, err)
			return instance, err
		}
		r.Recorder.Eventf(machineScope.KKMachine, corev1.EventTypeNormal, "SuccessfulCleaned", "Clean instance %q", instance.Name)
	}
	return instance, nil
}

func (r *Reconciler) reconcileNormal(ctx context.Context, machineScope *scope.MachineScope, _ pkg.ClusterScoper,
	kkInstanceScope scope.KKInstanceScope) (ctrl.Result, error) {
	machineScope.Info("Reconcile KKMachine normal")

	// If the KKMachine is in an error state, return early.
	if machineScope.HasFailed() {
		machineScope.Info("Error state detected, skipping reconciliation")
		return ctrl.Result{}, nil
	}

	if machineScope.KKMachine.Labels == nil {
		machineScope.KKMachine.Labels = make(map[string]string)
	}

	machineScope.KKMachine.Labels[infrav1.KKClusterLabelName] = machineScope.InfraCluster.InfraClusterName()

	if !machineScope.Cluster.Status.InfrastructureReady {
		machineScope.Info("Cluster infrastructure is not ready yet")
		conditions.MarkFalse(machineScope.KKMachine, infrav1.InstanceReadyCondition, infrav1.WaitingForClusterInfrastructureReason, clusterv1.ConditionSeverityInfo, "")
		return ctrl.Result{}, nil
	}

	// Make sure bootstrap data is available and populated.
	if machineScope.Machine.Spec.Bootstrap.DataSecretName == nil {
		machineScope.Info("Bootstrap data secret reference is not yet available")
		conditions.MarkFalse(machineScope.KKMachine, infrav1.InstanceReadyCondition, infrav1.WaitingForBootstrapDataReason, clusterv1.ConditionSeverityInfo, "")
		return ctrl.Result{}, nil
	}

	// Find existing instance
	instance, err := r.findInstance(ctx, machineScope)
	if err != nil {
		machineScope.Error(err, "unable to find instance")
		conditions.MarkUnknown(machineScope.KKMachine, infrav1.InstanceReadyCondition, infrav1.InstanceNotFoundReason, err.Error())
		return ctrl.Result{}, err
	}

	// If the KKMachine doesn't have our finalizer, add it.
	controllerutil.AddFinalizer(machineScope.KKMachine, infrav1.MachineFinalizer)
	// Register the finalizer after first read operation from KK to avoid orphaning KK resources on delete
	if err := machineScope.PatchObject(); err != nil {
		machineScope.Error(err, "unable to patch object")
		return ctrl.Result{}, err
	}

	// Create new instance from KKCluster since providerId is nils.
	if instance == nil {
		// Avoid a flickering condition between InstanceBootstrapStarted and InstanceBootstrapFailed if there's a persistent failure with createInstance
		if conditions.GetReason(machineScope.KKMachine, infrav1.InstanceReadyCondition) != infrav1.InstanceBootstrapFailedReason {
			conditions.MarkFalse(machineScope.KKMachine, infrav1.InstanceReadyCondition, infrav1.InstanceBootstrapStartedReason, clusterv1.ConditionSeverityInfo, "")
			if patchErr := machineScope.PatchObject(); err != nil {
				machineScope.Error(patchErr, "failed to patch conditions")
				return ctrl.Result{}, patchErr
			}
		}

		if !r.mutex.TryLock() {
			machineScope.V(4).Info("Waiting for the last KKInstance to be created")
			return ctrl.Result{RequeueAfter: 2 * time.Second}, nil
		}
		instance, err = r.createInstance(ctx, machineScope, kkInstanceScope)
		r.mutex.Unlock()
		if err != nil {
			machineScope.Error(err, "unable to create kkInstance")
			r.Recorder.Eventf(machineScope.KKMachine, corev1.EventTypeWarning, "FailedCreate", "Failed to create kkInstance: %v", err)
			conditions.MarkFalse(machineScope.KKMachine, infrav1.InstanceReadyCondition, infrav1.InstanceBootstrapFailedReason,
				clusterv1.ConditionSeverityError, err.Error())
			return ctrl.Result{}, err
		}
	}

	// Make sure Spec.ProviderID and Spec.InstanceID are always set.
	machineScope.SetProviderID(instance.Name, machineScope.Cluster.Name)
	machineScope.SetInstanceID(instance.Name)

	existingInstanceState := machineScope.GetInstanceState()
	machineScope.SetInstanceState(instance.Status.State)

	// Proceed to reconcile the KKMachine state.
	if existingInstanceState == nil || *existingInstanceState != instance.Status.State {
		machineScope.Info("KubeKey instance state changed", "state", instance.Status.State, "instance-id", *machineScope.GetInstanceID())
	}

	switch instance.Status.State {
	case "":
		machineScope.SetNotReady()
		conditions.MarkFalse(machineScope.KKMachine, infrav1.InstanceReadyCondition, infrav1.InstanceNotReadyReason, clusterv1.ConditionSeverityWarning, "")
		return ctrl.Result{Requeue: true}, nil
	case infrav1.InstanceStatePending:
		machineScope.SetNotReady()
		conditions.MarkFalse(machineScope.KKMachine, infrav1.InstanceReadyCondition, infrav1.InstanceNotReadyReason, clusterv1.ConditionSeverityWarning, "")
		return ctrl.Result{Requeue: true}, nil
	case infrav1.InstanceStateBootstrapping:
		machineScope.SetNotReady()
		conditions.MarkFalse(machineScope.KKMachine, infrav1.InstanceReadyCondition, infrav1.InstanceNotReadyReason, clusterv1.ConditionSeverityWarning, "")
		return ctrl.Result{Requeue: true}, nil
	case infrav1.InstanceStateCleaning, infrav1.InstanceStateCleaned:
		machineScope.SetNotReady()
		conditions.MarkFalse(machineScope.KKMachine, infrav1.InstanceReadyCondition, infrav1.InstanceCleanedReason, clusterv1.ConditionSeverityWarning, "")
		return ctrl.Result{Requeue: true}, nil
	case infrav1.InstanceStateRunning:
		if err := r.setNodeProviderID(ctx, machineScope, instance); err != nil {
			machineScope.Error(err, "failed to patch the Kubernetes node with the machine providerID")
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}

		var addresses []clusterv1.MachineAddress
		privateIPAddress := clusterv1.MachineAddress{
			Type:    clusterv1.MachineInternalIP,
			Address: instance.Spec.InternalAddress,
		}
		addresses = append(addresses, privateIPAddress)
		machineScope.SetAddresses(addresses)

		machineScope.SetReady()
		conditions.MarkTrue(machineScope.KKMachine, infrav1.InstanceReadyCondition)
	case infrav1.InstanceStateInPlaceUpgrading:
		machineScope.SetNotReady()
		conditions.MarkFalse(machineScope.KKMachine, infrav1.InstanceReadyCondition, infrav1.InstanceInPlaceUpgradingReason, clusterv1.ConditionSeverityWarning, "")
		return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
	default:
		machineScope.SetNotReady()
		machineScope.Info("KubeKey instance state is undefined", "state", instance.Status.State, "instance-id", *machineScope.GetInstanceID())
		r.Recorder.Eventf(machineScope.KKMachine, corev1.EventTypeWarning, "InstanceUnhandledState", "KubeKey instance state is undefined")
		machineScope.SetFailureReason(capierrors.UpdateMachineError)
		machineScope.SetFailureMessage(errors.Errorf("KubeKey instance state %q is undefined", instance.Status.State))
		conditions.MarkUnknown(machineScope.KKMachine, infrav1.InstanceReadyCondition, "", "")
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) findInstance(ctx context.Context, machineScope *scope.MachineScope) (*infrav1.KKInstance, error) {
	machineScope.V(4).Info("Find KubeKey instance")

	kkInstance := &infrav1.KKInstance{}

	// Parse the ProviderID.
	pid, err := noderefutil.NewProviderID(machineScope.GetProviderID())
	if err != nil {
		if errors.Is(err, noderefutil.ErrEmptyProviderID) {
			machineScope.Info("KKMachine does not have an instance id")
			return nil, nil
		}
		return nil, errors.Wrapf(err, "failed to parse Spec.ProviderID")
	}

	machineScope.V(4).Info("KKMachine has an instance id", "instance-id", pid.ID())
	// If the ProviderID is populated, describe the instance using the ID.
	id := pointer.String(pid.ID())

	obj := client.ObjectKey{
		Namespace: machineScope.KKMachine.Namespace,
		Name:      *id,
	}
	if err := r.Client.Get(ctx, obj, kkInstance); err != nil {
		return nil, err
	}
	// The only case where the instance is nil here is when the providerId is empty and instance could not be found by tags.
	return kkInstance, nil
}

func (r *Reconciler) indexKKMachineByInstanceID(o client.Object) []string {
	kkMachine, ok := o.(*infrav1.KKMachine)
	if !ok {
		return nil
	}

	if kkMachine.Spec.InstanceID != nil {
		return []string{*kkMachine.Spec.InstanceID}
	}

	return nil
}

// KKClusterToKKMachines is a handler.ToRequestsFunc to be used to enqeue requests for reconciliation
// of KKMachines.
func (r *Reconciler) KKClusterToKKMachines(log logr.Logger) handler.MapFunc {
	return func(o client.Object) []ctrl.Request {
		c, ok := o.(*infrav1.KKCluster)
		if !ok {
			panic(fmt.Sprintf("Expected a KKCluster but got a %T", o))
		}

		log := log.WithValues("objectMapper", "kkClusterToKKMachine", "namespace", c.Namespace, "kkCluster", c.Name)

		// Don't handle deleted KKClusters
		if !c.ObjectMeta.DeletionTimestamp.IsZero() {
			log.V(4).Info("KKCluster has a deletion timestamp, skipping mapping.")
			return nil
		}

		cluster, err := cutil.GetOwnerCluster(context.TODO(), r.Client, c.ObjectMeta)
		switch {
		case apierrors.IsNotFound(err) || cluster == nil:
			log.V(4).Info("Cluster for KKCluster not found, skipping mapping.")
			return nil
		case err != nil:
			log.Error(err, "Failed to get owning cluster, skipping mapping.")
			return nil
		}

		return r.requestsForCluster(log, cluster.Namespace, cluster.Name)
	}
}

func (r *Reconciler) requeueKKMachinesForUnpausedCluster(log logr.Logger) handler.MapFunc {
	return func(o client.Object) []ctrl.Request {
		c, ok := o.(*clusterv1.Cluster)
		if !ok {
			panic(fmt.Sprintf("Expected a Cluster but got a %T", o))
		}

		log := log.WithValues("objectMapper", "clusterToKKMachine", "namespace", c.Namespace, "cluster", c.Name)

		// Don't handle deleted clusters
		if !c.ObjectMeta.DeletionTimestamp.IsZero() {
			log.V(4).Info("Cluster has a deletion timestamp, skipping mapping.")
			return nil
		}

		return r.requestsForCluster(log, c.Namespace, c.Name)
	}
}

func (r *Reconciler) requestsForCluster(log logr.Logger, namespace, name string) []ctrl.Request {
	labels := map[string]string{clusterv1.ClusterLabelName: name}
	machineList := &clusterv1.MachineList{}
	if err := r.Client.List(context.TODO(), machineList, client.InNamespace(namespace), client.MatchingLabels(labels)); err != nil {
		log.Error(err, "Failed to get owned Machines, skipping mapping.")
		return nil
	}

	result := make([]ctrl.Request, 0, len(machineList.Items))
	for _, m := range machineList.Items {
		log.WithValues("machine", m.Name)
		if m.Spec.InfrastructureRef.GroupVersionKind().Kind != "KKMachine" {
			log.V(4).Info("Machine has an InfrastructureRef for a different type, will not add to reconciliation request.")
			continue
		}
		if m.Spec.InfrastructureRef.Name == "" {
			log.V(4).Info("Machine has an InfrastructureRef with an empty name, will not add to reconciliation request.")
			continue
		}
		log.WithValues("kkMachine", m.Spec.InfrastructureRef.Name)
		log.V(4).Info("Adding KKMachine to reconciliation request.")
		result = append(result, ctrl.Request{NamespacedName: client.ObjectKey{Namespace: m.Namespace, Name: m.Spec.InfrastructureRef.Name}})
	}
	return result
}
