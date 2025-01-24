package infrastructure

import (
	"context"
	"errors"
	"fmt"

	capkkinfrav1beta1 "github.com/kubesphere/kubekey/api/capkk/infrastructure/v1beta1"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	clusterv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	clusterutil "sigs.k8s.io/cluster-api/util"
	clusterannotations "sigs.k8s.io/cluster-api/util/annotations"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/cluster-api/util/secret"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	ctrlcontroller "sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/util"
)

const (
	kkMachineControllerName = "kkmachine"
)

// KKMachineReconciler reconciles a KKMachine object.
// One KKMachine should have one Pipeline running in time.
type KKMachineReconciler struct {
	*runtime.Scheme
	ctrlclient.Client
	record.EventRecorder
}

// Name implements controllers.controller.
func (r *KKMachineReconciler) Name() string {
	return kkMachineControllerName
}

// SetupWithManager implements controllers.controller.
func (r *KKMachineReconciler) SetupWithManager(mgr ctrl.Manager, o ctrlcontroller.Options) error {
	r.Scheme = mgr.GetScheme()
	r.Client = mgr.GetClient()
	r.EventRecorder = mgr.GetEventRecorderFor(r.Name())

	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(o).
		For(&capkkinfrav1beta1.KKMachine{}).
		Complete(r)
}

// Reconcile implements controllers.controller.
func (r *KKMachineReconciler) Reconcile(ctx context.Context, req reconcile.Request) (_ reconcile.Result, retErr error) {
	kkmachine := &capkkinfrav1beta1.KKMachine{}
	err := r.Client.Get(ctx, req.NamespacedName, kkmachine)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	helper, err := patch.NewHelper(kkmachine, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}
	defer func() {
		if err := helper.Patch(ctx, kkmachine); err != nil {
			retErr = errors.Join(retErr, err)
		}
	}()

	// Fetch KKCluster.
	kkcluster := &capkkinfrav1beta1.KKCluster{}
	if err := util.GetOwnerFromObject(ctx, r.Scheme, r.Client, kkmachine, kkcluster); err != nil {
		return reconcile.Result{}, err
	}

	// Fetch the Cluster.
	cluster, err := clusterutil.GetOwnerCluster(ctx, r.Client, kkcluster.ObjectMeta)
	if err != nil {
		return reconcile.Result{}, err
	}
	if cluster == nil {
		klog.V(5).InfoS("Cluster has not yet set OwnerRef")

		return reconcile.Result{}, nil
	}

	// skip if cluster is paused.
	if clusterannotations.IsPaused(cluster, kkmachine) {
		klog.InfoS("cluster or kkmachine is marked as paused. Won't reconcile")

		return reconcile.Result{}, nil
	}

	// Add finalizer first if not set to avoid the race condition between init and delete.
	// Note: Finalizers in general can only be added when the deletionTimestamp is not set.
	if kkmachine.ObjectMeta.DeletionTimestamp.IsZero() && !controllerutil.ContainsFinalizer(kkmachine, capkkinfrav1beta1.KKMachineFinalizer) {
		controllerutil.AddFinalizer(kkmachine, capkkinfrav1beta1.KKMachineFinalizer)

		return ctrl.Result{}, nil
	}

	// Handle deleted clusters
	if !kkmachine.DeletionTimestamp.IsZero() {
		return reconcile.Result{}, r.reconcileDelete(ctx, kkmachine, kkcluster, cluster)
	}

	// Handle non-deleted clusters
	return reconcile.Result{}, r.reconcileNormal(ctx, kkmachine, kkcluster, cluster)
}

func (r *KKMachineReconciler) reconcileDelete(ctx context.Context, kkmachine *capkkinfrav1beta1.KKMachine, kkcluster *capkkinfrav1beta1.KKCluster, cluster *clusterv1beta1.Cluster) error {
	// deal orphan delete_node pipeline
	pipelineList := &kkcorev1.PipelineList{}
	if err := util.GetObjectListFromOwner(ctx, r.Scheme, r.Client, kkmachine, pipelineList, ctrlclient.MatchingFields{
		kkcorev1.PipelineFieldPlaybook: _const.CAPKKPlaybookDeleteNode,
	}); err != nil {
		return err
	}
	for _, p := range pipelineList.Items {
		pipeline := ptr.To(p)
		if kkmachine.Annotations[capkkinfrav1beta1.DeleteNodePipelineAnnotation] == pipeline.Name {
			continue
		}
		if (pipeline.Status.Phase == kkcorev1.PipelinePhaseSucceeded || pipeline.Status.Phase == kkcorev1.PipelinePhaseFailed) &&
			controllerutil.ContainsFinalizer(pipeline, capkkinfrav1beta1.PipelineCompletedFinalizer) {
			// remove pipeline finalizer
			pipeline := pipeline.DeepCopy()
			controllerutil.RemoveFinalizer(pipeline, capkkinfrav1beta1.PipelineCompletedFinalizer)
			if err := r.Client.Patch(ctx, pipeline, ctrlclient.MergeFrom(pipeline)); err != nil {
				return err
			}
		}
	}

	if kkmachine.Annotations[capkkinfrav1beta1.DeleteNodePipelineAnnotation] == "" {
		// should create delete_node pipeline
		return r.createDeleteNodePipeline(ctx, kkmachine, kkcluster, cluster)
	}

	deletePipeline := &kkcorev1.Pipeline{}
	if err := r.Client.Get(ctx, ctrlclient.ObjectKey{Namespace: kkmachine.Namespace, Name: kkmachine.Annotations[capkkinfrav1beta1.DeleteNodePipelineAnnotation]}, deletePipeline); err != nil {
		if apierrors.IsNotFound(err) {
			// maybe delete by user. skip
			return nil
		}

		return err
	}

	if deletePipeline.Status.Phase == kkcorev1.PipelinePhaseSucceeded && controllerutil.ContainsFinalizer(deletePipeline, capkkinfrav1beta1.DeleteNodePipelineAnnotation) {
		// remove pipeline finalizer
		pipeline := deletePipeline.DeepCopy()
		controllerutil.RemoveFinalizer(deletePipeline, capkkinfrav1beta1.DeleteNodePipelineAnnotation)
		if err := r.Client.Patch(ctx, pipeline, ctrlclient.MergeFrom(deletePipeline)); err != nil {
			return err
		}
		// remove kkmachine finalizer
		controllerutil.RemoveFinalizer(kkmachine, capkkinfrav1beta1.KKMachineFinalizer)
	}

	return r.Client.Delete(ctx, deletePipeline)
}

// reconcileNormal handles normal reconcile.
// when dataSecret or certificates files changed. KCP will RollingUpdate machine (create new machines to replace old machines)
// so the sync file should contains in add_node pipeline.
func (r *KKMachineReconciler) reconcileNormal(ctx context.Context, kkmachine *capkkinfrav1beta1.KKMachine, kkcluster *capkkinfrav1beta1.KKCluster, cluster *clusterv1beta1.Cluster) error {
	pipelineName := kkmachine.Annotations[capkkinfrav1beta1.AddNodePipelineAnnotation]
	if pipelineName == "" {
		// should create pipeline
		if err := r.createAddNodePipeline(ctx, kkmachine, kkcluster, cluster); err != nil {
			return err
		}
	}
	// check pipeline status
	pipeline := &kkcorev1.Pipeline{}
	if err := r.Client.Get(ctx, ctrlclient.ObjectKey{Namespace: kkmachine.Namespace, Name: pipelineName}, pipeline); err != nil {
		if apierrors.IsNotFound(err) {
			// the pipeline has not found.
			r.EventRecorder.Eventf(kkmachine, corev1.EventTypeWarning, "AddNodeFailed", "add node pipeline:%s not found", pipelineName)

			return nil
		}

		return err
	}

	switch pipeline.Status.Phase {
	case kkcorev1.PipelinePhaseSucceeded:
		// set machine to ready
		kkmachine.Status.Ready = true
	case kkcorev1.PipelinePhaseFailed:
		// set machine to not ready
		kkmachine.Status.Ready = false
		kkmachine.Status.FailureReason = capkkinfrav1beta1.KKMachineFailedReasonAddNodeFailed
		kkmachine.Status.FailureMessage = fmt.Sprintf("add_node pipeline %q run failed", pipelineName)
	}

	return nil
}

func (r *KKMachineReconciler) createAddNodePipeline(ctx context.Context, kkmachine *capkkinfrav1beta1.KKMachine, kkcluster *capkkinfrav1beta1.KKCluster, cluster *clusterv1beta1.Cluster) error {
	if ok, _ := checkIfPipelineCompleted(ctx, r.Scheme, r.Client, kkmachine); !ok {
		return nil
	}
	// add data secret to pipeline.config
	// Fetch owner machine
	machine := &clusterv1beta1.Machine{}
	if err := util.GetOwnerFromObject(ctx, r.Scheme, r.Client, kkmachine, machine); err != nil {
		klog.Warning("cannot get owner machine, waiting")

		return nil
	}
	// todo when install offline. should mount workdir to pipeline.
	volumes, volumeMounts := getVolumeMountsFromEnv()
	// mount kubeconfig
	volumeMounts = append(volumeMounts, corev1.VolumeMount{
		Name:      "kubeconfig",
		MountPath: _const.CAPKKKubeconfigPath,
	})
	volumes = append(volumes, corev1.Volume{
		Name: "kubeconfig",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: secret.Name(cluster.Name, secret.Kubeconfig),
			},
		},
	})
	if machine.Spec.Bootstrap.DataSecretName != nil {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "cloud-config",
			MountPath: _const.CAPKKCloudConfigPath,
		})
		volumes = append(volumes, corev1.Volume{
			Name: "cloud-config",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: *machine.Spec.Bootstrap.DataSecretName,
				},
			},
		})
	}
	if machine.Spec.Bootstrap.DataSecretName != nil {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "cloud-config",
			MountPath: _const.CAPKKCloudConfigPath,
		})
		volumes = append(volumes, corev1.Volume{
			Name: "cloud-config",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: *machine.Spec.Bootstrap.DataSecretName,
				},
			},
		})
	}

	config, err := r.generateConfig(kkmachine, kkcluster, cluster)
	if err != nil {
		klog.ErrorS(err, "get default config error, use default config", "kubeVersion", kkcluster.Spec.KubeVersion)
	}
	if err := config.SetValue("workdir", _const.CAPKKWorkdir); err != nil {
		return fmt.Errorf("failed to set \"workdir\" in config error: %w", err)
	}
	if err := config.SetValue("provider_id", *kkmachine.Spec.ProviderID); err != nil {
		return fmt.Errorf("failed to set \"provider_id\" in config error: %w", err)
	}

	pipeline := &kkcorev1.Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: kkmachine.Name + "-",
			Namespace:    kkmachine.Namespace,
			Annotations: map[string]string{
				capkkinfrav1beta1.DeleteNodePipelineAnnotation: kkmachine.Name,
			},
		},
		Spec: kkcorev1.PipelineSpec{
			Project: kkcorev1.PipelineProject{
				Addr: _const.CAPKKProjectdir,
			},
			Playbook:     _const.CAPKKPlaybookAddNode,
			InventoryRef: util.ObjectRef(r.Scheme, kkmachine),
			Config:       *config,
			VolumeMounts: volumeMounts,
			Volumes:      volumes,
		},
	}
	if err := controllerutil.SetOwnerReference(kkmachine, pipeline, r.Scheme); err != nil {
		return err
	}
	controllerutil.AddFinalizer(pipeline, capkkinfrav1beta1.PipelineCompletedFinalizer)
	if err := r.Create(ctx, pipeline); err != nil {
		return err
	}
	// add pipeline name to kkmachine
	kkmachine.Annotations[capkkinfrav1beta1.AddNodePipelineAnnotation] = pipeline.Name

	return nil
}

func (r *KKMachineReconciler) createDeleteNodePipeline(ctx context.Context, kkmachine *capkkinfrav1beta1.KKMachine, kkcluster *capkkinfrav1beta1.KKCluster, cluster *clusterv1beta1.Cluster) error {
	if ok, _ := checkIfPipelineCompleted(ctx, r.Scheme, r.Client, kkmachine); !ok {
		return nil
	}
	config, err := r.generateConfig(kkmachine, kkcluster, cluster)
	if err != nil {
		klog.ErrorS(err, "get default config error, use default config", "kubeVersion", kkcluster.Spec.KubeVersion)
	}
	volumes, volumeMounts := getVolumeMountsFromEnv()
	// mount kubeconfig
	volumeMounts = append(volumeMounts, corev1.VolumeMount{
		Name:      "kubeconfig",
		MountPath: _const.CAPKKKubeconfigPath,
	})
	volumes = append(volumes, corev1.Volume{
		Name: "kubeconfig",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: secret.Name(cluster.Name, secret.Kubeconfig),
			},
		},
	})
	pipeline := &kkcorev1.Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: kkmachine.Name + "-",
			Namespace:    kkmachine.Namespace,
			Annotations: map[string]string{
				capkkinfrav1beta1.DeleteNodePipelineAnnotation: kkmachine.Name,
			},
		},
		Spec: kkcorev1.PipelineSpec{
			Project: kkcorev1.PipelineProject{
				Addr: _const.CAPKKProjectdir,
			},
			Playbook:     _const.CAPKKPlaybookDeleteNode,
			InventoryRef: util.ObjectRef(r.Scheme, kkmachine),
			Config:       *config,
			VolumeMounts: volumeMounts,
			Volumes:      volumes,
		},
	}
	if err := controllerutil.SetOwnerReference(kkmachine, pipeline, r.Scheme); err != nil {
		return err
	}
	controllerutil.AddFinalizer(pipeline, capkkinfrav1beta1.PipelineCompletedFinalizer)
	if err := r.Create(ctx, pipeline); err != nil {
		return err
	}
	kkmachine.Annotations[capkkinfrav1beta1.DeleteNodePipelineAnnotation] = pipeline.Name

	return nil
}

func (r *KKMachineReconciler) generateConfig(kkmachine *capkkinfrav1beta1.KKMachine, kkcluster *capkkinfrav1beta1.KKCluster, cluster *clusterv1beta1.Cluster) (*kkcorev1.Config, error) {
	var config *kkcorev1.Config
	if kkmachine.Spec.Config.Raw != nil {
		config = &kkcorev1.Config{
			Spec: kkmachine.Spec.Config,
		}
	} else {
		c, err := getDefaultConfig(kkcluster.Spec.KubeVersion)
		if err != nil {
			klog.ErrorS(err, "get default config error, use default config", "kubeVersion", kkcluster.Spec.KubeVersion)
		}
		config = c
	}

	if err := config.SetValue("workdir", _const.CAPKKWorkdir); err != nil {
		return nil, fmt.Errorf("failed to set \"workdir\" in config error: %w", err)
	}
	if err := config.SetValue("provider_id", *kkmachine.Spec.ProviderID); err != nil {
		return nil, fmt.Errorf("failed to set \"provider_id\" in config error: %w", err)
	}

	switch kkcluster.Spec.ControlPlaneEndpointType {
	case "", capkkinfrav1beta1.ControlPlaneEndpointTypeVIP:
		// should set vip addr to config
		if err := config.SetValue("kubernetes.control_plane_endpoint.kube_vip.address", cluster.Spec.ControlPlaneEndpoint.Host); err != nil {
			return nil, fmt.Errorf("failed to set \"kubernetes.kube_vip.address\" in config error: %w", err)
		}
		if err := config.SetValue("kubernetes.control_plane_endpoint.type", "kube_vip"); err != nil {
			return nil, fmt.Errorf("failed to set \"kubernetes.kube_vip.enabled\" in config error: %w", err)
		}
	case capkkinfrav1beta1.ControlPlaneEndpointTypeDNS:
		if err := config.SetValue("kubernetes.control_plane_endpoint.host", cluster.Spec.ControlPlaneEndpoint.Host); err != nil {
			return nil, fmt.Errorf("failed to set \"kubernetes.kube_vip.address\" in config error: %w", err)
		}
		if err := config.SetValue("kubernetes.control_plane_endpoint.type", "dns"); err != nil {
			return nil, fmt.Errorf("failed to set \"kubernetes.kube_vip.enabled\" in config error: %w", err)
		}
	}
	if err := config.SetValue("kubernetes.cluster_name", cluster.Name); err != nil {
		return nil, fmt.Errorf("failed to set \"cluster_name\" in config error: %w", err)
	}
	if err := config.SetValue("kubernetes.roles", kkmachine.Spec.Roles); err != nil {
		return nil, fmt.Errorf("failed to set \"kubernetes.roles\" in config error: %w", err)
	}
	if err := config.SetValue("cluster_network", cluster.Spec.ClusterNetwork); err != nil {
		return nil, fmt.Errorf("failed to set \"cluster_network\" in config error: %w", err)
	}

	return config, nil
}
