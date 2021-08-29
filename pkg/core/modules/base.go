package modules

import (
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/connector"
)

type BaseModule struct {
	Name      string
	Cache     *cache.Cache
	RootCache *cache.Cache
	Runtime   connector.Runtime
}

func (b *BaseModule) Default(runtime connector.Runtime, rootCache *cache.Cache, moduleCache *cache.Cache) {
	if b.Name == "" {
		b.Name = DefaultModuleName
	}

	b.Runtime = runtime
	b.RootCache = rootCache
	b.Cache = moduleCache
}

func (b *BaseModule) Init() {
}

func (b *BaseModule) Is() string {
	return BaseModuleType
}

func (b *BaseModule) Run() error {
	return nil
}

func (b *BaseModule) AutoAssert() {
}
