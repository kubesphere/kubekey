package infrastructure

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"os"

	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/util"
)

const (
	defaultGroupControlPlane = "kube_control_plane"
	defaultGroupWorker       = "kube_worker"
	defaultClusterGroup      = "k8s_cluster"
)

func getControlPlaneGroupName() string {
	groupName := os.Getenv(_const.ENV_CAPKK_GROUP_CONTROLPLANE)
	if groupName == "" {
		groupName = defaultGroupControlPlane
	}

	return groupName
}

func getWorkerGroupName() string {
	groupName := os.Getenv(_const.ENV_CAPKK_GROUP_WORKER)
	if groupName == "" {
		groupName = defaultGroupWorker
	}

	return groupName
}

func getVolumeMountsFromEnv() ([]corev1.Volume, []corev1.VolumeMount) {
	volumeMounts := make([]corev1.VolumeMount, 0)
	volumes := make([]corev1.Volume, 0)

	if binaryPVC := os.Getenv(_const.ENV_CAPKK_VOLUME_BINARY); binaryPVC != "" {
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
	if projectPVC := os.Getenv(_const.ENV_CAPKK_VOLUME_PROJECT); projectPVC != "" {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "project",
			MountPath: _const.CAPKKProjectdir,
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

	return volumes, volumeMounts
}

// checkIfPipelineCompleted determines if all pipelines associated with the given owner are completed.
// At any given time, there should be at most one pipeline running for each owner.
func checkIfPipelineCompleted(ctx context.Context, scheme *runtime.Scheme, client ctrlclient.Client, owner ctrlclient.Object) (bool, error) {
	pipelineList := &kkcorev1.PipelineList{}
	if err := util.GetObjectListFromOwner(ctx, scheme, client, owner, pipelineList); err != nil {
		return false, err
	}
	for _, pipeline := range pipelineList.Items {
		if pipeline.Status.Phase != kkcorev1.PipelinePhaseFailed && pipeline.Status.Phase != kkcorev1.PipelinePhaseSucceeded {
			return false, nil
		}
	}

	return true, nil
}

// kubeVersionConfigs is the default config for each kube_version
//
//go:embed versions
var kubeVersionConfigs embed.FS

// getDefaultConfig get default config by kubeVersion.
func getDefaultConfig(kubeVersion string) (*kkcorev1.Config, error) {
	config := &kkcorev1.Config{}
	if kubeVersion == "" {
		return config, errors.New("kubeVersion or config is empty")
	}

	data, err := kubeVersionConfigs.ReadFile(fmt.Sprintf("versions/%s.yaml", kubeVersion))
	if err != nil {
		return config, fmt.Errorf("read default config file error: %w", err)
	}
	if err := yaml.Unmarshal(data, config); err != nil {
		return config, fmt.Errorf("unmarshal config file error: %w", err)
	}

	return config, nil
}
