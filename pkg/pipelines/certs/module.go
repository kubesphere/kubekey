package certs

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/core/modules"
	"github.com/kubesphere/kubekey/pkg/core/prepare"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"github.com/kubesphere/kubekey/pkg/pipelines/kubernetes"
	"github.com/pkg/errors"
	"os"
	"text/tabwriter"
)

type CheckCertsModule struct {
	common.KubeModule
}

func (c *CheckCertsModule) Init() {
	c.Name = "CheckCertsModule"

	check := &modules.RemoteTask{
		Name:     "CheckClusterCerts",
		Desc:     "check cluster certs",
		Hosts:    c.Runtime.GetHostsByRole(common.Master),
		Action:   new(ListClusterCerts),
		Parallel: true,
	}

	c.Tasks = []modules.Task{
		check,
	}
}

type PrintClusterCertsModule struct {
	common.KubeModule
}

func (p *PrintClusterCertsModule) Init() {
	p.Name = "PrintClusterCertsModule"
	p.Desc = "display cluster certs form"
}

func (p *PrintClusterCertsModule) Run() error {
	certificates := make([]*Certificate, 0)
	caCertificates := make([]*CaCertificate, 0)

	for _, host := range p.Runtime.GetHostsByRole(common.Master) {
		certs, ok := host.GetCache().Get(common.Certificate)
		if !ok {
			return errors.New("get certificate failed by pipeline cache")
		}
		ca, ok := host.GetCache().Get(common.CaCertificate)
		if !ok {
			return errors.New("get ca certificate failed by pipeline cache")
		}
		hostCertificates := certs.([]*Certificate)
		hostCaCertificates := ca.([]*CaCertificate)
		certificates = append(certificates, hostCertificates...)
		caCertificates = append(caCertificates, hostCaCertificates...)
	}

	w := tabwriter.NewWriter(os.Stdout, 10, 4, 3, ' ', 0)
	_, _ = fmt.Fprintln(w, "CERTIFICATE\tEXPIRES\tRESIDUAL TIME\tCERTIFICATE AUTHORITY\tNODE")
	for _, cert := range certificates {
		s := fmt.Sprintf("%s\t%s\t%s\t%s\t%-8v",
			cert.Name,
			cert.Expires,
			cert.Residual,
			cert.AuthorityName,
			cert.NodeName,
		)

		_, _ = fmt.Fprintln(w, s)
		continue
	}
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "CERTIFICATE AUTHORITY\tEXPIRES\tRESIDUAL TIME\tNODE")
	for _, caCert := range caCertificates {
		c := fmt.Sprintf("%s\t%s\t%s\t%-8v",
			caCert.AuthorityName,
			caCert.Expires,
			caCert.Residual,
			caCert.NodeName,
		)

		_, _ = fmt.Fprintln(w, c)
		continue
	}

	_ = w.Flush()
	return nil
}

type RenewCertsModule struct {
	common.KubeModule
}

func (r *RenewCertsModule) Init() {
	r.Name = "RenewCertsModule"

	renew := &modules.RemoteTask{
		Name:     "RenewCerts",
		Desc:     "Renew control-plane certs",
		Hosts:    r.Runtime.GetHostsByRole(common.Master),
		Action:   new(RenewCerts),
		Parallel: false,
		Retry:    5,
	}

	copyKubeConfig := &modules.RemoteTask{
		Name:     "CopyKubeConfig",
		Desc:     "Copy admin.conf to ~/.kube/config",
		Hosts:    r.Runtime.GetHostsByRole(common.Master),
		Action:   new(kubernetes.CopyKubeConfigForControlPlane),
		Parallel: true,
		Retry:    2,
	}

	fetchKubeConfig := &modules.RemoteTask{
		Name:     "FetchKubeConfig",
		Desc:     "Fetch kube config file from control-plane",
		Hosts:    r.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.OnlyFirstMaster),
		Action:   new(FetchKubeConfig),
		Parallel: true,
	}

	syncKubeConfig := &modules.RemoteTask{
		Name:  "SyncKubeConfig",
		Desc:  "synchronize kube config to worker",
		Hosts: r.Runtime.GetHostsByRole(common.Worker),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyWorker),
		},
		Action:   new(SyneKubeConfigToWorker),
		Parallel: true,
		Retry:    3,
	}

	r.Tasks = []modules.Task{
		renew,
		copyKubeConfig,
		fetchKubeConfig,
		syncKubeConfig,
	}
}
