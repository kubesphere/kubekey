package pipeline

type Pipeline struct {
	Modules []Module
}

func (p *Pipeline) Start() error {
	for i := range p.Modules {
		for j := range p.Modules[i].Tasks {
			task := p.Modules[i].Tasks[j]
			if err := task.Execute(); err != nil {
				return err
			}
		}
	}
	return nil
}
