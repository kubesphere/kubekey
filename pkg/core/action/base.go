package action

import (
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/connector"
)

type BaseAction struct {
	Cache       *cache.Cache
	RootCache   *cache.Cache
	RuntimeConf connector.Runtime
}

func (b *BaseAction) Init(cache *cache.Cache, rootCache *cache.Cache, runtimeConf connector.Runtime) {
	b.Cache = cache
	b.RootCache = rootCache
	b.RuntimeConf = runtimeConf
}

func (b *BaseAction) Execute(runtime connector.Runtime) error {
	return nil
}

func (b *BaseAction) AutoAssert() {

}
