package module

import (
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/ending"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/kubesphere/kubekey/pkg/core/task"
	"github.com/pkg/errors"
)

type BaseTaskModule struct {
	BaseModule
	Tasks []task.Interface
}

func (b *BaseTaskModule) Init() {
	if b.Name == "" {
		b.Name = DefaultTaskModuleName
	}
}

func (b *BaseTaskModule) Is() string {
	return TaskModuleType
}

func (b *BaseTaskModule) Run(result *ending.ModuleResult) {
	for i := range b.Tasks {
		t := b.Tasks[i]
		t.Init(b.Runtime.(connector.Runtime), b.ModuleCache, b.PipelineCache)

		logger.Log.Infof("[%s] %s", b.Name, t.GetDesc())

		res := t.Execute()
		for j := range res.ActionResults {
			ac := res.ActionResults[j]
			logger.Log.Infof("%s: [%s]", ac.Status.String(), ac.Host.GetName())
			result.AppendHostResult(ac)
		}
		if res.IsFailed() {
			result.ErrResult(errors.Wrapf(res.CombineErr(), "Module[%s] exec failed", b.Name))
			return
		}
	}
	result.NormalResult()
}
