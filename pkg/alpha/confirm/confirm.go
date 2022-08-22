package confirm

import (
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/task"
)

type UpgradeKsConfirmModule struct {
	common.KubeModule
	Skip bool
}

func (u *UpgradeKsConfirmModule) IsSkip() bool {
	return u.Skip
}

func (u *UpgradeKsConfirmModule) Init() {
	u.Name = "UpgradeKsConfirmModule"
	u.Desc = "Display upgrade kubesphere confirmation form"

	display := &task.LocalTask{
		Name:   "ConfirmForm",
		Desc:   "Display confirmation form",
		Action: new(UpgradeConfirm),
	}

	u.Tasks = []task.Interface{
		display,
	}
}
