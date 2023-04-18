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

package dns

import (
	"path/filepath"

	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/action"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/prepare"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/task"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/util"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/images"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/plugins/dns/templates"
)

type ClusterDNSModule struct {
	common.KubeModule
}

func (c *ClusterDNSModule) Init() {
	c.Name = "ClusterDNSModule"
	c.Desc = "Deploy cluster dns"

	generateCoreDNSSvc := &task.RemoteTask{
		Name:  "GenerateCoreDNSSvc",
		Desc:  "Generate coredns service",
		Hosts: c.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&CoreDNSExist{Not: true},
		},
		Action: &action.Template{
			Template: templates.CorednsService,
			Dst:      filepath.Join(common.KubeConfigDir, templates.CorednsService.Name()),
			Data: util.Data{
				"ClusterIP": c.KubeConf.Cluster.CorednsClusterIP(),
			},
		},
		Parallel: true,
	}

	override := &task.RemoteTask{
		Name:  "OverrideCoreDNSService",
		Desc:  "Override coredns service",
		Hosts: c.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&CoreDNSExist{Not: true},
		},
		Action:   new(OverrideCoreDNS),
		Parallel: true,
	}

	generateNodeLocalDNS := &task.RemoteTask{
		Name:  "GenerateNodeLocalDNS",
		Desc:  "Generate nodelocaldns",
		Hosts: c.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(EnableNodeLocalDNS),
		},
		Action: &action.Template{
			Template: templates.NodeLocalDNSService,
			Dst:      filepath.Join(common.KubeConfigDir, templates.NodeLocalDNSService.Name()),
			Data: util.Data{
				"NodelocaldnsImage": images.GetImage(c.Runtime, c.KubeConf, "k8s-dns-node-cache").ImageName(),
			},
		},
		Parallel: true,
	}

	applyNodeLocalDNS := &task.RemoteTask{
		Name:  "DeployNodeLocalDNS",
		Desc:  "Deploy nodelocaldns",
		Hosts: c.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(EnableNodeLocalDNS),
		},
		Action:   new(DeployNodeLocalDNS),
		Parallel: true,
		Retry:    5,
	}

	generateNodeLocalDNSConfigMap := &task.RemoteTask{
		Name:  "GenerateNodeLocalDNSConfigMap",
		Desc:  "Generate nodelocaldns configmap",
		Hosts: c.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(EnableNodeLocalDNS),
			new(NodeLocalDNSConfigMapNotExist),
		},
		Action:   new(GenerateNodeLocalDNSConfigMap),
		Parallel: true,
	}

	applyNodeLocalDNSConfigMap := &task.RemoteTask{
		Name:  "ApplyNodeLocalDNSConfigMap",
		Desc:  "Apply nodelocaldns configmap",
		Hosts: c.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(EnableNodeLocalDNS),
			new(NodeLocalDNSConfigMapNotExist),
		},
		Action:   new(ApplyNodeLocalDNSConfigMap),
		Parallel: true,
		Retry:    5,
	}

	c.Tasks = []task.Interface{
		generateCoreDNSSvc,
		override,
		generateNodeLocalDNS,
		applyNodeLocalDNS,
		generateNodeLocalDNSConfigMap,
		applyNodeLocalDNSConfigMap,
	}
}
