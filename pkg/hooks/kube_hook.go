package hooks

import (
	kubekeycontroller "github.com/kubesphere/kubekey/controllers/kubekey"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/hook"
)

type UpdateCRStatusHook struct {
	hook.Base
}

func (u *UpdateCRStatusHook) Try() error {
	kubeRuntime := u.Runtime.(*common.KubeRuntime)

	if !kubeRuntime.Arg.InCluster {
		return nil
	}

	conf := &common.KubeConf{
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
