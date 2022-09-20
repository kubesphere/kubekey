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

package images

import (
	"github.com/kubesphere/kubekey/pkg/alpha/binary"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/task"
	"github.com/kubesphere/kubekey/pkg/images"
	"github.com/kubesphere/kubekey/pkg/kubernetes"
)

type UpgradeImagesModule struct {
	common.KubeModule
}

func (p *UpgradeImagesModule) Init() {
	p.Name = "UpgradeImagesModule"
	p.Desc = "pull the images that cluster need"

	pull := &task.RemoteTask{
		Name:     "PullImages",
		Desc:     "Start to pull images on all nodes",
		Hosts:    p.Runtime.GetHostsByRole(common.K8s),
		Prepare:  new(kubernetes.NotEqualPlanVersion),
		Action:   new(images.PullImage),
		Parallel: true,
	}

	p.Tasks = []task.Interface{
		pull,
	}
}

type setBinaryCacheModule struct {
	common.KubeModule
}

func (p *setBinaryCacheModule) Init() {
	p.Name = "setBinaryCacheModule"
	p.Desc = "set the docker and containerd binary paths in cache"

	setBinaryCache := &task.LocalTask{
		Name:   "SetBinaryCache",
		Desc:   "Set Binary Path in PipelineCache",
		Action: &binary.GetBinaryPath{Binaries: []string{"docker", "containerd", "runc", "crictl"}},
	}

	p.Tasks = []task.Interface{
		setBinaryCache,
	}
}
