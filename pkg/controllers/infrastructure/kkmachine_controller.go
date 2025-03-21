package infrastructure

import (
	"context"
	"embed"
	"fmt"
	"time"

	"github.com/cockroachdb/errors"
	capkkinfrav1beta1 "github.com/kubesphere/kubekey/api/capkk/infrastructure/v1beta1"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	clusterv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	ctrlcontroller "sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/yaml"

	"github.com/kubesphere/kubekey/v4/cmd/controller-manager/app/options"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/util"
)

// KKMachineReconciler reconciles a KKMachine object.
// One KKMachine should have one Playbook running in time.
type KKMachineReconciler struct {
	ctrlclient.Client
	record.EventRecorder
	restConfig *rest.Config
	lockType   string
}

var _ options.Controller = &KKMachineReconciler{}
var _ reconcile.Reconciler = &KKMachineReconciler{}

// kubeVersionConfigs is the default config for each kube_version
//
//go:embed versions
var kubeVersionConfigs embed.FS

// Name implements controllers.controller.
func (r *KKMachineReconciler) Name() string {
	return "kkmachine-reconciler"
}

// SetupWithManager implements controllers.controller.
func (r *KKMachineReconciler) SetupWithManager(mgr ctrl.Manager, o options.ControllerManagerServerOptions) error {
	r.Client = mgr.GetClient()
	r.EventRecorder = mgr.GetEventRecorderFor(r.Name())
	r.restConfig = mgr.GetConfig()
	r.lockType = o.LeaderElectionResourceLock
	if r.lockType == "" {
		r.lockType = resourcelock.LeasesResourceLock
	}

	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(ctrlcontroller.Options{
			MaxConcurrentReconciles: o.MaxConcurrentReconciles,
		}).
		For(&capkkinfrav1beta1.KKMachine{}).
		// Watches playbook to sync kkmachine.
		Watches(&kkcorev1.Playbook{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj ctrlclient.Object) []reconcile.Request {
			kkmachine := &capkkinfrav1beta1.KKMachine{}
			if err := util.GetOwnerFromObject(ctx, r.Client, obj, kkmachine); err == nil {
				return []ctrl.Request{{NamespacedName: ctrlclient.ObjectKeyFromObject(kkmachine)}}
			}

			return nil
		})).
		Complete(r)
}

// Reconcile implements controllers.controller.
func (r *KKMachineReconciler) Reconcile(ctx context.Context, req reconcile.Request) (_ reconcile.Result, retErr error) {
	kkmachine := &capkkinfrav1beta1.KKMachine{}
	if err := r.Client.Get(ctx, req.NamespacedName, kkmachine); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, errors.Wrapf(err, "failed to get kkmachine %q", req.String())
	}
	clusterName := kkmachine.Labels[clusterv1beta1.ClusterNameLabel]
	if clusterName == "" {
		klog.V(5).InfoS("kkmachine is not belong cluster", "kkmachine", req.String())

		return ctrl.Result{}, nil
	}
	scope, err := newClusterScope(ctx, r.Client, reconcile.Request{NamespacedName: types.NamespacedName{
		Namespace: req.Namespace,
		Name:      clusterName,
	}})
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}
	if err := scope.newPatchHelper(kkmachine); err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}
	defer func() {
		if err := scope.PatchHelper.Patch(ctx, kkmachine); err != nil {
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
	if kkmachine.ObjectMeta.DeletionTimestamp.IsZero() && !controllerutil.ContainsFinalizer(kkmachine, capkkinfrav1beta1.KKMachineFinalizer) {
		controllerutil.AddFinalizer(kkmachine, capkkinfrav1beta1.KKMachineFinalizer)

		return ctrl.Result{}, nil
	}

	if !kkmachine.DeletionTimestamp.IsZero() {
		return reconcile.Result{}, errors.WithStack(r.reconcileDelete(ctx, scope, kkmachine))
	}

	machine := &clusterv1beta1.Machine{}
	if err := util.GetOwnerFromObject(ctx, r.Client, kkmachine, machine); err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "failed to get machine from kkmachine %q", ctrlclient.ObjectKeyFromObject(machine))
	}
	kkmachine.Spec.Version = machine.Spec.Version

	if kkmachine.Spec.ProviderID == nil {
		klog.InfoS("kkmachine has not providerID, waiting for inventory to set", "kkmachine", kkmachine.Name)

		return reconcile.Result{}, nil
	}
	// should waiting control plane ready when kkmachine is worker
	if machine.Spec.Bootstrap.DataSecretName == nil {
		klog.InfoS("waiting cloud-config ready...", "kkmachine", kkmachine.Name)

		return reconcile.Result{RequeueAfter: 10 * time.Second}, nil
	}

	return reconcile.Result{}, errors.WithStack(r.reconcileNormal(ctx, scope, kkmachine, machine))
}

// reconcileDelete handles delete reconcile.
func (r *KKMachineReconciler) reconcileDelete(ctx context.Context, scope *clusterScope, kkmachine *capkkinfrav1beta1.KKMachine) error {
	// check if addNodePlaybook has created
	addNodePlaybookName := kkmachine.Annotations[capkkinfrav1beta1.AddNodePlaybookAnnotation]
	delNodePlaybookName := kkmachine.Annotations[capkkinfrav1beta1.DeleteNodePlaybookAnnotation]
	addNodePlaybook, delNodePlaybook, err := r.getPlaybook(ctx, scope, kkmachine)
	if err != nil {
		return errors.WithStack(err)
	}
	switch {
	case addNodePlaybookName == "" && delNodePlaybookName == "":
		// the kkmachine has not executor any playbook, delete direct.
		controllerutil.RemoveFinalizer(kkmachine, capkkinfrav1beta1.KKMachineFinalizer)
	case addNodePlaybookName != "" && delNodePlaybookName == "":
		// should waiting addNodePlaybook completed and create deleteNodePlaybook
		if addNodePlaybook == nil || // addNodePlaybook has been deleted
			(addNodePlaybook.Status.Phase == kkcorev1.PlaybookPhaseSucceeded || addNodePlaybook.Status.Phase == kkcorev1.PlaybookPhaseFailed) { // addNodePlaybook has completed
			return r.createDeleteNodePlaybook(ctx, scope, kkmachine)
		}
		// should waiting addNodePlaybook completed
		return nil
	case addNodePlaybookName != "" && delNodePlaybookName != "":
		if addNodePlaybook != nil && addNodePlaybook.DeletionTimestamp.IsZero() {
			return r.Client.Delete(ctx, addNodePlaybook)
		}
		if delNodePlaybook != nil && delNodePlaybook.DeletionTimestamp.IsZero() {
			if delNodePlaybook.Status.Phase == kkcorev1.PlaybookPhaseSucceeded {
				return r.Client.Delete(ctx, delNodePlaybook)
			}
			// should waiting delNodePlaybook completed
			return nil
		}
	}

	if addNodePlaybook == nil && delNodePlaybook == nil {
		// Delete finalizer.
		controllerutil.RemoveFinalizer(kkmachine, capkkinfrav1beta1.KKMachineFinalizer)
	}

	return nil
}

// getPlaybook get addNodePlaybook and delNodePlaybook from kkmachine.Annotations.
func (r *KKMachineReconciler) getPlaybook(ctx context.Context, scope *clusterScope, kkmachine *capkkinfrav1beta1.KKMachine) (*kkcorev1.Playbook, *kkcorev1.Playbook, error) {
	var addNodePlaybook, delNodePlaybook *kkcorev1.Playbook
	if name, ok := kkmachine.Annotations[capkkinfrav1beta1.AddNodePlaybookAnnotation]; ok && name != "" {
		addNodePlaybook = &kkcorev1.Playbook{}
		if err := r.Client.Get(ctx, ctrlclient.ObjectKey{Namespace: scope.Namespace, Name: name}, addNodePlaybook); err != nil {
			if !apierrors.IsNotFound(err) {
				// maybe delete by user. skip
				return nil, nil, errors.Wrapf(err, "failed to get addNode playbook from kkmachine %q with annotation %q", ctrlclient.ObjectKeyFromObject(kkmachine), capkkinfrav1beta1.AddNodePlaybookAnnotation)
			}
			addNodePlaybook = nil
		}
	}
	if name, ok := kkmachine.Annotations[capkkinfrav1beta1.DeleteNodePlaybookAnnotation]; ok && name != "" {
		delNodePlaybook = &kkcorev1.Playbook{}
		if err := r.Client.Get(ctx, ctrlclient.ObjectKey{Namespace: scope.Namespace, Name: name}, delNodePlaybook); err != nil {
			if !apierrors.IsNotFound(err) {
				// maybe delete by user. skip
				return nil, nil, errors.Wrapf(err, "failed to get delNode playbook from kkmachine %q with annotation %q", ctrlclient.ObjectKeyFromObject(kkmachine), capkkinfrav1beta1.DeleteNodePlaybookAnnotation)
			}
			delNodePlaybook = nil
		}
	}

	return addNodePlaybook, delNodePlaybook, nil
}

// reconcileNormal handles normal reconcile.
// when dataSecret or certificates files changed. KCP will RollingUpdate machine (create new machines to replace old machines)
// so the sync file should contains in add_node playbook.
func (r *KKMachineReconciler) reconcileNormal(ctx context.Context, scope *clusterScope, kkmachine *capkkinfrav1beta1.KKMachine, machine *clusterv1beta1.Machine) error {
	playbookName := kkmachine.Annotations[capkkinfrav1beta1.AddNodePlaybookAnnotation]
	if playbookName == "" {
		kkmachine.Status.Ready = false
		kkmachine.Status.FailureReason = ""
		kkmachine.Status.FailureMessage = ""
		// should create playbook
		return r.createAddNodePlaybook(ctx, scope, kkmachine, machine)
	}
	// check playbook status
	playbook := &kkcorev1.Playbook{}
	if err := r.Client.Get(ctx, ctrlclient.ObjectKey{Namespace: scope.Namespace, Name: playbookName}, playbook); err != nil {
		if apierrors.IsNotFound(err) {
			// the playbook has not found.
			r.EventRecorder.Eventf(kkmachine, corev1.EventTypeWarning, "AddNodeFailed", "add node playbook: %q not found", playbookName)

			return nil
		}

		return errors.Wrapf(err, "failed to get playbook %s/%s", scope.Namespace, playbookName)
	}

	switch playbook.Status.Phase {
	case kkcorev1.PlaybookPhaseSucceeded:
		// set machine to ready
		kkmachine.Status.Ready = true
		kkmachine.Status.FailureReason = ""
		kkmachine.Status.FailureMessage = ""
	case kkcorev1.PlaybookPhaseFailed:
		// set machine to not ready
		kkmachine.Status.Ready = false
		kkmachine.Status.FailureReason = capkkinfrav1beta1.KKMachineFailedReasonAddNodeFailed
		kkmachine.Status.FailureMessage = fmt.Sprintf("add_node playbook %q run failed", playbookName)
	}

	return nil
}

func (r *KKMachineReconciler) createAddNodePlaybook(ctx context.Context, scope *clusterScope, kkmachine *capkkinfrav1beta1.KKMachine, machine *clusterv1beta1.Machine) error {
	if ok, _ := scope.ifPlaybookCompleted(ctx, kkmachine); !ok {
		return nil
	}
	volumes, volumeMounts := scope.getVolumeMounts(ctx)
	// mount cloud-config
	volumeMounts = append(volumeMounts, corev1.VolumeMount{
		Name:      "cloud-config",
		MountPath: _const.CAPKKCloudConfigPath,
		ReadOnly:  true,
	})
	volumes = append(volumes, corev1.Volume{
		Name: "cloud-config",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: *machine.Spec.Bootstrap.DataSecretName,
			},
		},
	})

	config, err := r.getConfig(scope, kkmachine)
	if err != nil {
		klog.ErrorS(err, "get default config error, use default config", "version", kkmachine.Spec.Version)
	}
	playbook := &kkcorev1.Playbook{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: kkmachine.Name + "-",
			Namespace:    scope.Namespace,
			Labels: map[string]string{
				clusterv1beta1.ClusterNameLabel: scope.Name,
			},
		},
		Spec: kkcorev1.PlaybookSpec{
			Project: kkcorev1.PlaybookProject{
				Addr: _const.CAPKKProjectdir,
			},
			Playbook:     _const.CAPKKPlaybookAddNode,
			InventoryRef: util.ObjectRef(r.Client, scope.Inventory),
			Config:       ptr.Deref(config, kkcorev1.Config{}),
			VolumeMounts: volumeMounts,
			Volumes:      volumes,
		},
	}
	if err := ctrl.SetControllerReference(kkmachine, playbook, r.Client.Scheme()); err != nil {
		return errors.Wrapf(err, "failed to set ownerReference from kkmachine %q to addNode playbook", ctrlclient.ObjectKeyFromObject(kkmachine))
	}
	if err := r.Client.Create(ctx, playbook); err != nil {
		return errors.Wrapf(err, "failed to create addNode playbook from kkmachine %q", ctrlclient.ObjectKeyFromObject(kkmachine))
	}
	// add playbook name to kkmachine
	kkmachine.Annotations[capkkinfrav1beta1.AddNodePlaybookAnnotation] = playbook.Name

	return nil
}

func (r *KKMachineReconciler) createDeleteNodePlaybook(ctx context.Context, scope *clusterScope, kkmachine *capkkinfrav1beta1.KKMachine) error {
	if ok, _ := scope.ifPlaybookCompleted(ctx, kkmachine); !ok {
		return nil
	}
	config, err := r.getConfig(scope, kkmachine)
	if err != nil {
		klog.ErrorS(err, "get default config error, use default config", "kubeVersion", kkmachine.Spec.Version)
	}
	volumes, volumeMounts := scope.getVolumeMounts(ctx)
	playbook := &kkcorev1.Playbook{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: kkmachine.Name + "-",
			Namespace:    scope.Namespace,
			Labels: map[string]string{
				clusterv1beta1.ClusterNameLabel: scope.Name,
			},
		},
		Spec: kkcorev1.PlaybookSpec{
			Project: kkcorev1.PlaybookProject{
				Addr: _const.CAPKKProjectdir,
			},
			Playbook:     _const.CAPKKPlaybookDeleteNode,
			InventoryRef: util.ObjectRef(r.Client, scope.Inventory),
			Config:       *config,
			VolumeMounts: volumeMounts,
			Volumes:      volumes,
		},
	}
	if err := ctrl.SetControllerReference(kkmachine, playbook, r.Client.Scheme()); err != nil {
		return errors.Wrapf(err, "failed to set ownerReference from kkmachine %q to delNode playbook", ctrlclient.ObjectKeyFromObject(kkmachine))
	}
	if err := r.Client.Create(ctx, playbook); err != nil {
		return errors.Wrapf(err, "failed to create delNode playbook from kkmachine %q", ctrlclient.ObjectKeyFromObject(kkmachine))
	}
	kkmachine.Annotations[capkkinfrav1beta1.DeleteNodePlaybookAnnotation] = playbook.Name

	return nil
}

// getConfig get default config for kkmachine.
func (r *KKMachineReconciler) getConfig(scope *clusterScope, kkmachine *capkkinfrav1beta1.KKMachine) (*kkcorev1.Config, error) {
	var config = &kkcorev1.Config{}
	if kkmachine.Spec.Config.Raw != nil {
		config = &kkcorev1.Config{
			Spec: kkmachine.Spec.Config,
		}
	} else {
		if kkmachine.Spec.Version == nil {
			return config, errors.New("kubeVersion or config is empty")
		}
		data, err := kubeVersionConfigs.ReadFile(fmt.Sprintf("versions/%s.yaml", *kkmachine.Spec.Version))
		if err != nil {
			return config, errors.Wrap(err, "failed to read default config file")
		}
		if err := yaml.Unmarshal(data, config); err != nil {
			return config, errors.Wrap(err, "failed to unmarshal config file")
		}
		klog.InfoS("get default config", "config", config)
	}

	if err := config.SetValue(_const.Workdir, _const.CAPKKWorkdir); err != nil {
		return config, errors.Wrapf(err, "failed to set %q in config", _const.Workdir)
	}
	if err := config.SetValue("node_name", _const.ProviderID2Host(scope.Name, kkmachine.Spec.ProviderID)); err != nil {
		return config, errors.Wrap(err, "failed to set \"node_name\" in config")
	}
	if err := config.SetValue("kube_version", kkmachine.Spec.Version); err != nil {
		return config, errors.Wrap(err, "failed to set \"kube_version\" in config")
	}
	if err := config.SetValue("kubernetes.cluster_name", scope.Cluster.Name); err != nil {
		return config, errors.Wrap(err, "failed to set \"kubernetes.cluster_name\" in config")
	}
	if err := config.SetValue("kubernetes.roles", kkmachine.Spec.Roles); err != nil {
		return config, errors.Wrap(err, "failed to set \"kubernetes.roles\" in config")
	}
	if err := config.SetValue("cluster_network", scope.Cluster.Spec.ClusterNetwork); err != nil {
		return config, errors.Wrap(err, "failed to set \"cluster_network\" in config")
	}

	switch scope.KKCluster.Spec.ControlPlaneEndpointType {
	case capkkinfrav1beta1.ControlPlaneEndpointTypeVIP:
		// should set vip addr to config
		if err := config.SetValue("kubernetes.control_plane_endpoint.kube_vip.address", scope.Cluster.Spec.ControlPlaneEndpoint.Host); err != nil {
			return config, errors.Wrap(err, "failed to set \"kubernetes.control_plane_endpoint.kube_vip.address\" in config")
		}
	case capkkinfrav1beta1.ControlPlaneEndpointTypeDNS:
		// do nothing
	default:
		return config, errors.New("unsupport ControlPlaneEndpointType")
	}
	if err := config.SetValue("kubernetes.control_plane_endpoint.host", scope.Cluster.Spec.ControlPlaneEndpoint.Host); err != nil {
		return config, errors.Wrap(err, "failed to set \"kubernetes.kube_vip.address\" in config")
	}
	if err := config.SetValue("kubernetes.control_plane_endpoint.type", scope.KKCluster.Spec.ControlPlaneEndpointType); err != nil {
		return config, errors.Wrap(err, "failed to set \"kubernetes.kube_vip.enabled\" in config")
	}

	return config, nil
}
