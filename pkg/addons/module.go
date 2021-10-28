package addons

import (
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/task"
)

type AddonsModule struct {
	common.KubeModule
	Skip bool
}

func (a *AddonsModule) IsSkip() bool {
	return a.Skip
}

func (a *AddonsModule) Init() {
	a.Name = "AddonsModule"
	a.Desc = "Install addons"

	install := &task.LocalTask{
		Name:   "InstallAddons",
		Desc:   "Install addons",
		Action: new(Install),
	}

	a.Tasks = []task.Interface{
		install,
	}
}
