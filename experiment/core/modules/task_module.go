package modules

import (
	"github.com/kubesphere/kubekey/experiment/core/cache"
	"github.com/kubesphere/kubekey/experiment/core/config"
	"github.com/kubesphere/kubekey/experiment/core/logger"
	"github.com/pkg/errors"
)

type BaseTaskModule struct {
	BaseModule
	Tasks []Task
}

func (t *BaseTaskModule) Default(runtime *config.Runtime, rootCache *cache.Cache) {
	if t.Name == "" {
		t.Name = DefaultTaskModuleName
	}

	t.Runtime = runtime
	t.RootCache = rootCache
	t.Cache = cache.NewCache()
}

func (t *BaseTaskModule) Init() {
}

func (t *BaseTaskModule) Is() string {
	return TaskModuleType
}

func (t *BaseTaskModule) Run() error {
	logger.Log.SetModule(t.Name)
	logger.Log.Info("Begin Run")
	for i := range t.Tasks {
		task := t.Tasks[i]
		task.Init(t.Runtime, t.Cache, t.RootCache)
		if err := task.Execute(); err != nil {
			return errors.Wrapf(err, "Module %s exec failed", t.Name)
		}
	}
	return nil
}
