package module

import (
	"context"
	"fmt"
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/common"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/ending"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/kubesphere/kubekey/pkg/core/prepare"
	"github.com/pkg/errors"
	"time"
)

type LocalTask struct {
	Name    string
	Desc    string
	Prepare prepare.Prepare
	Action  action.Action
	Retry   int
	Delay   time.Duration
	Timeout time.Duration

	PipelineCache *cache.Cache
	ModuleCache   *cache.Cache
	Runtime       connector.Runtime
	tag           string
	IgnoreError   bool
	TaskResult    *ending.TaskResult
}

func (l *LocalTask) GetDesc() string {
	return l.Desc
}

func (l *LocalTask) Init(runtime connector.Runtime, moduleCache *cache.Cache, pipelineCache *cache.Cache) {
	l.ModuleCache = moduleCache
	l.PipelineCache = pipelineCache
	l.Runtime = runtime
	l.Default()
}

func (l *LocalTask) Default() {
	l.TaskResult = ending.NewTaskResult()
	if l.Name == "" {
		l.Name = DefaultTaskName
	}

	if l.Prepare == nil {
		l.Prepare = new(prepare.BasePrepare)
	}

	if l.Action == nil {
		l.TaskResult.AppendErr(nil, errors.New("the action is nil"))
		return
	}

	if l.Retry < 1 {
		l.Retry = 1
	}

	if l.Delay <= 0 {
		l.Delay = 5 * time.Second
	}

	if l.Timeout <= 0 {
		l.Timeout = DefaultTimeout * time.Minute
	}
}

func (l *LocalTask) Execute() *ending.TaskResult {
	if l.TaskResult.IsFailed() {
		return l.TaskResult
	}

	host := &connector.BaseHost{
		Name: common.LocalHost,
	}

	selfRuntime := l.Runtime.Copy()
	l.RunWithTimeout(selfRuntime, host)

	if l.TaskResult.IsFailed() {
		l.TaskResult.ErrResult()
		return l.TaskResult
	}

	l.TaskResult.NormalResult()
	return l.TaskResult
}

func (l *LocalTask) RunWithTimeout(runtime connector.Runtime, host connector.Host) {
	ctx, cancel := context.WithTimeout(context.Background(), l.Timeout)
	defer cancel()

	errCh := make(chan error)
	defer close(errCh)

	go l.Run(runtime, host, errCh)
	select {
	case <-ctx.Done():
		l.TaskResult.AppendErr(host, fmt.Errorf("execute task timeout, Timeout=%d", l.Timeout))
	case e := <-errCh:
		if e != nil {
			l.TaskResult.AppendErr(host, e)
		}
	}
}

func (l *LocalTask) Run(runtime connector.Runtime, host connector.Host, errCh chan error) {
	runtime.SetRunner(&connector.Runner{
		Conn: nil,
		//Debug: runtime.Arg.Debug,
		Host: host,
	})

	l.Prepare.Init(l.ModuleCache, l.PipelineCache)
	l.Prepare.AutoAssert(runtime)
	if ok, e := l.WhenWithRetry(runtime, host); !ok {
		if e != nil {
			errCh <- e
			return
		} else {
			l.TaskResult.AppendSkip(host)
			errCh <- nil
			return
		}
	}

	l.Action.Init(l.ModuleCache, l.PipelineCache)
	l.Action.AutoAssert(runtime)
	if err := l.ExecuteWithRetry(runtime, host); err != nil {
		errCh <- err
		return
	}
	l.TaskResult.AppendSuccess(host)
	errCh <- nil
}

func (l *LocalTask) WhenWithRetry(runtime connector.Runtime, host connector.Host) (bool, error) {
	pass := false
	err := fmt.Errorf("pre-check exec failed after %d retires", l.Retry)
	for i := 0; i < l.Retry; i++ {
		if res, e := l.When(runtime); e != nil {
			logger.Log.Messagef(host.GetName(), e.Error())
			logger.Log.Infof("retry: [%s]", host.GetName())

			if i == l.Retry-1 {
				err = errors.New(err.Error() + e.Error())
				continue
			}
			time.Sleep(l.Delay)
			continue
		} else {
			err = nil
			pass = res
			break
		}
	}

	return pass, err
}

func (l *LocalTask) When(runtime connector.Runtime) (bool, error) {
	if l.Prepare == nil {
		return true, nil
	}
	if ok, err := l.Prepare.PreCheck(runtime); err != nil {
		return false, err
	} else if !ok {
		return false, nil
	} else {
		return true, nil
	}
}

func (l *LocalTask) ExecuteWithRetry(runtime connector.Runtime, host connector.Host) error {
	err := fmt.Errorf("[%s] exec failed after %d retires: ", l.Name, l.Retry)
	for i := 0; i < l.Retry; i++ {
		e := l.Action.Execute(runtime)
		if e != nil {
			logger.Log.Messagef(host.GetName(), e.Error())
			logger.Log.Infof("retry: [%s]", host.GetName())

			if i == l.Retry-1 {
				err = errors.New(err.Error() + e.Error())
				continue
			}
			time.Sleep(l.Delay)
			continue
		} else {
			err = nil
			break
		}
	}
	return err
}
