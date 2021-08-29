package common

import "github.com/kubesphere/kubekey/pkg/core/action"

type KubeAction struct {
	action.BaseAction
	KubeConf *KubeRuntime
}

func (k *KubeAction) AutoAssert() {
	conf := k.RuntimeConf.(*KubeRuntime)
	k.KubeConf = conf
}
