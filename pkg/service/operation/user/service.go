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

package user

import (
	"github.com/kubesphere/kubekey/v3/pkg/clients/ssh"
)

// Service holds a collection of interfaces.
// The interfaces are broken down like this to group functions together.
type Service struct {
	SSHClient ssh.Interface
	Name      string
	Desc      string
}

// NewService returns a new service given the remote instance Linux user.
func NewService(sshClient ssh.Interface, name, desc string) *Service {
	return &Service{
		SSHClient: sshClient,
		Name:      name,
		Desc:      desc,
	}
}
