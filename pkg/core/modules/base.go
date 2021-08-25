package modules

import (
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/config"
)

type BaseModule struct {
	Name      string
	Cache     *cache.Cache
	RootCache *cache.Cache
	Runtime   *config.Runtime
}

func (t *BaseModule) Default(runtime *config.Runtime, rootCache *cache.Cache, moduleCache *cache.Cache) {
	if t.Name == "" {
		t.Name = DefaultModuleName
	}

	t.Runtime = runtime
	t.RootCache = rootCache
	t.Cache = moduleCache
}

func (t *BaseModule) Init() {
}

func (t *BaseModule) Is() string {
	return BaseModuleType
}

func (t *BaseModule) Run() error {
	return nil
}
