package images

import (
	"github.com/kubesphere/kubekey/pkg/core/modules"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
)

type ImageModule struct {
	common.KubeModule
	Skip bool
}

func (i *ImageModule) IsSkip() bool {
	return i.Skip
}

func (i *ImageModule) Init() {
	i.Name = "ImageModule"

	pull := &modules.RemoteTask{
		Name:     "PullImages",
		Desc:     "start to pull images on all nodes",
		Hosts:    i.Runtime.GetAllHosts(),
		Action:   new(PullImage),
		Parallel: true,
	}

	i.Tasks = []modules.Task{
		pull,
	}
}
