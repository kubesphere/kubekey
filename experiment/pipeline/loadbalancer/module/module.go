package module

import (
	"github.com/kubesphere/kubekey/experiment/core/action"
	"github.com/kubesphere/kubekey/experiment/core/config"
	"github.com/kubesphere/kubekey/experiment/core/pipeline"
	"github.com/kubesphere/kubekey/experiment/core/prepare"
	"github.com/kubesphere/kubekey/experiment/core/util"
	"github.com/kubesphere/kubekey/experiment/pipeline/loadbalancer/module/templates"
	"strconv"
)

type HaproxyModule struct {
	pipeline.BaseTaskModule
}

func (h *HaproxyModule) Init() {
	h.Name = "InternalLoadbalancer"

	makeConfigDir := pipeline.Task{
		Name:     "MakeHaproxyConfigDir",
		Hosts:    h.Runtime.WorkerNodes,
		Prepare:  new(prepare.OnlyWorker),
		Action:   new(haproxyPreparatoryWork),
		Parallel: true,
	}

	haproxyCfg := pipeline.Task{
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
	getMd5Sum := pipeline.Task{
		Name:     "GetChecksumFromConfig",
		Hosts:    h.Runtime.WorkerNodes,
		Prepare:  new(prepare.OnlyWorker),
		Action:   new(getChecksum),
		Parallel: true,
	}

	haproxyManifestK3s := pipeline.Task{
		Name:  "GenerateHaproxyManifestK3s",
		Hosts: h.Runtime.WorkerNodes,
		Prepare: &prepare.PrepareCollection{
			Prepares: []prepare.Prepare{
				new(prepare.OnlyWorker),
				new(prepare.OnlyK3s),
			}},
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

	haproxyManifestK8s := pipeline.Task{
		Name:  "GenerateHaproxyManifest",
		Hosts: h.Runtime.WorkerNodes,
		Prepare: &prepare.PrepareCollection{
			Prepares: []prepare.Prepare{
				new(prepare.OnlyWorker),
				new(prepare.OnlyKubernetes),
			}},
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

	// UpdateKubeletConfig Update server field in kubelet.conf
	// When create a HA cluster by internal LB, we will set the server filed to 127.0.0.1:6443 (default) which in kubelet.conf.
	// Because of that, the control plone node's kubelet connect the local api-server.
	// And the work node's kubelet connect 127.0.0.1:6443 (default) that is proxy by the node's local nginx.
	updateKubeletConfig := pipeline.Task{
		Name:     "UpdateKubeletConfig",
		Hosts:    h.Runtime.K8sNodes,
		Prepare:  new(updateKubeletPrapre),
		Action:   new(updateKubelet),
		Parallel: true,
	}

	h.Tasks = []pipeline.Task{
		makeConfigDir,
		haproxyCfg,
		getMd5Sum,
		haproxyManifestK3s,
		haproxyManifestK8s,
		updateKubeletConfig,
	}
}

func masterNodeStr(runtime *config.Runtime) []string {
	masterNodes := make([]string, len(runtime.MasterNodes))
	for i, node := range runtime.MasterNodes {
		masterNodes[i] = node.Name + " " + node.InternalAddress + ":" + strconv.Itoa(runtime.Cluster.ControlPlaneEndpoint.Port)
	}
	return masterNodes
}
