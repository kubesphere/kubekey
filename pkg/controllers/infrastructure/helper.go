package infrastructure

import (
	"context"

	"github.com/cockroachdb/errors"
	capkkinfrav1beta1 "github.com/kubesphere/kubekey/api/capkk/infrastructure/v1beta1"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/utils/ptr"
	kubeadmv1beta2 "sigs.k8s.io/cluster-api/api/controlplane/kubeadm/v1beta2"
	clusterv1beta1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	clusterannotations "sigs.k8s.io/cluster-api/util/annotations"
	"sigs.k8s.io/cluster-api/util/secret"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/controllers/util"
)

const (
	defaultGroupWorker  = "kube_worker"
	defaultClusterGroup = "k8s_cluster"
)

func getControlPlaneGroupName() string {
	return _const.Getenv(_const.CapkkGroupControlPlane)
}

func getWorkerGroupName() string {
	return _const.Getenv(_const.CapkkGroupWorker)
}

// the cluster resource in kubernetes. only contains the single resource for cluster.
type clusterScope struct {
	client ctrlclient.Client

	reconcile.Request
	Cluster           *clusterv1beta1.Cluster
	ControlPlane      *unstructured.Unstructured
	MachineDeployment *clusterv1beta1.MachineDeployment
	KKCluster         *capkkinfrav1beta1.KKCluster
	// Optional
	Inventory *kkcorev1.Inventory
	// Optional
	*util.PatchHelper
}

func newClusterScope(ctx context.Context, client ctrlclient.Client, clusterReq reconcile.Request) (*clusterScope, error) {
	var scope = &clusterScope{
		client:            client,
		Request:           clusterReq,
		Cluster:           &clusterv1beta1.Cluster{},
		ControlPlane:      &unstructured.Unstructured{},
		MachineDeployment: &clusterv1beta1.MachineDeployment{},
		KKCluster:         &capkkinfrav1beta1.KKCluster{},
		Inventory:         &kkcorev1.Inventory{},
	}
	// Cluster
	scope.Cluster = &clusterv1beta1.Cluster{}
	if err := client.Get(ctx, scope.NamespacedName, scope.Cluster); err != nil {
		// must hve scope
		return scope, errors.Wrapf(err, "failed to get cluster with scope %q", scope.String())
	}
	// KKCluster
	if err := client.Get(ctx, ctrlclient.ObjectKey{
		Namespace: scope.Cluster.GetNamespace(),
		Name:      scope.Cluster.Spec.InfrastructureRef.Name,
	}, scope.KKCluster); err != nil {
		return scope, errors.Wrapf(err, "failed to get kkcluster with scope %q", scope.String())
	}
	// ControlPlane
	scope.ControlPlane.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   scope.Cluster.Spec.ControlPlaneRef.APIGroup,
		Version: kubeadmv1beta2.GroupVersion.Version,
		Kind:    scope.Cluster.Spec.ControlPlaneRef.Kind,
	})
	if err := client.Get(ctx, ctrlclient.ObjectKey{
		Namespace: scope.Cluster.GetNamespace(),
		Name:      scope.Cluster.Spec.ControlPlaneRef.Name,
	}, scope.ControlPlane); err != nil && !apierrors.IsNotFound(err) {
		return scope, errors.Wrapf(err, "failed to get control-plane with scope %q", scope.String())
	}
	// MachineDeployment
	mdlist := &clusterv1beta1.MachineDeploymentList{}
	if err := util.GetObjectListFromOwner(ctx, client, scope.Cluster, mdlist); err == nil && len(mdlist.Items) == 1 {
		scope.MachineDeployment = ptr.To(mdlist.Items[0])
	}
	// inventory
	invlist := &kkcorev1.InventoryList{}
	if err := util.GetObjectListFromOwner(ctx, client, scope.KKCluster, invlist); err == nil && len(invlist.Items) == 1 {
		scope.Inventory = ptr.To(invlist.Items[0])
	}

	return scope, nil
}

func (p *clusterScope) newPatchHelper(obj ...ctrlclient.Object) error {
	helper, err := util.NewPatchHelper(p.client, obj...)
	if err != nil {
		return err
	}
	p.PatchHelper = helper

	return nil
}

func (p *clusterScope) isPaused() bool {
	return clusterannotations.IsPaused(p.Cluster, p.KKCluster)
}

// checkIfPlaybookCompleted determines if all playbooks associated with the given owner are completed.
// At any given time, there should be at most one playbook running for each owner.
func (p *clusterScope) ifPlaybookCompleted(ctx context.Context, owner ctrlclient.Object) (bool, error) {
	playbookList := &kkcorev1.PlaybookList{}
	if err := util.GetObjectListFromOwner(ctx, p.client, owner, playbookList); err != nil {
		return false, err
	}
	for _, playbook := range playbookList.Items {
		if playbook.Status.Phase != kkcorev1.PlaybookPhaseFailed && playbook.Status.Phase != kkcorev1.PlaybookPhaseSucceeded {
			return false, nil
		}
	}

	return true, nil
}

func (p *clusterScope) getVolumeMounts(ctx context.Context) ([]corev1.Volume, []corev1.VolumeMount) {
	volumeMounts := make([]corev1.VolumeMount, 0)
	volumes := make([]corev1.Volume, 0)

	if binaryPVC := _const.Getenv(_const.CapkkVolumeBinary); binaryPVC != "" {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "kubekey",
			MountPath: _const.CAPKKBinarydir,
		})
		volumes = append(volumes, corev1.Volume{
			Name: "kubekey",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: binaryPVC,
				},
			},
		})
	}
	if projectPVC := _const.Getenv(_const.CapkkVolumeProject); projectPVC != "" {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "project",
			MountPath: _const.CAPKKProjectdir,
			ReadOnly:  true,
		})
		volumes = append(volumes, corev1.Volume{
			Name: "project",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: projectPVC,
				},
			},
		})
	}
	if workdirPVC := _const.Getenv(_const.CapkkVolumeWorkdir); workdirPVC != "" {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "workdir",
			MountPath: _const.CAPKKWorkdir,
		})
		volumes = append(volumes, corev1.Volume{
			Name: "workdir",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: workdirPVC,
				},
			},
		})
	}

	// mount ssh privatekey
	if sshName, ok := p.KKCluster.Annotations[capkkinfrav1beta1.KKClusterSSHPrivateKeyAnnotation]; ok {
		if sshName == "" {
			sshName = secret.Name(p.Cluster.Name, "ssh")
		}
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "ssh",
			MountPath: "/root/.ssh",
			ReadOnly:  true,
		})
		volumes = append(volumes, corev1.Volume{
			Name: "ssh",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: sshName,
				},
			},
		})
	}

	if err := p.client.Get(ctx, ctrlclient.ObjectKey{
		Namespace: p.Namespace,
		Name:      secret.Name(p.Cluster.Name, secret.Kubeconfig),
	}, &corev1.Secret{}); err == nil {
		// mount kubeconfig
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "kubeconfig",
			MountPath: _const.CAPKKCloudKubeConfigPath,
			ReadOnly:  true,
		})
		volumes = append(volumes, corev1.Volume{
			Name: "kubeconfig",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: secret.Name(p.Cluster.Name, secret.Kubeconfig),
				},
			},
		})
	}

	return volumes, volumeMounts
}
