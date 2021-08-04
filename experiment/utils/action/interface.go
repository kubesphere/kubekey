package action

import (
	"github.com/kubesphere/kubekey/experiment/utils/config"
	"github.com/kubesphere/kubekey/experiment/utils/ending"
	"github.com/kubesphere/kubekey/experiment/utils/vars"
)

type Action interface {
	Execute(vars vars.Vars) (result *ending.Result)
	Init(mgr *config.Manager)
}
