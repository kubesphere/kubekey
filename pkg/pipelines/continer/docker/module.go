package docker

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/modules"
	"github.com/kubesphere/kubekey/pkg/core/prepare"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"github.com/kubesphere/kubekey/pkg/pipelines/continer/docker/templates"
	"strings"
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
				"Mirrors":            mirrors(d.KubeConf),
				"InsecureRegistries": insecureRegistries(d.KubeConf),
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

func mirrors(kubeConf *common.KubeConf) string {
	var mirrors string
	if kubeConf.Cluster.Registry.RegistryMirrors != nil {
		var mirrorsArr []string
		for _, mirror := range kubeConf.Cluster.Registry.RegistryMirrors {
			mirrorsArr = append(mirrorsArr, fmt.Sprintf("\"%s\"", mirror))
		}
		mirrors = strings.Join(mirrorsArr, ", ")
	}
	return mirrors
}

func insecureRegistries(kubeConf *common.KubeConf) string {
	var insecureRegistries string
	if kubeConf.Cluster.Registry.InsecureRegistries != nil {
		var registriesArr []string
		for _, repo := range kubeConf.Cluster.Registry.InsecureRegistries {
			registriesArr = append(registriesArr, fmt.Sprintf("\"%s\"", repo))
		}
		insecureRegistries = strings.Join(registriesArr, ", ")
	}
	return insecureRegistries
}
