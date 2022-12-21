/*
 Copyright 2021 The KubeSphere Authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package pipeline

import (
	"fmt"
	"os"
	"sync"

	"github.com/pkg/errors"

	"github.com/kubesphere/kubekey/v2/pkg/core/cache"
	"github.com/kubesphere/kubekey/v2/pkg/core/connector"
	"github.com/kubesphere/kubekey/v2/pkg/core/ending"
	"github.com/kubesphere/kubekey/v2/pkg/core/logger"
	"github.com/kubesphere/kubekey/v2/pkg/core/module"
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
	SpecHosts       int
	PipelineCache   *cache.Cache
	ModuleCachePool sync.Pool
	ModulePostHooks []module.PostHookInterface
}

func (p *Pipeline) Init() error {
	fmt.Print(logo)
	p.PipelineCache = cache.NewCache()
	p.SpecHosts = len(p.Runtime.GetAllHosts())
	//if err := p.Runtime.GenerateWorkDir(); err != nil {
	//	return err
	//}
	//if err := p.Runtime.InitLogger(); err != nil {
	//	return err
	//}
	return nil
}

func (p *Pipeline) Start() error {
	if err := p.Init(); err != nil {
		return errors.Wrapf(err, "Pipeline[%s] execute failed", p.Name)
	}
	for i := range p.Modules {
		m := p.Modules[i]
		if m.IsSkip() {
			continue
		}

		moduleCache := p.newModuleCache()
		m.Default(p.Runtime, p.PipelineCache, moduleCache)
		m.AutoAssert()
		m.Init()
		for j := range p.ModulePostHooks {
			m.AppendPostHook(p.ModulePostHooks[j])
		}

		res := p.RunModule(m)
		err := m.CallPostHook(res)
		if res.IsFailed() {
			return errors.Wrapf(res.CombineResult, "Pipeline[%s] execute failed", p.Name)
		}
		if err != nil {
			return errors.Wrapf(err, "Pipeline[%s] execute failed", p.Name)
		}
		p.releaseModuleCache(moduleCache)
	}
	p.releasePipelineCache()

	// close ssh connect
	for _, host := range p.Runtime.GetAllHosts() {
		p.Runtime.GetConnector().Close(host)
	}

	if p.SpecHosts != len(p.Runtime.GetAllHosts()) {
		return errors.Errorf("Pipeline[%s] execute failed: there are some error in your spec hosts", p.Name)
	}
	logger.Log.Infof("Pipeline[%s] execute successfully", p.Name)
	return nil
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
