package pipeline

import "github.com/kubesphere/kubekey/experiment/utils/cache"

type Module interface {
	Is() string
	Run() error
}

type TaskModule struct {
	Tasks []Task
	Pool  *cache.Cache
}

func NewTaskModule(tasks []Task) *TaskModule {
	return &TaskModule{
		Tasks: tasks,
	}
}

func (t *TaskModule) Is() string {
	return "task"
}

func (t *TaskModule) Run() error {
	modulePool := t.InitPool()
	for i := range t.Tasks {
		task := t.Tasks[i]
		task.Pool = modulePool
		if err := task.Execute(); err != nil {
			return err
		}
	}
	return nil
}

func (t *TaskModule) InitPool() *cache.Cache {
	return cache.NewPool()
}
