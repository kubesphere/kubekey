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

type Task struct {
	Name        string
	Hosts       []kubekeyapiv1alpha1.HostCfg
	Action      action.Action
	Cache       *cache.Cache
	Log         logger.KubeKeyLog
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
	t.Log = *log
	t.Log.SetTask(t.Name)
	t.Cache = cache
	t.Default()
}

func (t *Task) Execute() error {
	t.Log.Info("Begin Run")
	if t.TaskResult.IsFailed() {
		return t.TaskResult.CombineErr()
	}

	wg := &sync.WaitGroup{}
	// todo: user can customize the pool size
	routinePool := make(chan struct{}, DefaultCon)
	defer close(routinePool)

	mgr := config.GetManager()
	for i := range t.Hosts {
		selfMgr := mgr.Copy()
		_ = t.ConfigureSelfManager(selfMgr, &t.Hosts[i], i)

		if t.Parallel {
			go t.Run(selfMgr, wg, routinePool)
		} else {
			t.Run(selfMgr, wg, routinePool)
		}
	}
	wg.Wait()

	if t.TaskResult.IsFailed() {
		return t.TaskResult.CombineErr()
	}
	t.TaskResult.NormalResult()
	return nil
}

func (t *Task) ConfigureSelfManager(mgr *config.Manager, host *kubekeyapiv1alpha1.HostCfg, index int) error {
	t.Log.SetNode(host.Name)
	mgr.Logger = t.Log

	conn, err := mgr.Connector.Connect(*host)
	if err != nil {
		t.TaskResult.AppendErr(errors.Wrapf(err, "Failed to connect to %s", host.Address))
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

func (t *Task) Run(mgr *config.Manager, wg *sync.WaitGroup, pool chan struct{}) {
	pool <- struct{}{}
	wg.Add(1)

	t.Prepare.Init(mgr, t.Cache)
	if ok := t.WhenWithRetry(mgr); !ok {
		return
	}

	t.Action.Init(mgr, t.Cache)
	t.ExecuteWithRetry(wg, pool, mgr)
}

func (t *Task) When(mgr *config.Manager) (bool, error) {
	if t.Prepare == nil {
		return true, nil
	}
	if ok, err := t.Prepare.PreCheck(); err != nil {
		mgr.Logger.Error(err)
		t.TaskResult.AppendErr(err)
		return false, err
	} else if !ok {
		return false, nil
	} else {
		return true, nil
	}
}

func (t *Task) WhenWithRetry(mgr *config.Manager) bool {
	pass := false
	timeout := true
	for i := 0; i < t.Retry; i++ {
		if res, err := t.When(mgr); err != nil {
			time.Sleep(t.Delay)
			continue
		} else {
			timeout = false
			pass = res
			break
		}
	}

	if timeout {
		mgr.Logger.Errorf("Execute task pre-check timeout, Timeout=%fs, after %d retries", t.Delay.Seconds(), t.Retry)
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

func (t *Task) Default() {
	t.TaskResult = ending.NewTaskResult()
	if t.Name == "" {
		t.Name = DefaultTaskName
	}

	if len(t.Hosts) < 1 {
		t.Hosts = []kubekeyapiv1alpha1.HostCfg{}
		t.TaskResult.AppendErr(errors.New("the length of task hosts is 0"))
		return
	}

	if t.Action == nil {
		t.TaskResult.AppendErr(errors.New("the action is nil"))
		return
	}

	if t.Retry < 1 {
		t.Retry = 1
	}

	if t.Delay <= 0 {
		t.Delay = 3
	}
}
