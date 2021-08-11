package pipeline

import (
	"github.com/kubesphere/kubekey/experiment/core/cache"
	"github.com/kubesphere/kubekey/experiment/core/config"
	"github.com/kubesphere/kubekey/experiment/core/logger"
)

type Module interface {
	Is() string
	Run() error
	Init()
}

type TaskModule struct {
	Name    string
	Tasks   []Task
	Cache   *cache.Cache
	Runtime *config.Runtime
}

func NewTaskModule(name string, runtime *config.Runtime, tasks []Task) *TaskModule {
	return &TaskModule{
		Name:    name,
		Runtime: runtime,
		Tasks:   tasks,
	}
}

func (t *TaskModule) Init() {
	logger.Log.SetModule(t.Name)
	t.Cache = cache.NewCache()
}

func (t *TaskModule) Is() string {
	return "task"
}

func (t *TaskModule) Run() error {
	logger.Log.Info("Begin Run")
	for i := range t.Tasks {
		task := t.Tasks[i]
		task.Init(t.Runtime, t.Cache)
		if err := task.Execute(); err != nil {
			return err
		}
	}
	return nil
}
