package loadbalancer

import (
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/config"
	"github.com/kubesphere/kubekey/pkg/core/modules"
	"github.com/kubesphere/kubekey/pkg/core/prepare"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/pipelines/loadbalancer/templates"
	"strconv"
)

type HaproxyModule struct {
	modules.BaseTaskModule
}

func (h *HaproxyModule) Init() {
	h.Name = "InternalLoadbalancer"

	makeConfigDir := modules.Task{
		Name:     "MakeHaproxyConfigDir",
		Hosts:    h.Runtime.WorkerNodes,
		Prepare:  new(prepare.OnlyWorker),
		Action:   new(haproxyPreparatoryWork),
		Parallel: true,
	}

	haproxyCfg := modules.Task{
		Name:    "GenerateHaproxyConfig",
		Hosts:   h.Runtime.WorkerNodes,
		Prepare: new(prepare.OnlyWorker),
		Action: &action.Template{
			Template: templates.HaproxyConfig,
			Dst:      "/etc/kubekey/haproxy/haproxy.cfg",
			Data: util.Data{
				"MasterNodes":                          masterNodeStr(h.Runtime),
				"LoadbalancerApiserverPort":            h.Runtime.Cluster.ControlPlaneEndpoint.Port,
				"LoadbalancerApiserverHealthcheckPort": 8081,
				"KubernetesType":                       h.Runtime.Cluster.Kubernetes.Type,
			},
		},
		Parallel: true,
	}

	// Calculation config md5 as the checksum.
	// It will make load balancer reload when config changes.
	getMd5Sum := modules.Task{
		Name:     "GetChecksumFromConfig",
		Hosts:    h.Runtime.WorkerNodes,
		Prepare:  new(prepare.OnlyWorker),
		Action:   new(getChecksum),
		Parallel: true,
	}

	haproxyManifestK3s := modules.Task{
		Name:  "GenerateHaproxyManifestK3s",
		Hosts: h.Runtime.WorkerNodes,
		Prepare: &prepare.PrepareCollection{
			new(prepare.OnlyWorker),
			new(prepare.OnlyK3s),
		},
		Action: &action.Template{
			Template: templates.HaproxyManifest,
			Dst:      "/etc/kubernetes/manifests",
			Data: util.Data{
				// todo: implement image module
				"HaproxyImage":                         "haproxy:2.3",
				"LoadbalancerApiserverHealthcheckPort": 8081,
				"Checksum":                             h.Cache.GetMustString("md5"),
			},
		},
		Parallel: true,
	}

	haproxyManifestK8s := modules.Task{
		Name:  "GenerateHaproxyManifest",
		Hosts: h.Runtime.WorkerNodes,
		Prepare: &prepare.PrepareCollection{
			new(prepare.OnlyWorker),
			new(prepare.OnlyKubernetes),
		},
		Action: &action.Template{
			Template: templates.HaproxyManifest,
			Dst:      "/etc/kubernetes/manifests",
			Data: util.Data{
				// todo: implement image module
				"HaproxyImage":                         "haproxy:2.3",
				"LoadbalancerApiserverHealthcheckPort": 8081,
				"Checksum":                             h.Cache.GetMustString("md5"),
			},
		},
		Parallel: true,
	}

	updateK3sConfig := modules.Task{
		Name:  "UpdateK3sConfig",
		Hosts: h.Runtime.WorkerNodes,
		Prepare: &prepare.PrepareCollection{
			new(prepare.OnlyK3s),
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
	updateKubeletConfig := modules.Task{
		Name:  "UpdateKubeletConfig",
		Hosts: h.Runtime.K8sNodes,
		Prepare: &prepare.PrepareCollection{
			new(prepare.OnlyKubernetes),
			new(updateKubeletPrepare),
		},
		Action:   new(updateKubelet),
		Parallel: true,
		Retry:    3,
	}

	// updateKubeproxyConfig is used to update kube-proxy configmap and restart tge kube-proxy pod.
	updateKubeproxyConfig := modules.Task{
		Name:  "UpdateKubeproxyConfig",
		Hosts: []*kubekeyapiv1alpha1.HostCfg{h.Runtime.MasterNodes[0]},
		Prepare: &prepare.PrepareCollection{
			new(prepare.OnlyKubernetes),
			new(prepare.OnlyFirstMaster),
			new(updateKubeproxyPrapre),
		},
		Action:   new(updateKubeproxy),
		Parallel: true,
		Retry:    3,
	}

	// UpdateHostsFile is used to update the '/etc/hosts'. Make the 'lb.kubesphere.local' address to set as 127.0.0.1.
	// All of the 'admin.conf' and '/.kube/config' will connect to 127.0.0.1:6443.
	updateHostsFile := modules.Task{
		Name:     "UpdateHostsFile",
		Hosts:    h.Runtime.K8sNodes,
		Prepare:  nil,
		Action:   new(updateHosts),
		Parallel: true,
		Retry:    3,
	}

	h.Tasks = []modules.Task{
		makeConfigDir,
		haproxyCfg,
		getMd5Sum,
		haproxyManifestK3s,
		haproxyManifestK8s,
		updateK3sConfig,
		updateKubeletConfig,
		updateKubeproxyConfig,
		updateHostsFile,
	}
}

func masterNodeStr(runtime *config.Runtime) []string {
	masterNodes := make([]string, len(runtime.MasterNodes))
	for i, node := range runtime.MasterNodes {
		masterNodes[i] = node.Name + " " + node.InternalAddress + ":" + strconv.Itoa(runtime.Cluster.ControlPlaneEndpoint.Port)
	}
	return masterNodes
}
