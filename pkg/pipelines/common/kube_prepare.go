package common

import (
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/prepare"
)

type KubePrepare struct {
	prepare.BasePrepare
	KubeConf *KubeRuntime
}

func (k *KubePrepare) AutoAssert() {
	conf := k.RuntimeConf.(*KubeRuntime)
	k.KubeConf = conf
}

type OnlyFirstMaster struct {
	KubePrepare
}

func (o *OnlyFirstMaster) PreCheck(runtime connector.Runtime) (bool, error) {
	if runtime.RemoteHost().IsRole(Master) &&
		runtime.RemoteHost().GetName() == runtime.GetHostsByRole(Master)[0].GetName() {
		return true, nil
	}
	return false, nil
}

type OnlyWorker struct {
	KubePrepare
}

func (o *OnlyWorker) PreCheck(runtime connector.Runtime) (bool, error) {
	if runtime.RemoteHost().IsRole(Worker) && !runtime.RemoteHost().IsRole(Master) {
		return true, nil
	}
	return false, nil
}

type OnlyK3s struct {
	KubePrepare
}

func (o *OnlyK3s) PreCheck(runtime connector.Runtime) (bool, error) {
	if o.KubeConf.Cluster.Kubernetes.Type == "k3s" {
		return true, nil
	}
	return false, nil
}

type OnlyKubernetes struct {
	KubePrepare
}

func (o *OnlyKubernetes) PreCheck(runtime connector.Runtime) (bool, error) {
	if o.KubeConf.Cluster.Kubernetes.Type != "k3s" {
		return true, nil
	}
	return false, nil
}
