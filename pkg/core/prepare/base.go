package prepare

import (
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/config"
)

type BasePrepare struct {
	Cache     *cache.Cache
	RootCache *cache.Cache
}

func (b *BasePrepare) Init(cache *cache.Cache, rootCache *cache.Cache) {
	b.Cache = cache
	b.RootCache = rootCache
}

func (b *BasePrepare) PreCheck(runtime *config.Runtime) (bool, error) {
	return true, nil
}

type PrepareCollection []Prepare

func (p *PrepareCollection) Init(cache *cache.Cache, rootCache *cache.Cache) {
	for _, v := range *p {
		v.Init(cache, rootCache)
	}
}

func (p *PrepareCollection) PreCheck(runtime *config.Runtime) (bool, error) {
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
