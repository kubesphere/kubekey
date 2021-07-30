package pipeline

import kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"

type Prepare interface {
	PreCheck(host *kubekeyapiv1alpha1.HostCfg) (bool, error)
}

// Condition struct is a Default implementation.
type Condition struct {
	Cond bool
}

func (c *Condition) PreCheck(host *kubekeyapiv1alpha1.HostCfg) (bool, error) {
	if c.Cond {
		return true, nil
	}
	return false, nil
}
