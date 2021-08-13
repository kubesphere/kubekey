package module

import (
	"github.com/kubesphere/kubekey/experiment/core/action"
	"github.com/kubesphere/kubekey/experiment/core/config"
	"github.com/kubesphere/kubekey/experiment/core/pipeline"
	"github.com/kubesphere/kubekey/experiment/core/prepare"
	"github.com/kubesphere/kubekey/experiment/core/vars"
	"os"
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
			TemplateName: "haproxy.cfg",
			Dst:          "/etc/kubekey/haproxy/haproxy.cfg",
			Data: map[string]interface{}{
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
			TemplateName: "haproxy.yaml",
			Dst:          "/etc/kubernetes/manifests",
			Data: map[string]interface{}{
				// todo: implement image module
				"HaproxyImage":                         "haproxy:2.3",
				"LoadbalancerApiserverHealthcheckPort": 8081,
				"Checksum":                             h.Cache.GetMustString("md5"),
			},
		},
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
			TemplateName: "haproxy.yaml",
			Dst:          "/etc/kubernetes/manifests",
			Data: map[string]interface{}{
				// todo: implement image module
				"HaproxyImage":                         "haproxy:2.3",
				"LoadbalancerApiserverHealthcheckPort": 8081,
				"Checksum":                             h.Cache.GetMustString("md5"),
			},
		},
	}

	h.Tasks = []pipeline.Task{
		makeConfigDir,
		haproxyCfg,
		getMd5Sum,
		haproxyManifestK3s,
		haproxyManifestK8s,
	}
}

type haproxyPreparatoryWork struct {
	action.BaseAction
}

func (h *haproxyPreparatoryWork) Execute(vars vars.Vars) error {
	if err := h.Runtime.Runner.MkDir("/etc/kubekey/haproxy"); err != nil {
		return err
	}
	if err := h.Runtime.Runner.Chmod("/etc/kubekey/haproxy", os.FileMode(0777)); err != nil {
		return err
	}
	return nil
}

type getChecksum struct {
	action.BaseAction
}

func (g *getChecksum) Execute(vars vars.Vars) error {
	md5Str, err := g.Runtime.Runner.FileMd5("/etc/kubekey/haproxy/haproxy.cfg")
	if err != nil {
		return err
	}
	g.Cache.Set("md5", md5Str)
	return nil
}

func masterNodeStr(runtime *config.Runtime) []string {
	masterNodes := make([]string, len(runtime.MasterNodes))
	for i, node := range runtime.MasterNodes {
		masterNodes[i] = node.Name + " " + node.InternalAddress + ":" + strconv.Itoa(runtime.Cluster.ControlPlaneEndpoint.Port)
	}
	return masterNodes
}
