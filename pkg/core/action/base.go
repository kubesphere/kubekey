package action

import (
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/connector"
)

type BaseAction struct {
	ModuleCache   *cache.Cache
	PipelineCache *cache.Cache
}

func (b *BaseAction) Init(moduleCache *cache.Cache, pipelineCache *cache.Cache) {
	b.ModuleCache = moduleCache
	b.PipelineCache = pipelineCache
}

func (b *BaseAction) Execute(runtime connector.Runtime) error {
	return nil
}

func (b *BaseAction) AutoAssert(runtime connector.Runtime) {

}
