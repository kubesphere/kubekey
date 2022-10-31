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
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	bootstraputil "k8s.io/cluster-bootstrap/token/util"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	bsutil "sigs.k8s.io/cluster-api/bootstrap/util"
	expv1 "sigs.k8s.io/cluster-api/exp/api/v1beta1"
	"sigs.k8s.io/cluster-api/feature"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/annotations"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/cluster-api/util/predicates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	infrabootstrapv1 "github.com/kubesphere/kubekey/bootstrap/k3s/api/v1beta1"
	"github.com/kubesphere/kubekey/bootstrap/k3s/pkg/cloudinit"
	"github.com/kubesphere/kubekey/bootstrap/k3s/pkg/locking"
	k3stypes "github.com/kubesphere/kubekey/bootstrap/k3s/pkg/types"
	kklog "github.com/kubesphere/kubekey/util/log"
	"github.com/kubesphere/kubekey/util/secret"
)

// InitLocker is a lock that is used around kubeadm init.
type InitLocker interface {
	Lock(ctx context.Context, cluster *clusterv1.Cluster, machine *clusterv1.Machine) bool
	Unlock(ctx context.Context, cluster *clusterv1.Cluster) bool
}

// +kubebuilder:rbac:groups=bootstrap.cluster.x-k8s.io,resources=k3sconfigs;k3sconfigs/status;k3sconfigs/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;clusters/status;machinesets;machines;machines/status;machinepools;machinepools/status,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=secrets;events;configmaps,verbs=get;list;watch;create;update;patch;delete

// K3sConfigReconciler reconciles a K3sConfig object
type K3sConfigReconciler struct {
	client.Client
	K3sInitLock InitLocker

	// WatchFilterValue is the label value used to filter events prior to reconciliation.
	WatchFilterValue string
}

// Scope is a scoped struct used during reconciliation.
type Scope struct {
	logr.Logger
	Config      *infrabootstrapv1.K3sConfig
	ConfigOwner *bsutil.ConfigOwner
	Cluster     *clusterv1.Cluster
}

// SetupWithManager sets up the controller with the Manager.
func (r *K3sConfigReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	if r.K3sInitLock == nil {
		r.K3sInitLock = locking.NewControlPlaneInitMutex(mgr.GetClient())
	}

	b := ctrl.NewControllerManagedBy(mgr).
		For(&infrabootstrapv1.K3sConfig{}).
		WithOptions(options).
		Watches(
			&source.Kind{Type: &clusterv1.Machine{}},
			handler.EnqueueRequestsFromMapFunc(r.MachineToBootstrapMapFunc),
		).WithEventFilter(predicates.ResourceNotPausedAndHasFilterLabel(ctrl.LoggerFrom(ctx), r.WatchFilterValue))

	if feature.Gates.Enabled(feature.MachinePool) {
		b = b.Watches(
			&source.Kind{Type: &expv1.MachinePool{}},
			handler.EnqueueRequestsFromMapFunc(r.MachinePoolToBootstrapMapFunc),
		).WithEventFilter(predicates.ResourceNotPausedAndHasFilterLabel(ctrl.LoggerFrom(ctx), r.WatchFilterValue))
	}

	c, err := b.Build(r)
	if err != nil {
		return errors.Wrap(err, "failed setting up with a controller manager")
	}

	err = c.Watch(
		&source.Kind{Type: &clusterv1.Cluster{}},
		handler.EnqueueRequestsFromMapFunc(r.ClusterToK3sConfigs),
		predicates.All(ctrl.LoggerFrom(ctx),
			predicates.ClusterUnpausedAndInfrastructureReady(ctrl.LoggerFrom(ctx)),
			predicates.ResourceHasFilterLabel(ctrl.LoggerFrom(ctx), r.WatchFilterValue),
		),
	)
	if err != nil {
		return errors.Wrap(err, "failed adding Watch for Clusters to controller manager")
	}

	return nil
}

// Reconcile handles K3sConfig events.
func (r *K3sConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, retErr error) {
	log := ctrl.LoggerFrom(ctx)

	// Lookup the kubeadm config
	config := &infrabootstrapv1.K3sConfig{}
	if err := r.Client.Get(ctx, req.NamespacedName, config); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get config")
		return ctrl.Result{}, err
	}

	// AddOwners adds the owners of K3sConfig as k/v pairs to the logger.
	// Specifically, it will add K3sControlPlane, MachineSet and MachineDeployment.
	ctx, log, err := kklog.AddOwners(ctx, r.Client, config)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Look up the owner of this k3s config if there is one
	configOwner, err := bsutil.GetConfigOwner(ctx, r.Client, config)
	if apierrors.IsNotFound(err) {
		// Could not find the owner yet, this is not an error and will rereconcile when the owner gets set.
		return ctrl.Result{}, nil
	}
	if err != nil {
		log.Error(err, "Failed to get owner")
		return ctrl.Result{}, err
	}
	if configOwner == nil {
		return ctrl.Result{}, nil
	}
	log = log.WithValues(configOwner.GetKind(), klog.KRef(configOwner.GetNamespace(), configOwner.GetName()), "resourceVersion", configOwner.GetResourceVersion())

	log = log.WithValues("Cluster", klog.KRef(configOwner.GetNamespace(), configOwner.ClusterName()))
	ctx = ctrl.LoggerInto(ctx, log)

	// Lookup the cluster the config owner is associated with
	cluster, err := util.GetClusterByName(ctx, r.Client, configOwner.GetNamespace(), configOwner.ClusterName())
	if err != nil {
		if errors.Cause(err) == util.ErrNoCluster {
			log.Info(fmt.Sprintf("%s does not belong to a cluster yet, waiting until it's part of a cluster", configOwner.GetKind()))
			return ctrl.Result{}, nil
		}

		if apierrors.IsNotFound(err) {
			log.Info("Cluster does not exist yet, waiting until it is created")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Could not get cluster with metadata")
		return ctrl.Result{}, err
	}

	if annotations.IsPaused(cluster, config) {
		log.Info("Reconciliation is paused for this object")
		return ctrl.Result{}, nil
	}

	scope := &Scope{
		Logger:      log,
		Config:      config,
		ConfigOwner: configOwner,
		Cluster:     cluster,
	}

	// Initialize the patch helper.
	patchHelper, err := patch.NewHelper(config, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Attempt to Patch the K3sConfig object and status after each reconciliation if no error occurs.
	defer func() {
		// always update the readyCondition; the summary is represented using the "1 of x completed" notation.
		conditions.SetSummary(config,
			conditions.WithConditions(
				bootstrapv1.DataSecretAvailableCondition,
				bootstrapv1.CertificatesAvailableCondition,
			),
		)
		// Patch ObservedGeneration only if the reconciliation completed successfully
		var patchOpts []patch.Option
		if retErr == nil {
			patchOpts = append(patchOpts, patch.WithStatusObservedGeneration{})
		}
		if err := patchHelper.Patch(ctx, config, patchOpts...); err != nil {
			log.Error(retErr, "Failed to patch config")
			if retErr == nil {
				retErr = err
			}
		}
	}()

	switch {
	// Wait for the infrastructure to be ready.
	case !cluster.Status.InfrastructureReady:
		log.Info("Cluster infrastructure is not ready, waiting")
		conditions.MarkFalse(config, bootstrapv1.DataSecretAvailableCondition, bootstrapv1.WaitingForClusterInfrastructureReason, clusterv1.ConditionSeverityInfo, "")
		return ctrl.Result{}, nil
	// Reconcile status for machines that already have a secret reference, but our status isn't up-to-date.
	// This case solves the pivoting scenario (or a backup restore) which doesn't preserve the status subresource on objects.
	case configOwner.DataSecretName() != nil && (!config.Status.Ready || config.Status.DataSecretName == nil):
		config.Status.Ready = true
		config.Status.DataSecretName = configOwner.DataSecretName()
		conditions.MarkTrue(config, bootstrapv1.DataSecretAvailableCondition)
		return ctrl.Result{}, nil
	// Status is ready means a config has been generated.
	case config.Status.Ready:
		return ctrl.Result{}, nil
	}

	// Note: can't use IsFalse here because we need to handle the absence of the condition as well as false.
	if !conditions.IsTrue(cluster, clusterv1.ControlPlaneInitializedCondition) {
		return r.handleClusterNotInitialized(ctx, scope)
	}

	// Every other case it's a join scenario
	// Nb. in this case ClusterConfiguration and InitConfiguration should not be defined by users, but in case of misconfigurations, CABPK3s simply ignore them

	// Unlock any locks that might have been set during init process
	r.K3sInitLock.Unlock(ctx, cluster)

	// if the .spec.cluster is missing, create a default one
	if config.Spec.Cluster == nil {
		log.Info("Creating default .spec.cluster")
		config.Spec.Cluster = &infrabootstrapv1.Cluster{}
	}

	// it's a control plane join
	if configOwner.IsControlPlaneMachine() {
		return r.joinControlplane(ctx, scope)
	}

	// It's a worker join
	return r.joinWorker(ctx, scope)
}

func (r *K3sConfigReconciler) handleClusterNotInitialized(ctx context.Context, scope *Scope) (_ ctrl.Result, retErr error) {
	// initialize the DataSecretAvailableCondition if missing.
	// this is required in order to avoid the condition's LastTransitionTime to flicker in case of errors surfacing
	// using the DataSecretGeneratedFailedReason
	if conditions.GetReason(scope.Config, bootstrapv1.DataSecretAvailableCondition) != bootstrapv1.DataSecretGenerationFailedReason {
		conditions.MarkFalse(scope.Config, bootstrapv1.DataSecretAvailableCondition, clusterv1.WaitingForControlPlaneAvailableReason, clusterv1.ConditionSeverityInfo, "")
	}

	// if it's NOT a control plane machine, requeue
	if !scope.ConfigOwner.IsControlPlaneMachine() {
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	// if the machine has not ClusterConfiguration and InitConfiguration, requeue
	if scope.Config.Spec.ServerConfiguration == nil && scope.Config.Spec.Cluster == nil {
		scope.Info("Control plane is not ready, requeing joining control planes until ready.")
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	machine := &clusterv1.Machine{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(scope.ConfigOwner.Object, machine); err != nil {
		return ctrl.Result{}, errors.Wrapf(err, "cannot convert %s to Machine", scope.ConfigOwner.GetKind())
	}

	// acquire the init lock so that only the first machine configured
	// as control plane get processed here
	// if not the first, requeue
	if !r.K3sInitLock.Lock(ctx, scope.Cluster, machine) {
		scope.Info("A control plane is already being initialized, requeing until control plane is ready")
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	defer func() {
		if retErr != nil {
			if !r.K3sInitLock.Unlock(ctx, scope.Cluster) {
				retErr = kerrors.NewAggregate([]error{retErr, errors.New("failed to unlock the kubeadm init lock")})
			}
		}
	}()

	scope.Info("Creating BootstrapData for the first control plane")

	if scope.Config.Spec.ServerConfiguration == nil {
		scope.Config.Spec.ServerConfiguration = &infrabootstrapv1.ServerConfiguration{}
	}

	// injects into config.ClusterConfiguration values from top level object
	r.reconcileTopLevelObjectSettings(ctx, scope.Cluster, machine, scope.Config)

	certificates := secret.NewCertificatesForInitialControlPlane()
	err := certificates.LookupOrGenerate(
		ctx,
		r.Client,
		util.ObjectKey(scope.Cluster),
		*metav1.NewControllerRef(scope.Config, infrabootstrapv1.GroupVersion.WithKind("K3sConfig")),
	)
	if err != nil {
		conditions.MarkFalse(scope.Config, bootstrapv1.CertificatesAvailableCondition, bootstrapv1.CertificatesGenerationFailedReason, clusterv1.ConditionSeverityWarning, err.Error())
		return ctrl.Result{}, err
	}

	conditions.MarkTrue(scope.Config, bootstrapv1.CertificatesAvailableCondition)

	t, err := r.generateAndStoreToken(ctx, scope)
	if err != nil {
		return ctrl.Result{}, err
	}

	initData, err := k3stypes.MarshalInitServerConfiguration(&scope.Config.Spec, t)
	if err != nil {
		scope.Error(err, "Failed to marshal server configuration")
		return ctrl.Result{}, err
	}

	files, err := r.resolveFiles(ctx, scope.Config)
	if err != nil {
		conditions.MarkFalse(scope.Config, bootstrapv1.DataSecretAvailableCondition, bootstrapv1.DataSecretGenerationFailedReason, clusterv1.ConditionSeverityWarning, err.Error())
		return ctrl.Result{}, err
	}

	initConfigFile := bootstrapv1.File{
		Path:        k3stypes.DefaultK3sConfigLocation,
		Content:     initData,
		Owner:       "root:root",
		Permissions: "0640",
	}

	controlPlaneInput := &cloudinit.ControlPlaneInput{
		BaseUserData: cloudinit.BaseUserData{
			AdditionalFiles: files,
			PreK3sCommands:  scope.Config.Spec.PreK3sCommands,
			PostK3sCommands: scope.Config.Spec.PostK3sCommands,
			ConfigFile:      initConfigFile,
		},
		Certificates: certificates,
	}

	bootstrapInitData, err := cloudinit.NewInitControlPlane(controlPlaneInput)
	if err != nil {
		scope.Error(err, "Failed to generate user data for bootstrap control plane")
		return ctrl.Result{}, err
	}

	if err := r.storeBootstrapData(ctx, scope, bootstrapInitData); err != nil {
		scope.Error(err, "Failed to store bootstrap data")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *K3sConfigReconciler) joinWorker(ctx context.Context, scope *Scope) (ctrl.Result, error) {
	scope.Info("Creating BootstrapData for the worker node")

	machine := &clusterv1.Machine{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(scope.ConfigOwner.Object, machine); err != nil {
		return ctrl.Result{}, errors.Wrapf(err, "cannot convert %s to Machine", scope.ConfigOwner.GetKind())
	}

	// injects into config.Spec values from top level object
	r.reconcileWorkerTopLevelObjectSettings(ctx, scope.Cluster, machine, scope.Config)

	// Ensure that agentConfiguration is properly set for joining node on the current cluster.
	if res, err := r.reconcileDiscovery(ctx, scope.Cluster, scope.Config); err != nil {
		return ctrl.Result{}, err
	} else if !res.IsZero() {
		return res, nil
	}

	if scope.Config.Spec.AgentConfiguration == nil {
		scope.Config.Spec.AgentConfiguration = &infrabootstrapv1.AgentConfiguration{}
	}

	joinWorkerData, err := k3stypes.MarshalJoinAgentConfiguration(&scope.Config.Spec)
	if err != nil {
		scope.Error(err, "Failed to marshal join configuration")
		return ctrl.Result{}, err
	}

	files, err := r.resolveFiles(ctx, scope.Config)
	if err != nil {
		conditions.MarkFalse(scope.Config, bootstrapv1.DataSecretAvailableCondition, bootstrapv1.DataSecretGenerationFailedReason, clusterv1.ConditionSeverityWarning, err.Error())
		return ctrl.Result{}, err
	}

	joinConfigFile := bootstrapv1.File{
		Path:        k3stypes.DefaultK3sConfigLocation,
		Content:     joinWorkerData,
		Owner:       "root:root",
		Permissions: "0640",
	}

	workerJoinInput := &cloudinit.NodeInput{
		BaseUserData: cloudinit.BaseUserData{
			AdditionalFiles: files,
			PreK3sCommands:  scope.Config.Spec.PreK3sCommands,
			PostK3sCommands: scope.Config.Spec.PostK3sCommands,
			ConfigFile:      joinConfigFile,
		},
	}

	cloudInitData, err := cloudinit.NewNode(workerJoinInput)
	if err != nil {
		scope.Error(err, "Failed to generate user data for bootstrap control plane")
		return ctrl.Result{}, err
	}

	if err := r.storeBootstrapData(ctx, scope, cloudInitData); err != nil {
		scope.Error(err, "Failed to store bootstrap data")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *K3sConfigReconciler) joinControlplane(ctx context.Context, scope *Scope) (ctrl.Result, error) {
	scope.Info("Creating BootstrapData for the joining control plane")

	if !scope.ConfigOwner.IsControlPlaneMachine() {
		return ctrl.Result{}, fmt.Errorf("%s is not a valid control plane kind, only Machine is supported", scope.ConfigOwner.GetKind())
	}

	if scope.Config.Spec.ServerConfiguration == nil {
		scope.Config.Spec.ServerConfiguration = &infrabootstrapv1.ServerConfiguration{}
	}

	machine := &clusterv1.Machine{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(scope.ConfigOwner.Object, machine); err != nil {
		return ctrl.Result{}, errors.Wrapf(err, "cannot convert %s to Machine", scope.ConfigOwner.GetKind())
	}

	// injects into config.ClusterConfiguration values from top level object
	r.reconcileTopLevelObjectSettings(ctx, scope.Cluster, machine, scope.Config)

	// Ensure that joinConfiguration.Discovery is properly set for joining node on the current cluster.
	if res, err := r.reconcileDiscovery(ctx, scope.Cluster, scope.Config); err != nil {
		return ctrl.Result{}, err
	} else if !res.IsZero() {
		return res, nil
	}

	joinData, err := k3stypes.MarshalJoinServerConfiguration(&scope.Config.Spec)
	if err != nil {
		scope.Error(err, "Failed to marshal join configuration")
		return ctrl.Result{}, err
	}

	files, err := r.resolveFiles(ctx, scope.Config)
	if err != nil {
		conditions.MarkFalse(scope.Config, bootstrapv1.DataSecretAvailableCondition, bootstrapv1.DataSecretGenerationFailedReason, clusterv1.ConditionSeverityWarning, err.Error())
		return ctrl.Result{}, err
	}

	joinConfigFile := bootstrapv1.File{
		Path:        k3stypes.DefaultK3sConfigLocation,
		Content:     joinData,
		Owner:       "root:root",
		Permissions: "0640",
	}

	controlPlaneJoinInput := &cloudinit.ControlPlaneInput{
		BaseUserData: cloudinit.BaseUserData{
			AdditionalFiles: files,
			PreK3sCommands:  scope.Config.Spec.PreK3sCommands,
			PostK3sCommands: scope.Config.Spec.PostK3sCommands,
			ConfigFile:      joinConfigFile,
		},
	}

	cloudInitData, err := cloudinit.NewJoinControlPlane(controlPlaneJoinInput)
	if err != nil {
		scope.Error(err, "Failed to generate user data for bootstrap control plane")
		return ctrl.Result{}, err
	}

	if err := r.storeBootstrapData(ctx, scope, cloudInitData); err != nil {
		scope.Error(err, "Failed to store bootstrap data")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *K3sConfigReconciler) generateAndStoreToken(ctx context.Context, scope *Scope) (string, error) {
	t, err := bootstraputil.GenerateBootstrapToken()
	if err != nil {
		return "", errors.Wrap(err, "unable to generate bootstrap token")
	}

	s := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-token", scope.Cluster.Name),
			Namespace: scope.Config.Namespace,
			Labels: map[string]string{
				clusterv1.ClusterLabelName: scope.Cluster.Name,
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: infrabootstrapv1.GroupVersion.String(),
					Kind:       "K3sConfig",
					Name:       scope.Config.Name,
					UID:        scope.Config.UID,
					Controller: pointer.Bool(true),
				},
			},
		},
		Data: map[string][]byte{
			"value": []byte(t),
		},
		Type: clusterv1.ClusterSecretType,
	}

	// as secret creation and scope.Config status patch are not atomic operations
	// it is possible that secret creation happens but the config.Status patches are not applied
	if err := r.Client.Create(ctx, s); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return "", errors.Wrapf(err, "failed to create token for K3sConfig %s/%s", scope.Config.Namespace, scope.Config.Name)
		}
		if err := r.Client.Update(ctx, s); err != nil {
			return "", errors.Wrapf(err, "failed to update bootstrap token secret for K3sConfig %s/%s", scope.Config.Namespace, scope.Config.Name)
		}
	}

	return t, nil
}

// resolveFiles maps .Spec.Files into cloudinit.Files, resolving any object references
// along the way.
func (r *K3sConfigReconciler) resolveFiles(ctx context.Context, cfg *infrabootstrapv1.K3sConfig) ([]bootstrapv1.File, error) {
	collected := make([]bootstrapv1.File, 0, len(cfg.Spec.Files))

	for i := range cfg.Spec.Files {
		in := cfg.Spec.Files[i]
		if in.ContentFrom != nil {
			data, err := r.resolveSecretFileContent(ctx, cfg.Namespace, in)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to resolve file source")
			}
			in.ContentFrom = nil
			in.Content = string(data)
		}
		collected = append(collected, in)
	}

	return collected, nil
}

// resolveSecretFileContent returns file content fetched from a referenced secret object.
func (r *K3sConfigReconciler) resolveSecretFileContent(ctx context.Context, ns string, source bootstrapv1.File) ([]byte, error) {
	s := &corev1.Secret{}
	key := types.NamespacedName{Namespace: ns, Name: source.ContentFrom.Secret.Name}
	if err := r.Client.Get(ctx, key, s); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, errors.Wrapf(err, "secret not found: %s", key)
		}
		return nil, errors.Wrapf(err, "failed to retrieve Secret %q", key)
	}
	data, ok := s.Data[source.ContentFrom.Secret.Key]
	if !ok {
		return nil, errors.Errorf("secret references non-existent secret key: %q", source.ContentFrom.Secret.Key)
	}
	return data, nil
}

// storeBootstrapData creates a new secret with the data passed in as input,
// sets the reference in the configuration status and ready to true.
func (r *K3sConfigReconciler) storeBootstrapData(ctx context.Context, scope *Scope, data []byte) error {
	log := ctrl.LoggerFrom(ctx)

	s := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      scope.Config.Name,
			Namespace: scope.Config.Namespace,
			Labels: map[string]string{
				clusterv1.ClusterLabelName: scope.Cluster.Name,
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: infrabootstrapv1.GroupVersion.String(),
					Kind:       "K3sConfig",
					Name:       scope.Config.Name,
					UID:        scope.Config.UID,
					Controller: pointer.Bool(true),
				},
			},
		},
		Data: map[string][]byte{
			"value": data,
		},
		Type: clusterv1.ClusterSecretType,
	}

	// as secret creation and scope.Config status patch are not atomic operations
	// it is possible that secret creation happens but the config.Status patches are not applied
	if err := r.Client.Create(ctx, s); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return errors.Wrapf(err, "failed to create bootstrap data secret for K3sConfig %s/%s", scope.Config.Namespace, scope.Config.Name)
		}
		log.Info("bootstrap data secret for K3sConfig already exists, updating", "Secret", klog.KObj(s))
		if err := r.Client.Update(ctx, s); err != nil {
			return errors.Wrapf(err, "failed to update bootstrap data secret for K3sConfig %s/%s", scope.Config.Namespace, scope.Config.Name)
		}
	}
	scope.Config.Status.DataSecretName = pointer.String(s.Name)
	scope.Config.Status.Ready = true
	conditions.MarkTrue(scope.Config, bootstrapv1.DataSecretAvailableCondition)
	return nil
}

func (r *K3sConfigReconciler) reconcileDiscovery(ctx context.Context, cluster *clusterv1.Cluster, config *infrabootstrapv1.K3sConfig) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	// if config already contains a file discovery configuration, respect it without further validations
	if config.Spec.Cluster.TokenFile != "" {
		return ctrl.Result{}, nil
	}

	// if BootstrapToken already contains an APIServerEndpoint, respect it; otherwise inject the APIServerEndpoint endpoint defined in cluster status
	apiServerEndpoint := config.Spec.Cluster.Server
	if apiServerEndpoint == "" {
		if !cluster.Spec.ControlPlaneEndpoint.IsValid() {
			log.V(1).Info("Waiting for Cluster Controller to set Cluster.Server")
			return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
		}

		apiServerEndpoint = cluster.Spec.ControlPlaneEndpoint.String()
		config.Spec.Cluster.Server = fmt.Sprintf("https://%s", apiServerEndpoint)
		log.V(3).Info("Altering Cluster.Server", "Server", apiServerEndpoint)
	}

	// if BootstrapToken already contains a token, respect it; otherwise create a new bootstrap token for the node to join
	if config.Spec.Cluster.Token == "" {
		s := &corev1.Secret{}
		obj := client.ObjectKey{
			Namespace: config.Namespace,
			Name:      fmt.Sprintf("%s-token", cluster.Name),
		}

		if err := r.Client.Get(ctx, obj, s); err != nil {
			return ctrl.Result{}, errors.Wrapf(err, "failed to get token for K3sConfig %s/%s", config.Namespace, config.Name)
		}

		config.Spec.Cluster.Token = string(s.Data["value"])
		log.V(3).Info("Altering Cluster.Token")
	}

	return ctrl.Result{}, nil
}

// MachineToBootstrapMapFunc is a handler.ToRequestsFunc to be used to enqueue
// request for reconciliation of K3sConfig.
func (r *K3sConfigReconciler) MachineToBootstrapMapFunc(o client.Object) []ctrl.Request {
	m, ok := o.(*clusterv1.Machine)
	if !ok {
		panic(fmt.Sprintf("Expected a Machine but got a %T", o))
	}

	var result []ctrl.Request
	if m.Spec.Bootstrap.ConfigRef != nil && m.Spec.Bootstrap.ConfigRef.GroupVersionKind() == infrabootstrapv1.GroupVersion.WithKind("K3sConfig") {
		name := client.ObjectKey{Namespace: m.Namespace, Name: m.Spec.Bootstrap.ConfigRef.Name}
		result = append(result, ctrl.Request{NamespacedName: name})
	}
	return result
}

// MachinePoolToBootstrapMapFunc is a handler.ToRequestsFunc to be used to enqueue
// request for reconciliation of K3sConfig.
func (r *K3sConfigReconciler) MachinePoolToBootstrapMapFunc(o client.Object) []ctrl.Request {
	m, ok := o.(*expv1.MachinePool)
	if !ok {
		panic(fmt.Sprintf("Expected a MachinePool but got a %T", o))
	}

	var result []ctrl.Request
	configRef := m.Spec.Template.Spec.Bootstrap.ConfigRef
	if configRef != nil && configRef.GroupVersionKind().GroupKind() == infrabootstrapv1.GroupVersion.WithKind("K3sConfig").GroupKind() {
		name := client.ObjectKey{Namespace: m.Namespace, Name: configRef.Name}
		result = append(result, ctrl.Request{NamespacedName: name})
	}
	return result
}

// ClusterToK3sConfigs is a handler.ToRequestsFunc to be used to enqueue
// requests for reconciliation of K3sConfig.
func (r *K3sConfigReconciler) ClusterToK3sConfigs(o client.Object) []ctrl.Request {
	var result []ctrl.Request

	c, ok := o.(*clusterv1.Cluster)
	if !ok {
		panic(fmt.Sprintf("Expected a Cluster but got a %T", o))
	}

	selectors := []client.ListOption{
		client.InNamespace(c.Namespace),
		client.MatchingLabels{
			clusterv1.ClusterLabelName: c.Name,
		},
	}

	machineList := &clusterv1.MachineList{}
	if err := r.Client.List(context.TODO(), machineList, selectors...); err != nil {
		return nil
	}

	for _, m := range machineList.Items {
		if m.Spec.Bootstrap.ConfigRef != nil &&
			m.Spec.Bootstrap.ConfigRef.GroupVersionKind().GroupKind() == infrabootstrapv1.GroupVersion.WithKind("K3sConfig").GroupKind() {
			name := client.ObjectKey{Namespace: m.Namespace, Name: m.Spec.Bootstrap.ConfigRef.Name}
			result = append(result, ctrl.Request{NamespacedName: name})
		}
	}

	if feature.Gates.Enabled(feature.MachinePool) {
		machinePoolList := &expv1.MachinePoolList{}
		if err := r.Client.List(context.TODO(), machinePoolList, selectors...); err != nil {
			return nil
		}

		for _, mp := range machinePoolList.Items {
			if mp.Spec.Template.Spec.Bootstrap.ConfigRef != nil &&
				mp.Spec.Template.Spec.Bootstrap.ConfigRef.GroupVersionKind().GroupKind() == infrabootstrapv1.GroupVersion.WithKind("K3sConfig").GroupKind() {
				name := client.ObjectKey{Namespace: mp.Namespace, Name: mp.Spec.Template.Spec.Bootstrap.ConfigRef.Name}
				result = append(result, ctrl.Request{NamespacedName: name})
			}
		}
	}

	return result
}

// reconcileTopLevelObjectSettings injects into config.ClusterConfiguration values from top level objects like cluster and machine.
// The implementation func respect user provided config values, but in case some of them are missing, values from top level objects are used.
func (r *K3sConfigReconciler) reconcileTopLevelObjectSettings(ctx context.Context, cluster *clusterv1.Cluster, machine *clusterv1.Machine, config *infrabootstrapv1.K3sConfig) {
	log := ctrl.LoggerFrom(ctx)

	// If there are no Network settings defined in ClusterConfiguration, use ClusterNetwork settings, if defined
	if cluster.Spec.ClusterNetwork != nil {
		if config.Spec.ServerConfiguration.Networking.ClusterDomain == "" && cluster.Spec.ClusterNetwork.ServiceDomain != "" {
			config.Spec.ServerConfiguration.Networking.ClusterDomain = cluster.Spec.ClusterNetwork.ServiceDomain
			log.V(3).Info("Altering ServerConfiguration.Networking.ClusterDomain", "ClusterDomain", config.Spec.ServerConfiguration.Networking.ClusterDomain)
		}
		if config.Spec.ServerConfiguration.Networking.ServiceCIDR == "" &&
			cluster.Spec.ClusterNetwork.Services != nil &&
			len(cluster.Spec.ClusterNetwork.Services.CIDRBlocks) > 0 {
			config.Spec.ServerConfiguration.Networking.ServiceCIDR = cluster.Spec.ClusterNetwork.Services.String()
			log.V(3).Info("Altering ServerConfiguration.Networking.ServiceCIDR", "ServiceCIDR", config.Spec.ServerConfiguration.Networking.ServiceCIDR)
		}
		if config.Spec.ServerConfiguration.Networking.ClusterCIDR == "" &&
			cluster.Spec.ClusterNetwork.Pods != nil &&
			len(cluster.Spec.ClusterNetwork.Pods.CIDRBlocks) > 0 {
			config.Spec.ServerConfiguration.Networking.ClusterCIDR = cluster.Spec.ClusterNetwork.Pods.String()
			log.V(3).Info("Altering ServerConfiguration.Networking.ClusterCIDR", "ClusterCIDR", config.Spec.ServerConfiguration.Networking.ClusterCIDR)
		}
	}

	// If there are no Version settings defined, use Version from machine, if defined
	if config.Spec.Version == "" && machine.Spec.Version != nil {
		config.Spec.Version = *machine.Spec.Version
		log.V(3).Info("Altering Spec.Version", "Version", config.Spec.Version)
	}
}

func (r *K3sConfigReconciler) reconcileWorkerTopLevelObjectSettings(ctx context.Context, _ *clusterv1.Cluster, machine *clusterv1.Machine, config *infrabootstrapv1.K3sConfig) {
	log := ctrl.LoggerFrom(ctx)

	// If there are no Version settings defined, use Version from machine, if defined
	if config.Spec.Version == "" && machine.Spec.Version != nil {
		config.Spec.Version = *machine.Spec.Version
		log.V(3).Info("Altering Spec.Version", "Version", config.Spec.Version)
	}
}
