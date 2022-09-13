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

package k3s

import (
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/prepare"
	"github.com/kubesphere/kubekey/pkg/core/task"
	"github.com/kubesphere/kubekey/pkg/k3s/templates"
	"path/filepath"
)

type StatusModule struct {
	common.KubeModule
}

func (s *StatusModule) Init() {
	s.Name = "StatusModule"
	s.Desc = "Get cluster status"

	cluster := NewK3sStatus()
	s.PipelineCache.GetOrSet(common.ClusterStatus, cluster)

	clusterStatus := &task.RemoteTask{
		Name:     "GetClusterStatus",
		Desc:     "Get k3s cluster status",
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
	i.Desc = "Install k3s cluster"

	syncBinary := &task.RemoteTask{
		Name:     "SyncKubeBinary",
		Desc:     "Synchronize k3s binaries",
		Hosts:    i.Runtime.GetHostsByRole(common.K8s),
		Prepare:  &NodeInCluster{Not: true},
		Action:   new(SyncKubeBinary),
		Parallel: true,
		Retry:    2,
	}

	killAllScript := &task.RemoteTask{
		Name:    "GenerateK3sKillAllScript",
		Desc:    "Generate k3s killall.sh script",
		Hosts:   i.Runtime.GetHostsByRole(common.K8s),
		Prepare: &NodeInCluster{Not: true},
		Action: &action.Template{
			Template: templates.K3sKillallScript,
			Dst:      filepath.Join("/usr/local/bin", templates.K3sKillallScript.Name()),
		},
		Parallel: true,
		Retry:    2,
	}

	uninstallScript := &task.RemoteTask{
		Name:    "GenerateK3sUninstallScript",
		Desc:    "Generate k3s uninstall script",
		Hosts:   i.Runtime.GetHostsByRole(common.K8s),
		Prepare: &NodeInCluster{Not: true},
		Action: &action.Template{
			Template: templates.K3sUninstallScript,
			Dst:      filepath.Join("/usr/local/bin", templates.K3sUninstallScript.Name()),
		},
		Parallel: true,
		Retry:    2,
	}

	chmod := &task.RemoteTask{
		Name:     "ChmodScript",
		Desc:     "Chmod +x k3s script ",
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
	i.Name = "K3sInitClusterModule"
	i.Desc = "Init k3s cluster"

	k3sService := &task.RemoteTask{
		Name:  "GenerateK3sService",
		Desc:  "Generate k3s Service",
		Hosts: i.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&ClusterIsExist{Not: true},
		},
		Action:   new(GenerateK3sService),
		Parallel: true,
	}

	k3sEnv := &task.RemoteTask{
		Name:  "GenerateK3sServiceEnv",
		Desc:  "Generate k3s service env",
		Hosts: i.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&ClusterIsExist{Not: true},
		},
		Action:   new(GenerateK3sServiceEnv),
		Parallel: true,
	}

	k3sRegistryConfig := &task.RemoteTask{
		Name:  "GenerateK3sRegistryConfig",
		Desc:  "Generate k3s registry config",
		Hosts: i.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&ClusterIsExist{Not: true},
			&UsePrivateRegstry{Not: false},
		},
		Action:   new(GenerateK3sRegistryConfig),
		Parallel: true,
	}

	enableK3s := &task.RemoteTask{
		Name:  "EnableK3sService",
		Desc:  "Enable k3s service",
		Hosts: i.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&ClusterIsExist{Not: true},
		},
		Action:   new(EnableK3sService),
		Parallel: true,
	}

	copyKubeConfig := &task.RemoteTask{
		Name:  "CopyKubeConfig",
		Desc:  "Copy k3s.yaml to ~/.kube/config",
		Hosts: i.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&ClusterIsExist{Not: true},
		},
		Action:   new(CopyK3sKubeConfig),
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
		k3sService,
		k3sEnv,
		k3sRegistryConfig,
		enableK3s,
		copyKubeConfig,
		addMasterTaint,
		addWorkerLabel,
	}
}

type JoinNodesModule struct {
	common.KubeModule
}

func (j *JoinNodesModule) Init() {
	j.Name = "K3sJoinNodesModule"
	j.Desc = "Join k3s nodes"

	k3sService := &task.RemoteTask{
		Name:  "GenerateK3sService",
		Desc:  "Generate k3s Service",
		Hosts: j.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&NodeInCluster{Not: true},
		},
		Action:   new(GenerateK3sService),
		Parallel: true,
	}

	k3sEnv := &task.RemoteTask{
		Name:  "GenerateK3sServiceEnv",
		Desc:  "Generate k3s service env",
		Hosts: j.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&NodeInCluster{Not: true},
		},
		Action:   new(GenerateK3sServiceEnv),
		Parallel: true,
	}

	k3sRegistryConfig := &task.RemoteTask{
		Name:  "GenerateK3sRegistryConfig",
		Desc:  "Generate k3s registry config",
		Hosts: j.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&NodeInCluster{Not: true},
			&UsePrivateRegstry{Not: false},
		},
		Action:   new(GenerateK3sRegistryConfig),
		Parallel: true,
	}

	enableK3s := &task.RemoteTask{
		Name:  "EnableK3sService",
		Desc:  "Enable k3s service",
		Hosts: j.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&NodeInCluster{Not: true},
		},
		Action:   new(EnableK3sService),
		Parallel: true,
	}

	copyKubeConfigForMaster := &task.RemoteTask{
		Name:  "CopyKubeConfig",
		Desc:  "Copy k3s.yaml to ~/.kube/config",
		Hosts: j.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			&NodeInCluster{Not: true},
		},
		Action:   new(CopyK3sKubeConfig),
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
		k3sService,
		k3sEnv,
		k3sRegistryConfig,
		enableK3s,
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
	d.Desc = "Delete k3s cluster"

	execScript := &task.RemoteTask{
		Name:     "ExecUninstallScript",
		Desc:     "Exec k3s uninstall script",
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
