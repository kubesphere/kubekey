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

package filesystem

import (
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/module"
	"github.com/kubesphere/kubekey/pkg/core/task"
)

type ChownModule struct {
	module.BaseTaskModule
}

func (c *ChownModule) Init() {
	c.Name = "ChownModule"
	c.Desc = "Change file and dir mode and owner"

	userKubeDir := &task.RemoteTask{
		Name:     "ChownFileAndDir",
		Desc:     "Chown user $HOME/.kube dir",
		Hosts:    c.Runtime.GetHostsByRole(common.K8s),
		Action:   &ChownFileAndDir{Path: "$HOME/.kube"},
		Parallel: true,
	}

	c.Tasks = []task.Interface{
		userKubeDir,
	}
}

type ChownWorkDirModule struct {
	common.ArtifactModule
}

func (c *ChownWorkDirModule) Init() {
	c.Name = "ChownWorkDirModule"
	c.Desc = "Change file and dir owner"

	userKubeDir := &task.LocalTask{
		Name:   "ChownFileAndDir",
		Desc:   "Chown ./kubekey dir",
		Action: &LocalTaskChown{Path: c.Runtime.GetWorkDir()},
	}

	output := &task.LocalTask{
		Name:   "Chown output file",
		Desc:   "Chown output file",
		Action: &LocalTaskChown{Path: c.Manifest.Arg.Output},
	}

	c.Tasks = []task.Interface{
		userKubeDir,
		output,
	}
}
