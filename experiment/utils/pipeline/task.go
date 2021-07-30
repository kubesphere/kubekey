package pipeline

import (
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/experiment/utils/action"
	"github.com/kubesphere/kubekey/experiment/utils/config"
	"github.com/kubesphere/kubekey/pkg/util/runner"
	"github.com/pkg/errors"
	"time"
)

type Vars map[string]interface{}

type Task struct {
	Name        string
	Hosts       []kubekeyapiv1alpha1.HostCfg
	Action      action.Action
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
}

func (t *Task) Execute() error {
	if t.Ending != nil {
		t.Ending = NewResult()
	}
	defer t.Ending.SetEndTime()

	for i := range t.Hosts {
		if ok, _ := t.When(&t.Hosts[i]); !ok {
			continue
		}

		_ = t.SetupManager(&t.Hosts[i], i)

		if t.Ending.GetErr() != nil && !t.IgnoreError {
			return t.Ending.GetErr()
		}

		if t.Parallel {
			go func() {}()
		} else {
			t.Action.Execute(&t.Hosts[i], t.Vars)
		}
	}

	return nil
}

func (t *Task) When(host *kubekeyapiv1alpha1.HostCfg) (bool, error) {
	if t.Prepare == nil {
		return true, nil
	}
	if ok, err := t.Prepare.PreCheck(host); err != nil {
		t.Ending.ErrResult(err)
		return false, err
	} else if !ok {
		return false, nil
	} else {
		return true, nil
	}
}

func (t *Task) SetupManager(host *kubekeyapiv1alpha1.HostCfg, index int) error {
	mgr := config.GetManager()

	conn, err := mgr.Connector.Connect(*host)
	if err != nil {
		t.Ending.ErrResult(err)
		return errors.Wrapf(err, "Failed to connect to %s", host.Address)
	}

	mgr.Runner = &runner.Runner{
		Conn:  conn,
		Debug: mgr.Debug,
		Host:  host,
		Index: index,
	}
	return nil
}
