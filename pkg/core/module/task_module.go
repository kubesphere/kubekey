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

package module

import (
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/ending"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/kubesphere/kubekey/pkg/core/task"
	"github.com/pkg/errors"
)

type BaseTaskModule struct {
	BaseModule
	Tasks []task.Interface
}

func (b *BaseTaskModule) Init() {
	if b.Name == "" {
		b.Name = DefaultTaskModuleName
	}
}

func (b *BaseTaskModule) Is() string {
	return TaskModuleType
}

func (b *BaseTaskModule) Run(result *ending.ModuleResult) {
	for i := range b.Tasks {
		t := b.Tasks[i]
		t.Init(b.Runtime.(connector.Runtime), b.ModuleCache, b.PipelineCache)

		logger.Log.Infof("[%s] %s", b.Name, t.GetDesc())
		res := t.Execute()
		for j := range res.ActionResults {
			ac := res.ActionResults[j]
			logger.Log.Infof("%s: [%s]", ac.Status.String(), ac.Host.GetName())
			result.AppendHostResult(ac)

			if _, ok := t.(*task.RemoteTask); ok {
				if b.Runtime.GetIgnoreErr() {
					if len(b.Runtime.GetAllHosts()) > 0 {
						if ac.GetStatus() == ending.FAILED {
							res.Status = ending.SUCCESS
							b.Runtime.DeleteHost(ac.Host)
						}
					} else {
						result.ErrResult(errors.Wrapf(res.CombineErr(), "Module[%s] exec failed", b.Name))
						return
					}
				}
			}
		}

		if res.IsFailed() {
			t.ExecuteRollback()
			result.ErrResult(errors.Wrapf(res.CombineErr(), "Module[%s] exec failed", b.Name))
			return
		}
	}
	result.NormalResult()
}
