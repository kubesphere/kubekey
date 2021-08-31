package common

import "github.com/kubesphere/kubekey/pkg/core/action"

type KubeAction struct {
	action.BaseAction
	KubeConf *KubeConf
}

func (k *KubeAction) AutoAssert() {
	kubeRuntime := k.Runtime.(*KubeRuntime)
	conf := &KubeConf{
		ClusterHosts: kubeRuntime.ClusterHosts,
		Cluster:      kubeRuntime.Cluster,
		Kubeconfig:   kubeRuntime.Kubeconfig,
		Conditions:   kubeRuntime.Conditions,
		ClientSet:    kubeRuntime.ClientSet,
		Arg:          kubeRuntime.Arg,
	}

	k.KubeConf = conf
}
