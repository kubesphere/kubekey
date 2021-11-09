package common

import (
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/connector"
)

type KubeAction struct {
	action.BaseAction
	KubeConf *KubeConf
}

func (k *KubeAction) AutoAssert(runtime connector.Runtime) {
	kubeRuntime := runtime.(*KubeRuntime)
	conf := &KubeConf{
		ClusterHosts: kubeRuntime.ClusterHosts,
		Cluster:      kubeRuntime.Cluster,
		ClusterName:  kubeRuntime.ClusterName,
		Kubeconfig:   kubeRuntime.Kubeconfig,
		ClientSet:    kubeRuntime.ClientSet,
		Arg:          kubeRuntime.Arg,
	}

	k.KubeConf = conf
}
