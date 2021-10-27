package hooks

import (
	kubekeycontroller "github.com/kubesphere/kubekey/controllers/kubekey"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/module"
)

type UpdateCRStatusHook struct {
	module.PostHook
}

func (u *UpdateCRStatusHook) Try() error {
	m := u.Module.(*module.BaseModule)
	kubeRuntime := m.Runtime.(*common.KubeRuntime)

	if !kubeRuntime.Arg.InCluster {
		return nil
	}

	if err := kubekeycontroller.UpdateClusterConditions(kubeRuntime, m.Desc, u.Result); err != nil {
		return err
	}
	return nil
}
