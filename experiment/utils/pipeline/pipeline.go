package pipeline

import (
	"github.com/kubesphere/kubekey/experiment/utils/logger"
	"github.com/pkg/errors"
)

type Pipeline struct {
	Modules []Module
	Log     *logger.KubeKeyLog
}

func (p *Pipeline) Start() error {
	for i := range p.Modules {
		m := p.Modules[i]
		m.Init(p.Log)
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
