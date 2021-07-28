package pipline

type Pipeline struct {
	Modules []Module
}

func (p *Pipeline) Start(vars *Vars) error {
	for i := range p.Modules {
		for j := range p.Modules[i].Tasks {
			task := p.Modules[i].Tasks[j]
			if ok, _ := task.When(); ok {
				if err := task.Execute(vars); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
