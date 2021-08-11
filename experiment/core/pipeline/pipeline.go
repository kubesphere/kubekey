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
		m.Init()
		switch m.Is() {
		case "task":
			if err := m.Run(); err != nil {
				return err
			}
		case "webserver":
			go m.Run()
		default:
			return errors.New("invalid module")
		}
	}
	return nil
}
