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

package directory

import (
	"os"

	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/clients/ssh"
	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/util/filesystem"
)

type Service struct {
	SSHClient ssh.Interface
	Path      string
	Mode      os.FileMode
}

func NewService(sshClient ssh.Interface, path string, mode os.FileMode) *Service {
	return &Service{
		SSHClient: sshClient,
		Path:      path,
		Mode:      checkFileMode(mode),
	}
}

func checkFileMode(mode os.FileMode) os.FileMode {
	if mode.Perm() == 0 {
		mode = os.FileMode(filesystem.FileMode0664)
	}

	if mode.Type() != os.ModeDir {
		mode = os.ModeDir | mode
	}

	return mode
}
