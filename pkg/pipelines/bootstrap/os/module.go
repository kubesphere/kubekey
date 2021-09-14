package os

import (
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/modules"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/pipelines/bootstrap/os/templates"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"path/filepath"
)

type ConfigureOSModule struct {
	common.KubeModule
}

func (c *ConfigureOSModule) Init() {
	c.Name = "ConfigureOSModule"

	initOS := &modules.Task{
		Name:     "InitOS",
		Desc:     "prepare to init OS",
		Hosts:    c.Runtime.GetAllHosts(),
		Action:   new(NodeConfigureOS),
		Parallel: true,
	}

	GenerateScript := &modules.Task{
		Name:  "GenerateScript",
		Desc:  "generate init os script",
		Hosts: c.Runtime.GetAllHosts(),
		Action: &action.Template{
			Template: templates.InitOsScriptTmpl,
			Dst:      filepath.Join(common.KubeScriptDir, "initOS.sh"),
			Data: util.Data{
				"Hosts": c.KubeConf.ClusterHosts,
			},
		},
		Parallel: true,
	}

	ExecScript := &modules.Task{
		Name:     "ExecScript",
		Desc:     "exec init os script",
		Hosts:    c.Runtime.GetAllHosts(),
		Action:   new(NodeExecScript),
		Parallel: true,
	}

	c.Tasks = []*modules.Task{
		initOS,
		GenerateScript,
		ExecScript,
	}
}

type ClearOSEnvironmentModule struct {
	common.KubeModule
}

func (c *ClearOSEnvironmentModule) Init() {
	c.Name = "ClearOSModule"

	resetNetworkConfig := &modules.Task{
		Name:     "ResetNetworkConfig",
		Desc:     "Reset os network config",
		Hosts:    c.Runtime.GetHostsByRole(common.K8s),
		Action:   new(ResetNetworkConfig),
		Parallel: true,
	}

	stopETCD := &modules.Task{
		Name:     "StopETCDService",
		Desc:     "Stop etcd service",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Action:   new(StopETCDService),
		Parallel: true,
	}

	removeFiles := &modules.Task{
		Name:     "RemoveFiles",
		Desc:     "Remove cluster files",
		Hosts:    c.Runtime.GetHostsByRole(common.K8s),
		Action:   new(RemoveFiles),
		Parallel: true,
	}

	daemonReload := &modules.Task{
		Name:     "DaemonReload",
		Desc:     "Systemd daemon reload",
		Hosts:    c.Runtime.GetHostsByRole(common.K8s),
		Action:   new(DaemonReload),
		Parallel: true,
	}

	c.Tasks = []*modules.Task{
		resetNetworkConfig,
		stopETCD,
		removeFiles,
		daemonReload,
	}
}

type InitDependenciesModule struct {
	common.KubeModule
}

func (i *InitDependenciesModule) Init() {
	i.Name = "InitDependenciesModule"

	getOSData := &modules.Task{
		Name:     "GetOSData",
		Desc:     "Get OS release",
		Hosts:    i.Runtime.GetAllHosts(),
		Action:   new(GetOSData),
		Parallel: true,
	}

	onlineInstall := &modules.Task{
		Name:     "OnlineInstallDependencies",
		Desc:     "Online install dependencies",
		Hosts:    i.Runtime.GetAllHosts(),
		Action:   new(OnlineInstallDependencies),
		Parallel: true,
	}

	offlineInstall := &modules.Task{
		Name:     "OnlineInstallDependencies",
		Desc:     "Online install dependencies",
		Hosts:    i.Runtime.GetAllHosts(),
		Action:   new(OfflineInstallDependencies),
		Parallel: true,
	}

	if i.KubeConf.Arg.SourcesDir == "" {
		i.Tasks = []*modules.Task{
			getOSData,
			onlineInstall,
		}
	} else {
		i.Tasks = []*modules.Task{
			getOSData,
			offlineInstall,
		}
	}
}
