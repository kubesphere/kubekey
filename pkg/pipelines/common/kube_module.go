package common

import (
	"github.com/kubesphere/kubekey/pkg/core/modules"
)

type KubeModule struct {
	modules.BaseTaskModule
	KubeConf *KubeRuntime
}

func (k *KubeModule) AutoAssert() {
	conf := k.Runtime.(*KubeRuntime)
	k.KubeConf = conf
}

type KubeCustomModule struct {
	modules.CustomModule
	KubeConf *KubeRuntime
}

func (k *KubeCustomModule) AutoAssert() {
	conf := k.Runtime.(*KubeRuntime)
	k.KubeConf = conf
}
