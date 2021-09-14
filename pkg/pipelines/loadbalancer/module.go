package loadbalancer

import (
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/modules"
	"github.com/kubesphere/kubekey/pkg/core/prepare"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"github.com/kubesphere/kubekey/pkg/pipelines/loadbalancer/templates"
	"path/filepath"
)

type HaproxyModule struct {
	common.KubeModule
	Skip bool
}

func (h *HaproxyModule) IsSkip() bool {
	return h.Skip
}

func (h *HaproxyModule) Init() {
	h.Name = "InternalLoadbalancer"

	//makeConfigDir := &modules.Task{
	//	Name:     "MakeHaproxyConfigDir",
	//	Desc:     "Make dir /etc/kubekey/haproxy",
	//	Hosts:    h.Runtime.GetHostsByRole(common.Worker),
	//	Prepare:  new(common.OnlyWorker),
	//	Action:   new(haproxyPreparatoryWork),
	//	Parallel: true,
	//}

	haproxyCfg := &modules.Task{
		Name:    "GenerateHaproxyConfig",
		Desc:    "Generate haproxy.cfg",
		Hosts:   h.Runtime.GetHostsByRole(common.Worker),
		Prepare: new(common.OnlyWorker),
		Action: &action.Template{
			Template: templates.HaproxyConfig,
			Dst:      filepath.Join(common.HaproxyDir, templates.HaproxyConfig.Name()),
			Data: util.Data{
				"MasterNodes":                          templates.MasterNodeStr(h.Runtime, h.KubeConf),
				"LoadbalancerApiserverPort":            h.KubeConf.Cluster.ControlPlaneEndpoint.Port,
				"LoadbalancerApiserverHealthcheckPort": 8081,
				"KubernetesType":                       h.KubeConf.Cluster.Kubernetes.Type,
			},
		},
		Parallel: true,
	}

	// Calculation config md5 as the checksum.
	// It will make load balancer reload when config changes.
	getMd5Sum := &modules.Task{
		Name:     "GetChecksumFromConfig",
		Desc:     "Calculate the MD5 value according to haproxy.cfg",
		Hosts:    h.Runtime.GetHostsByRole(common.Worker),
		Prepare:  new(common.OnlyWorker),
		Action:   new(getChecksum),
		Parallel: true,
	}

	// todo: 拆分k3s和k8s的module，分别为不太的task数组
	haproxyManifestK3s := &modules.Task{
		Name:  "GenerateHaproxyManifestK3s",
		Hosts: h.Runtime.GetHostsByRole(common.Worker),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyWorker),
			new(common.OnlyK3s),
		},
		Action:   new(GenerateHaproxyManifest),
		Parallel: true,
	}

	haproxyManifestK8s := &modules.Task{
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

	updateK3sConfig := &modules.Task{
		Name:  "UpdateK3sConfig",
		Hosts: h.Runtime.GetHostsByRole(common.Worker),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyK3s),
			new(updateK3sPrepare),
		},
		Action:   new(updateK3s),
		Parallel: true,
		Retry:    3,
	}

	// UpdateKubeletConfig Update server field in kubelet.conf
	// When create a HA cluster by internal LB, we will set the server filed to 127.0.0.1:6443 (default) which in kubelet.conf.
	// Because of that, the control plone node's kubelet connect the local api-server.
	// And the work node's kubelet connect 127.0.0.1:6443 (default) that is proxy by the node's local nginx.
	updateKubeletConfig := &modules.Task{
		Name:  "UpdateKubeletConfig",
		Desc:  "Update kubelet config",
		Hosts: h.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyKubernetes),
			new(updateKubeletPrepare),
		},
		Action:   new(updateKubelet),
		Parallel: true,
		Retry:    3,
	}

	// updateKubeProxyConfig is used to update kube-proxy configmap and restart tge kube-proxy pod.
	updateKubeProxyConfig := &modules.Task{
		Name:  "UpdateKubeProxyConfig",
		Desc:  "Update kube-proxy configmap",
		Hosts: []connector.Host{h.Runtime.GetHostsByRole(common.Master)[0]},
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyKubernetes),
			new(common.OnlyFirstMaster),
			new(updateKubeProxyPrapre),
		},
		Action:   new(updateKubeproxy),
		Parallel: true,
		Retry:    3,
	}

	// UpdateHostsFile is used to update the '/etc/hosts'. Make the 'lb.kubesphere.local' address to set as 127.0.0.1.
	// All of the 'admin.conf' and '/.kube/config' will connect to 127.0.0.1:6443.
	updateHostsFile := &modules.Task{
		Name:     "UpdateHostsFile",
		Desc:     "Update /etc/hosts",
		Hosts:    h.Runtime.GetHostsByRole(common.K8s),
		Action:   new(updateHosts),
		Parallel: true,
		Retry:    3,
	}

	h.Tasks = []*modules.Task{
		//makeConfigDir,
		haproxyCfg,
		getMd5Sum,
		haproxyManifestK3s,
		haproxyManifestK8s,
		updateK3sConfig,
		updateKubeletConfig,
		updateKubeProxyConfig,
		updateHostsFile,
	}
}
