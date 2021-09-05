package etcd

import (
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/modules"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"github.com/kubesphere/kubekey/pkg/pipelines/etcd/templates"
	"path/filepath"
)

type ETCDPreCheckModule struct {
	common.KubeModule
}

func (e *ETCDPreCheckModule) Init() {
	e.Name = "ETCDPreCheckModule"
	getStatus := &modules.Task{
		Name:     "GetETCDStatus",
		Desc:     "get etcd status",
		Hosts:    e.Runtime.GetHostsByRole(common.ETCD),
		Action:   new(GetStatus),
		Parallel: false,
		Retry:    0,
	}
	e.Tasks = []*modules.Task{
		getStatus,
	}
}

type ETCDModule struct {
	common.KubeModule
}

func (e *ETCDModule) Init() {
	e.Name = "ETCDModule"

	generateCertsScript := &modules.Task{
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
	generateOpenSSLConf := &modules.Task{
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

	execCertsScript := &modules.Task{
		Name:     "ExecCertsScript",
		Desc:     "exec certs script",
		Hosts:    e.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  new(FirstETCDNode),
		Action:   new(ExecCertsScript),
		Parallel: true,
		Retry:    1,
	}

	syncCertsFile := &modules.Task{
		Name:     "SyncCertsFile",
		Desc:     "synchronize certs file",
		Hosts:    e.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  &FirstETCDNode{Not: true},
		Action:   new(SyncCertsFile),
		Parallel: true,
		Retry:    1,
	}

	syncCertsToMaster := &modules.Task{
		Name:     "SyncCertsFileToMaster",
		Desc:     "synchronize certs file to master",
		Hosts:    e.Runtime.GetHostsByRole(common.Master),
		Prepare:  &common.OnlyETCD{Not: true},
		Action:   new(SyncCertsFile),
		Parallel: true,
		Retry:    1,
	}

	installETCDBinary := &modules.Task{
		Name:     "InstallETCDBinary",
		Desc:     "install etcd using binary",
		Hosts:    e.Runtime.GetHostsByRole(common.ETCD),
		Action:   new(InstallETCDBinary),
		Parallel: true,
		Retry:    1,
	}

	generateETCDService := &modules.Task{
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

	accessAddress := &modules.Task{
		Name:     "GenerateAccessAddress",
		Desc:     "generate access address",
		Hosts:    e.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  new(FirstETCDNode),
		Action:   new(GenerateAccessAddress),
		Parallel: true,
		Retry:    1,
	}

	existETCDHealthCheck := &modules.Task{
		Name:     "ExistETCDHealthCheck",
		Desc:     "health check on exist etcd",
		Hosts:    e.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  new(NodeETCDExist),
		Action:   new(HealthCheck),
		Parallel: true,
		Retry:    20,
	}

	generateETCDConfig := &modules.Task{
		Name:     "GenerateETCDConfig",
		Desc:     "generate etcd.env config on new etcd",
		Hosts:    e.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  &NodeETCDExist{Not: true},
		Action:   &GenerateConfig{},
		Parallel: false,
	}

	joinMember := &modules.Task{
		Name:     "JoinETCDMember",
		Desc:     "join etcd member",
		Hosts:    e.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  &NodeETCDExist{Not: true},
		Action:   new(JoinMember),
		Parallel: false,
	}

	restart := &modules.Task{
		Name:     "RestartETCD",
		Desc:     "restart etcd",
		Hosts:    e.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  &NodeETCDExist{Not: true},
		Action:   new(RestartETCD),
		Parallel: true,
	}

	newETCDNodeHealthCheck := &modules.Task{
		Name:     "NewETCDNodeHealthCheck",
		Desc:     "health check on new etcd",
		Hosts:    e.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  &NodeETCDExist{Not: true},
		Action:   new(HealthCheck),
		Parallel: true,
		Retry:    20,
	}

	checkMember := &modules.Task{
		Name:     "CheckETCDMember",
		Desc:     "check etcd member",
		Hosts:    e.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  &NodeETCDExist{Not: true},
		Action:   new(CheckMember),
		Parallel: true,
	}

	allRefreshETCDConfig := &modules.Task{
		Name:     "AllRefreshETCDConfig",
		Desc:     "refresh etcd.env config on all etcd",
		Hosts:    e.Runtime.GetHostsByRole(common.ETCD),
		Action:   new(RefreshConfig),
		Parallel: false,
	}

	refreshETCDConfigToExist := &modules.Task{
		Name:     "RefreshETCDConfigToExist",
		Desc:     "refresh etcd.env config to exist mode on all etcd",
		Hosts:    e.Runtime.GetHostsByRole(common.ETCD),
		Action:   &RefreshConfig{ToExisting: true},
		Parallel: false,
	}

	allETCDNodeHealthCheck := &modules.Task{
		Name:     "AllETCDNodeHealthCheck",
		Desc:     "health check on all etcd",
		Hosts:    e.Runtime.GetHostsByRole(common.ETCD),
		Action:   new(HealthCheck),
		Parallel: true,
		Retry:    20,
	}

	backupETCD := &modules.Task{
		Name:     "BackupETCD",
		Desc:     "backup etcd data regularly",
		Hosts:    e.Runtime.GetHostsByRole(common.ETCD),
		Action:   new(BackupETCD),
		Parallel: true,
	}

	if v, ok := e.RootCache.Get(ETCDCluster); ok {
		cluster := v.(EtcdCluster)
		if cluster.clusterExist {
			e.Tasks = []*modules.Task{
				generateCertsScript,
				generateOpenSSLConf,
				execCertsScript,
				syncCertsFile,
				syncCertsToMaster,
				installETCDBinary,
				generateETCDService,
				accessAddress,
				existETCDHealthCheck,
				generateETCDConfig,
				joinMember,
				restart,
				newETCDNodeHealthCheck,
				checkMember,
				allRefreshETCDConfig,
				allETCDNodeHealthCheck,
				backupETCD,
			}
		} else {
			e.Tasks = []*modules.Task{
				generateCertsScript,
				generateOpenSSLConf,
				execCertsScript,
				syncCertsFile,
				syncCertsToMaster,
				installETCDBinary,
				generateETCDService,
				accessAddress,
				existETCDHealthCheck,
				generateETCDConfig,
				allRefreshETCDConfig,
				restart,
				allETCDNodeHealthCheck,
				refreshETCDConfigToExist,
				allETCDNodeHealthCheck,
				backupETCD,
			}
		}
	}
}
