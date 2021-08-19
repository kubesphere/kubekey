package pipeline

import (
	"github.com/kubesphere/kubekey/experiment/core/cache"
	"github.com/kubesphere/kubekey/experiment/core/config"
	"github.com/kubesphere/kubekey/experiment/core/logger"
	"github.com/kubesphere/kubekey/experiment/core/modules"
	"github.com/pkg/errors"
)

type Pipeline struct {
	Name          string
	Modules       []modules.Module
	Runtime       *config.Runtime
	PipelineCache *cache.Cache
}

func (p *Pipeline) Init() {
	p.PipelineCache = cache.NewCache()
}

func (p *Pipeline) Start() error {
	logger.Log.SetPipeline(p.Name)
	logger.Log.Info("Begin Run")
	p.Init()
	for i := range p.Modules {
		m := p.Modules[i]
		m.Default(p.Runtime, p.PipelineCache)
		m.Init()
		switch m.Is() {
		case modules.TaskModuleType:
			if err := m.Run(); err != nil {
				return errors.Wrapf(err, "Pipeline %s exec failed", p.Name)
			}
		case modules.ServerModuleType:
			go m.Run()
		default:
			if err := m.Run(); err != nil {
				return errors.Wrapf(err, "Pipeline %s exec failed", p.Name)
			}
		}
		logger.Log.Info("Success")
		logger.Log.Flush()
	}
	return nil
}
