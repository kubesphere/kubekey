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

package customscripts

import (
	"fmt"

	kubekeyapiv1alpha2 "github.com/kubesphere/kubekey/cmd/kk/apis/kubekey/v1alpha2"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/core/module"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/core/task"
)

type CustomScriptsModule struct {
	module.BaseTaskModule
	Phase   string
	Scripts []kubekeyapiv1alpha2.CustomScripts
}

func (m *CustomScriptsModule) Init() {
	m.Name = fmt.Sprintf("CustomScriptsModule Phase:%s", m.Phase)
	m.Desc = "Exec custom shell scripts for each nodes."

	for idx, script := range m.Scripts {

		taskName := fmt.Sprintf("Phase:%s(%d/%d) script:%s", m.Phase, idx, len(m.Scripts), script.Name)
		taskDir := fmt.Sprintf("%s-%d-script", m.Phase, idx)
		task := &task.RemoteTask{
			Name:     taskName,
			Desc:     taskName,
			Hosts:    m.Runtime.GetAllHosts(),
			Action:   &CustomScriptTask{taskDir: taskDir, script: script},
			Parallel: true,
			Retry: 1,
		}

		m.Tasks = append(m.Tasks, task)
	}
}
