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

package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/record"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	cutil "sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/annotations"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/cluster-api/util/predicates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	infrav1 "github.com/kubesphere/kubekey/api/v1beta1"
	"github.com/kubesphere/kubekey/pkg/clients/ssh"
	"github.com/kubesphere/kubekey/pkg/scope"
	"github.com/kubesphere/kubekey/pkg/service"
	"github.com/kubesphere/kubekey/pkg/service/binary"
	"github.com/kubesphere/kubekey/pkg/service/bootstrap"
	"github.com/kubesphere/kubekey/pkg/service/containermanager"
	"github.com/kubesphere/kubekey/pkg/service/provisioning"
	"github.com/kubesphere/kubekey/pkg/service/repository"
	"github.com/kubesphere/kubekey/pkg/util"
)

const (
	defaultRequeueWait        = 30 * time.Second
	defaultKKInstanceInterval = 5 * time.Second
	defaultKKInstanceTimeout  = 10 * time.Minute
)

// KKInstanceReconciler reconciles a KKInstance object
type KKInstanceReconciler struct {
	client.Client
	Scheme                  *runtime.Scheme
	Recorder                record.EventRecorder
	sshClientFactory        func(scope *scope.InstanceScope) ssh.Interface
	bootstrapFactory        func(sshClient ssh.Interface, scope scope.LBScope, instanceScope *scope.InstanceScope) service.Bootstrap
	repositoryFactory       func(sshClient ssh.Interface, scope scope.KKInstanceScope, instanceScope *scope.InstanceScope) service.Repository
	binaryFactory           func(sshClient ssh.Interface, scope scope.KKInstanceScope, instanceScope *scope.InstanceScope) service.BinaryService
	containerManagerFactory func(sshClient ssh.Interface, scope scope.KKInstanceScope, instanceScope *scope.InstanceScope) service.ContainerManager
	provisioningFactory     func(sshClient ssh.Interface, format bootstrapv1.Format) service.Provisioning
	WatchFilterValue        string
	DataDir                 string

	WaitKKInstanceInterval time.Duration
	WaitKKInstanceTimeout  time.Duration
}

func (r *KKInstanceReconciler) getSSHClient(scope *scope.InstanceScope) ssh.Interface {
	if r.sshClientFactory != nil {
		return r.sshClientFactory(scope)
	}
	return ssh.NewClient(scope.KKInstance.Spec.Address, scope.KKInstance.Spec.Auth, &scope.Logger)
}

func (r *KKInstanceReconciler) getBootstrapService(sshClient ssh.Interface, scope scope.LBScope, instanceScope *scope.InstanceScope) service.Bootstrap {
	if r.bootstrapFactory != nil {
		return r.bootstrapFactory(sshClient, scope, instanceScope)
	}
	return bootstrap.NewService(sshClient, scope, instanceScope)
}

func (r *KKInstanceReconciler) getRepositoryService(sshClient ssh.Interface, scope scope.KKInstanceScope, instanceScope *scope.InstanceScope) service.Repository {
	if r.repositoryFactory != nil {
		return r.repositoryFactory(sshClient, scope, instanceScope)
	}
	return repository.NewService(sshClient, scope, instanceScope)
}

func (r *KKInstanceReconciler) getBinaryService(sshClient ssh.Interface, scope scope.KKInstanceScope, instanceScope *scope.InstanceScope) service.BinaryService {
	if r.binaryFactory != nil {
		return r.binaryFactory(sshClient, scope, instanceScope)
	}
	return binary.NewService(sshClient, scope, instanceScope)
}

func (r *KKInstanceReconciler) getContainerManager(sshClient ssh.Interface, scope scope.KKInstanceScope, instanceScope *scope.InstanceScope) service.ContainerManager {
	if r.containerManagerFactory != nil {
		return r.containerManagerFactory(sshClient, scope, instanceScope)
	}
	return containermanager.NewService(sshClient, scope, instanceScope)
}

func (r *KKInstanceReconciler) getProvisioningService(sshClient ssh.Interface, format bootstrapv1.Format) service.Provisioning {
	if r.provisioningFactory != nil {
		return r.provisioningFactory(sshClient, format)
	}
	return provisioning.NewService(sshClient, format)
}

// SetupWithManager sets up the controller with the Manager.
func (r *KKInstanceReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	log := ctrl.LoggerFrom(ctx)

	if r.WaitKKInstanceInterval.Nanoseconds() == 0 {
		r.WaitKKInstanceInterval = defaultKKInstanceInterval
	}
	if r.WaitKKInstanceTimeout.Nanoseconds() == 0 {
		r.WaitKKInstanceTimeout = defaultKKInstanceTimeout
	}

	c, err := ctrl.NewControllerManagedBy(mgr).
		WithOptions(options).
		For(&infrav1.KKInstance{}).
		Watches(
			&source.Kind{Type: &infrav1.KKMachine{}},
			handler.EnqueueRequestsFromMapFunc(r.KKMachineToKKInstanceMapFunc(log)),
		).
		Watches(
			&source.Kind{Type: &infrav1.KKCluster{}},
			handler.EnqueueRequestsFromMapFunc(r.KKClusterToKKInstances(log)),
		).
		WithEventFilter(predicates.ResourceNotPausedAndHasFilterLabel(log, r.WatchFilterValue)).
		WithEventFilter(
			predicate.Funcs{
				// Avoid reconciling if the event triggering the reconciliation is related to incremental status updates
				// for KKInstance resources only
				UpdateFunc: func(e event.UpdateEvent) bool {
					log.V(5).Info("KKInstance controller update predicate")
					if _, ok := e.ObjectOld.(*infrav1.KKInstance); !ok {
						log.V(5).Info(fmt.Sprintf("gvk is %s, not equale KKInstance", e.ObjectOld.GetObjectKind().GroupVersionKind().Kind))
						return true
					}

					oldInstance := e.ObjectOld.(*infrav1.KKInstance).DeepCopy()
					newInstance := e.ObjectNew.(*infrav1.KKInstance).DeepCopy()

					oldInstance.Status = infrav1.KKInstanceStatus{}
					newInstance.Status = infrav1.KKInstanceStatus{}

					oldInstance.ObjectMeta.ResourceVersion = ""
					newInstance.ObjectMeta.ResourceVersion = ""

					if cmp.Equal(oldInstance, newInstance) {
						log.V(4).Info("oldInstance and newInstance are equaled, skip")
						return false
					}
					log.V(4).Info("oldInstance and newInstance are not equaled, allowing further processing")
					return true
				},
			},
		).
		Build(r)
	if err != nil {
		return err
	}

	err = c.Watch(
		&source.Kind{Type: &clusterv1.Cluster{}},
		handler.EnqueueRequestsFromMapFunc(r.requeueKKInstancesForUnpausedCluster(log)),
		predicates.ClusterUnpausedAndInfrastructureReady(log),
	)
	if err != nil {
		return err
	}

	return nil
}

//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=kkinstances,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=kkinstances/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=kkinstances/finalizers,verbs=update

func (r *KKInstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, retErr error) {
	log := ctrl.LoggerFrom(ctx)
	// Fetch the KKInstance.
	kkInstance := &infrav1.KKInstance{}
	err := r.Get(ctx, req.NamespacedName, kkInstance)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Fetch the KKMachine.
	kkMachine, err := util.GetOwnerKKMachine(ctx, r.Client, kkInstance.ObjectMeta)
	if err != nil {
		return ctrl.Result{}, err
	}
	if kkMachine == nil {
		log.Info("KKMachine Controller has not yet set OwnerRef")
		return ctrl.Result{}, nil
	}

	log = log.WithValues("kkMachine", kkMachine.Name)

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
	cluster, err := cutil.GetClusterFromMetadata(ctx, r.Client, kkInstance.ObjectMeta)
	if err != nil {
		log.Info("KKInstance is missing cluster label or cluster does not exist")
		return ctrl.Result{}, nil
	}

	if annotations.IsPaused(cluster, kkInstance) {
		log.Info("KKInstance or linked Cluster is marked as paused. Won't reconcile")
		return ctrl.Result{}, nil
	}

	log = log.WithValues("cluster", cluster.Name)

	infraCluster, err := util.GetInfraCluster(ctx, r.Client, log, cluster, "kkinstance", r.DataDir)
	if err != nil {
		return ctrl.Result{}, errors.Wrapf(err, "error getting infra provider cluster object")
	}
	if infraCluster == nil {
		log.Info("KKCluster is not ready yet")
		return ctrl.Result{}, nil
	}

	instanceScope, err := scope.NewInstanceScope(scope.InstanceScopeParams{
		Client:       r.Client,
		Cluster:      cluster,
		Machine:      machine,
		InfraCluster: infraCluster,
		KKInstance:   kkInstance,
	})
	if err != nil {
		log.Error(err, "failed to create instance scope")
		return ctrl.Result{}, err
	}

	if err := r.reconcilePing(ctx, instanceScope); err != nil {
		return ctrl.Result{}, errors.Wrapf(err, "failed to ping remote instance [%s]", kkInstance.Spec.Address)
	}

	// Always close the scope when exiting this function, so we can persist any KKInstance changes.
	defer func() {
		if err := instanceScope.Close(); err != nil && retErr == nil {
			log.Error(err, "failed to patch object")
			retErr = err
		}
	}()

	if !kkInstance.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, instanceScope, infraCluster)
	}

	return r.reconcileNormal(ctx, instanceScope, infraCluster, infraCluster)
}

func (r *KKInstanceReconciler) reconcileDelete(ctx context.Context, instanceScope *scope.InstanceScope, lbScope scope.LBScope) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	log.V(4).Info("Reconcile KKInstance delete")

	if conditions.Get(instanceScope.KKInstance, infrav1.KKInstanceDeletingBootstrapCondition) == nil {
		conditions.MarkFalse(instanceScope.KKInstance, infrav1.KKInstanceDeletingBootstrapCondition,
			infrav1.CleaningReason, clusterv1.ConditionSeverityInfo, "Cleaning the node before deletion")
	}

	if err := instanceScope.PatchObject(); err != nil {
		instanceScope.Error(err, "unable to patch object")
		return ctrl.Result{}, err
	}

	sshClient := r.getSSHClient(instanceScope)
	if err := r.reconcileDeletingBootstrap(ctx, sshClient, instanceScope, lbScope); err != nil {
		instanceScope.Error(err, "failed to reconcile deleting bootstrap")
		return ctrl.Result{}, nil
	}
	conditions.MarkTrue(instanceScope.KKInstance, infrav1.KKInstanceDeletingBootstrapCondition)
	instanceScope.SetState(infrav1.InstanceStateCleaned)
	instanceScope.Info("Reconcile KKInstance delete successful")
	controllerutil.RemoveFinalizer(instanceScope.KKInstance, infrav1.InstanceFinalizer)
	return ctrl.Result{}, nil
}

func (r *KKInstanceReconciler) reconcileNormal(ctx context.Context, instanceScope *scope.InstanceScope, lbScope scope.LBScope, kkInstanceScope scope.KKInstanceScope) (ctrl.Result, error) {
	instanceScope.Info("Reconcile KKInstance normal")

	// If the KKInstance is in an error state, return early.
	if instanceScope.HasFailed() {
		instanceScope.Info("Error state detected, skipping reconciliation")
		return ctrl.Result{}, nil
	}

	// If the KKMachine doesn't have our finalizer, add it.
	controllerutil.AddFinalizer(instanceScope.KKInstance, infrav1.InstanceFinalizer)
	// Register the finalizer after first read operation from KK to avoid orphaning KK resources on delete
	if err := instanceScope.PatchObject(); err != nil {
		instanceScope.Error(err, "unable to patch object")
		return ctrl.Result{}, err
	}

	sshClient := r.getSSHClient(instanceScope)

	phases := []func(context.Context, ssh.Interface, *scope.InstanceScope, scope.KKInstanceScope, scope.LBScope) error{
		r.reconcileBootstrap,
		r.reconcileRepository,
		r.reconcileBinaryService,
		r.reconcileContainerManager,
		r.reconcileProvisioning,
	}

	for _, phase := range phases {
		pollErr := wait.PollImmediate(r.WaitKKInstanceInterval, r.WaitKKInstanceTimeout, func() (done bool, err error) {
			if err := phase(ctx, sshClient, instanceScope, kkInstanceScope, lbScope); err != nil {
				return false, err
			}
			return true, nil
		})
		if pollErr != nil {
			instanceScope.Error(pollErr, "failed to reconcile phase")
			return ctrl.Result{RequeueAfter: defaultRequeueWait}, pollErr
		}
	}

	instanceScope.SetState(infrav1.InstanceStateRunning)
	instanceScope.Info("Reconcile KKInstance normal successful")
	return ctrl.Result{}, nil
}

// KKClusterToKKInstances is a handler.ToRequestsFunc to be used to enqeue requests for reconciliation of KKInstance.
func (r *KKInstanceReconciler) KKClusterToKKInstances(log logr.Logger) handler.MapFunc {
	log.V(4).Info("KKClusterToKKInstances")
	return func(o client.Object) []ctrl.Request {
		c, ok := o.(*infrav1.KKCluster)
		if !ok {
			panic(fmt.Sprintf("Expected a KKCluster but got a %T", o))
		}

		log := log.WithValues("objectMapper", "kkClusterToKKInstance", "namespace", c.Namespace, "kkCluster", c.Name)

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

func (r *KKInstanceReconciler) requeueKKInstancesForUnpausedCluster(log logr.Logger) handler.MapFunc {
	log.V(4).Info("requeueKKInstancesForUnpausedCluster")
	return func(o client.Object) []ctrl.Request {
		c, ok := o.(*clusterv1.Cluster)
		if !ok {
			panic(fmt.Sprintf("Expected a Cluster but got a %T", o))
		}

		log := log.WithValues("objectMapper", "clusterToKKInstance", "namespace", c.Namespace, "cluster", c.Name)

		// Don't handle deleted clusters
		if !c.ObjectMeta.DeletionTimestamp.IsZero() {
			log.V(4).Info("Cluster has a deletion timestamp, skipping mapping.")
			return nil
		}

		return r.requestsForCluster(log, c.Namespace, c.Name)
	}
}

func (r *KKInstanceReconciler) requestsForCluster(log logr.Logger, namespace, name string) []ctrl.Request {
	labels := map[string]string{clusterv1.ClusterLabelName: name}
	kkMachineList := &infrav1.KKMachineList{}
	if err := r.Client.List(context.TODO(), kkMachineList, client.InNamespace(namespace), client.MatchingLabels(labels)); err != nil {
		log.Error(err, "Failed to get owned Machines, skipping mapping.")
		return nil
	}

	result := make([]ctrl.Request, 0, len(kkMachineList.Items))
	for _, m := range kkMachineList.Items {
		log.WithValues("kkmachine", m.Name)

		if m.Spec.InstanceID == nil {
			log.V(4).Info("KKMachine does not have a providerID, will not add to reconciliation request.")
			continue
		}

		log.WithValues("kkInstance", m.Spec.InstanceID)
		log.V(4).Info("Adding KKInstance to reconciliation request.")
		result = append(result, ctrl.Request{NamespacedName: client.ObjectKey{Namespace: m.Namespace, Name: *m.Spec.InstanceID}})
	}
	return result
}

// KKMachineToKKInstanceMapFunc returns a handler.ToRequestsFunc that watches for
// KKMachine events and returns reconciliation requests for an KKInstance object.
func (r *KKInstanceReconciler) KKMachineToKKInstanceMapFunc(log logr.Logger) handler.MapFunc {
	log.V(4).Info("KKMachineToKKInstanceMapFunc")
	return func(o client.Object) []reconcile.Request {
		m, ok := o.(*infrav1.KKMachine)
		if !ok {
			return nil
		}

		if m.Spec.InstanceID == nil {
			return nil
		}

		return []reconcile.Request{
			{
				NamespacedName: client.ObjectKey{
					Namespace: m.Namespace,
					Name:      *m.Spec.InstanceID,
				},
			},
		}
	}
}
