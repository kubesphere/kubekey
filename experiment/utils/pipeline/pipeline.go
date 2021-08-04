package pipeline

import "github.com/pkg/errors"

type Pipeline struct {
	Modules []Module
}

func (p *Pipeline) Start() error {
	for i := range p.Modules {
		m := p.Modules[i]
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
