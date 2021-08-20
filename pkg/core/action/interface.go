package action

import (
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/config"
	"github.com/kubesphere/kubekey/pkg/core/ending"
	"github.com/kubesphere/kubekey/pkg/core/vars"
)

type Action interface {
	Execute(vars vars.Vars) (err error)
	Init(mgr *config.Runtime, cache *cache.Cache, rootCache *cache.Cache)
	WrapResult(err error) *ending.Result
}
