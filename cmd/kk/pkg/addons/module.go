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

package addons

import (
	"fmt"

	kubekeyapiv1alpha2 "github.com/kubesphere/kubekey/v3/cmd/kk/apis/kubekey/v1alpha2"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/task"
)

type AddonsModule struct {
	common.KubeModule
	Skip bool
}

func (a *AddonsModule) IsSkip() bool {
	return a.Skip
}

func (a *AddonsModule) Init() {
	a.Name = "AddonsModule"
	a.Desc = "Install addons"

	install := &task.LocalTask{
		Name:   "InstallAddons",
		Desc:   "Install addons",
		Action: new(Install),
	}

	a.Tasks = []task.Interface{
		install,
	}
}

type AddonModule struct {
	common.KubeModule
	addon *kubekeyapiv1alpha2.Addon
}

func (s *AddonModule) Init() {
	s.Name = "AddonModule"
	s.Desc = fmt.Sprintf("Install addon %s", s.addon.Name)

	install := &task.LocalTask{
		Name:   "InstallAddon",
		Desc:   fmt.Sprintf("Install addon %s", s.addon.Name),
		Action: &InstallAddon{addon: s.addon},
	}

	s.Tasks = []task.Interface{
		install,
	}
}
