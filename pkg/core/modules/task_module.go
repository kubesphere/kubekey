package modules

import (
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/config"
	"github.com/pkg/errors"
)

type BaseTaskModule struct {
	BaseModule
	Tasks []Task
}

func (b *BaseTaskModule) Default(runtime *config.Runtime, rootCache *cache.Cache, moduleCache *cache.Cache) {
	if b.Name == "" {
		b.Name = DefaultTaskModuleName
	}

	b.Runtime = runtime
	b.RootCache = rootCache
	b.Cache = moduleCache
}

func (b *BaseTaskModule) Init() {
}

func (b *BaseTaskModule) Is() string {
	return TaskModuleType
}

func (b *BaseTaskModule) Run() error {
	for i := range b.Tasks {
		task := b.Tasks[i]
		task.Init(b.Name, b.Runtime, b.Cache, b.RootCache)
		if err := task.Execute(); err != nil {
			return errors.Wrapf(err, "Module[%s] exec failed", b.Name)
		}
	}
	return nil
}
