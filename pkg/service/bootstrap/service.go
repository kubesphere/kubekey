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

package bootstrap

import (
	"os"
	"text/template"

	"github.com/kubesphere/kubekey/pkg/clients/ssh"
	"github.com/kubesphere/kubekey/pkg/scope"
	"github.com/kubesphere/kubekey/pkg/service/operation"
	"github.com/kubesphere/kubekey/pkg/service/operation/directory"
	"github.com/kubesphere/kubekey/pkg/service/operation/file"
	"github.com/kubesphere/kubekey/pkg/service/operation/user"
)

// Service holds a collection of interfaces.
// The interfaces are broken down like this to group functions together.
type Service struct {
	sshClient     ssh.Interface
	scope         scope.LBScope
	instanceScope *scope.InstanceScope

	userFactory      func(sshClient ssh.Interface, name, desc string) operation.User
	directoryFactory func(sshClient ssh.Interface, path string, mode os.FileMode) operation.Directory
	templateFactory  func(sshClient ssh.Interface, template *template.Template, data file.Data, dst string) (operation.Template, error)
}

// NewService returns a new service given the remote instance kubekey build-in bootstrap provision client.
func NewService(sshClient ssh.Interface, scope scope.LBScope, instanceScope *scope.InstanceScope) *Service {
	return &Service{
		sshClient:     sshClient,
		scope:         scope,
		instanceScope: instanceScope,
	}
}

func (s *Service) getUserService(name, desc string) operation.User {
	if s.userFactory != nil {
		return s.userFactory(s.sshClient, name, desc)
	}
	return user.NewService(s.sshClient, name, desc)
}

func (s *Service) getDirectoryService(path string, mode os.FileMode) operation.Directory {
	if s.directoryFactory != nil {
		return s.directoryFactory(s.sshClient, path, mode)
	}
	return directory.NewService(s.sshClient, path, mode)
}

func (s *Service) getTemplateService(template *template.Template, data file.Data, dst string) (operation.Template, error) {
	if s.templateFactory != nil {
		return s.templateFactory(s.sshClient, template, data, dst)
	}
	return file.NewTemplate(s.sshClient, s.scope.RootFs(), template, data, dst)
}
