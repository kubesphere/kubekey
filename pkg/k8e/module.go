/*
 Copyright 2022 The KubeSphere Authors.

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

package k8e

import (
	"path/filepath"

	"github.com/kubesphere/kubekey/v2/pkg/common"
	"github.com/kubesphere/kubekey/v2/pkg/core/action"
	"github.com/kubesphere/kubekey/v2/pkg/core/prepare"
	"github.com/kubesphere/kubekey/v2/pkg/core/task"
	"github.com/kubesphere/kubekey/v2/pkg/k8e/templates"
)

type StatusModule struct {
	common.KubeModule
}

func (s *StatusModule) Init() {
	s.Name = "StatusModule"
	s.Desc = "Get cluster status"

	cluster := NewK8eStatus()
	s.PipelineCache.GetOrSet(common.ClusterStatus, cluster)

	clusterStatus := &task.RemoteTask{
		Name:     "GetClusterStatus",
		Desc:     "Get K8e cluster status",
		Hosts:    s.Runtime.GetHostsByRole(common.Master),
		Action:   new(GetClusterStatus),
		Parallel: false,
	}

	s.Tasks = []task.Interface{
		clusterStatus,
	}
}

type InstallKubeBinariesModule struct {
	common.KubeModule
}

func (i *InstallKubeBinariesModule) Init() {
	i.Name = "InstallKubeBinariesModule"
	i.Desc = "Install k8e cluster"

	syncBinary := &task.RemoteTask{
		Name:     "SyncKubeBinary",
		Desc:     "Synchronize k8e binaries",
		Hosts:    i.Runtime.GetHostsByRole(common.K8s),
		Prepare:  &NodeInCluster{Not: true},
		Action:   new(SyncKubeBinary),
		Parallel: true,
		Retry:    2,
	}

	killAllScript := &task.RemoteTask{
		Name:    "GenerateK8eKillAllScript",
		Desc:    "Generate k8e killall.sh script",
		Hosts:   i.Runtime.GetHostsByRole(common.K8s),
		Prepare: &NodeInCluster{Not: true},
		Action: &action.Template{
			Template: templates.K8eKillallScript,
			Dst:      filepath.Join("/usr/local/bin", templates.K8eKillallScript.Name()),
		},
		Parallel: true,
		Retry:    2,
	}

	uninstallScript := &task.RemoteTask{
		Name:    "GenerateK8eUninstallScript",
		Desc:    "Generate k8e uninstall script",
		Hosts:   i.Runtime.GetHostsByRole(common.K8s),
		Prepare: &NodeInCluster{Not: true},
		Action: &action.Template{
			Template: templates.K8eUninstallScript,
			Dst:      filepath.Join("/usr/local/bin", templates.K8eUninstallScript.Name()),
		},
		Parallel: true,
		Retry:    2,
	}

	chmod := &task.RemoteTask{
		Name:     "ChmodScript",
		Desc:     "Chmod +x k8e script ",
		Hosts:    i.Runtime.GetHostsByRole(common.K8s),
		Prepare:  &NodeInCluster{Not: true},
		Action:   new(ChmodScript),
		Parallel: true,
		Retry:    2,
	}

	i.Tasks = []task.Interface{
		syncBinary,
		killAllScript,
		uninstallScript,
		chmod,
	}
}

type InitClusterModule struct {
	common.KubeModule
}

func (i *InitClusterModule) Init() {
	i.Name = "K8eInitClusterModule"
	i.Desc = "Init k8e cluster"

	k8eService := &task.RemoteTask{
		Name:  "GenerateK8eService",
		Desc:  "Generate k8e Service",
		Hosts: i.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&ClusterIsExist{Not: true},
		},
		Action:   new(GenerateK8eService),
		Parallel: true,
	}

	k8eEnv := &task.RemoteTask{
		Name:  "GenerateK8eServiceEnv",
		Desc:  "Generate k8e service env",
		Hosts: i.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&ClusterIsExist{Not: true},
		},
		Action:   new(GenerateK8eServiceEnv),
		Parallel: true,
	}

	enableK8e := &task.RemoteTask{
		Name:  "EnableK8eService",
		Desc:  "Enable k8e service",
		Hosts: i.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&ClusterIsExist{Not: true},
		},
		Action:   new(EnableK8eService),
		Parallel: true,
	}

	copyKubeConfig := &task.RemoteTask{
		Name:  "CopyKubeConfig",
		Desc:  "Copy k8e.yaml to ~/.kube/config",
		Hosts: i.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&ClusterIsExist{Not: true},
		},
		Action:   new(CopyK8eKubeConfig),
		Parallel: true,
	}

	addMasterTaint := &task.RemoteTask{
		Name:  "AddMasterTaint",
		Desc:  "Add master taint",
		Hosts: i.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&ClusterIsExist{Not: true},
			&common.IsWorker{Not: true},
		},
		Action:   new(AddMasterTaint),
		Parallel: true,
		Retry:    5,
	}

	addWorkerLabel := &task.RemoteTask{
		Name:  "AddWorkerLabel",
		Desc:  "Add worker label",
		Hosts: i.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&ClusterIsExist{Not: true},
			new(common.IsWorker),
		},
		Action:   new(AddWorkerLabel),
		Parallel: true,
		Retry:    5,
	}

	i.Tasks = []task.Interface{
		k8eService,
		k8eEnv,
		enableK8e,
		copyKubeConfig,
		addMasterTaint,
		addWorkerLabel,
	}
}

type JoinNodesModule struct {
	common.KubeModule
}

func (j *JoinNodesModule) Init() {
	j.Name = "K8eJoinNodesModule"
	j.Desc = "Join k8e nodes"

	k8eService := &task.RemoteTask{
		Name:  "GenerateK8eService",
		Desc:  "Generate k8e Service",
		Hosts: j.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&NodeInCluster{Not: true},
		},
		Action:   new(GenerateK8eService),
		Parallel: true,
	}

	k8eEnv := &task.RemoteTask{
		Name:  "GenerateK8eServiceEnv",
		Desc:  "Generate k8e service env",
		Hosts: j.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&NodeInCluster{Not: true},
		},
		Action:   new(GenerateK8eServiceEnv),
		Parallel: true,
	}

	enableK8e := &task.RemoteTask{
		Name:  "EnableK8eService",
		Desc:  "Enable k8e service",
		Hosts: j.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&NodeInCluster{Not: true},
		},
		Action:   new(EnableK8eService),
		Parallel: true,
	}

	copyKubeConfigForMaster := &task.RemoteTask{
		Name:  "CopyKubeConfig",
		Desc:  "Copy k8e.yaml to ~/.kube/config",
		Hosts: j.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			&NodeInCluster{Not: true},
		},
		Action:   new(CopyK8eKubeConfig),
		Parallel: true,
	}

	syncKubeConfigToWorker := &task.RemoteTask{
		Name:  "SyncKubeConfigToWorker",
		Desc:  "Synchronize kube config to worker",
		Hosts: j.Runtime.GetHostsByRole(common.Worker),
		Prepare: &prepare.PrepareCollection{
			&NodeInCluster{Not: true},
			new(common.OnlyWorker),
		},
		Action:   new(SyncKubeConfigToWorker),
		Parallel: true,
	}

	addMasterTaint := &task.RemoteTask{
		Name:  "AddMasterTaint",
		Desc:  "Add master taint",
		Hosts: j.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			&NodeInCluster{Not: true},
			&common.IsWorker{Not: true},
		},
		Action:   new(AddMasterTaint),
		Parallel: true,
		Retry:    5,
	}

	addWorkerLabel := &task.RemoteTask{
		Name:  "AddWorkerLabel",
		Desc:  "Add worker label",
		Hosts: j.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&NodeInCluster{Not: true},
			new(common.IsWorker),
		},
		Action:   new(AddWorkerLabel),
		Parallel: true,
		Retry:    5,
	}

	j.Tasks = []task.Interface{
		k8eService,
		k8eEnv,
		enableK8e,
		copyKubeConfigForMaster,
		syncKubeConfigToWorker,
		addMasterTaint,
		addWorkerLabel,
	}
}

type DeleteClusterModule struct {
	common.KubeModule
}

func (d *DeleteClusterModule) Init() {
	d.Name = "DeleteClusterModule"
	d.Desc = "Delete k8e cluster"

	execScript := &task.RemoteTask{
		Name:     "ExecUninstallScript",
		Desc:     "Exec k8e uninstall script",
		Hosts:    d.Runtime.GetHostsByRole(common.K8s),
		Action:   new(ExecUninstallScript),
		Parallel: true,
	}

	d.Tasks = []task.Interface{
		execScript,
	}
}

type SaveKubeConfigModule struct {
	common.KubeModule
}

func (s *SaveKubeConfigModule) Init() {
	s.Name = "SaveKubeConfigModule"
	s.Desc = "Save kube config file as a configmap"

	save := &task.LocalTask{
		Name:   "SaveKubeConfig",
		Desc:   "Save kube config as a configmap",
		Action: new(SaveKubeConfig),
	}

	s.Tasks = []task.Interface{
		save,
	}
}
