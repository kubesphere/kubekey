package binaries

import (
	"github.com/kubesphere/kubekey/pkg/core/modules"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
)

type NodeBinariesModule struct {
	common.KubeModule
}

func (n *NodeBinariesModule) Init() {
	n.Name = "NodeBinariesModule"

	download := &modules.LocalTask{
		Name:   "DownloadBinaries",
		Desc:   "Download installation binaries",
		Action: new(Download),
	}

	n.Tasks = []modules.Task{
		download,
	}
}

//func (n *NodeBinariesModule) Run() error {
//	cfg := n.KubeConf.Cluster
//
//	var kubeVersion string
//	if cfg.Kubernetes.Version == "" {
//		kubeVersion = kubekeyapiv1alpha1.DefaultKubeVersion
//	} else {
//		kubeVersion = cfg.Kubernetes.Version
//	}
//
//	archMap := make(map[string]bool)
//	for _, host := range n.KubeConf.Cluster.Hosts {
//		switch host.Arch {
//		case "amd64":
//			archMap["amd64"] = true
//		case "arm64":
//			archMap["arm64"] = true
//		default:
//			return errors.New(fmt.Sprintf("Unsupported architecture: %s", host.Arch))
//		}
//	}
//
//	for arch := range archMap {
//		binariesDir := filepath.Join(n.Runtime.GetWorkDir(), kubeVersion, arch)
//		if err := util.CreateDir(binariesDir); err != nil {
//			return errors.Wrap(err, "Failed to create download target dir")
//		}
//
//		switch cfg.Kubernetes.Type {
//		case "k3s":
//			if err := K3sFilesDownloadHTTP(n.KubeConf, binariesDir, kubeVersion, arch); err != nil {
//				return err
//			}
//		default:
//			if err := K8sFilesDownloadHTTP(n.KubeConf, binariesDir, kubeVersion, arch, n.PipelineCache); err != nil {
//				return err
//			}
//		}
//	}
//	return nil
//}
