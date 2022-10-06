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

package loadbalancer

import (
	"path/filepath"

	kubekeyapiv1alpha2 "github.com/kubesphere/kubekey/cmd/kk/apis/kubekey/v1alpha2"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/core/action"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/core/connector"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/core/prepare"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/core/task"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/core/util"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/loadbalancer/templates"
)

type HaproxyModule struct {
	common.KubeModule
	Skip bool
}

func (h *HaproxyModule) IsSkip() bool {
	return h.Skip
}

func (h *HaproxyModule) Init() {
	h.Name = "InternalLoadbalancerModule"
	h.Desc = "Install internal load balancer"

	haproxyCfg := &task.RemoteTask{
		Name:    "GenerateHaproxyConfig",
		Desc:    "Generate haproxy.cfg",
		Hosts:   h.Runtime.GetHostsByRole(common.Worker),
		Prepare: new(common.OnlyWorker),
		Action: &action.Template{
			Template: templates.HaproxyConfig,
			Dst:      filepath.Join(common.HaproxyDir, templates.HaproxyConfig.Name()),
			Data: util.Data{
				"MasterNodes":                          templates.MasterNodeStr(h.Runtime, h.KubeConf),
				"LoadbalancerApiserverPort":            kubekeyapiv1alpha2.DefaultApiserverPort,
				"LoadbalancerApiserverHealthcheckPort": 8081,
				"KubernetesType":                       h.KubeConf.Cluster.Kubernetes.Type,
			},
		},
		Parallel: true,
	}

	// Calculation config md5 as the checksum.
	// It will make load balancer reload when config changes.
	getMd5Sum := &task.RemoteTask{
		Name:     "GetChecksumFromConfig",
		Desc:     "Calculate the MD5 value according to haproxy.cfg",
		Hosts:    h.Runtime.GetHostsByRole(common.Worker),
		Prepare:  new(common.OnlyWorker),
		Action:   new(GetChecksum),
		Parallel: true,
	}

	haproxyManifestK8s := &task.RemoteTask{
		Name:  "GenerateHaproxyManifest",
		Desc:  "Generate haproxy manifest",
		Hosts: h.Runtime.GetHostsByRole(common.Worker),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyWorker),
			new(common.OnlyKubernetes),
		},
		Action:   new(GenerateHaproxyManifest),
		Parallel: true,
	}

	// UpdateKubeletConfig Update server field in kubelet.conf
	// When create a HA cluster by internal LB, we will set the server filed to 127.0.0.1:6443 (default) which in kubelet.conf.
	// Because of that, the control plone node's kubelet connect the local api-server.
	// And the work node's kubelet connect 127.0.0.1:6443 (default) that is proxy by the node's local nginx.
	updateKubeletConfig := &task.RemoteTask{
		Name:  "UpdateKubeletConfig",
		Desc:  "Update kubelet config",
		Hosts: h.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyKubernetes),
			new(updateKubeletPrepare),
		},
		Action:   new(UpdateKubelet),
		Parallel: true,
		Retry:    3,
	}

	// updateKubeProxyConfig is used to update kube-proxy configmap and restart tge kube-proxy pod.
	updateKubeProxyConfig := &task.RemoteTask{
		Name:  "UpdateKubeProxyConfig",
		Desc:  "Update kube-proxy configmap",
		Hosts: []connector.Host{h.Runtime.GetHostsByRole(common.Master)[0]},
		Prepare: &prepare.PrepareCollection{
			new(common.EnableKubeProxy),
			new(common.OnlyKubernetes),
			new(common.OnlyFirstMaster),
			new(updateKubeProxyPrapre),
		},
		Action:   new(UpdateKubeProxy),
		Parallel: true,
		Retry:    3,
	}

	// UpdateHostsFile is used to update the '/etc/hosts'. Make the 'lb.kubesphere.local' address to set as 127.0.0.1.
	// All of the 'admin.conf' and '/.kube/config' will connect to 127.0.0.1:6443.
	updateHostsFile := &task.RemoteTask{
		Name:     "UpdateHostsFile",
		Desc:     "Update /etc/hosts",
		Hosts:    h.Runtime.GetHostsByRole(common.K8s),
		Action:   new(UpdateHosts),
		Parallel: true,
		Retry:    3,
	}

	h.Tasks = []task.Interface{
		haproxyCfg,
		getMd5Sum,
		haproxyManifestK8s,
		updateKubeletConfig,
		updateKubeProxyConfig,
		updateHostsFile,
	}
}

type KubevipModule struct {
	common.KubeModule
	Skip bool
}

func (k *KubevipModule) IsSkip() bool {
	return k.Skip
}

func (k *KubevipModule) Init() {
	k.Name = "InternalLoadbalancerModule"
	k.Desc = "Install internal load balancer"

	checkVIPAddress := &task.RemoteTask{
		Name:     "CheckVIPAddress",
		Desc:     "Check VIP Address",
		Hosts:    k.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.OnlyFirstMaster),
		Action:   new(CheckVIPAddress),
		Parallel: true,
	}

	getInterface := &task.RemoteTask{
		Name:     "GetNodeInterface",
		Desc:     "Get Node Interface",
		Hosts:    k.Runtime.GetHostsByRole(common.Master),
		Action:   new(GetInterfaceName),
		Parallel: true,
	}

	kubevipManifestOnlyFirstMaster := &task.RemoteTask{
		Name:     "GenerateKubevipManifest",
		Desc:     "Generate kubevip manifest at first master",
		Hosts:    k.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.OnlyFirstMaster),
		Action:   new(GenerateKubevipManifest),
		Parallel: true,
	}

	kubevipManifestNotFirstMaster := &task.RemoteTask{
		Name:     "GenerateKubevipManifest",
		Desc:     "Generate kubevip manifest at other master",
		Hosts:    k.Runtime.GetHostsByRole(common.Master),
		Prepare:  &common.OnlyFirstMaster{Not: true},
		Action:   new(GenerateKubevipManifest),
		Parallel: true,
	}

	if exist, _ := k.BaseModule.PipelineCache.GetMustBool(common.ClusterExist); exist {
		k.Tasks = []task.Interface{
			checkVIPAddress,
			getInterface,
			kubevipManifestNotFirstMaster,
		}
	} else {
		k.Tasks = []task.Interface{
			checkVIPAddress,
			getInterface,
			kubevipManifestOnlyFirstMaster,
		}
	}
}

type K3sHaproxyModule struct {
	common.KubeModule
	Skip bool
}

func (k *K3sHaproxyModule) IsSkip() bool {
	return k.Skip
}

func (k *K3sHaproxyModule) Init() {
	k.Name = "InternalLoadbalancerModule"
	k.Name = "Install internal load balancer"

	haproxyCfg := &task.RemoteTask{
		Name:    "GenerateHaproxyConfig",
		Desc:    "Generate haproxy.cfg",
		Hosts:   k.Runtime.GetHostsByRole(common.Worker),
		Prepare: new(common.OnlyWorker),
		Action: &action.Template{
			Template: templates.HaproxyConfig,
			Dst:      filepath.Join(common.HaproxyDir, templates.HaproxyConfig.Name()),
			Data: util.Data{
				"MasterNodes":                          templates.MasterNodeStr(k.Runtime, k.KubeConf),
				"LoadbalancerApiserverPort":            k.KubeConf.Cluster.ControlPlaneEndpoint.Port,
				"LoadbalancerApiserverHealthcheckPort": 8081,
				"KubernetesType":                       k.KubeConf.Cluster.Kubernetes.Type,
			},
		},
		Parallel: true,
	}

	// Calculation config md5 as the checksum.
	// It will make load balancer reload when config changes.
	getMd5Sum := &task.RemoteTask{
		Name:     "GetChecksumFromConfig",
		Desc:     "Calculate the MD5 value according to haproxy.cfg",
		Hosts:    k.Runtime.GetHostsByRole(common.Worker),
		Prepare:  new(common.OnlyWorker),
		Action:   new(GetChecksum),
		Parallel: true,
	}

	haproxyManifestK3s := &task.RemoteTask{
		Name:  "GenerateHaproxyManifestK3s",
		Desc:  "Generate haproxy manifest",
		Hosts: k.Runtime.GetHostsByRole(common.Worker),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyWorker),
			new(common.OnlyK3s),
		},
		Action:   new(GenerateK3sHaproxyManifest),
		Parallel: true,
	}

	updateK3sConfig := &task.RemoteTask{
		Name:  "UpdateK3sConfig",
		Desc:  "Update k3s config",
		Hosts: k.Runtime.GetHostsByRole(common.Worker),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyK3s),
			new(updateK3sPrepare),
		},
		Action:   new(UpdateK3s),
		Parallel: true,
		Retry:    3,
	}

	// UpdateHostsFile is used to update the '/etc/hosts'. Make the 'lb.kubesphere.local' address to set as 127.0.0.1.
	// All of the 'admin.conf' and '/.kube/config' will connect to 127.0.0.1:6443.
	updateHostsFile := &task.RemoteTask{
		Name:     "UpdateHostsFile",
		Desc:     "Update /etc/hosts",
		Hosts:    k.Runtime.GetHostsByRole(common.K8s),
		Action:   new(UpdateHosts),
		Parallel: true,
		Retry:    3,
	}

	k.Tasks = []task.Interface{
		haproxyCfg,
		getMd5Sum,
		haproxyManifestK3s,
		updateK3sConfig,
		updateHostsFile,
	}
}

type K3sKubevipModule struct {
	common.KubeModule
	Skip bool
}

func (k *K3sKubevipModule) IsSkip() bool {
	return k.Skip
}

func (k *K3sKubevipModule) Init() {
	k.Name = "InternalLoadbalancerModule"
	k.Name = "Install internal load balancer"

	checkVIPAddress := &task.RemoteTask{
		Name:     "CheckVIPAddress",
		Desc:     "Check VIP Address",
		Hosts:    k.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.OnlyFirstMaster),
		Action:   new(CheckVIPAddress),
		Parallel: true,
	}

	createManifestsFolder := &task.RemoteTask{
		Name:     "CreateManifestsFolder",
		Desc:     "Create Manifests Folder",
		Hosts:    k.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.OnlyFirstMaster),
		Action:   new(CreateManifestsFolder),
		Parallel: true,
	}

	getInterface := &task.RemoteTask{
		Name:     "GetNodeInterface",
		Desc:     "Get Node Interface",
		Hosts:    k.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.OnlyFirstMaster),
		Action:   new(GetInterfaceName),
		Parallel: true,
	}

	kubevipDaemonsetK3s := &task.RemoteTask{
		Name:     "GenerateKubevipManifest",
		Desc:     "Generate kubevip daemoset",
		Hosts:    k.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.OnlyFirstMaster),
		Action:   new(GenerateK3sKubevipDaemonset),
		Parallel: true,
	}

	k.Tasks = []task.Interface{
		checkVIPAddress,
		createManifestsFolder,
		getInterface,
		kubevipDaemonsetK3s,
	}
}

type DeleteVIPModule struct {
	common.KubeModule
	Skip bool
}

func (k *DeleteVIPModule) IsSkip() bool {
	return k.Skip
}

func (k *DeleteVIPModule) Init() {
	k.Name = "DeleteVIPModule"
	k.Desc = "Delete VIP"

	getInterface := &task.RemoteTask{
		Name:     "GetNodeInterface",
		Desc:     "Get Node Interface",
		Hosts:    k.Runtime.GetHostsByRole(common.Master),
		Action:   new(GetInterfaceName),
		Parallel: true,
	}

	DeleteVIP := &task.RemoteTask{
		Name:     "Delete VIP",
		Desc:     "Delete the VIP",
		Hosts:    k.Runtime.GetHostsByRole(common.Master),
		Action:   new(DeleteVIP),
		Parallel: true,
	}

	k.Tasks = []task.Interface{
		getInterface,
		DeleteVIP,
	}
}
