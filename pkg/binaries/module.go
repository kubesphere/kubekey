package binaries

import (
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/module"
)

type NodeBinariesModule struct {
	common.KubeModule
}

func (n *NodeBinariesModule) Init() {
	n.Name = "NodeBinariesModule"
	n.Desc = "Download installation binaries"

	download := &module.LocalTask{
		Name:   "DownloadBinaries",
		Desc:   "Download installation binaries",
		Action: new(Download),
	}

	n.Tasks = []module.Task{
		download,
	}
}

type K3sNodeBinariesModule struct {
	common.KubeModule
}

func (k *K3sNodeBinariesModule) Init() {
	k.Name = "K3sNodeBinariesModule"
	k.Desc = "Download installation binaries"

	download := &module.LocalTask{
		Name:   "DownloadBinaries",
		Desc:   "Download installation binaries",
		Action: new(K3sDownload),
	}

	k.Tasks = []module.Task{
		download,
	}
}
