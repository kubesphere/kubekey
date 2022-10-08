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

package prepare

import "github.com/kubesphere/kubekey/cmd/kk/pkg/core/connector"

type FileExist struct {
	BasePrepare
	FilePath string
	Not      bool
}

func (f *FileExist) PreCheck(runtime connector.Runtime) (bool, error) {
	exist, err := runtime.GetRunner().FileExist(f.FilePath)
	if err != nil {
		return false, err
	}
	if f.Not {
		return !exist, nil
	}
	return exist, nil
}
