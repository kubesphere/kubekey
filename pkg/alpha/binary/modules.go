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
	"github.com/kubesphere/kubekey/pkg/binaries"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/task"
	"github.com/kubesphere/kubekey/pkg/kubernetes"
)

type UpgradeBinaryModule struct {
	common.KubeModule
}

func (p *UpgradeBinaryModule) Init() {
	p.Name = "UpgradeBinaryModule"
	p.Desc = "Download the binary and synchronize kubernetes binaries"

	download := &task.LocalTask{
		Name:    "DownloadBinaries",
		Desc:    "Download installation binaries",
		Prepare: new(kubernetes.NotEqualPlanVersion),
		Action:  new(binaries.Download),
	}

	syncBinary := &task.RemoteTask{
		Name:     "SyncKubeBinary",
		Desc:     "Synchronize kubernetes binaries",
		Hosts:    p.Runtime.GetHostsByRole(common.K8s),
		Prepare:  new(kubernetes.NotEqualPlanVersion),
		Action:   new(kubernetes.SyncKubeBinary),
		Parallel: true,
		Retry:    2,
	}

	p.Tasks = []task.Interface{
		download,
		syncBinary,
	}
}
