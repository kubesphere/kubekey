package prepare

import (
	"github.com/kubesphere/kubekey/experiment/core/cache"
	"github.com/kubesphere/kubekey/experiment/core/config"
)

type BasePrepare struct {
	Runtime *config.Runtime
	Cache   *cache.Cache
}

func (b *BasePrepare) Init(runtime *config.Runtime, cache *cache.Cache) {
	b.Runtime = runtime
	b.Cache = cache
}

func (b *BasePrepare) PreCheck() (bool, error) {
	return true, nil
}

func (b *BasePrepare) Not(x bool) (bool, error) {
	return !x, nil
}

type PrepareCollection []Prepare

func (p *PrepareCollection) Init(runtime *config.Runtime, cache *cache.Cache) {
	for _, v := range *p {
		v.Init(runtime, cache)
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
