package action

import (
	"github.com/kubesphere/kubekey/experiment/utils/cache"
	"github.com/kubesphere/kubekey/experiment/utils/config"
	"github.com/kubesphere/kubekey/experiment/utils/ending"
	"github.com/kubesphere/kubekey/experiment/utils/vars"
)

type Action interface {
	Execute(vars vars.Vars) (err error)
	Init(mgr *config.Manager, pool *cache.Cache)
	WrapResult(err error) *ending.Result
}
