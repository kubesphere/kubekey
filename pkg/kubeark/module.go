package kubeark

import (
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/task"
)

type KubearkModule struct {
	common.KubeModule
	Skip bool
}

func (k *KubearkModule) IsSkip() bool {
	return k.Skip
}

func (k *KubearkModule) Init() {
	k.Name = "KubearkModule"
	k.Desc = "Install Kubeark"

	kubearkGenerateManifest := &task.RemoteTask{
		Name:     "GenerateKubearkManifest",
		Desc:     "Generate Kubeark manifest at other master",
		Hosts:    k.Runtime.GetHostsByRole(common.Master),
		Action:   new(GenerateKubearkManifest),
		Parallel: true,
	}

	applyKubeark := &task.RemoteTask{
		Name:     "DeployKubeark",
		Desc:     "Deploy Kubeark",
		Hosts:    k.Runtime.GetHostsByRole(common.Master),
		Action:   new(DeployKubeark),
		Parallel: true,
		Retry:    5,
	}

	k.Tasks = []task.Interface{
		kubearkGenerateManifest,
		applyKubeark,
	}

}
