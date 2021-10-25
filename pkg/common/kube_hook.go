package common

import (
	kubekeycontroller "github.com/kubesphere/kubekey/controllers/kubekey"
	"github.com/kubesphere/kubekey/pkg/core/hook"
)

type UpdateCRStatusHook struct {
	hook.Base
}

func (u *UpdateCRStatusHook) Try() error {
	kubeRuntime := u.Runtime.(*KubeRuntime)

	if !kubeRuntime.Arg.InCluster {
		return nil
	}

	conf := &KubeConf{
		ClusterHosts: kubeRuntime.ClusterHosts,
		Cluster:      kubeRuntime.Cluster,
		Kubeconfig:   kubeRuntime.Kubeconfig,
		Conditions:   kubeRuntime.Conditions,
		ClientSet:    kubeRuntime.ClientSet,
		Arg:          kubeRuntime.Arg,
	}
	if err := kubekeycontroller.UpdateClusterConditions(conf, u.Desc, u.Result); err != nil {
		return err
	}
	return nil
}
