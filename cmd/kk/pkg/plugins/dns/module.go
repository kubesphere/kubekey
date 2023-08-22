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

	generateCorednsConfigMap := &task.RemoteTask{
		Name:  "GenerateCorednsConfigMap",
		Desc:  "Generate coredns configmap",
		Hosts: c.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
		},
		Action: &action.Template{
			Template: templates.CorednsConfigMap,
			Dst:      filepath.Join(common.KubeConfigDir, templates.CorednsConfigMap.Name()),
			Data: util.Data{
				"DNSEtcHosts":        c.KubeConf.Cluster.DNS.DNSEtcHosts,
				"ExternalZones":      c.KubeConf.Cluster.DNS.CoreDNS.ExternalZones,
				"AdditionalConfigs":  c.KubeConf.Cluster.DNS.CoreDNS.AdditionalConfigs,
				"RewriteBlock":       c.KubeConf.Cluster.DNS.CoreDNS.RewriteBlock,
				"ClusterDomain":      c.KubeConf.Cluster.Kubernetes.DNSDomain,
				"UpstreamDNSServers": c.KubeConf.Cluster.DNS.CoreDNS.UpstreamDNSServers,
			},
		},
		Parallel: true,
	}

	applyCorednsConfigMap := &task.RemoteTask{
		Name:  "ApplyCorednsConfigMap",
		Desc:  "Apply coredns configmap",
		Hosts: c.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
		},
		Action:   new(ApplyCorednsConfigMap),
		Parallel: true,
		Retry:    5,
	}

	generateCoreDNS := &task.RemoteTask{
		Name:  "GenerateCoreDNS",
		Desc:  "Generate coredns manifests",
		Hosts: c.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
		},
		Action:   new(GenerateCorednsmanifests),
		Parallel: true,
	}

	deployCoredns := &task.RemoteTask{
		Name:  "DeployCoreDNS",
		Desc:  "Deploy coredns",
		Hosts: c.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
		},
		Action:   new(DeployCoreDNS),
		Parallel: true,
	}

	generateNodeLocalDNSConfigMap := &task.RemoteTask{
		Name:  "GenerateNodeLocalDNSConfigMap",
		Desc:  "Generate nodelocaldns configmap",
		Hosts: c.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(EnableNodeLocalDNS),
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
		},
		Action:   new(ApplyNodeLocalDNSConfigMap),
		Parallel: true,
		Retry:    5,
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
				"DNSEtcHosts":       c.KubeConf.Cluster.DNS.DNSEtcHosts,
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

	c.Tasks = []task.Interface{
		generateCorednsConfigMap,
		applyCorednsConfigMap,
		generateCoreDNS,
		deployCoredns,
		generateNodeLocalDNSConfigMap,
		applyNodeLocalDNSConfigMap,
		generateNodeLocalDNS,
		applyNodeLocalDNS,
	}
}
