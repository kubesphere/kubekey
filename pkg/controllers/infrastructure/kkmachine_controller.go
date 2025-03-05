package infrastructure

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"time"

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
// One KKMachine should have one Pipeline running in time.
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
		// Watches pipeline to sync kkmachine.
		Watches(&kkcorev1.Pipeline{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj ctrlclient.Object) []reconcile.Request {
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

		return ctrl.Result{}, err
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
		return ctrl.Result{}, err
	}
	if err := scope.newPatchHelper(kkmachine); err != nil {
		return ctrl.Result{}, err
	}
	defer func() {
		if err := scope.PatchHelper.Patch(ctx, kkmachine); err != nil {
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
	if kkmachine.ObjectMeta.DeletionTimestamp.IsZero() && !controllerutil.ContainsFinalizer(kkmachine, capkkinfrav1beta1.KKMachineFinalizer) {
		controllerutil.AddFinalizer(kkmachine, capkkinfrav1beta1.KKMachineFinalizer)

		return ctrl.Result{}, nil
	}

	if !kkmachine.DeletionTimestamp.IsZero() {
		return reconcile.Result{}, r.reconcileDelete(ctx, scope, kkmachine)
	}

	machine := &clusterv1beta1.Machine{}
	if err := util.GetOwnerFromObject(ctx, r.Client, kkmachine, machine); err != nil {
		return reconcile.Result{}, err
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

	return reconcile.Result{}, r.reconcileNormal(ctx, scope, kkmachine, machine)
}

// reconcileDelete handles delete reconcile.
func (r *KKMachineReconciler) reconcileDelete(ctx context.Context, scope *clusterScope, kkmachine *capkkinfrav1beta1.KKMachine) error {
	// check if addNodePipeline has created
	addNodePipelineName := kkmachine.Annotations[capkkinfrav1beta1.AddNodePipelineAnnotation]
	delNodePipelineName := kkmachine.Annotations[capkkinfrav1beta1.DeleteNodePipelineAnnotation]
	addNodePipeline, delNodePipeline, err := r.getPipeline(ctx, scope, kkmachine)
	if err != nil {
		return err
	}
	switch {
	case addNodePipelineName == "" && delNodePipelineName == "":
		// the kkmachine has not executor any pipeline, delete direct.
		controllerutil.RemoveFinalizer(kkmachine, capkkinfrav1beta1.KKMachineFinalizer)
	case addNodePipelineName != "" && delNodePipelineName == "":
		// should waiting addNodePipeline completed and create deleteNodePipeline
		if addNodePipeline == nil || // addNodePipeline has been deleted
			(addNodePipeline.Status.Phase == kkcorev1.PipelinePhaseSucceeded || addNodePipeline.Status.Phase == kkcorev1.PipelinePhaseFailed) { // addNodePipeline has completed
			return r.createDeleteNodePipeline(ctx, scope, kkmachine)
		}
		// should waiting addNodePipeline completed
		return nil
	case addNodePipelineName != "" && delNodePipelineName != "":
		if addNodePipeline != nil && addNodePipeline.DeletionTimestamp.IsZero() {
			return r.Client.Delete(ctx, addNodePipeline)
		}
		if delNodePipeline != nil && delNodePipeline.DeletionTimestamp.IsZero() {
			if delNodePipeline.Status.Phase == kkcorev1.PipelinePhaseSucceeded {
				return r.Client.Delete(ctx, delNodePipeline)
			}
			// should waiting delNodePipeline completed
			return nil
		}
	}

	if addNodePipeline == nil && delNodePipeline == nil {
		// Delete finalizer.
		controllerutil.RemoveFinalizer(kkmachine, capkkinfrav1beta1.KKMachineFinalizer)
	}

	return nil
}

// getPipeline get addNodePipeline and delNodePipeline from kkmachine.Annotations.
func (r *KKMachineReconciler) getPipeline(ctx context.Context, scope *clusterScope, kkmachine *capkkinfrav1beta1.KKMachine) (*kkcorev1.Pipeline, *kkcorev1.Pipeline, error) {
	var addNodePipeline, delNodePipeline *kkcorev1.Pipeline
	if name, ok := kkmachine.Annotations[capkkinfrav1beta1.AddNodePipelineAnnotation]; ok && name != "" {
		addNodePipeline = &kkcorev1.Pipeline{}
		if err := r.Client.Get(ctx, ctrlclient.ObjectKey{Namespace: scope.Namespace, Name: name}, addNodePipeline); err != nil {
			if !apierrors.IsNotFound(err) {
				// maybe delete by user. skip
				return nil, nil, err
			}
			addNodePipeline = nil
		}
	}
	if name, ok := kkmachine.Annotations[capkkinfrav1beta1.DeleteNodePipelineAnnotation]; ok && name != "" {
		delNodePipeline = &kkcorev1.Pipeline{}
		if err := r.Client.Get(ctx, ctrlclient.ObjectKey{Namespace: scope.Namespace, Name: name}, delNodePipeline); err != nil {
			if !apierrors.IsNotFound(err) {
				// maybe delete by user. skip
				return nil, nil, err
			}
			delNodePipeline = nil
		}
	}

	return addNodePipeline, delNodePipeline, nil
}

// reconcileNormal handles normal reconcile.
// when dataSecret or certificates files changed. KCP will RollingUpdate machine (create new machines to replace old machines)
// so the sync file should contains in add_node pipeline.
func (r *KKMachineReconciler) reconcileNormal(ctx context.Context, scope *clusterScope, kkmachine *capkkinfrav1beta1.KKMachine, machine *clusterv1beta1.Machine) error {
	pipelineName := kkmachine.Annotations[capkkinfrav1beta1.AddNodePipelineAnnotation]
	if pipelineName == "" {
		kkmachine.Status.Ready = false
		kkmachine.Status.FailureReason = ""
		kkmachine.Status.FailureMessage = ""
		// should create pipeline
		return r.createAddNodePipeline(ctx, scope, kkmachine, machine)
	}
	// check pipeline status
	pipeline := &kkcorev1.Pipeline{}
	if err := r.Client.Get(ctx, ctrlclient.ObjectKey{Namespace: scope.Namespace, Name: pipelineName}, pipeline); err != nil {
		if apierrors.IsNotFound(err) {
			// the pipeline has not found.
			r.EventRecorder.Eventf(kkmachine, corev1.EventTypeWarning, "AddNodeFailed", "add node pipeline: %q not found", pipelineName)

			return nil
		}

		return err
	}

	switch pipeline.Status.Phase {
	case kkcorev1.PipelinePhaseSucceeded:
		// set machine to ready
		kkmachine.Status.Ready = true
		kkmachine.Status.FailureReason = ""
		kkmachine.Status.FailureMessage = ""
	case kkcorev1.PipelinePhaseFailed:
		// set machine to not ready
		kkmachine.Status.Ready = false
		kkmachine.Status.FailureReason = capkkinfrav1beta1.KKMachineFailedReasonAddNodeFailed
		kkmachine.Status.FailureMessage = fmt.Sprintf("add_node pipeline %q run failed", pipelineName)
	}

	return nil
}

func (r *KKMachineReconciler) createAddNodePipeline(ctx context.Context, scope *clusterScope, kkmachine *capkkinfrav1beta1.KKMachine, machine *clusterv1beta1.Machine) error {
	if ok, _ := scope.ifPipelineCompleted(ctx, kkmachine); !ok {
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
	pipeline := &kkcorev1.Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: kkmachine.Name + "-",
			Namespace:    scope.Namespace,
			Labels: map[string]string{
				clusterv1beta1.ClusterNameLabel: scope.Name,
			},
		},
		Spec: kkcorev1.PipelineSpec{
			Project: kkcorev1.PipelineProject{
				Addr: _const.CAPKKProjectdir,
			},
			Playbook:     _const.CAPKKPlaybookAddNode,
			InventoryRef: util.ObjectRef(r.Client, scope.Inventory),
			Config:       ptr.Deref(config, kkcorev1.Config{}),
			VolumeMounts: volumeMounts,
			Volumes:      volumes,
		},
	}
	if err := ctrl.SetControllerReference(kkmachine, pipeline, r.Client.Scheme()); err != nil {
		return err
	}
	if err := r.Client.Create(ctx, pipeline); err != nil {
		return err
	}
	// add pipeline name to kkmachine
	kkmachine.Annotations[capkkinfrav1beta1.AddNodePipelineAnnotation] = pipeline.Name

	return nil
}

func (r *KKMachineReconciler) createDeleteNodePipeline(ctx context.Context, scope *clusterScope, kkmachine *capkkinfrav1beta1.KKMachine) error {
	if ok, _ := scope.ifPipelineCompleted(ctx, kkmachine); !ok {
		return nil
	}
	config, err := r.getConfig(scope, kkmachine)
	if err != nil {
		klog.ErrorS(err, "get default config error, use default config", "kubeVersion", kkmachine.Spec.Version)
	}
	volumes, volumeMounts := scope.getVolumeMounts(ctx)
	pipeline := &kkcorev1.Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: kkmachine.Name + "-",
			Namespace:    scope.Namespace,
			Labels: map[string]string{
				clusterv1beta1.ClusterNameLabel: scope.Name,
			},
		},
		Spec: kkcorev1.PipelineSpec{
			Project: kkcorev1.PipelineProject{
				Addr: _const.CAPKKProjectdir,
			},
			Playbook:     _const.CAPKKPlaybookDeleteNode,
			InventoryRef: util.ObjectRef(r.Client, scope.Inventory),
			Config:       *config,
			VolumeMounts: volumeMounts,
			Volumes:      volumes,
		},
	}
	if err := ctrl.SetControllerReference(kkmachine, pipeline, r.Client.Scheme()); err != nil {
		return err
	}
	if err := r.Client.Create(ctx, pipeline); err != nil {
		return err
	}
	kkmachine.Annotations[capkkinfrav1beta1.DeleteNodePipelineAnnotation] = pipeline.Name

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
			return config, fmt.Errorf("read default config file error: %w", err)
		}
		if err := yaml.Unmarshal(data, config); err != nil {
			return config, fmt.Errorf("unmarshal config file error: %w", err)
		}
		klog.InfoS("get default config", "config", config)
	}

	if err := config.SetValue(_const.Workdir, _const.CAPKKWorkdir); err != nil {
		return config, fmt.Errorf("failed to set %q in config error: %w", _const.Workdir, err)
	}
	if err := config.SetValue("node_name", _const.ProviderID2Host(scope.Name, kkmachine.Spec.ProviderID)); err != nil {
		return config, fmt.Errorf("failed to set \"node_name\" in config error: %w", err)
	}
	if err := config.SetValue("kube_version", kkmachine.Spec.Version); err != nil {
		return config, fmt.Errorf("failed to set \"kube_version\" in config error: %w", err)
	}
	if err := config.SetValue("kubernetes.cluster_name", scope.Cluster.Name); err != nil {
		return config, fmt.Errorf("failed to set \"kubernetes.cluster_name\" in config error: %w", err)
	}
	if err := config.SetValue("kubernetes.roles", kkmachine.Spec.Roles); err != nil {
		return config, fmt.Errorf("failed to set \"kubernetes.roles\" in config error: %w", err)
	}
	if err := config.SetValue("cluster_network", scope.Cluster.Spec.ClusterNetwork); err != nil {
		return config, fmt.Errorf("failed to set \"cluster_network\" in config error: %w", err)
	}

	switch scope.KKCluster.Spec.ControlPlaneEndpointType {
	case capkkinfrav1beta1.ControlPlaneEndpointTypeVIP:
		// should set vip addr to config
		if err := config.SetValue("kubernetes.control_plane_endpoint.kube_vip.address", scope.Cluster.Spec.ControlPlaneEndpoint.Host); err != nil {
			return config, fmt.Errorf("failed to set \"kubernetes.control_plane_endpoint.kube_vip.address\" in config error: %w", err)
		}
	case capkkinfrav1beta1.ControlPlaneEndpointTypeDNS:
		// do nothing
	default:
		return config, errors.New("unsupport ControlPlaneEndpointType")
	}
	if err := config.SetValue("kubernetes.control_plane_endpoint.host", scope.Cluster.Spec.ControlPlaneEndpoint.Host); err != nil {
		return config, fmt.Errorf("failed to set \"kubernetes.kube_vip.address\" in config error: %w", err)
	}
	if err := config.SetValue("kubernetes.control_plane_endpoint.type", scope.KKCluster.Spec.ControlPlaneEndpointType); err != nil {
		return config, fmt.Errorf("failed to set \"kubernetes.kube_vip.enabled\" in config error: %w", err)
	}

	return config, nil
}
