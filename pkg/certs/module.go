package certs

import (
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/prepare"
	"github.com/kubesphere/kubekey/pkg/core/task"
	"github.com/kubesphere/kubekey/pkg/kubernetes"
)

type CheckCertsModule struct {
	common.KubeModule
}

func (c *CheckCertsModule) Init() {
	c.Name = "CheckCertsModule"
	c.Desc = "Check cluster certs"

	check := &task.RemoteTask{
		Name:     "CheckClusterCerts",
		Desc:     "Check cluster certs",
		Hosts:    c.Runtime.GetHostsByRole(common.Master),
		Action:   new(ListClusterCerts),
		Parallel: true,
	}

	c.Tasks = []task.Interface{
		check,
	}
}

type PrintClusterCertsModule struct {
	common.KubeModule
}

func (p *PrintClusterCertsModule) Init() {
	p.Name = "PrintClusterCertsModule"
	p.Desc = "Display cluster certs form"

	display := &task.LocalTask{
		Name:   "DisplayCertsForm",
		Desc:   "Display cluster certs form",
		Action: new(DisplayForm),
	}

	p.Tasks = []task.Interface{
		display,
	}
}

type RenewCertsModule struct {
	common.KubeModule
}

func (r *RenewCertsModule) Init() {
	r.Name = "RenewCertsModule"
	r.Desc = "Renew control-plane certs"

	renew := &task.RemoteTask{
		Name:     "RenewCerts",
		Desc:     "Renew control-plane certs",
		Hosts:    r.Runtime.GetHostsByRole(common.Master),
		Action:   new(RenewCerts),
		Parallel: false,
		Retry:    5,
	}

	copyKubeConfig := &task.RemoteTask{
		Name:     "CopyKubeConfig",
		Desc:     "Copy admin.conf to ~/.kube/config",
		Hosts:    r.Runtime.GetHostsByRole(common.Master),
		Action:   new(kubernetes.CopyKubeConfigForControlPlane),
		Parallel: true,
		Retry:    2,
	}

	fetchKubeConfig := &task.RemoteTask{
		Name:     "FetchKubeConfig",
		Desc:     "Fetch kube config file from control-plane",
		Hosts:    r.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.OnlyFirstMaster),
		Action:   new(FetchKubeConfig),
		Parallel: true,
	}

	syncKubeConfig := &task.RemoteTask{
		Name:  "SyncKubeConfig",
		Desc:  "Synchronize kube config to worker",
		Hosts: r.Runtime.GetHostsByRole(common.Worker),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyWorker),
		},
		Action:   new(SyneKubeConfigToWorker),
		Parallel: true,
		Retry:    3,
	}

	r.Tasks = []task.Interface{
		renew,
		copyKubeConfig,
		fetchKubeConfig,
		syncKubeConfig,
	}
}
