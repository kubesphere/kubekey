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

package etcd

import (
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/task"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/etcd/templates"
	"path/filepath"
)

type PreCheckModule struct {
	common.KubeModule
}

func (p *PreCheckModule) Init() {
	p.Name = "ETCDPreCheckModule"
	p.Desc = "Get ETCD cluster status"

	getStatus := &task.RemoteTask{
		Name:     "GetETCDStatus",
		Desc:     "Get etcd status",
		Hosts:    p.Runtime.GetHostsByRole(common.ETCD),
		Action:   new(GetStatus),
		Parallel: false,
		Retry:    0,
	}
	p.Tasks = []task.Interface{
		getStatus,
	}
}

type CertsModule struct {
	common.KubeModule
}

func (c *CertsModule) Init() {
	c.Name = "CertsModule"
	c.Desc = "Sign ETCD cluster certs"

	generateCertsScript := &task.RemoteTask{
		Name:    "GenerateCertsScript",
		Desc:    "Generate certs script",
		Hosts:   c.Runtime.GetHostsByRole(common.ETCD),
		Prepare: new(FirstETCDNode),
		Action: &action.Template{
			Template: templates.EtcdSslScript,
			Dst:      filepath.Join(common.ETCDCertDir, templates.EtcdSslScript.Name()),
			Data: util.Data{
				"Masters": templates.GenerateHosts(c.Runtime.GetHostsByRole(common.ETCD)),
				"Hosts":   templates.GenerateHosts(c.Runtime.GetHostsByRole(common.Master)),
			},
		},
		Parallel: true,
		Retry:    1,
	}

	dnsList, ipList := templates.DNSAndIp(c.KubeConf)
	generateOpenSSLConf := &task.RemoteTask{
		Name:    "GenerateOpenSSLConf",
		Desc:    "Generate OpenSSL config",
		Hosts:   c.Runtime.GetHostsByRole(common.ETCD),
		Prepare: new(FirstETCDNode),
		Action: &action.Template{
			Template: templates.ETCDOpenSSLConf,
			Dst:      filepath.Join(common.ETCDCertDir, templates.ETCDOpenSSLConf.Name()),
			Data: util.Data{
				"Dns": dnsList,
				"Ips": ipList,
			},
		},
		Parallel: true,
		Retry:    1,
	}

	execCertsScript := &task.RemoteTask{
		Name:     "ExecCertsScript",
		Desc:     "Exec certs script",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  new(FirstETCDNode),
		Action:   new(ExecCertsScript),
		Parallel: true,
		Retry:    1,
	}

	syncCertsFile := &task.RemoteTask{
		Name:     "SyncCertsFile",
		Desc:     "Synchronize certs file",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  &FirstETCDNode{Not: true},
		Action:   new(SyncCertsFile),
		Parallel: true,
		Retry:    1,
	}

	syncCertsToMaster := &task.RemoteTask{
		Name:     "SyncCertsFileToMaster",
		Desc:     "Synchronize certs file to master",
		Hosts:    c.Runtime.GetHostsByRole(common.Master),
		Prepare:  &common.OnlyETCD{Not: true},
		Action:   new(SyncCertsFile),
		Parallel: true,
		Retry:    1,
	}

	c.Tasks = []task.Interface{
		generateCertsScript,
		generateOpenSSLConf,
		execCertsScript,
		syncCertsFile,
		syncCertsToMaster,
	}
}

type InstallETCDBinaryModule struct {
	common.KubeModule
}

func (i *InstallETCDBinaryModule) Init() {
	i.Name = "InstallETCDBinaryModule"
	i.Desc = "Install ETCD cluster"

	installETCDBinary := &task.RemoteTask{
		Name:     "InstallETCDBinary",
		Desc:     "Install etcd using binary",
		Hosts:    i.Runtime.GetHostsByRole(common.ETCD),
		Action:   new(InstallETCDBinary),
		Parallel: true,
		Retry:    1,
	}

	generateETCDService := &task.RemoteTask{
		Name:  "GenerateETCDService",
		Desc:  "Generate etcd service",
		Hosts: i.Runtime.GetHostsByRole(common.ETCD),
		Action: &action.Template{
			Template: templates.ETCDService,
			Dst:      "/etc/systemd/system/etcd.service",
		},
		Parallel: true,
		Retry:    1,
	}

	accessAddress := &task.RemoteTask{
		Name:     "GenerateAccessAddress",
		Desc:     "Generate access address",
		Hosts:    i.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  new(FirstETCDNode),
		Action:   new(GenerateAccessAddress),
		Parallel: true,
		Retry:    1,
	}

	i.Tasks = []task.Interface{
		installETCDBinary,
		generateETCDService,
		accessAddress,
	}
}

type ConfigureModule struct {
	common.KubeModule
}

func (e *ConfigureModule) Init() {
	e.Name = "ETCDConfigureModule"
	e.Desc = "Configure ETCD cluster"

	if v, ok := e.PipelineCache.Get(common.ETCDCluster); ok {
		cluster := v.(*EtcdCluster)
		if !cluster.clusterExist {
			e.Tasks = handleNewCluster(e)
		} else {
			e.Tasks = handleExistCluster(e)
		}
	}
}

func handleNewCluster(c *ConfigureModule) []task.Interface {

	existETCDHealthCheck := &task.RemoteTask{
		Name:     "ExistETCDHealthCheck",
		Desc:     "Health check on exist etcd",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  new(NodeETCDExist),
		Action:   new(HealthCheck),
		Parallel: true,
		Retry:    20,
	}

	generateETCDConfig := &task.RemoteTask{
		Name:     "GenerateETCDConfig",
		Desc:     "Generate etcd.env config on new etcd",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  &NodeETCDExist{Not: true},
		Action:   new(GenerateConfig),
		Parallel: false,
	}

	allRefreshETCDConfig := &task.RemoteTask{
		Name:     "AllRefreshETCDConfig",
		Desc:     "Refresh etcd.env config on all etcd",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Action:   new(RefreshConfig),
		Parallel: false,
	}

	restart := &task.RemoteTask{
		Name:     "RestartETCD",
		Desc:     "Restart etcd",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  &NodeETCDExist{Not: true},
		Action:   new(RestartETCD),
		Parallel: true,
	}

	allETCDNodeHealthCheck := &task.RemoteTask{
		Name:     "AllETCDNodeHealthCheck",
		Desc:     "Health check on all etcd",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Action:   new(HealthCheck),
		Parallel: true,
		Retry:    20,
	}

	refreshETCDConfigToExist := &task.RemoteTask{
		Name:     "RefreshETCDConfigToExist",
		Desc:     "Refresh etcd.env config to exist mode on all etcd",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Action:   &RefreshConfig{ToExisting: true},
		Parallel: false,
	}

	tasks := []task.Interface{
		existETCDHealthCheck,
		generateETCDConfig,
		allRefreshETCDConfig,
		restart,
		allETCDNodeHealthCheck,
		refreshETCDConfigToExist,
		allETCDNodeHealthCheck,
	}
	return tasks
}

func handleExistCluster(c *ConfigureModule) []task.Interface {

	existETCDHealthCheck := &task.RemoteTask{
		Name:     "ExistETCDHealthCheck",
		Desc:     "Health check on exist etcd",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  new(NodeETCDExist),
		Action:   new(HealthCheck),
		Parallel: true,
		Retry:    20,
	}

	generateETCDConfig := &task.RemoteTask{
		Name:     "GenerateETCDConfig",
		Desc:     "Generate etcd.env config on new etcd",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  &NodeETCDExist{Not: true},
		Action:   new(GenerateConfig),
		Parallel: false,
	}

	joinMember := &task.RemoteTask{
		Name:     "JoinETCDMember",
		Desc:     "Join etcd member",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  &NodeETCDExist{Not: true},
		Action:   new(JoinMember),
		Parallel: false,
	}

	restart := &task.RemoteTask{
		Name:     "RestartETCD",
		Desc:     "Restart etcd",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  &NodeETCDExist{Not: true},
		Action:   new(RestartETCD),
		Parallel: true,
	}

	newETCDNodeHealthCheck := &task.RemoteTask{
		Name:     "NewETCDNodeHealthCheck",
		Desc:     "Health check on new etcd",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  &NodeETCDExist{Not: true},
		Action:   new(HealthCheck),
		Parallel: true,
		Retry:    20,
	}

	checkMember := &task.RemoteTask{
		Name:     "CheckETCDMember",
		Desc:     "Check etcd member",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  &NodeETCDExist{Not: true},
		Action:   new(CheckMember),
		Parallel: true,
	}

	allRefreshETCDConfig := &task.RemoteTask{
		Name:     "AllRefreshETCDConfig",
		Desc:     "Refresh etcd.env config on all etcd",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Action:   new(RefreshConfig),
		Parallel: false,
	}

	allETCDNodeHealthCheck := &task.RemoteTask{
		Name:     "AllETCDNodeHealthCheck",
		Desc:     "Health check on all etcd",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Action:   new(HealthCheck),
		Parallel: true,
		Retry:    20,
	}

	tasks := []task.Interface{
		existETCDHealthCheck,
		generateETCDConfig,
		joinMember,
		restart,
		newETCDNodeHealthCheck,
		checkMember,
		allRefreshETCDConfig,
		allETCDNodeHealthCheck,
	}
	return tasks
}

type BackupModule struct {
	common.KubeModule
}

func (b *BackupModule) Init() {
	b.Name = "ETCDBackupModule"
	b.Desc = "Backup ETCD cluster data"

	backupETCD := &task.RemoteTask{
		Name:     "BackupETCD",
		Desc:     "Backup etcd data regularly",
		Hosts:    b.Runtime.GetHostsByRole(common.ETCD),
		Action:   new(BackupETCD),
		Parallel: true,
	}

	b.Tasks = []task.Interface{
		backupETCD,
	}
}
