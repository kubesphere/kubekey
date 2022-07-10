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

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	cutil "sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/annotations"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/cluster-api/util/predicates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	infrav1 "github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/api/v1beta1"
	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/clients/ssh"
	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/scope"
	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/service"
	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/service/bootstrap"
	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/util"
)

const (
	defaultRequeueWait = 30 * time.Second
)

// KKInstanceReconciler reconciles a KKInstance object
type KKInstanceReconciler struct {
	client.Client
	Scheme           *runtime.Scheme
	Recorder         record.EventRecorder
	bootstrapFactory func(sshClient ssh.Interface, scope scope.LBScope) service.Bootstrap
	WatchFilterValue string

	Dialer *ssh.Dialer
}

func (r *KKInstanceReconciler) getSSHClient(scope *scope.InstanceScope) (ssh.Interface, error) {
	return r.Dialer.Connect(scope.KKInstance.Spec.Address, &scope.KKInstance.Spec.Auth)
}

func (r *KKInstanceReconciler) getBootstrapService(sshClient ssh.Interface, scope scope.LBScope) service.Bootstrap {
	if r.bootstrapFactory != nil {
		return r.bootstrapFactory(sshClient, scope)
	}
	return bootstrap.NewService(sshClient, scope)
}

// SetupWithManager sets up the controller with the Manager.
func (r *KKInstanceReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	log := ctrl.LoggerFrom(ctx)
	kkClusterToKKInstances := r.KKClusterToKKInstances(log)

	c, err := ctrl.NewControllerManagedBy(mgr).
		WithOptions(options).
		For(&infrav1.KKInstance{}).
		Watches(
			&source.Kind{Type: &infrav1.KKMachine{}},
			handler.EnqueueRequestsFromMapFunc(r.KKMachineToKKInstanceMapFunc()),
		).
		Watches(
			&source.Kind{Type: &infrav1.KKCluster{}},
			handler.EnqueueRequestsFromMapFunc(kkClusterToKKInstances),
		).
		WithEventFilter(predicates.ResourceNotPausedAndHasFilterLabel(log, r.WatchFilterValue)).
		WithEventFilter(
			predicate.Funcs{
				// Avoid reconciling if the event triggering the reconciliation is related to incremental status updates
				// for KKInstance resources only
				UpdateFunc: func(e event.UpdateEvent) bool {
					if e.ObjectOld.GetObjectKind().GroupVersionKind().Kind != "KKInstance" {
						return true
					}

					oldInstance := e.ObjectOld.(*infrav1.KKInstance).DeepCopy()
					newInstance := e.ObjectNew.(*infrav1.KKInstance).DeepCopy()

					oldInstance.Status = infrav1.KKInstanceStatus{}
					newInstance.Status = infrav1.KKInstanceStatus{}

					oldInstance.ObjectMeta.ResourceVersion = ""
					newInstance.ObjectMeta.ResourceVersion = ""

					return !cmp.Equal(oldInstance, newInstance)
				},
			},
		).
		Build(r)
	if err != nil {
		return err
	}

	requeueKKInstancesForUnpausedCluster := r.requeueKKInstancesForUnpausedCluster(log)
	err = c.Watch(
		&source.Kind{Type: &clusterv1.Cluster{}},
		handler.EnqueueRequestsFromMapFunc(requeueKKInstancesForUnpausedCluster),
		predicates.ClusterUnpausedAndInfrastructureReady(log),
	)
	if err != nil {
		return err
	}

	if r.Dialer == nil {
		r.Dialer = ssh.NewDialer()
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

	infraCluster, err := util.GetInfraCluster(ctx, r.Client, cluster, kkMachine, "kkinstance")
	if err != nil {
		return ctrl.Result{}, errors.Wrapf(err, "error getting infra provider cluster object")
	}
	if infraCluster == nil {
		log.Info("KKCluster is not ready yet")
		return ctrl.Result{}, nil
	}

	if err := r.Dialer.Ping(kkInstance.Spec.Address, &kkInstance.Spec.Auth, 3); err != nil {
		return ctrl.Result{}, errors.Wrapf(err, "failed to ping remote instance [%s]", kkInstance.Spec.Address)
	}
	defer r.Dialer.Close(kkInstance.Spec.Address)

	instanceScope, err := scope.NewInstanceScope(scope.InstanceScopeParams{
		Client:       r.Client,
		Cluster:      cluster,
		InfraCluster: infraCluster,
		KKMachine:    kkMachine,
		KKInstance:   kkInstance,
	})
	if err != nil {
		log.Error(err, "failed to create instance scope")
		return ctrl.Result{}, err
	}

	// Always close the scope when exiting this function, so we can persist any KKInstance changes.
	defer func() {
		if err := instanceScope.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()

	if !kkInstance.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, instanceScope)
	}

	return r.reconcileNormal(ctx, instanceScope, infraCluster)
}

func (r *KKInstanceReconciler) reconcileDelete(ctx context.Context, instanceScope *scope.InstanceScope) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	log.V(4).Info("Reconcile KKInstance delete")

	return ctrl.Result{}, nil
}

func (r *KKInstanceReconciler) reconcileNormal(ctx context.Context, instanceScope *scope.InstanceScope, lbScope scope.LBScope) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	log.V(4).Info("Reconcile KKInstance normal")

	// If the KKInstance is in an error state, return early.
	if instanceScope.HasFailed() {
		log.Info("Error state detected, skipping reconciliation")
		return ctrl.Result{}, nil
	}

	if !instanceScope.Cluster.Status.InfrastructureReady {
		log.Info("Cluster infrastructure is not ready yet")
		conditions.MarkFalse(instanceScope.KKMachine, infrav1.InstanceReadyCondition, infrav1.WaitingForClusterInfrastructureReason, clusterv1.ConditionSeverityInfo, "")
		return ctrl.Result{}, nil
	}

	sshClient, err := r.getSSHClient(instanceScope)
	if err != nil {
		log.Info("failed to get remote ssh client", "instance", instanceScope.KKInstance.Name)
		return ctrl.Result{}, errors.Wrapf(err, "failed to get remote instance [%s] ssh client", instanceScope.InternalAddress())
	}

	if err := r.reconcileBootstrap(ctx, sshClient, instanceScope, lbScope); err != nil {
		log.Error(err, "failed to reconcile bootstrap")
		conditions.MarkFalse(instanceScope.KKInstance, infrav1.InstanceReadyCondition, infrav1.InstanceBootstrapFailedReason, clusterv1.ConditionSeverityWarning, "")
		return ctrl.Result{RequeueAfter: defaultRequeueWait}, err
	}
	instanceScope.SetState(infrav1.InstanceStateRunning)

	// parse cloud-init file
	// todo:

	// init os environment
	// todo:

	// download binaries
	// todo:

	// install container runtime
	// todo:

	return ctrl.Result{}, nil
}

// KKClusterToKKInstances is a handler.ToRequestsFunc to be used to enqeue requests for reconciliation of KKInstance.
func (r *KKInstanceReconciler) KKClusterToKKInstances(log logr.Logger) handler.MapFunc {
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

func (r *KKInstanceReconciler) KKMachineToKKInstanceMapFunc() handler.MapFunc {
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
