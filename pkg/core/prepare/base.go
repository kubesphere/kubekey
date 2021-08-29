package prepare

import (
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/connector"
)

type BasePrepare struct {
	Cache       *cache.Cache
	RootCache   *cache.Cache
	RuntimeConf connector.Runtime
}

func (b *BasePrepare) Init(cache *cache.Cache, rootCache *cache.Cache, runtimeConf connector.Runtime) {
	b.Cache = cache
	b.RootCache = rootCache
	b.RuntimeConf = runtimeConf
}

func (b *BasePrepare) PreCheck(runtime connector.Runtime) (bool, error) {
	return true, nil
}

func (b *BasePrepare) AutoAssert() {
}

type PrepareCollection []Prepare

func (p *PrepareCollection) Init(cache *cache.Cache, rootCache *cache.Cache, runtimeConf connector.Runtime) {
	for _, v := range *p {
		v.Init(cache, rootCache, runtimeConf)
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

func (p *PrepareCollection) AutoAssert() {
	for _, v := range *p {
		v.AutoAssert()
	}
}
