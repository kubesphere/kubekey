package pipline

import (
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"time"
)

type Vars map[string]interface{}

type Task struct {
	Name        string
	Hosts       []kubekeyapiv1alpha1.HostCfg
	Action      Action
	Env         []map[string]string
	Vars        Vars
	tag         string
	Parallel    bool
	Prepare     Prepare
	IgnoreError bool
	Retry       int
	Delay       time.Time
	Serial      string
	Ending      Ending

	CurrentNode kubekeyapiv1alpha1.HostCfg
}

func (t *Task) Execute(vars *Vars) error {
	if t.Ending.GetErr() != nil {
		return t.Ending.GetErr()
	}
	for i := range t.Hosts {
		if t.Parallel {
			go func() {}()
		} else {
			t.Action.Execute(t.Hosts[i], vars)
		}
	}

	return nil
}

func (t *Task) When() (bool, error) {
	if t.Prepare == nil {
		return true, nil
	}
	if ok, err := t.Prepare.PreCheck(); err != nil {
		t.Ending = NewResultWithErr(err)
		return false, err
	} else if !ok {
		return false, nil
	} else {
		return true, nil
	}
}
