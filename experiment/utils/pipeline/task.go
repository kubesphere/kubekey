package pipeline

import (
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/experiment/utils/action"
	"github.com/kubesphere/kubekey/experiment/utils/cache"
	"github.com/kubesphere/kubekey/experiment/utils/config"
	"github.com/kubesphere/kubekey/experiment/utils/ending"
	"github.com/kubesphere/kubekey/experiment/utils/logger"
	"github.com/kubesphere/kubekey/experiment/utils/prepare"
	"github.com/kubesphere/kubekey/experiment/utils/runner"
	"github.com/kubesphere/kubekey/experiment/utils/vars"
	"github.com/pkg/errors"
	"sync"
	"time"
)

const (
	DefaultTimeout = 120
	DefaultCon     = 10
)

type Task struct {
	Name        string
	Manager     *config.Manager
	Hosts       []kubekeyapiv1alpha1.HostCfg
	Action      action.Action
	Cache       *cache.Cache
	Log         *logger.KubeKeyLog
	Vars        vars.Vars
	tag         string
	Parallel    bool
	Prepare     prepare.Prepare
	IgnoreError bool
	Retry       int
	Delay       time.Duration
	Serial      string
	TaskResult  *ending.TaskResult
}

func (t *Task) Init(log *logger.KubeKeyLog, cache *cache.Cache) {
	t.Log = log
	t.Log.SetTask(t.Name)
	t.Cache = cache
}

func (t *Task) Execute() error {
	t.Log.Info("Begin Run")
	if t.TaskResult != nil {
		t.TaskResult = ending.NewTaskResult()
	}

	wg := &sync.WaitGroup{}
	// todo: user can customize the pool size
	routinePool := make(chan struct{}, DefaultCon)
	defer close(routinePool)

	for i := range t.Hosts {
		mgr := config.GetManager()
		selfMgr := mgr.Copy()
		t.Log.SetNode(t.Hosts[i].Name)
		selfMgr.Logger = t.Log

		_ = t.SetupManager(selfMgr, &t.Hosts[i], i)

		t.Prepare.Init(mgr, t.Cache)
		if ok := t.WhenWithRetry(); !ok {
			continue
		}

		if t.TaskResult.IsFailed() {
			return t.TaskResult.CombineErr()
		}

		t.Action.Init(selfMgr, t.Cache)
		routinePool <- struct{}{}
		wg.Add(1)
		if t.Parallel {
			go t.ExecuteWithRetry(wg, routinePool, mgr)
		} else {
			t.ExecuteWithRetry(wg, routinePool, mgr)
		}
	}
	wg.Wait()

	if t.TaskResult.IsFailed() {
		return t.TaskResult.CombineErr()
	}
	t.TaskResult.NormalResult()
	return nil
}

func (t *Task) SetupManager(mgr *config.Manager, host *kubekeyapiv1alpha1.HostCfg, index int) error {
	conn, err := mgr.Connector.Connect(*host)
	if err != nil {
		t.TaskResult.AppendErr(errors.Wrapf(err, "Failed to connect to %s", host.Address))
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

func (t *Task) When() (bool, error) {
	if t.Prepare == nil {
		return true, nil
	}
	if ok, err := t.Prepare.PreCheck(); err != nil {
		t.Manager.Logger.Error(err)
		t.TaskResult.AppendErr(err)
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

func (t *Task) ExecuteWithTimer(wg *sync.WaitGroup, pool chan struct{}, resChan chan string, mgr *config.Manager) ending.Ending {
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

	res := t.Action.WrapResult(t.Action.Execute(t.Vars))
	return res
}

func (t *Task) ExecuteWithRetry(wg *sync.WaitGroup, pool chan struct{}, mgr *config.Manager) {
	resChan := make(chan string)
	defer close(resChan)

	if t.Retry < 1 {
		t.Retry = 1
	}
	var end ending.Ending
	for i := 0; i < t.Retry; i++ {
		end = t.ExecuteWithTimer(wg, pool, resChan, mgr)
		if end.GetErr() != nil {
			mgr.Logger.Error(end.GetErr())
			time.Sleep(t.Delay)
			continue
		} else {
			break
		}
	}

	if end != nil {
		t.TaskResult.AppendEnding(end, mgr.Runner.Host.Name)
		if end.GetErr() != nil {
			t.TaskResult.AppendErr(end.GetErr())
		}
	}
	resChan <- "done"
}
