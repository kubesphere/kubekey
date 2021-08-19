package modules

import (
	"github.com/kubesphere/kubekey/experiment/core/cache"
	"github.com/kubesphere/kubekey/experiment/core/config"
	"github.com/kubesphere/kubekey/experiment/core/logger"
)

type BaseModule struct {
	Name      string
	Cache     *cache.Cache
	RootCache *cache.Cache
	Runtime   *config.Runtime
}

func (t *BaseModule) Default(runtime *config.Runtime, rootCache *cache.Cache) {
	if t.Name == "" {
		t.Name = DefaultModuleName
	}

	logger.Log.SetModule(t.Name)
	t.Runtime = runtime
	t.RootCache = rootCache
	t.Cache = cache.NewCache()
}

func (t *BaseModule) Init() {
}

func (t *BaseModule) Is() string {
	return BaseModuleType
}

func (t *BaseModule) Run() error {
	logger.Log.Info("Begin Run")
	return nil
}
