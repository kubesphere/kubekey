package modules

import (
	"context"
	"fmt"
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/config"
	"github.com/kubesphere/kubekey/pkg/core/ending"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/kubesphere/kubekey/pkg/core/prepare"
	"github.com/kubesphere/kubekey/pkg/core/runner"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/core/vars"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"time"
)

type Task struct {
	Name        string
	Desc        string
	Hosts       []*kubekeyapiv1alpha1.HostCfg
	Prepare     prepare.Prepare
	Action      action.Action
	Vars        vars.Vars
	Parallel    bool
	Retry       int
	Delay       time.Duration
	Concurrency float64

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

func (t *Task) Execute() *ending.TaskResult {
	if t.TaskResult.IsFailed() {
		return t.TaskResult
	}

	routinePool := make(chan struct{}, t.calculateConcurrency())
	defer close(routinePool)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*DefaultTimeout)
	g, ctx := errgroup.WithContext(ctx)
	defer cancel()

	var err error
	for i := range t.Hosts {
		selfRuntime := t.Runtime.Copy()
		selfHost := t.Hosts[i].Copy()

		if t.Parallel {
			g.Go(func() error {
				if err := t.RunWithTimeout(ctx, selfRuntime, selfHost, i, routinePool); err != nil {
					return err
				}
				return nil
			})
		} else {
			if err = t.RunWithTimeout(ctx, selfRuntime, selfHost, i, routinePool); err != nil {
				break
			}
		}
	}
	if e := g.Wait(); e != nil {
		err = e
	}

	if err != nil {
		t.TaskResult.AppendErr(err)
		t.TaskResult.ErrResult()
		return t.TaskResult
	}

	t.TaskResult.NormalResult()
	return t.TaskResult
}

func (t *Task) RunWithTimeout(ctx context.Context, runtime *config.Runtime, host *kubekeyapiv1alpha1.HostCfg, index int, pool chan struct{}) error {
	pool <- struct{}{}

	errCh := make(chan error)
	defer close(errCh)
	go t.Run(runtime, host, index, errCh)
	select {
	case <-ctx.Done():
		<-pool
		return fmt.Errorf("execute task timeout, Timeout=%dm", DefaultTimeout)
	case e := <-errCh:
		<-pool
		return e
	}
}

func (t *Task) Run(runtime *config.Runtime, host *kubekeyapiv1alpha1.HostCfg, index int, errCh chan error) {
	if err := t.ConfigureSelfRuntime(runtime, host, index); err != nil {
		logger.Log.Errorf("failed: [%s]", host.Name)
		errCh <- err
		return
	}

	t.Prepare.Init(t.Cache, t.RootCache)
	if ok, e := t.WhenWithRetry(runtime); !ok {
		if e != nil {
			logger.Log.Errorf("failed: [%s]", host.Name)
			errCh <- e
			return
		} else {
			logger.Log.Errorf("skipped: [%s]", host.Name)
			errCh <- nil
			return
		}
	}

	t.Action.Init(t.Cache, t.RootCache)
	if err := t.ExecuteWithRetry(runtime); err != nil {
		logger.Log.Errorf("failed: [%s]", host.Name)
		errCh <- err
		return
	}
	logger.Log.Errorf("success: [%s]", runtime.Runner.Host.Name)
	errCh <- nil
}

func (t *Task) ConfigureSelfRuntime(runtime *config.Runtime, host *kubekeyapiv1alpha1.HostCfg, index int) error {
	conn, err := runtime.Connector.Connect(*host)
	if err != nil {
		return errors.Wrapf(err, "failed to connect to %s", host.Address)
	}

	runtime.Runner = &runner.Runner{
		Conn:  conn,
		Debug: runtime.Arg.Debug,
		Host:  host,
		Index: index,
	}
	return nil
}

func (t *Task) When(runtime *config.Runtime) (bool, error) {
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

func (t *Task) WhenWithRetry(runtime *config.Runtime) (bool, error) {
	pass := false
	err := fmt.Errorf("pre-check exec failed after %d retires", t.Retry)
	for i := 0; i < t.Retry; i++ {
		if res, e := t.When(runtime); e != nil {
			logger.Log.Infof("retry: [%s]", runtime.Runner.Host.Name)
			logger.Log.Error(e)
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

func (t *Task) ExecuteWithRetry(runtime *config.Runtime) error {
	err := fmt.Errorf("action exec failed after %d retires", t.Retry)
	for i := 0; i < t.Retry; i++ {
		e := t.Action.Execute(runtime, t.Vars)
		if e != nil {
			logger.Log.Infof("retry: [%s]", runtime.Runner.Host.Name)
			logger.Log.Error(e)
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
