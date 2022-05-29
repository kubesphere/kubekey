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

package hooks

import (
	"github.com/kubesphere/kubekey/cmd/kk/internal/common"
	kubekeycontroller "github.com/kubesphere/kubekey/controllers/kubekey"
	"github.com/kubesphere/kubekey/util/workflow/module"
)

type UpdateCRStatusHook struct {
	module.PostHook
}

func (u *UpdateCRStatusHook) Try() error {
	m := u.Module.(*module.BaseModule)
	kubeRuntime := m.Runtime.(*common.KubeRuntime)

	if !kubeRuntime.Arg.InCluster {
		return nil
	}

	if err := kubekeycontroller.UpdateClusterConditions(kubeRuntime, m.Desc, u.Result); err != nil {
		return err
	}
	return nil
}
