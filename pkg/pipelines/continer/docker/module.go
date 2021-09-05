package docker

import (
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/modules"
	"github.com/kubesphere/kubekey/pkg/core/prepare"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"github.com/kubesphere/kubekey/pkg/pipelines/continer/docker/templates"
)

type DockerModule struct {
	common.KubeModule
	Skip bool
}

func (d *DockerModule) IsSkip() bool {
	return d.Skip
}

func (d *DockerModule) Init() {
	d.Name = "DockerModule"

	install := &modules.Task{
		Name:     "InstallDocker",
		Desc:     "install docker",
		Hosts:    d.Runtime.GetAllHosts(),
		Action:   new(InstallDocker),
		Parallel: true,
	}

	generateConfig := &modules.Task{
		Name:  "GenerateDockerConfig",
		Desc:  "generate docker config",
		Hosts: d.Runtime.GetAllHosts(),
		Prepare: &prepare.FileExist{
			FilePath: "/etc/docker/daemon.json",
			Not:      true,
		},
		Action: &action.Template{
			Template: templates.DockerConfigTempl,
			Dst:      "/etc/docker/daemon.json",
			Data: util.Data{
				"Mirrors":            templates.Mirrors(d.KubeConf),
				"InsecureRegistries": templates.InsecureRegistries(d.KubeConf),
			},
		},
		Parallel: true,
	}

	reload := &modules.Task{
		Name:     "ReloadDockerConfig",
		Desc:     "reload docker config",
		Hosts:    d.Runtime.GetAllHosts(),
		Action:   new(ReloadDocker),
		Parallel: true,
	}

	d.Tasks = []*modules.Task{
		install,
		generateConfig,
		reload,
	}
}
