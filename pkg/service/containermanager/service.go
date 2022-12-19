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

package containermanager

import (
	"embed"
	"time"

	"github.com/kubesphere/kubekey/v3/pkg/clients/ssh"
	"github.com/kubesphere/kubekey/v3/pkg/scope"
	"github.com/kubesphere/kubekey/v3/pkg/service/operation/file"
)

//go:embed templates
var f embed.FS

// Service holds a collection of interfaces.
// The interfaces are broken down like this to group functions together.
type Service interface {
	Type() string
	Version() string
	IsExist() bool
	Get(timeout time.Duration) error
	Install() error
}

// NewService returns a new service given the remote instance container manager client.
func NewService(sshClient ssh.Interface, scope scope.KKInstanceScope, instanceScope *scope.InstanceScope) Service {
	switch instanceScope.ContainerManager().Type {
	case file.ContainerdID:
		return NewContainerdService(sshClient, scope, instanceScope)
	case file.DockerID:
		return NewDockerService(sshClient, scope, instanceScope)
	default:
		return NewContainerdService(sshClient, scope, instanceScope)
	}
}
