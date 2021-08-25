package binaries

import "github.com/kubesphere/kubekey/pkg/core/modules"

type NodeBinariesModule struct {
	modules.BaseTaskModule
}

func (n *NodeBinariesModule) Init() {
	n.Name = "NodeBinariesModule"

}
