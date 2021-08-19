package action

import (
	"github.com/kubesphere/kubekey/experiment/core/cache"
	"github.com/kubesphere/kubekey/experiment/core/config"
	"github.com/kubesphere/kubekey/experiment/core/ending"
	"github.com/kubesphere/kubekey/experiment/core/vars"
)

type Action interface {
	Execute(vars vars.Vars) (err error)
	Init(mgr *config.Runtime, cache *cache.Cache, rootCache *cache.Cache)
	WrapResult(err error) *ending.Result
}
