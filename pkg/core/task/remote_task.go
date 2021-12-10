/*
 Copyright 2021 The KubeSphere Authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package task

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/ending"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/kubesphere/kubekey/pkg/core/prepare"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/pkg/errors"
)

type RemoteTask struct {
	Name        string
	Desc        string
	Hosts       []connector.Host
	Prepare     prepare.Prepare
	Action      action.Action
	Parallel    bool
	Retry       int
	Delay       time.Duration
	Timeout     time.Duration
	Concurrency float64

	PipelineCache *cache.Cache
	ModuleCache   *cache.Cache
	Runtime       connector.Runtime
	tag           string
	IgnoreError   bool
	TaskResult    *ending.TaskResult
}

func (t *RemoteTask) GetDesc() string {
	return t.Desc
}

func (t *RemoteTask) Init(runtime connector.Runtime, moduleCache *cache.Cache, pipelineCache *cache.Cache) {
	t.ModuleCache = moduleCache
	t.PipelineCache = pipelineCache
	t.Runtime = runtime
	t.Default()
}

func (t *RemoteTask) Execute() *ending.TaskResult {
	if t.TaskResult.IsFailed() {
		return t.TaskResult
	}

	routinePool := make(chan struct{}, DefaultCon)
	defer close(routinePool)

	ctx, cancel := context.WithTimeout(context.Background(), t.Timeout)
	defer cancel()
	wg := &sync.WaitGroup{}
	for i := range t.Hosts {
		if t.Runtime.HostIsDeprecated(t.Hosts[i]) || t.Hosts[i] == nil {
			continue
		}
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

	if t.TaskResult.IsFailed() {
		t.TaskResult.ErrResult()
		return t.TaskResult
	}

	t.TaskResult.NormalResult()
	return t.TaskResult
}

func (t *RemoteTask) RunWithTimeout(ctx context.Context, runtime connector.Runtime, host connector.Host, index int,
	wg *sync.WaitGroup, pool chan struct{}) {

	pool <- struct{}{}

	errCh := make(chan error)
	defer close(errCh)

	go t.Run(ctx, runtime, host, index, errCh)
	select {
	case <-ctx.Done():
		t.TaskResult.AppendErr(host, fmt.Errorf("execute task timeout, Timeout=%d", t.Timeout))
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

func (t *RemoteTask) Run(ctx context.Context, runtime connector.Runtime, host connector.Host, index int, errCh chan error) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := t.ConfigureSelfRuntime(runtime, host, index); err != nil {
			errCh <- err
			return
		}

		t.Prepare.Init(t.ModuleCache, t.PipelineCache)
		t.Prepare.AutoAssert(runtime)
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

		t.Action.Init(t.ModuleCache, t.PipelineCache)
		t.Action.AutoAssert(runtime)
		if err := t.ExecuteWithRetry(runtime); err != nil {
			errCh <- err
			return
		}
		t.TaskResult.AppendSuccess(host)
		errCh <- nil
		return
	}
}

func (t *RemoteTask) ConfigureSelfRuntime(runtime connector.Runtime, host connector.Host, index int) error {
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

func (t *RemoteTask) When(runtime connector.Runtime) (bool, error) {
	if t.Prepare == nil {
		return true, nil
	}
	if ok, err := t.Prepare.PreCheck(runtime); err != nil {
		return false, err
	} else if !ok {
		return false, nil
	}
	return true, nil
}

func (t *RemoteTask) WhenWithRetry(runtime connector.Runtime) (bool, error) {
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

func (t *RemoteTask) ExecuteWithRetry(runtime connector.Runtime) error {
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

func (t *RemoteTask) Default() {
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

	if t.Timeout <= 0 {
		t.Timeout = DefaultTimeout * time.Minute
	}

	if t.Concurrency <= 0 || t.Concurrency > 1 {
		t.Concurrency = 1
	}
}

func (t *RemoteTask) calculateConcurrency() int {
	num := t.Concurrency * float64(len(t.Hosts))
	res := int(util.Round(num, 0))
	if res < 1 {
		res = 1
	}
	return res
}
