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

package container

import (
	"path/filepath"
	"strings"

	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/container/templates"
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/kubesphere/kubekey/pkg/core/prepare"
	"github.com/kubesphere/kubekey/pkg/core/task"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/images"
	"github.com/kubesphere/kubekey/pkg/kubernetes"
)

type InstallContainerModule struct {
	common.KubeModule
	Skip bool
}

func (i *InstallContainerModule) IsSkip() bool {
	return i.Skip
}

func (i *InstallContainerModule) Init() {
	i.Name = "InstallContainerModule"
	i.Desc = "Install container manager"

	switch i.KubeConf.Cluster.Kubernetes.ContainerManager {
	case common.Docker:
		i.Tasks = InstallDocker(i)
	case common.Conatinerd:
		i.Tasks = InstallContainerd(i)
	case common.Crio:
		// TODO: Add the steps of cri-o's installation.
	case common.Isula:
		// TODO: Add the steps of iSula's installation.
	default:
		logger.Log.Fatalf("Unsupported container runtime: %s", strings.TrimSpace(i.KubeConf.Cluster.Kubernetes.ContainerManager))
	}
}

func InstallDocker(m *InstallContainerModule) []task.Interface {
	syncBinaries := &task.RemoteTask{
		Name:  "SyncDockerBinaries",
		Desc:  "Sync docker binaries",
		Hosts: m.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&kubernetes.NodeInCluster{Not: true},
			&DockerExist{Not: true},
		},
		Action:   new(SyncDockerBinaries),
		Parallel: true,
		Retry:    2,
	}

	generateContainerdService := &task.RemoteTask{
		Name:  "GenerateContainerdService",
		Desc:  "Generate containerd service",
		Hosts: m.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&kubernetes.NodeInCluster{Not: true},
			&DockerExist{Not: true},
		},
		Action: &action.Template{
			Template: templates.ContainerdService,
			Dst:      filepath.Join("/etc/systemd/system", templates.ContainerdService.Name()),
		},
		Parallel: true,
	}

	enableContainerd := &task.RemoteTask{
		Name:  "EnableContainerd",
		Desc:  "Enable containerd",
		Hosts: m.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&kubernetes.NodeInCluster{Not: true},
			&DockerExist{Not: true},
		},
		Action:   new(EnableContainerd),
		Parallel: true,
	}

	generateDockerService := &task.RemoteTask{
		Name:  "GenerateDockerService",
		Desc:  "Generate docker service",
		Hosts: m.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&kubernetes.NodeInCluster{Not: true},
			&DockerExist{Not: true},
		},
		Action: &action.Template{
			Template: templates.DockerService,
			Dst:      filepath.Join("/etc/systemd/system", templates.DockerService.Name()),
		},
		Parallel: true,
	}

	generateDockerConfig := &task.RemoteTask{
		Name:  "GenerateDockerConfig",
		Desc:  "Generate docker config",
		Hosts: m.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&kubernetes.NodeInCluster{Not: true},
			&DockerExist{Not: true},
		},
		Action: &action.Template{
			Template: templates.DockerConfig,
			Dst:      filepath.Join("/etc/docker/", templates.DockerConfig.Name()),
			Data: util.Data{
				"Mirrors":            templates.Mirrors(m.KubeConf),
				"InsecureRegistries": templates.InsecureRegistries(m.KubeConf),
			},
		},
		Parallel: true,
	}

	enableDocker := &task.RemoteTask{
		Name:  "EnableDocker",
		Desc:  "Enable docker",
		Hosts: m.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&kubernetes.NodeInCluster{Not: true},
			&DockerExist{Not: true},
		},
		Action:   new(EnableDocker),
		Parallel: true,
	}

	return []task.Interface{
		syncBinaries,
		generateContainerdService,
		enableContainerd,
		generateDockerService,
		generateDockerConfig,
		enableDocker,
	}
}

func InstallContainerd(m *InstallContainerModule) []task.Interface {
	syncCrictlBinaries := &task.RemoteTask{
		Name:  "SyncCrictlBinaries",
		Desc:  "Sync crictl binaries",
		Hosts: m.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&kubernetes.NodeInCluster{Not: true},
			&CrictlExist{Not: true},
		},
		Action:   new(SyncCrictlBinaries),
		Parallel: true,
		Retry:    2,
	}

	syncDockerBinaries := &task.RemoteTask{
		Name:  "SyncDockerBinaries",
		Desc:  "Sync docker binaries",
		Hosts: m.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&kubernetes.NodeInCluster{Not: true},
			&ContainerdExist{Not: true},
		},
		Action:   new(SyncDockerBinaries),
		Parallel: true,
		Retry:    2,
	}

	generateContainerdService := &task.RemoteTask{
		Name:  "GenerateContainerdService",
		Desc:  "Generate containerd service",
		Hosts: m.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&kubernetes.NodeInCluster{Not: true},
			&ContainerdExist{Not: true},
		},
		Action: &action.Template{
			Template: templates.ContainerdService,
			Dst:      filepath.Join("/etc/systemd/system", templates.ContainerdService.Name()),
		},
		Parallel: true,
	}

	generateContainerdConfig := &task.RemoteTask{
		Name:  "GenerateContainerdConfig",
		Desc:  "Generate containerd config",
		Hosts: m.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&kubernetes.NodeInCluster{Not: true},
			&ContainerdExist{Not: true},
		},
		Action: &action.Template{
			Template: templates.ContainerdConfig,
			Dst:      filepath.Join("/etc/containerd/", templates.ContainerdConfig.Name()),
			Data: util.Data{
				"Mirrors":            templates.Mirrors(m.KubeConf),
				"InsecureRegistries": templates.InsecureRegistries(m.KubeConf),
				"SandBoxImage":       images.GetImage(m.Runtime, m.KubeConf, "pause").ImageName(),
			},
		},
		Parallel: true,
	}

	generateCrictlConfig := &task.RemoteTask{
		Name:  "GenerateCrictlConfig",
		Desc:  "Generate crictl config",
		Hosts: m.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&kubernetes.NodeInCluster{Not: true},
			&ContainerdExist{Not: true},
		},
		Action: &action.Template{
			Template: templates.CrictlConfig,
			Dst:      filepath.Join("/etc/", templates.CrictlConfig.Name()),
			Data: util.Data{
				"Endpoint": m.KubeConf.Cluster.Kubernetes.ContainerRuntimeEndpoint,
			},
		},
		Parallel: true,
	}

	enableContainerd := &task.RemoteTask{
		Name:  "EnableContainerd",
		Desc:  "Enable containerd",
		Hosts: m.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&kubernetes.NodeInCluster{Not: true},
			&ContainerdExist{Not: true},
		},
		Action:   new(EnableContainerd),
		Parallel: true,
	}

	generateDockerService := &task.RemoteTask{
		Name:  "GenerateDockerService",
		Desc:  "Generate docker service",
		Hosts: m.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&kubernetes.NodeInCluster{Not: true},
			&ContainerdExist{Not: true},
		},
		Action: &action.Template{
			Template: templates.DockerService,
			Dst:      filepath.Join("/etc/systemd/system", templates.DockerService.Name()),
		},
		Parallel: true,
	}

	generateDockerConfig := &task.RemoteTask{
		Name:  "GenerateDockerConfig",
		Desc:  "Generate docker config",
		Hosts: m.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&kubernetes.NodeInCluster{Not: true},
			&ContainerdExist{Not: true},
		},
		Action: &action.Template{
			Template: templates.DockerConfig,
			Dst:      filepath.Join("/etc/docker/", templates.DockerConfig.Name()),
			Data: util.Data{
				"Mirrors":            templates.Mirrors(m.KubeConf),
				"InsecureRegistries": templates.InsecureRegistries(m.KubeConf),
			},
		},
		Parallel: true,
	}

	enableDocker := &task.RemoteTask{
		Name:  "EnableDocker",
		Desc:  "Enable docker",
		Hosts: m.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&kubernetes.NodeInCluster{Not: true},
			&ContainerdExist{Not: true},
		},
		Action:   new(EnableDocker),
		Parallel: true,
	}

	return []task.Interface{
		syncCrictlBinaries,
		syncDockerBinaries,
		generateContainerdService,
		generateContainerdConfig,
		generateCrictlConfig,
		enableContainerd,
		generateDockerService,
		generateDockerConfig,
		enableDocker,
	}
}
