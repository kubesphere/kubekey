package modules

import (
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/pkg/errors"
)

type BaseTaskModule struct {
	BaseModule
	Tasks []Task
}

func (b *BaseTaskModule) Init() {
	if b.Name == "" {
		b.Name = DefaultTaskModuleName
	}
}

func (b *BaseTaskModule) Is() string {
	return TaskModuleType
}

func (b *BaseTaskModule) Run() error {
	for i := range b.Tasks {
		task := b.Tasks[i]
		task.Init(b.Name, b.Runtime.(connector.Runtime), b.ModuleCache, b.PipelineCache)
		if res := task.Execute(); res.IsFailed() {
			return errors.Wrapf(res.CombineErr(), "Module[%s] exec failed", b.Name)
		}
	}
	return nil
}
