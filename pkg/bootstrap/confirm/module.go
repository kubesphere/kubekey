package confirm

import (
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/module"
)

type InstallConfirmModule struct {
	common.KubeModule
	Skip bool
}

func (i *InstallConfirmModule) IsSkip() bool {
	return i.Skip
}

func (i *InstallConfirmModule) Init() {
	i.Name = "ConfirmModule"
	i.Desc = "display confirmation form"

	display := &module.LocalTask{
		Name:   "ConfirmForm",
		Desc:   "Display confirmation form",
		Action: new(InstallationConfirm),
	}

	i.Tasks = []module.Task{
		display,
	}
}

type DeleteClusterConfirmModule struct {
	common.KubeModule
}

func (d *DeleteClusterConfirmModule) Init() {
	d.Name = "DeleteClusterConfirmModule"
	d.Desc = "Display delete confirmation form"

	display := &module.LocalTask{
		Name:   "ConfirmForm",
		Desc:   "Display confirmation form",
		Action: &DeleteConfirm{Content: "cluster"},
	}

	d.Tasks = []module.Task{
		display,
	}
}

type DeleteNodeConfirmModule struct {
	common.KubeModule
}

func (d *DeleteNodeConfirmModule) Init() {
	d.Name = "DeleteNodeConfirmModule"
	d.Desc = "Display delete node confirmation form"

	display := &module.LocalTask{
		Name:   "ConfirmForm",
		Desc:   "Display confirmation form",
		Action: &DeleteConfirm{Content: "node"},
	}

	d.Tasks = []module.Task{
		display,
	}
}

type UpgradeConfirmModule struct {
	common.KubeModule
	Skip bool
}

func (u *UpgradeConfirmModule) IsSkip() bool {
	return u.Skip
}

func (u *UpgradeConfirmModule) Init() {
	u.Name = "UpgradeConfirmModule"
	u.Desc = "Display upgrade confirmation form"

	display := &module.LocalTask{
		Name:   "ConfirmForm",
		Desc:   "Display confirmation form",
		Action: new(UpgradeConfirm),
	}

	u.Tasks = []module.Task{
		display,
	}
}
