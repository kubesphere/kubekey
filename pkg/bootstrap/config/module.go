package config

import (
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/task"
)

type ModifyConfigModule struct {
	common.KubeModule
}

func (m *ModifyConfigModule) Init() {
	m.Name = "ModifyConfigModule"
	m.Desc = "Modify the KubeKey config file"

	modify := &task.LocalTask{
		Name:   "ModifyConfig",
		Desc:   "Modify the KubeKey config file",
		Action: new(ModifyConfig),
	}

	m.Tasks = []task.Interface{
		modify,
	}
}
