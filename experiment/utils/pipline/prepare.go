package pipline

type Prepare interface {
	PreCheck() (bool, error)
}

// Condition struct is a Default implementation.
type Condition struct {
	Cond bool
}

func (c *Condition) PreCheck() (bool, error) {
	if c.Cond {
		return true, nil
	}
	return false, nil
}
