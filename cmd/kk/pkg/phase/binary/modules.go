/*
 Copyright 2022 The KubeSphere Authors.

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

package binary

import (
	"github.com/kubesphere/kubekey/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/core/task"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/kubernetes"
)

type SyncBinaryModule struct {
	common.KubeModule
}

func (p *SyncBinaryModule) Init() {
	p.Name = "SyncBinaryModule"
	p.Desc = "synchronize kubernetes binaries"

	syncBinary := &task.RemoteTask{
		Name:     "SyncKubeBinary",
		Desc:     "Synchronize kubernetes binaries",
		Hosts:    p.Runtime.GetHostsByRole(common.K8s),
		Action:   new(kubernetes.SyncKubeBinary),
		Parallel: true,
		Retry:    2,
	}

	p.Tasks = []task.Interface{
		syncBinary,
	}
}
