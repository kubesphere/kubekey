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
	"path/filepath"

	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/bootstrap/os/templates"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/action"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/prepare"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/task"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/util"
)

type ConfigureOSModule struct {
	common.KubeModule
	Skip bool
}

func (c *ConfigureOSModule) IsSkip() bool {
	return c.Skip
}

func (c *ConfigureOSModule) Init() {
	c.Name = "ConfigureOSModule"
	c.Desc = "Init os dependencies"

	getOSData := &task.RemoteTask{
		Name:     "GetOSData",
		Desc:     "Get OS release",
		Hosts:    c.Runtime.GetAllHosts(),
		Action:   new(GetOSData),
		Parallel: true,
	}

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
				"Hosts":       templates.GenerateHosts(c.Runtime, c.KubeConf),
				"IPv6Support": templates.EnabledIPv6(c.KubeConf),
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

	ConfigureNtpServer := &task.RemoteTask{
		Name:     "ConfigureNtpServer",
		Desc:     "configure the ntp server for each node",
		Hosts:    c.Runtime.GetAllHosts(),
		Prepare:  new(NodeConfigureNtpCheck),
		Action:   new(NodeConfigureNtpServer),
		Parallel: true,
	}

	c.Tasks = []task.Interface{
		getOSData,
		initOS,
		GenerateScript,
		ExecScript,
		ConfigureNtpServer,
	}
}

type ClearNodeOSModule struct {
	common.KubeModule
}

func (c *ClearNodeOSModule) Init() {
	c.Name = "ClearNodeOSModule"

	stopKubelet := &task.RemoteTask{
		Name:     "StopKubelet",
		Desc:     "Stop Kubelet",
		Hosts:    c.Runtime.GetHostsByRole(common.Worker),
		Prepare:  new(DeleteNode),
		Action:   new(StopKubelet),
		Parallel: true,
	}

	resetNetworkConfig := &task.RemoteTask{
		Name:     "ResetNetworkConfig",
		Desc:     "Reset os network config",
		Hosts:    c.Runtime.GetHostsByRole(common.Worker),
		Prepare:  new(DeleteNode),
		Action:   new(ResetNetworkConfig),
		Parallel: true,
	}

	removeFiles := &task.RemoteTask{
		Name:     "RemoveFiles",
		Desc:     "Remove node files",
		Hosts:    c.Runtime.GetHostsByRole(common.Worker),
		Prepare:  new(DeleteNode),
		Action:   new(RemoveNodeFiles),
		Parallel: true,
	}

	daemonReload := &task.RemoteTask{
		Name:     "DaemonReload",
		Desc:     "Systemd daemon reload",
		Hosts:    c.Runtime.GetHostsByRole(common.Worker),
		Prepare:  new(DeleteNode),
		Action:   new(DaemonReload),
		Parallel: true,
	}

	c.Tasks = []task.Interface{
		stopKubelet,
		resetNetworkConfig,
		removeFiles,
		daemonReload,
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

	uninstallETCD := &task.RemoteTask{
		Name:  "UninstallETCD",
		Desc:  "Uninstall etcd",
		Hosts: c.Runtime.GetHostsByRole(common.ETCD),
		Prepare: &prepare.PrepareCollection{
			new(EtcdTypeIsKubeKey),
		},
		Action:   new(UninstallETCD),
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
		uninstallETCD,
		removeFiles,
		daemonReload,
	}
}

type RepositoryOnlineModule struct {
	common.KubeModule
	Skip bool
}

func (r *RepositoryOnlineModule) IsSkip() bool {
	return r.Skip
}

func (r *RepositoryOnlineModule) Init() {
	r.Name = "RepositoryOnlineModule"

	getOSData := &task.RemoteTask{
		Name:     "GetOSData",
		Desc:     "Get OS release",
		Hosts:    r.Runtime.GetAllHosts(),
		Action:   new(GetOSData),
		Parallel: true,
	}

	newRepo := &task.RemoteTask{
		Name:     "NewRepoClient",
		Desc:     "New repository client",
		Hosts:    r.Runtime.GetAllHosts(),
		Action:   new(NewRepoClient),
		Parallel: true,
		Retry:    1,
	}

	install := &task.RemoteTask{
		Name:     "InstallPackage",
		Desc:     "Install packages",
		Hosts:    r.Runtime.GetAllHosts(),
		Action:   new(InstallPackage),
		Parallel: true,
		Retry:    1,
	}

	r.Tasks = []task.Interface{
		getOSData,
		newRepo,
		install,
	}
}

type RepositoryModule struct {
	common.KubeModule
	Skip bool
}

func (r *RepositoryModule) IsSkip() bool {
	return r.Skip
}

func (r *RepositoryModule) Init() {
	r.Name = "RepositoryModule"
	r.Desc = "Install local repository"

	getOSData := &task.RemoteTask{
		Name:     "GetOSData",
		Desc:     "Get OS release",
		Hosts:    r.Runtime.GetAllHosts(),
		Action:   new(GetOSData),
		Parallel: true,
	}

	sync := &task.RemoteTask{
		Name:     "SyncRepositoryISOFile",
		Desc:     "Sync repository iso file to all nodes",
		Hosts:    r.Runtime.GetAllHosts(),
		Action:   new(SyncRepositoryFile),
		Parallel: true,
		Retry:    2,
	}

	mount := &task.RemoteTask{
		Name:     "MountISO",
		Desc:     "Mount iso file",
		Hosts:    r.Runtime.GetAllHosts(),
		Action:   new(MountISO),
		Parallel: true,
		Retry:    1,
	}

	newRepo := &task.RemoteTask{
		Name:     "NewRepoClient",
		Desc:     "New repository client",
		Hosts:    r.Runtime.GetAllHosts(),
		Action:   new(NewRepoClient),
		Parallel: true,
		Retry:    1,
		Rollback: new(RollbackUmount),
	}

	backup := &task.RemoteTask{
		Name:     "BackupOriginalRepository",
		Desc:     "Backup original repository",
		Hosts:    r.Runtime.GetAllHosts(),
		Action:   new(BackupOriginalRepository),
		Parallel: true,
		Retry:    1,
		Rollback: new(RecoverBackupSuccessNode),
	}

	add := &task.RemoteTask{
		Name:     "AddLocalRepository",
		Desc:     "Add local repository",
		Hosts:    r.Runtime.GetAllHosts(),
		Action:   new(AddLocalRepository),
		Parallel: true,
		Retry:    1,
		Rollback: new(RecoverRepository),
	}

	install := &task.RemoteTask{
		Name:     "InstallPackage",
		Desc:     "Install packages",
		Hosts:    r.Runtime.GetAllHosts(),
		Action:   new(InstallPackage),
		Parallel: true,
		Retry:    1,
		Rollback: new(RecoverRepository),
	}

	reset := &task.RemoteTask{
		Name:     "ResetRepository",
		Desc:     "Reset repository to the original repository",
		Hosts:    r.Runtime.GetAllHosts(),
		Action:   new(ResetRepository),
		Parallel: true,
		Retry:    1,
	}

	umount := &task.RemoteTask{
		Name:     "UmountISO",
		Desc:     "Umount ISO file",
		Hosts:    r.Runtime.GetAllHosts(),
		Action:   new(UmountISO),
		Parallel: true,
	}

	r.Tasks = []task.Interface{
		getOSData,
		sync,
		mount,
		newRepo,
		backup,
		add,
		install,
		reset,
		umount,
	}
}
