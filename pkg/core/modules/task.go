package modules

import (
	"context"
	"fmt"
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/ending"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/kubesphere/kubekey/pkg/core/prepare"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/pkg/errors"
	"sync"
	"time"
)

type Task struct {
	Name        string
	Desc        string
	Hosts       []connector.Host
	Prepare     prepare.Prepare
	Action      action.Action
	Parallel    bool
	Retry       int
	Delay       time.Duration
	Concurrency float64

	PipelineCache *cache.Cache
	ModuleCache   *cache.Cache
	Runtime       connector.Runtime
	tag           string
	IgnoreError   bool
	TaskResult    *ending.TaskResult
}

func (t *Task) Init(moduleName string, runtime connector.Runtime, moduleCache *cache.Cache, pipelineCache *cache.Cache) {
	t.ModuleCache = moduleCache
	t.PipelineCache = pipelineCache
	t.Runtime = runtime
	t.Default()

	logger.Log.Infof("[%s] %s", moduleName, t.Desc)
}

func (t *Task) Execute() *ending.TaskResult {
	if t.TaskResult.IsFailed() {
		return t.TaskResult
	}

	routinePool := make(chan struct{}, t.calculateConcurrency())
	defer close(routinePool)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*DefaultTimeout)
	defer cancel()
	wg := &sync.WaitGroup{}

	for i := range t.Hosts {
		selfRuntime := t.Runtime.Copy()
		selfHost := t.Hosts[i].Copy()

		wg.Add(1)
		if t.Parallel {
			go t.RunWithTimeout(ctx, selfRuntime, selfHost, i, wg, routinePool)
		} else {
			t.RunWithTimeout(ctx, selfRuntime, selfHost, i, wg, routinePool)
		}
	}
	wg.Wait()

	for _, res := range t.TaskResult.ActionResults {
		logger.Log.Infof("%s: [%s]", res.Status.String(), res.Host.GetName())
	}

	if t.TaskResult.IsFailed() {
		t.TaskResult.ErrResult()
		return t.TaskResult
	}

	t.TaskResult.NormalResult()
	return t.TaskResult
}

func (t *Task) RunWithTimeout(ctx context.Context, runtime connector.Runtime, host connector.Host, index int,
	wg *sync.WaitGroup, pool chan struct{}) {

	pool <- struct{}{}

	errCh := make(chan error)
	defer close(errCh)
	go t.Run(runtime, host, index, errCh)
	select {
	case <-ctx.Done():
		t.TaskResult.AppendErr(host, fmt.Errorf("execute task timeout, Timeout=%dm", DefaultTimeout))
		<-pool
		wg.Done()
	case e := <-errCh:
		if e != nil {
			t.TaskResult.AppendErr(host, e)
		}
		<-pool
		wg.Done()
	}
}

func (t *Task) Run(runtime connector.Runtime, host connector.Host, index int, errCh chan error) {
	if err := t.ConfigureSelfRuntime(runtime, host, index); err != nil {
		errCh <- err
		return
	}

	t.Prepare.Init(t.ModuleCache, t.PipelineCache, runtime)
	t.Prepare.AutoAssert()
	if ok, e := t.WhenWithRetry(runtime); !ok {
		if e != nil {
			errCh <- e
			return
		} else {
			t.TaskResult.AppendSkip(host)
			errCh <- nil
			return
		}
	}

	t.Action.Init(t.ModuleCache, t.PipelineCache, runtime)
	t.Action.AutoAssert()
	if err := t.ExecuteWithRetry(runtime); err != nil {
		errCh <- err
		return
	}
	t.TaskResult.AppendSuccess(host)
	errCh <- nil
}

func (t *Task) ConfigureSelfRuntime(runtime connector.Runtime, host connector.Host, index int) error {
	conn, err := runtime.GetConnector().Connect(host)
	if err != nil {
		return errors.Wrapf(err, "failed to connect to %s", host.GetAddress())
	}

	r := &connector.Runner{
		Conn: conn,
		//Debug: runtime.Arg.Debug,
		Host:  host,
		Index: index,
	}
	runtime.SetRunner(r)
	return nil
}

func (t *Task) When(runtime connector.Runtime) (bool, error) {
	if t.Prepare == nil {
		return true, nil
	}
	if ok, err := t.Prepare.PreCheck(runtime); err != nil {
		return false, err
	} else if !ok {
		return false, nil
	} else {
		return true, nil
	}
}

func (t *Task) WhenWithRetry(runtime connector.Runtime) (bool, error) {
	pass := false
	err := fmt.Errorf("pre-check exec failed after %d retires", t.Retry)
	for i := 0; i < t.Retry; i++ {
		if res, e := t.When(runtime); e != nil {
			logger.Log.Messagef(runtime.RemoteHost().GetName(), e.Error())
			logger.Log.Infof("retry: [%s]", runtime.GetRunner().Host.GetName())

			if i == t.Retry-1 {
				err = errors.New(err.Error() + e.Error())
				continue
			}
			time.Sleep(t.Delay)
			continue
		} else {
			err = nil
			pass = res
			break
		}
	}

	return pass, err
}

func (t *Task) ExecuteWithRetry(runtime connector.Runtime) error {
	err := fmt.Errorf("[%s] exec failed after %d retires: ", t.Name, t.Retry)
	for i := 0; i < t.Retry; i++ {
		e := t.Action.Execute(runtime)
		if e != nil {
			logger.Log.Messagef(runtime.RemoteHost().GetName(), e.Error())
			logger.Log.Infof("retry: [%s]", runtime.GetRunner().Host.GetName())

			if i == t.Retry-1 {
				err = errors.New(err.Error() + e.Error())
				continue
			}
			time.Sleep(t.Delay)
			continue
		} else {
			err = nil
			break
		}
	}
	return err
}

func (t *Task) Default() {
	t.TaskResult = ending.NewTaskResult()
	if t.Name == "" {
		t.Name = DefaultTaskName
	}

	if len(t.Hosts) < 1 {
		t.Hosts = []connector.Host{}
		t.TaskResult.AppendErr(nil, errors.New("the length of task hosts is 0"))
		return
	}

	if t.Prepare == nil {
		t.Prepare = new(prepare.BasePrepare)
	}

	if t.Action == nil {
		t.TaskResult.AppendErr(nil, errors.New("the action is nil"))
		return
	}

	if t.Retry < 1 {
		t.Retry = 3
	}

	if t.Delay <= 0 {
		t.Delay = 5 * time.Second
	}

	if t.Concurrency <= 0 || t.Concurrency > 1 {
		t.Concurrency = 1
	}
}

func (t *Task) calculateConcurrency() int {
	num := t.Concurrency * float64(len(t.Hosts))
	res := int(util.Round(num, 0))
	if res < 1 {
		res = 1
	}
	return res
}
