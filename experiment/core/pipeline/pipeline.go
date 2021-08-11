package pipeline

import (
	"github.com/kubesphere/kubekey/experiment/core/config"
	"github.com/kubesphere/kubekey/experiment/core/logger"
	"github.com/pkg/errors"
)

type Pipeline struct {
	Modules []Module
	Runtime *config.Runtime
}

func (p *Pipeline) Start() error {
	logger.Log.Info("Begin Run")
	for i := range p.Modules {
		m := p.Modules[i]
		m.Default(p.Runtime)
		m.Init()
		switch m.Is() {
		case TaskModule:
			if err := m.Run(); err != nil {
				return err
			}
		case ServerModule:
			go m.Run()
		default:
			return errors.New("invalid module")
		}
	}
	return nil
}
