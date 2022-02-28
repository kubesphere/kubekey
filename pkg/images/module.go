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
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/task"
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

//
//type LoadModule struct {
//	common.KubeModule
//}
//
//func (l *LoadModule) IsSkip() bool {
//	return l.Skip
//}
//
//func (l *LoadModule) Init() {
//	l.Name = "LoadModule"
//	l.Desc = "Load local tar file onto local image registry"
//
//	load := &task.LocalTask{
//		Name:   "LoadModule",
//		Desc:   "Load local tar file onto local image registry",
//		Action: new(PullImage),
//	}
//
//	l.Tasks = []task.Interface{
//		load,
//	}
//}

type PushModule struct {
	common.KubeModule
	Skip      bool
	ImagePath string
}

func (p *PushModule) IsSkip() bool {
	return p.Skip
}

func (p *PushModule) Init() {
	p.Name = "PushModule"
	p.Desc = "Push images on all nodes"

	push := &task.LocalTask{
		Name:   "PushImages",
		Desc:   "Push images to private registry",
		Action: &PushImage{ImagesPath: p.ImagePath},
	}

	p.Tasks = []task.Interface{
		push,
	}
}
