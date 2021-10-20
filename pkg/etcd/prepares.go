package etcd

import (
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/pkg/errors"
	"strings"
)

type FirstETCDNode struct {
	common.KubePrepare
	Not bool
}

func (f *FirstETCDNode) PreCheck(runtime connector.Runtime) (bool, error) {
	v, ok := f.PipelineCache.Get(common.ETCDCluster)
	if !ok {
		return false, errors.New("get etcd cluster status by pipeline cache failed")
	}
	cluster := v.(*EtcdCluster)

	if (!cluster.clusterExist && runtime.GetHostsByRole(common.ETCD)[0].GetName() == runtime.RemoteHost().GetName()) ||
		(cluster.clusterExist && strings.Contains(cluster.peerAddresses[0], runtime.RemoteHost().GetInternalAddress())) {
		return !f.Not, nil
	}
	return f.Not, nil
}

type NodeETCDExist struct {
	common.KubePrepare
	Not bool
}

func (n *NodeETCDExist) PreCheck(runtime connector.Runtime) (bool, error) {
	host := runtime.RemoteHost()
	if v, ok := host.GetCache().GetMustBool(common.ETCDExist); ok {
		if v {
			return !n.Not, nil
		} else {
			return n.Not, nil
		}
	} else {
		return false, errors.New("get etcd node status by host label failed")
	}
}
