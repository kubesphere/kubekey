package pipeline

import (
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/experiment/utils/action"
	"github.com/kubesphere/kubekey/experiment/utils/config"
	"github.com/kubesphere/kubekey/experiment/utils/runner"
	"github.com/pkg/errors"
	"sync"
	"time"
)

const (
	DefaultTimeout = 120
	DefaultCon     = 10
)

type Vars map[string]interface{}

type Task struct {
	Name        string
	Manager     *config.Manager
	Hosts       []kubekeyapiv1alpha1.HostCfg
	Action      action.Action
	Env         []map[string]string
	Vars        Vars
	tag         string
	Parallel    bool
	Prepare     Prepare
	IgnoreError bool
	Retry       int
	Delay       time.Duration
	Serial      string
	TaskResult  *TaskResult
}

func (t *Task) Execute() error {
	if t.TaskResult != nil {
		t.TaskResult = NewTaskResult()
	}

	wg := &sync.WaitGroup{}
	// todo: user can customize the pool size
	pool := make(chan struct{}, DefaultCon)
	defer close(pool)

	for i := range t.Hosts {
		mgr := config.GetManager()
		selfMgr := mgr.Copy()
		selfMgr.Logger = selfMgr.Logger.WithField("node", t.Hosts[i].Address)

		_ = t.SetupManager(selfMgr, &t.Hosts[i], i)

		if ok := t.WhenWithRetry(); !ok {
			continue
		}

		if t.TaskResult.IsFailed() {
			return t.TaskResult.CombineErr()
		}

		t.Action.Init(selfMgr)

		pool <- struct{}{}
		wg.Add(1)
		if t.Parallel {
			go t.ExecuteWithRetry(wg, pool, mgr)
		} else {
			t.ExecuteWithRetry(wg, pool, mgr)
		}
	}
	wg.Wait()

	if t.TaskResult.IsFailed() {
		return t.TaskResult.CombineErr()
	}
	return nil
}

func (t *Task) When() (bool, error) {
	if t.Prepare == nil {
		return true, nil
	}
	if ok, err := t.Prepare.PreCheck(t.Manager.Runner.Host); err != nil {
		t.Manager.Logger.Error(err)
		t.TaskResult.AppendErr(err)
		t.TaskResult.ErrResult()
		return false, err
	} else if !ok {
		return false, nil
	} else {
		return true, nil
	}
}

func (t *Task) WhenWithRetry() bool {
	pass := false
	timeout := true
	for i := 0; i < t.Retry; i++ {
		if res, err := t.When(); err != nil {
			time.Sleep(t.Delay)
			continue
		} else {
			timeout = false
			pass = res
			break
		}
	}

	if timeout {
		t.Manager.Logger.Errorf("Execute task pre-check timeout, Timeout=%fs, after %d retries", t.Delay.Seconds(), t.Retry)
	}
	return pass
}

func (t *Task) SetupManager(mgr *config.Manager, host *kubekeyapiv1alpha1.HostCfg, index int) error {
	conn, err := mgr.Connector.Connect(*host)
	if err != nil {
		t.TaskResult.AppendErr(errors.Wrapf(err, "Failed to connect to %s", host.Address))
		t.TaskResult.ErrResult()
		return errors.Wrapf(err, "Failed to connect to %s", host.Address)
	}

	t.Manager.Runner = &runner.Runner{
		Conn:  conn,
		Debug: mgr.Debug,
		Host:  host,
		Index: index,
	}
	return nil
}

func (t *Task) ExecuteWithTimer(wg *sync.WaitGroup, pool chan struct{}, resChan chan string, mgr *config.Manager) Ending {
	// generate a timer
	go func(result chan string, pool chan struct{}) {
		select {
		case <-result:
		case <-time.After(time.Minute * DefaultTimeout):
			mgr.Logger.Fatalf("Execute task timeout, Timeout=%ds", DefaultTimeout)
		}
		<-pool
		wg.Done()
	}(resChan, pool)

	res := t.Action.Execute(t.Vars)
	return res
}

func (t *Task) ExecuteWithRetry(wg *sync.WaitGroup, pool chan struct{}, mgr *config.Manager) {
	resChan := make(chan string)
	defer close(resChan)

	if t.Retry < 1 {
		t.Retry = 1
	}
	var ending Ending
	for i := 0; i < t.Retry; i++ {
		ending = t.ExecuteWithTimer(wg, pool, resChan, mgr)
		if ending.GetErr() != nil {
			mgr.Logger.Error(ending.GetErr())
			time.Sleep(t.Delay)
			continue
		} else {
			break
		}
	}

	if ending != nil {
		t.TaskResult.AppendEnding(ending, mgr.Runner.Host.Name)
		if ending.GetErr() != nil {
			t.TaskResult.AppendErr(ending.GetErr())
		}
	}
	resChan <- "done"
}
