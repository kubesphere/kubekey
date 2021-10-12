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

type K3sNodeBinariesModule struct {
	common.KubeModule
}

func (k *K3sNodeBinariesModule) Init() {
	k.Name = "K3sNodeBinariesModule"

	download := &modules.LocalTask{
		Name:   "DownloadBinaries",
		Desc:   "Download installation binaries",
		Action: new(K3sDownload),
	}

	k.Tasks = []modules.Task{
		download,
	}
}
