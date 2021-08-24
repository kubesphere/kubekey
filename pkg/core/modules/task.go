package modules

import (
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/config"
	"github.com/kubesphere/kubekey/pkg/core/ending"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/kubesphere/kubekey/pkg/core/prepare"
	"github.com/kubesphere/kubekey/pkg/core/runner"
	"github.com/kubesphere/kubekey/pkg/core/vars"
	"github.com/pkg/errors"
	"sync"
	"time"
)

type Task struct {
	Name     string
	Desc     string
	Hosts    []*kubekeyapiv1alpha1.HostCfg
	Prepare  prepare.Prepare
	Action   action.Action
	Vars     vars.Vars
	Parallel bool
	Retry    int
	Delay    time.Duration
	Serial   string

	RootCache   *cache.Cache
	Cache       *cache.Cache
	Runtime     *config.Runtime
	tag         string
	IgnoreError bool
	TaskResult  *ending.TaskResult
}

func (t *Task) Init(moduleName string, runtime *config.Runtime, cache *cache.Cache, rootCache *cache.Cache) {
	t.Cache = cache
	t.RootCache = rootCache
	t.Runtime = runtime
	t.Default()

	logger.Log.Infof("[%s] %s", moduleName, t.Desc)
}

// todo: maybe should redesign the ending
func (t *Task) Execute() error {
	if t.TaskResult.IsFailed() {
		return t.TaskResult.CombineErr()
	}

	wg := &sync.WaitGroup{}
	// todo: user can customize the pool size
	routinePool := make(chan struct{}, DefaultCon)
	defer close(routinePool)

	for i := range t.Hosts {
		selfRuntime := t.Runtime.Copy()
		selfHost := t.Hosts[i].Copy()
		_ = t.ConfigureSelfRuntime(selfRuntime, selfHost, i)
		if t.TaskResult.IsFailed() {
			logger.Log.Errorf("failed: [%s]", selfHost.Name)
			return t.TaskResult.CombineErr()
		}

		if t.Parallel {
			wg.Add(1)
			go t.Run(selfRuntime, wg, routinePool)
		} else {
			wg.Add(1)
			t.Run(selfRuntime, wg, routinePool)
		}
	}
	wg.Wait()

	if t.TaskResult.IsFailed() {
		return t.TaskResult.CombineErr()
	}
	t.TaskResult.NormalResult()
	return nil
}

func (t *Task) ConfigureSelfRuntime(runtime *config.Runtime, host *kubekeyapiv1alpha1.HostCfg, index int) error {

	conn, err := runtime.Connector.Connect(*host)
	if err != nil {
		t.TaskResult.AppendErr(errors.Wrapf(err, "Failed to connect to %s", host.Address))
		return errors.Wrapf(err, "Failed to connect to %s", host.Address)
	}

	runtime.Runner = &runner.Runner{
		Conn:  conn,
		Debug: runtime.Arg.Debug,
		Host:  host,
		Index: index,
	}
	return nil
}

func (t *Task) Run(runtime *config.Runtime, wg *sync.WaitGroup, pool chan struct{}) {
	// todo: check if it's ok when parallel.
	pool <- struct{}{}

	t.Prepare.Init(t.Cache, t.RootCache)
	if ok := t.WhenWithRetry(runtime); !ok {
		return
	}

	t.Action.Init(t.Cache, t.RootCache)
	t.ExecuteWithRetry(wg, pool, runtime)
	if t.TaskResult.IsFailed() {
		logger.Log.Errorf("failed: [%s]", runtime.Runner.Host.Name)
	} else {
		logger.Log.Infof("success: [%s]", runtime.Runner.Host.Name)
	}
}

func (t *Task) When(runtime *config.Runtime) (bool, error) {
	if t.Prepare == nil {
		return true, nil
	}
	if ok, err := t.Prepare.PreCheck(runtime); err != nil {
		t.TaskResult.AppendErr(err)
		return false, err
	} else if !ok {
		return false, nil
	} else {
		return true, nil
	}
}

func (t *Task) WhenWithRetry(runtime *config.Runtime) bool {
	pass := false
	timeout := true
	for i := 0; i < t.Retry; i++ {
		if res, err := t.When(runtime); err != nil {
			time.Sleep(t.Delay)
			continue
		} else {
			timeout = false
			pass = res
			break
		}
	}

	if timeout {
		logger.Log.Errorf("Execute task pre-check timeout, Timeout=%fs, after %d retries", t.Delay.Seconds(), t.Retry)
	}
	return pass
}

func (t *Task) ExecuteWithRetry(wg *sync.WaitGroup, pool chan struct{}, runtime *config.Runtime) {
	resChan := make(chan string)
	defer close(resChan)

	go func(result chan string, pool chan struct{}) {
		select {
		case <-result:
		case <-time.After(time.Minute * DefaultTimeout):
			logger.Log.Fatalf("Execute task timeout, Timeout=%ds", DefaultTimeout)
		}
		wg.Done()
		<-pool
	}(resChan, pool)

	var end ending.Ending
	for i := 0; i < t.Retry; i++ {
		res := t.Action.WrapResult(t.Action.Execute(runtime, t.Vars))
		end = res
		if end.GetErr() != nil {
			logger.Log.Error(end.GetErr())
			time.Sleep(t.Delay)
			continue
		} else {
			break
		}
	}

	if end != nil {
		t.TaskResult.AppendEnding(end, runtime.Runner.Host.Name)
		if end.GetErr() != nil {
			t.TaskResult.AppendErr(end.GetErr())
		}
	} else {
		t.TaskResult.ErrResult()
	}
	resChan <- "done"
}

func (t *Task) Default() {
	t.TaskResult = ending.NewTaskResult()
	if t.Name == "" {
		t.Name = DefaultTaskName
	}

	if len(t.Hosts) < 1 {
		t.Hosts = []*kubekeyapiv1alpha1.HostCfg{}
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
		t.Delay = 3 * time.Second
	}
}
