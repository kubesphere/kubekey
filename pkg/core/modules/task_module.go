package modules

import (
	cache2 "github.com/kubesphere/kubekey/pkg/core/cache"
	config2 "github.com/kubesphere/kubekey/pkg/core/config"
	logger2 "github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/pkg/errors"
)

type BaseTaskModule struct {
	BaseModule
	Tasks []Task
}

func (t *BaseTaskModule) Default(runtime *config2.Runtime, rootCache *cache2.Cache) {
	if t.Name == "" {
		t.Name = DefaultTaskModuleName
	}

	t.Runtime = runtime
	t.RootCache = rootCache
	t.Cache = cache2.NewCache()
}

func (t *BaseTaskModule) Init() {
}

func (t *BaseTaskModule) Is() string {
	return TaskModuleType
}

func (t *BaseTaskModule) Run() error {
	logger2.Log.SetModule(t.Name)
	logger2.Log.Info("Begin Run")
	for i := range t.Tasks {
		task := t.Tasks[i]
		task.Init(t.Runtime, t.Cache, t.RootCache)
		if err := task.Execute(); err != nil {
			return errors.Wrapf(err, "Module %s exec failed", t.Name)
		}
	}
	return nil
}
