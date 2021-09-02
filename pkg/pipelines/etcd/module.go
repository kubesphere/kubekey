package etcd

import (
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/modules"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"github.com/kubesphere/kubekey/pkg/pipelines/etcd/templates"
	"path/filepath"
)

type ETCDModule struct {
	common.KubeModule
}

func (e *ETCDModule) Init() {
	e.Name = "ETCDModule"

	getStatus := modules.Task{
		Name:     "GetETCDStatus",
		Desc:     "get etcd status",
		Hosts:    e.Runtime.GetHostsByRole(common.ETCD),
		Action:   new(GetStatus),
		Parallel: false,
		Retry:    0,
	}

	generateCertsScript := modules.Task{
		Name:    "GenerateCertsScript",
		Desc:    "generate certs script",
		Hosts:   e.Runtime.GetHostsByRole(common.ETCD),
		Prepare: new(FirstETCDNode),
		Action: &action.Template{
			Template: templates.EtcdSslScript,
			Dst:      filepath.Join(common.ETCDCertDir, "make-ssl-etcd.sh"),
			Data: util.Data{
				"Masters": templates.GenerateHosts(e.Runtime.GetHostsByRole(common.ETCD)),
				"Hosts":   templates.GenerateHosts(e.Runtime.GetHostsByRole(common.Master)),
			},
		},
		Parallel: true,
		Retry:    1,
	}

	dnsList, ipList := templates.DNSAndIp(e.KubeConf)
	generateOpenSSLConf := modules.Task{
		Name:    "GenerateOpenSSLConf",
		Desc:    "generate OpenSSL config",
		Hosts:   e.Runtime.GetHostsByRole(common.ETCD),
		Prepare: new(FirstETCDNode),
		Action: &action.Template{
			Template: templates.ETCDOpenSSLConf,
			Dst:      filepath.Join(common.ETCDCertDir, "openssl.conf"),
			Data: util.Data{
				"Dns": dnsList,
				"Ips": ipList,
			},
		},
		Parallel: true,
		Retry:    1,
	}

	execCertsScript := modules.Task{
		Name:     "ExecCertsScript",
		Desc:     "exec certs script",
		Hosts:    e.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  new(FirstETCDNode),
		Action:   new(ExecCertsScript),
		Parallel: true,
		Retry:    1,
	}

	syncCertsFile := modules.Task{
		Name:     "SyncCertsFile",
		Desc:     "synchronize certs file",
		Hosts:    e.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  &FirstETCDNode{Not: true},
		Action:   new(SyncCertsFile),
		Parallel: true,
		Retry:    1,
	}

	syncCertsToMaster := modules.Task{
		Name:     "SyncCertsFileToMaster",
		Desc:     "synchronize certs file to master",
		Hosts:    e.Runtime.GetHostsByRole(common.Master),
		Prepare:  &common.OnlyETCD{Not: true},
		Action:   new(SyncCertsFile),
		Parallel: true,
		Retry:    1,
	}

	installETCDBinary := modules.Task{
		Name:     "InstallETCDBinary",
		Desc:     "install etcd using binary",
		Hosts:    e.Runtime.GetHostsByRole(common.ETCD),
		Action:   new(InstallETCDBinary),
		Parallel: true,
		Retry:    1,
	}

	generateETCDService := modules.Task{
		Name:  "GenerateETCDService",
		Desc:  "generate etcd service",
		Hosts: e.Runtime.GetHostsByRole(common.ETCD),
		Action: &action.Template{
			Template: templates.ETCDService,
			Dst:      "/etc/systemd/system/etcd.service",
		},
		Parallel: true,
		Retry:    1,
	}

	accessAddress := modules.Task{
		Name:     "generateAccessAddress",
		Desc:     "generate access address",
		Hosts:    e.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  new(FirstETCDNode),
		Action:   new(GenerateAccessAddress),
		Parallel: true,
		Retry:    1,
	}

	e.Tasks = []modules.Task{
		getStatus,
		generateCertsScript,
		generateOpenSSLConf,
		execCertsScript,
		syncCertsFile,
		syncCertsToMaster,
		installETCDBinary,
		generateETCDService,
		accessAddress,
	}
}
