package os

import (
	"github.com/kubesphere/kubekey/pkg/bootstrap/os/templates"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/module"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"path/filepath"
)

type ConfigureOSModule struct {
	common.KubeModule
}

func (c *ConfigureOSModule) Init() {
	c.Name = "ConfigureOSModule"

	initOS := &module.RemoteTask{
		Name:     "InitOS",
		Desc:     "Prepare to init OS",
		Hosts:    c.Runtime.GetAllHosts(),
		Action:   new(NodeConfigureOS),
		Parallel: true,
	}

	GenerateScript := &module.RemoteTask{
		Name:  "GenerateScript",
		Desc:  "Generate init os script",
		Hosts: c.Runtime.GetAllHosts(),
		Action: &action.Template{
			Template: templates.InitOsScriptTmpl,
			Dst:      filepath.Join(common.KubeScriptDir, "initOS.sh"),
			Data: util.Data{
				"Hosts": templates.GenerateHosts(c.Runtime, c.KubeConf),
			},
		},
		Parallel: true,
	}

	ExecScript := &module.RemoteTask{
		Name:     "ExecScript",
		Desc:     "Exec init os script",
		Hosts:    c.Runtime.GetAllHosts(),
		Action:   new(NodeExecScript),
		Parallel: true,
	}

	c.Tasks = []module.Task{
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

	resetNetworkConfig := &module.RemoteTask{
		Name:     "ResetNetworkConfig",
		Desc:     "Reset os network config",
		Hosts:    c.Runtime.GetHostsByRole(common.K8s),
		Action:   new(ResetNetworkConfig),
		Parallel: true,
	}

	stopETCD := &module.RemoteTask{
		Name:     "StopETCDService",
		Desc:     "Stop etcd service",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Action:   new(StopETCDService),
		Parallel: true,
	}

	removeFiles := &module.RemoteTask{
		Name:     "RemoveFiles",
		Desc:     "Remove cluster files",
		Hosts:    c.Runtime.GetHostsByRole(common.K8s),
		Action:   new(RemoveFiles),
		Parallel: true,
	}

	daemonReload := &module.RemoteTask{
		Name:     "DaemonReload",
		Desc:     "Systemd daemon reload",
		Hosts:    c.Runtime.GetHostsByRole(common.K8s),
		Action:   new(DaemonReload),
		Parallel: true,
	}

	c.Tasks = []module.Task{
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

	getOSData := &module.RemoteTask{
		Name:     "GetOSData",
		Desc:     "Get OS release",
		Hosts:    i.Runtime.GetAllHosts(),
		Action:   new(GetOSData),
		Parallel: true,
	}

	onlineInstall := &module.RemoteTask{
		Name:     "OnlineInstallDependencies",
		Desc:     "Online install dependencies",
		Hosts:    i.Runtime.GetAllHosts(),
		Action:   new(OnlineInstallDependencies),
		Parallel: true,
	}

	offlineInstall := &module.RemoteTask{
		Name:     "OnlineInstallDependencies",
		Desc:     "Offline install dependencies",
		Hosts:    i.Runtime.GetAllHosts(),
		Action:   new(OfflineInstallDependencies),
		Parallel: true,
	}

	if i.KubeConf.Arg.SourcesDir == "" {
		i.Tasks = []module.Task{
			getOSData,
			onlineInstall,
		}
	} else {
		i.Tasks = []module.Task{
			getOSData,
			offlineInstall,
		}
	}
}
