package pipline

type Pipline struct {
	TaskList []Task
}

func (p *Pipline) Start(vars *Vars) error {
	for _, task := range p.TaskList {
		task.Execute(vars)
	}
	return nil
}

