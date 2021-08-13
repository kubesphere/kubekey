package prepare

import (
	"github.com/kubesphere/kubekey/experiment/core/cache"
	"github.com/kubesphere/kubekey/experiment/core/config"
)

type BasePrepare struct {
	runtime *config.Runtime
	Cache   *cache.Cache
}

func (b *BasePrepare) Init(runtime *config.Runtime, cache *cache.Cache) {
	b.runtime = runtime
	b.Cache = cache
}

func (b *BasePrepare) PreCheck() (bool, error) {
	return true, nil
}

func (b *BasePrepare) Not(x bool) (bool, error) {
	return !x, nil
}

type PrepareCollection struct {
	Prepares []Prepare
}

func (p *PrepareCollection) Init(runtime *config.Runtime, cache *cache.Cache) {
	for i := range p.Prepares {
		p.Prepares[i].Init(runtime, cache)
	}
}

func (p *PrepareCollection) PreCheck() (bool, error) {
	for i := range p.Prepares {
		res, err := p.Prepares[i].PreCheck()
		if err != nil {
			return false, err
		}
		if res == false {
			return false, nil
		}
	}
	return true, nil
}
