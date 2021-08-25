package prepare

import "github.com/kubesphere/kubekey/pkg/core/config"

// Condition struct is a Default implementation.
type Condition struct {
	BasePrepare
	Cond bool
}

func (c *Condition) PreCheck() (bool, error) {
	if c.Cond {
		return true, nil
	}
	return false, nil
}

type OnlyFirstMaster struct {
	BasePrepare
}

func (o *OnlyFirstMaster) PreCheck(runtime *config.Runtime) (bool, error) {
	if runtime.Runner.Host.IsMaster && runtime.Runner.Host.Name == runtime.MasterNodes[0].Name {
		return true, nil
	}
	return false, nil
}

type OnlyWorker struct {
	BasePrepare
}

func (o *OnlyWorker) PreCheck(runtime *config.Runtime) (bool, error) {
	if runtime.Runner.Host.IsWorker && !runtime.Runner.Host.IsMaster {
		return true, nil
	}
	return false, nil
}

type OnlyK3s struct {
	BasePrepare
}

func (o *OnlyK3s) PreCheck(runtime *config.Runtime) (bool, error) {
	if runtime.Cluster.Kubernetes.Type == "k3s" {
		return true, nil
	}
	return false, nil
}

type OnlyKubernetes struct {
	BasePrepare
}

func (o *OnlyKubernetes) PreCheck(runtime *config.Runtime) (bool, error) {
	if runtime.Cluster.Kubernetes.Type != "k3s" {
		return true, nil
	}
	return false, nil
}
