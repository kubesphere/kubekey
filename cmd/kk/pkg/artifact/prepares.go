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

package artifact

import (
	"fmt"

	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/connector"
)

type EnableDownload struct {
	common.ArtifactPrepare
}

func (e *EnableDownload) PreCheck(_ connector.Runtime) (bool, error) {
	for _, sys := range e.Manifest.Spec.OperatingSystems {
		if sys.Repository.Iso.LocalPath == "" && sys.Repository.Iso.Url != "" {
			return true, nil
		}
	}
	return false, nil
}

type Md5AreEqual struct {
	common.KubePrepare
	Not bool
}

func (m *Md5AreEqual) PreCheck(_ connector.Runtime) (bool, error) {
	equal, ok := m.ModuleCache.GetMustBool("md5AreEqual")
	if !ok {
		return false, fmt.Errorf("get md5 equal value from module cache failed")
	}

	if equal {
		return !m.Not, nil
	}
	return m.Not, nil
}
