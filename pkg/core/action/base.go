package action

import (
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/config"
	"github.com/kubesphere/kubekey/pkg/core/vars"
)

type BaseAction struct {
	Cache     *cache.Cache
	RootCache *cache.Cache
}

func (b *BaseAction) Init(cache *cache.Cache, rootCache *cache.Cache) {
	b.Cache = cache
	b.RootCache = rootCache
}

func (b *BaseAction) Execute(runtime *config.Runtime, vars vars.Vars) error {
	return nil
}
