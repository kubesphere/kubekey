package pipeline

import (
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/modules"
	"github.com/pkg/errors"
	"sync"
)

type Pipeline struct {
	Name            string
	Modules         []modules.Module
	Runtime         connector.Runtime
	PipelineCache   *cache.Cache
	ModuleCachePool sync.Pool
}

func (p *Pipeline) Init() {
	p.PipelineCache = cache.NewCache()
}

func (p *Pipeline) Start() error {
	p.Init()
	for i := range p.Modules {
		m := p.Modules[i]
		if err := p.RunModule(m); err != nil {
			return errors.Wrapf(err, "Pipeline[%s] exec failed", p.Name)
		}
	}
	return nil
}

func (p *Pipeline) RunModule(m modules.Module) error {
	moduleCache := p.newModuleCache()
	defer p.releaseModuleCache(moduleCache)
	m.Default(p.Runtime, p.PipelineCache, moduleCache)
	m.Init()
	m.AutoAssert()
	switch m.Is() {
	case modules.TaskModuleType:
		if err := m.Run(); err != nil {
			return err
		}
	case modules.ServerModuleType:
		go m.Run()
	default:
		if err := m.Run(); err != nil {
			return err
		}
	}
	return nil
}

func (p *Pipeline) newModuleCache() *cache.Cache {
	moduleCache, ok := p.ModuleCachePool.Get().(*cache.Cache)
	if ok {
		return moduleCache
	}
	return cache.NewCache()
}

func (p *Pipeline) releaseModuleCache(c *cache.Cache) {
	c.Clean()
	p.ModuleCachePool.Put(c)
}
