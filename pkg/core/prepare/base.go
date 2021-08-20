package prepare

import (
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/config"
)

type BasePrepare struct {
	Runtime   *config.Runtime
	Cache     *cache.Cache
	RootCache *cache.Cache
}

func (b *BasePrepare) Init(runtime *config.Runtime, cache *cache.Cache, rootCache *cache.Cache) {
	b.Runtime = runtime
	b.Cache = cache
	b.RootCache = rootCache
}

func (b *BasePrepare) PreCheck() (bool, error) {
	return true, nil
}

type PrepareCollection []Prepare

func (p *PrepareCollection) Init(runtime *config.Runtime, cache *cache.Cache, rootCache *cache.Cache) {
	for _, v := range *p {
		v.Init(runtime, cache, rootCache)
	}
}

func (p *PrepareCollection) PreCheck() (bool, error) {
	for _, v := range *p {
		res, err := v.PreCheck()
		if err != nil {
			return false, err
		}
		if res == false {
			return false, nil
		}
	}
	return true, nil
}
