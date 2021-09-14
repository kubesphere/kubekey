package precheck

import (
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/kubesphere/kubekey/pkg/core/modules"
	"github.com/kubesphere/kubekey/pkg/core/prepare"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"github.com/pkg/errors"
)

type NodePreCheckModule struct {
	common.KubeModule
	Skip bool
}

func (n *NodePreCheckModule) IsSkip() bool {
	return n.Skip
}

func (n *NodePreCheckModule) Init() {
	n.Name = "NodePreCheckModule"

	preCheck := &modules.Task{
		Name:  "NodePreCheck",
		Desc:  "a pre-check on nodes",
		Hosts: n.Runtime.GetAllHosts(),
		Prepare: &prepare.FastPrepare{
			Inject: func(runtime connector.Runtime) (bool, error) {
				if len(n.Runtime.GetHostsByRole(common.ETCD))%2 == 0 {
					logger.Log.Error("The number of etcd is even. Please configure it to be odd.")
					return false, errors.New("the number of etcd is even")
				}
				return true, nil
			}},
		Action:   new(NodePreCheck),
		Parallel: true,
	}

	n.Tasks = []*modules.Task{
		preCheck,
	}
}
