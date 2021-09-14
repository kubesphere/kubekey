package action

import (
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/connector"
)

type BaseAction struct {
	ModuleCache   *cache.Cache
	PipelineCache *cache.Cache
	Runtime       connector.Runtime
}

func (b *BaseAction) Init(moduleCache *cache.Cache, pipelineCache *cache.Cache, runtime connector.Runtime) {
	b.ModuleCache = moduleCache
	b.PipelineCache = pipelineCache
	b.Runtime = runtime
}

func (b *BaseAction) Execute(runtime connector.Runtime) error {
	return nil
}

func (b *BaseAction) AutoAssert() {

}
