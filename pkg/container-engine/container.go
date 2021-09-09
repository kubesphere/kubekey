package container_engine

import (
	"fmt"
	kubekeycontroller "github.com/kubesphere/kubekey/controllers/kubekey"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func InstallerContainerRuntime(mgr *manager.Manager) error {
	mgr.Logger.Infoln("Installing Container Runtime ...")
	switch strings.TrimSpace(mgr.Cluster.Kubernetes.ContainerManager) {
	case "docker", "":
		if err := mgr.RunTaskOnAllNodes(installDockerOnNode, true); err != nil {
			return err
		}
	case "containerd":
		if err := mgr.RunTaskOnAllNodes(installContainerdOnNode, true); err != nil {
			return err
		}
	case "crio":
		// TODO: Add the steps of cri-o's installation.
	case "isula":
		// TODO: Add the steps of iSula's installation.
	default:
		return errors.New(fmt.Sprintf("Unsupported container runtime: %s", strings.TrimSpace(mgr.Cluster.Kubernetes.ContainerManager)))
	}

	if mgr.InCluster {
		if err := kubekeycontroller.UpdateClusterConditions(mgr, "Init nodes", mgr.Conditions[0].StartTime, metav1.Now(), true, 1); err != nil {
			return err
		}
	}

	return nil
}
