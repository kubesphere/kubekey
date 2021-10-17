package modules

import (
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/logger"
)

type BaseModule struct {
	Name          string
	Desc          string
	Skip          bool
	ModuleCache   *cache.Cache
	PipelineCache *cache.Cache
	Runtime       connector.ModuleRuntime
}

func (b *BaseModule) IsSkip() bool {
	return b.Skip
}

func (b *BaseModule) Default(runtime connector.Runtime, pipelineCache *cache.Cache, moduleCache *cache.Cache) {
	b.Runtime = runtime
	b.PipelineCache = pipelineCache
	b.ModuleCache = moduleCache
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

func (b *BaseModule) Until() (*bool, error) {
	return nil, nil
}

func (b *BaseModule) Slogan() {
	if b.Desc != "" {
		logger.Log.Infof("[%s] %s", b.Name, b.Desc)
	}
}

func (b *BaseModule) AutoAssert() {
}
