package prepare

import (
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/connector"
)

type BasePrepare struct {
	ModuleCache   *cache.Cache
	PipelineCache *cache.Cache
}

func (b *BasePrepare) Init(moduleCache *cache.Cache, pipelineCache *cache.Cache) {
	b.ModuleCache = moduleCache
	b.PipelineCache = pipelineCache
}

func (b *BasePrepare) PreCheck(runtime connector.Runtime) (bool, error) {
	return true, nil
}

func (b *BasePrepare) AutoAssert(runtime connector.Runtime) {
}

type PrepareCollection []Prepare

func (p *PrepareCollection) Init(cache *cache.Cache, rootCache *cache.Cache) {
	for _, v := range *p {
		v.Init(cache, rootCache)
	}
}

func (p *PrepareCollection) PreCheck(runtime connector.Runtime) (bool, error) {
	for _, v := range *p {
		res, err := v.PreCheck(runtime)
		if err != nil {
			return false, err
		}
		if res == false {
			return false, nil
		}
	}
	return true, nil
}

func (p *PrepareCollection) AutoAssert(runtime connector.Runtime) {
	for _, v := range *p {
		v.AutoAssert(runtime)
	}
}
