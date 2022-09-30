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

package images

import (
	"github.com/kubesphere/kubekey/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/core/task"
)

type PullModule struct {
	common.KubeModule
	Skip bool
}

func (p *PullModule) IsSkip() bool {
	return p.Skip
}

func (p *PullModule) Init() {
	p.Name = "PullModule"
	p.Desc = "Pull images on all nodes"

	pull := &task.RemoteTask{
		Name:     "PullImages",
		Desc:     "Start to pull images on all nodes",
		Hosts:    p.Runtime.GetAllHosts(),
		Action:   new(PullImage),
		Parallel: true,
	}

	p.Tasks = []task.Interface{
		pull,
	}
}

type CopyImagesToLocalModule struct {
	common.ArtifactModule
}

func (c *CopyImagesToLocalModule) Init() {
	c.Name = "CopyImagesToLocalModule"
	c.Desc = "Copy images to a local OCI path from registries"

	copyImage := &task.LocalTask{
		Name:   "SaveImages",
		Desc:   "Copy images to a local OCI path from registries",
		Action: new(SaveImages),
	}

	c.Tasks = []task.Interface{
		copyImage,
	}
}

type CopyImagesToRegistryModule struct {
	common.KubeModule
	Skip      bool
	ImagePath string
}

func (c *CopyImagesToRegistryModule) IsSkip() bool {
	return c.Skip
}

func (c *CopyImagesToRegistryModule) Init() {
	c.Name = "CopyImagesToRegistryModule"
	c.Desc = "Copy images to a private registry from an artifact OCI path"

	copyImage := &task.LocalTask{
		Name:   "CopyImagesToRegistry",
		Desc:   "Copy images to a private registry from an artifact OCI Path",
		Action: &CopyImagesToRegistry{ImagesPath: c.ImagePath},
	}

	pushManifest := &task.LocalTask{
		Name:   "PushManifest",
		Desc:   "Push multi-arch manifest to private registry",
		Action: new(PushManifest),
	}

	c.Tasks = []task.Interface{
		copyImage,
		pushManifest,
	}
}
