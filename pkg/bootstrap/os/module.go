/*
 Copyright 2021 The KubeSphere Authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package os

import (
	"github.com/kubesphere/kubekey/pkg/bootstrap/os/templates"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/task"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"path/filepath"
)

type ConfigureOSModule struct {
	common.KubeModule
}

func (c *ConfigureOSModule) Init() {
	c.Name = "ConfigureOSModule"
	c.Desc = "Init os dependencies"

	initOS := &task.RemoteTask{
		Name:     "InitOS",
		Desc:     "Prepare to init OS",
		Hosts:    c.Runtime.GetAllHosts(),
		Action:   new(NodeConfigureOS),
		Parallel: true,
	}

	GenerateScript := &task.RemoteTask{
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

	ExecScript := &task.RemoteTask{
		Name:     "ExecScript",
		Desc:     "Exec init os script",
		Hosts:    c.Runtime.GetAllHosts(),
		Action:   new(NodeExecScript),
		Parallel: true,
	}

	c.Tasks = []task.Interface{
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

	resetNetworkConfig := &task.RemoteTask{
		Name:     "ResetNetworkConfig",
		Desc:     "Reset os network config",
		Hosts:    c.Runtime.GetHostsByRole(common.K8s),
		Action:   new(ResetNetworkConfig),
		Parallel: true,
	}

	stopETCD := &task.RemoteTask{
		Name:     "StopETCDService",
		Desc:     "Stop etcd service",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Action:   new(StopETCDService),
		Parallel: true,
	}

	removeFiles := &task.RemoteTask{
		Name:     "RemoveFiles",
		Desc:     "Remove cluster files",
		Hosts:    c.Runtime.GetHostsByRole(common.K8s),
		Action:   new(RemoveFiles),
		Parallel: true,
	}

	daemonReload := &task.RemoteTask{
		Name:     "DaemonReload",
		Desc:     "Systemd daemon reload",
		Hosts:    c.Runtime.GetHostsByRole(common.K8s),
		Action:   new(DaemonReload),
		Parallel: true,
	}

	c.Tasks = []task.Interface{
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

	getOSData := &task.RemoteTask{
		Name:     "GetOSData",
		Desc:     "Get OS release",
		Hosts:    i.Runtime.GetAllHosts(),
		Action:   new(GetOSData),
		Parallel: true,
	}

	onlineInstall := &task.RemoteTask{
		Name:     "OnlineInstallDependencies",
		Desc:     "Online install dependencies",
		Hosts:    i.Runtime.GetAllHosts(),
		Action:   new(OnlineInstallDependencies),
		Parallel: true,
	}

	offlineInstall := &task.RemoteTask{
		Name:     "OnlineInstallDependencies",
		Desc:     "Offline install dependencies",
		Hosts:    i.Runtime.GetAllHosts(),
		Action:   new(OfflineInstallDependencies),
		Parallel: true,
	}

	if i.KubeConf.Arg.SourcesDir == "" {
		i.Tasks = []task.Interface{
			getOSData,
			onlineInstall,
		}
	} else {
		i.Tasks = []task.Interface{
			getOSData,
			offlineInstall,
		}
	}
}
