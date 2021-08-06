package pipeline

import (
	"github.com/kubesphere/kubekey/experiment/utils/cache"
	"github.com/kubesphere/kubekey/experiment/utils/logger"
)

type Module interface {
	Is() string
	Run() error
	Init(log *logger.KubeKeyLog)
}

type TaskModule struct {
	Name  string
	Tasks []Task
	Cache *cache.Cache
	Log   *logger.KubeKeyLog
}

func NewTaskModule(name string, tasks []Task) *TaskModule {
	return &TaskModule{
		Name:  name,
		Tasks: tasks,
	}
}

func (t *TaskModule) Init(log *logger.KubeKeyLog) {
	t.Log = log
	t.Log.SetModule(t.Name)
	t.Cache = cache.NewCache()
}

func (t *TaskModule) Is() string {
	return "task"
}

func (t *TaskModule) Run() error {
	t.Log.Info("Begin Run")
	for i := range t.Tasks {
		task := t.Tasks[i]
		task.Init(t.Log, t.Cache)
		if err := task.Execute(); err != nil {
			return err
		}
	}
	return nil
}
