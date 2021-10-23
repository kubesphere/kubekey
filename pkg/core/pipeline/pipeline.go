package pipeline

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/ending"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/kubesphere/kubekey/pkg/core/module"
	"github.com/pkg/errors"
	"os"
	"sync"
)

var logo = `

 _   __      _          _   __           
| | / /     | |        | | / /           
| |/ / _   _| |__   ___| |/ /  ___ _   _ 
|    \| | | | '_ \ / _ \    \ / _ \ | | |
| |\  \ |_| | |_) |  __/ |\  \  __/ |_| |
\_| \_/\__,_|_.__/ \___\_| \_/\___|\__, |
                                    __/ |
                                   |___/

`

type Pipeline struct {
	Name            string
	Modules         []module.Module
	Runtime         connector.Runtime
	PipelineCache   *cache.Cache
	ModuleCachePool sync.Pool
}

func (p *Pipeline) Init() error {
	fmt.Print(logo)
	p.PipelineCache = cache.NewCache()
	if err := p.Runtime.GenerateWorkDir(); err != nil {
		return err
	}
	if err := p.Runtime.InitLogger(); err != nil {
		return err
	}
	return nil
}

func (p *Pipeline) Start() error {
	if err := p.Init(); err != nil {
		return errors.Wrapf(err, "Pipeline[%s] exec failed", p.Name)
	}
	for i := range p.Modules {
		m := p.Modules[i]
		if m.IsSkip() {
			continue
		}

		p.InitModule(m)
		res := p.RunModule(m)
		err := m.CallPostHook(res)
		if res.IsFailed() {
			return errors.Wrapf(res.CombineResult, "Pipeline[%s] exec failed", p.Name)
		}
		if err != nil {
			return errors.Wrapf(err, "Pipeline[%s] exec failed", p.Name)
		}
	}
	p.releasePipelineCache()
	logger.Log.Infof("Pipeline[%s] execute successful", p.Name)
	return nil
}

func (p *Pipeline) InitModule(m module.Module) {
	moduleCache := p.newModuleCache()
	defer p.releaseModuleCache(moduleCache)
	m.Default(p.Runtime, p.PipelineCache, moduleCache)
	m.AutoAssert()
	m.Init()
	m.RegisterHooks()
}

func (p *Pipeline) RunModule(m module.Module) *ending.ModuleResult {
	m.Slogan()

	result := ending.NewModuleResult()
	for {
		switch m.Is() {
		case module.TaskModuleType:
			m.Run(result)
			if result.IsFailed() {
				return result
			}

		case module.GoroutineModuleType:
			go func() {
				m.Run(result)
				if result.IsFailed() {
					os.Exit(1)
				}
			}()
		default:
			m.Run(result)
			if result.IsFailed() {
				return result
			}
		}

		stop, err := m.Until()
		if err != nil {
			result.LocalErrResult(err)
			return result
		}
		if stop == nil || *stop == true {
			break
		}
	}
	return result
}

func (p *Pipeline) newModuleCache() *cache.Cache {
	moduleCache, ok := p.ModuleCachePool.Get().(*cache.Cache)
	if ok {
		return moduleCache
	}
	return cache.NewCache()
}

func (p *Pipeline) releasePipelineCache() {
	p.PipelineCache.Clean()
}

func (p *Pipeline) releaseModuleCache(c *cache.Cache) {
	c.Clean()
	p.ModuleCachePool.Put(c)
}
