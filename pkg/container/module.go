package container

import (
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/container/templates"
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/kubesphere/kubekey/pkg/core/module"
	"github.com/kubesphere/kubekey/pkg/core/prepare"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/images"
	"github.com/kubesphere/kubekey/pkg/kubernetes"
	"path/filepath"
	"strings"
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

func InstallDocker(m *InstallContainerModule) []module.Task {
	syncBinaries := &module.RemoteTask{
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

	generateContainerdService := &module.RemoteTask{
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

	enableContainerd := &module.RemoteTask{
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

	generateDockerService := &module.RemoteTask{
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

	generateDOckerConfig := &module.RemoteTask{
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

	enableDocker := &module.RemoteTask{
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

	return []module.Task{
		syncBinaries,
		generateContainerdService,
		enableContainerd,
		generateDockerService,
		generateDOckerConfig,
		enableDocker,
	}
}

func InstallContainerd(m *InstallContainerModule) []module.Task {
	syncCrictlBinaries := &module.RemoteTask{
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

	syncDockerBinaries := &module.RemoteTask{
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

	generateContainerdService := &module.RemoteTask{
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

	generateContainerdConfig := &module.RemoteTask{
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

	generateCrictlConfig := &module.RemoteTask{
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

	enableContainerd := &module.RemoteTask{
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

	generateDockerService := &module.RemoteTask{
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

	generateDockerConfig := &module.RemoteTask{
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

	enableDocker := &module.RemoteTask{
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

	return []module.Task{
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
