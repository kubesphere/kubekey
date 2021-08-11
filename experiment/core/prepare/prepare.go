package prepare

// Condition struct is a Default implementation.
type Condition struct {
	BasePrepare
	Cond bool
}

func (c *Condition) PreCheck() (bool, error) {
	if c.Cond {
		return true, nil
	}
	return false, nil
}

type OnlyFirstMaster struct {
	BasePrepare
}

func (o *OnlyFirstMaster) PreCheck() (bool, error) {
	if o.mgr.Runner.Host.IsMaster && o.mgr.Runner.Host.Name == o.mgr.MasterNodes[0].Name {
		return true, nil
	}
	return false, nil
}
