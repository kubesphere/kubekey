package pipeline

import (
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/experiment/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/experiment/core/action"
	"github.com/kubesphere/kubekey/experiment/core/cache"
	"github.com/kubesphere/kubekey/experiment/core/config"
	"github.com/kubesphere/kubekey/experiment/core/ending"
	"github.com/kubesphere/kubekey/experiment/core/logger"
	"github.com/kubesphere/kubekey/experiment/core/prepare"
	"github.com/kubesphere/kubekey/experiment/core/runner"
	"github.com/kubesphere/kubekey/experiment/core/vars"
	"github.com/pkg/errors"
	"sync"
	"time"
)

type Task struct {
	Name     string
	Hosts    []kubekeyapiv1alpha1.HostCfg
	Prepare  prepare.Prepare
	Action   action.Action
	Vars     vars.Vars
	Parallel bool
	Retry    int
	Delay    time.Duration
	Serial   string

	Cache       *cache.Cache
	Runtime     *config.Runtime
	tag         string
	IgnoreError bool
	TaskResult  *ending.TaskResult
}

func (t *Task) Init(runtime *config.Runtime, cache *cache.Cache) {
	logger.Log.SetTask(t.Name)
	t.Cache = cache
	t.Runtime = runtime
	t.Default()
}

func (t *Task) Execute() error {
	logger.Log.Info("Begin Run")
	if t.TaskResult.IsFailed() {
		return t.TaskResult.CombineErr()
	}

	wg := &sync.WaitGroup{}
	// todo: user can customize the pool size
	routinePool := make(chan struct{}, DefaultCon)
	defer close(routinePool)

	for i := range t.Hosts {
		selfRuntime := t.Runtime.Copy()
		_ = t.ConfigureSelfRuntime(selfRuntime, &t.Hosts[i], i)

		if t.Parallel {
			go t.Run(selfRuntime, wg, routinePool)
		} else {
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
	logger.Log.SetNode(runtime.Runner.Host.Name)
	pool <- struct{}{}
	wg.Add(1)

	t.Prepare.Init(runtime, t.Cache)
	if ok := t.WhenWithRetry(); !ok {
		return
	}

	t.Action.Init(runtime, t.Cache)
	t.ExecuteWithRetry(wg, pool, runtime)
}

func (t *Task) When() (bool, error) {
	if t.Prepare == nil {
		return true, nil
	}
	if ok, err := t.Prepare.PreCheck(); err != nil {
		logger.Log.Error(err)
		t.TaskResult.AppendErr(errors.Wrapf(err, "task %s precheck failed", t.Name))
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
		logger.Log.Errorf("Execute task pre-check timeout, Timeout=%fs, after %d retries", t.Delay.Seconds(), t.Retry)
	}
	return pass
}

func (t *Task) ExecuteWithTimer(wg *sync.WaitGroup, pool chan struct{}, resChan chan string, runtime *config.Runtime) ending.Ending {
	// generate a timer
	go func(result chan string, pool chan struct{}) {
		select {
		case <-result:
		case <-time.After(time.Minute * DefaultTimeout):
			logger.Log.Fatalf("Execute task timeout, Timeout=%ds", DefaultTimeout)
		}
		<-pool
		wg.Done()
	}(resChan, pool)

	res := t.Action.WrapResult(t.Action.Execute(t.Vars))
	return res
}

func (t *Task) ExecuteWithRetry(wg *sync.WaitGroup, pool chan struct{}, runtime *config.Runtime) {
	resChan := make(chan string)
	defer close(resChan)

	var end ending.Ending
	for i := 0; i < t.Retry; i++ {
		end = t.ExecuteWithTimer(wg, pool, resChan, runtime)
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
			t.TaskResult.AppendErr(errors.Wrapf(end.GetErr(), "task %s exec failed", t.Name))
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
		t.Delay = 3 * time.Second
	}
}
