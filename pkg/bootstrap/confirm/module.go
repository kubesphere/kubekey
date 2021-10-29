package confirm

import (
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/task"
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
	i.Desc = "Display confirmation form"

	display := &task.LocalTask{
		Name:   "ConfirmForm",
		Desc:   "Display confirmation form",
		Action: new(InstallationConfirm),
	}

	i.Tasks = []task.Interface{
		display,
	}
}

type DeleteClusterConfirmModule struct {
	common.KubeModule
}

func (d *DeleteClusterConfirmModule) Init() {
	d.Name = "DeleteClusterConfirmModule"
	d.Desc = "Display delete confirmation form"

	display := &task.LocalTask{
		Name:   "ConfirmForm",
		Desc:   "Display confirmation form",
		Action: &DeleteConfirm{Content: "cluster"},
	}

	d.Tasks = []task.Interface{
		display,
	}
}

type DeleteNodeConfirmModule struct {
	common.KubeModule
}

func (d *DeleteNodeConfirmModule) Init() {
	d.Name = "DeleteNodeConfirmModule"
	d.Desc = "Display delete node confirmation form"

	display := &task.LocalTask{
		Name:   "ConfirmForm",
		Desc:   "Display confirmation form",
		Action: &DeleteConfirm{Content: "node"},
	}

	d.Tasks = []task.Interface{
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

	display := &task.LocalTask{
		Name:   "ConfirmForm",
		Desc:   "Display confirmation form",
		Action: new(UpgradeConfirm),
	}

	u.Tasks = []task.Interface{
		display,
	}
}
