package modules

import (
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/logger"
)

type BaseModule struct {
	Name      string
	Desc      string
	Skip      bool
	Cache     *cache.Cache
	RootCache *cache.Cache
	Runtime   connector.Runtime
}

func (b *BaseModule) IsSkip() bool {
	return b.Skip
}

func (b *BaseModule) Default(runtime connector.Runtime, rootCache *cache.Cache, moduleCache *cache.Cache) {
	b.Runtime = runtime
	b.RootCache = rootCache
	b.Cache = moduleCache
}

func (b *BaseModule) Init() {
	if b.Name == "" {
		b.Name = DefaultModuleName
	}
}

func (b *BaseModule) Is() string {
	return BaseModuleType
}

func (b *BaseModule) Run() error {
	return nil
}

func (b *BaseModule) Slogan() {
	if b.Desc != "" {
		logger.Log.Infof("[%s] %s", b.Name, b.Desc)
	}
}

func (b *BaseModule) AutoAssert() {
}
