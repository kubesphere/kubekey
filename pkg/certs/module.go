package certs

import (
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/module"
	"github.com/kubesphere/kubekey/pkg/core/prepare"
	"github.com/kubesphere/kubekey/pkg/kubernetes"
)

type CheckCertsModule struct {
	common.KubeModule
}

func (c *CheckCertsModule) Init() {
	c.Name = "CheckCertsModule"

	check := &module.RemoteTask{
		Name:     "CheckClusterCerts",
		Desc:     "Check cluster certs",
		Hosts:    c.Runtime.GetHostsByRole(common.Master),
		Action:   new(ListClusterCerts),
		Parallel: true,
	}

	c.Tasks = []module.Task{
		check,
	}
}

type PrintClusterCertsModule struct {
	common.KubeModule
}

func (p *PrintClusterCertsModule) Init() {
	p.Name = "PrintClusterCertsModule"
	p.Desc = "Display cluster certs form"

	display := &module.LocalTask{
		Name:   "DisplayCertsForm",
		Desc:   "Display cluster certs form",
		Action: new(DisplayForm),
	}

	p.Tasks = []module.Task{
		display,
	}
}

type RenewCertsModule struct {
	common.KubeModule
}

func (r *RenewCertsModule) Init() {
	r.Name = "RenewCertsModule"

	renew := &module.RemoteTask{
		Name:     "RenewCerts",
		Desc:     "Renew control-plane certs",
		Hosts:    r.Runtime.GetHostsByRole(common.Master),
		Action:   new(RenewCerts),
		Parallel: false,
		Retry:    5,
	}

	copyKubeConfig := &module.RemoteTask{
		Name:     "CopyKubeConfig",
		Desc:     "Copy admin.conf to ~/.kube/config",
		Hosts:    r.Runtime.GetHostsByRole(common.Master),
		Action:   new(kubernetes.CopyKubeConfigForControlPlane),
		Parallel: true,
		Retry:    2,
	}

	fetchKubeConfig := &module.RemoteTask{
		Name:     "FetchKubeConfig",
		Desc:     "Fetch kube config file from control-plane",
		Hosts:    r.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.OnlyFirstMaster),
		Action:   new(FetchKubeConfig),
		Parallel: true,
	}

	syncKubeConfig := &module.RemoteTask{
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

	r.Tasks = []module.Task{
		renew,
		copyKubeConfig,
		fetchKubeConfig,
		syncKubeConfig,
	}
}
