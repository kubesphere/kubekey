package images

import (
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/module"
)

type PullModule struct {
	common.KubeModule
	Skip bool
}

func (p *PullModule) IsSkip() bool {
	return p.Skip
}

func (p *PullModule) Init() {
	p.Name = "PullModule"
	p.Desc = "Pull images on all nodes"

	pull := &module.RemoteTask{
		Name:     "PullImages",
		Desc:     "Start to pull images on all nodes",
		Hosts:    p.Runtime.GetAllHosts(),
		Action:   new(PullImage),
		Parallel: true,
	}

	p.Tasks = []module.Task{
		pull,
	}
}
