package kubernetes

import (
	"path/filepath"

	"github.com/kubesphere/kubekey/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/core/action"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/core/task"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/kubernetes"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/kubernetes/templates"
)

type InstallKubeletModule struct {
	common.KubeModule
}

func (i *InstallKubeletModule) Init() {
	i.Name = "InstallKubeletModule"
	i.Desc = "Install kubernetes cluster"

	syncKubelet := &task.RemoteTask{
		Name:     "SyncKubelet",
		Desc:     "Synchronize kubelet",
		Hosts:    i.Runtime.GetHostsByRole(common.K8s),
		Prepare:  &kubernetes.NodeInCluster{Not: true},
		Action:   new(kubernetes.SyncKubelet),
		Parallel: true,
		Retry:    2,
	}

	generateKubeletService := &task.RemoteTask{
		Name:    "GenerateKubeletService",
		Desc:    "Generate kubelet service",
		Hosts:   i.Runtime.GetHostsByRole(common.K8s),
		Prepare: &kubernetes.NodeInCluster{Not: true},
		Action: &action.Template{
			Template: templates.KubeletService,
			Dst:      filepath.Join("/etc/systemd/system/", templates.KubeletService.Name()),
		},
		Parallel: true,
		Retry:    2,
	}

	enableKubelet := &task.RemoteTask{
		Name:     "EnableKubelet",
		Desc:     "Enable kubelet service",
		Hosts:    i.Runtime.GetHostsByRole(common.K8s),
		Prepare:  &kubernetes.NodeInCluster{Not: true},
		Action:   new(kubernetes.EnableKubelet),
		Parallel: true,
		Retry:    5,
	}

	generateKubeletEnv := &task.RemoteTask{
		Name:     "GenerateKubeletEnv",
		Desc:     "Generate kubelet env",
		Hosts:    i.Runtime.GetHostsByRole(common.K8s),
		Prepare:  &kubernetes.NodeInCluster{Not: true},
		Action:   new(kubernetes.GenerateKubeletEnv),
		Parallel: true,
		Retry:    2,
	}

	i.Tasks = []task.Interface{
		syncKubelet,
		generateKubeletService,
		enableKubelet,
		generateKubeletEnv,
	}
}
