package binaries

import (
	"fmt"
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
)

type NodeBinariesModule struct {
	common.KubeCustomModule
}

func (n *NodeBinariesModule) Init() {
	n.Name = "NodeBinariesModule"
	n.Desc = "Download Installation Files"
}

func (n *NodeBinariesModule) Run() error {
	cfg := n.KubeConf.Cluster
	currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return errors.Wrap(err, "Failed to get current directory")
	}

	var kubeVersion string
	if cfg.Kubernetes.Version == "" {
		kubeVersion = kubekeyapiv1alpha1.DefaultKubeVersion
	} else {
		kubeVersion = cfg.Kubernetes.Version
	}

	archMap := make(map[string]bool)
	for _, host := range n.KubeConf.Cluster.Hosts {
		switch host.Arch {
		case "amd64":
			archMap["amd64"] = true
		case "arm64":
			archMap["arm64"] = true
		default:
			return errors.New(fmt.Sprintf("Unsupported architecture: %s", host.Arch))
		}
	}

	for arch := range archMap {
		binariesDir := fmt.Sprintf("%s/%s/%s/%s", currentDir, kubekeyapiv1alpha1.DefaultPreDir, kubeVersion, arch)
		if err := util.CreateDir(binariesDir); err != nil {
			return errors.Wrap(err, "Failed to create download target dir")
		}

		switch cfg.Kubernetes.Type {
		case "k3s":
			if err := K3sFilesDownloadHTTP(n.KubeConf, binariesDir, kubeVersion, arch); err != nil {
				return err
			}
		default:
			if err := K8sFilesDownloadHTTP(n.KubeConf, binariesDir, kubeVersion, arch); err != nil {
				return err
			}
		}
	}
	return nil
}
