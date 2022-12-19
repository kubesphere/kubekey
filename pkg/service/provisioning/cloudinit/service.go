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

package cloudinit

import (
	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"

	"github.com/kubesphere/kubekey/v3/pkg/clients/ssh"
	"github.com/kubesphere/kubekey/v3/pkg/service/provisioning/commands"
)

// Service holds a collection of interfaces.
// The interfaces are broken down like this to group functions together.
type Service struct {
	SSHClient ssh.Interface
}

// NewService returns a new service.
func NewService(sshClient ssh.Interface) *Service {
	return &Service{
		SSHClient: sshClient,
	}
}

// RawBootstrapDataToProvisioningCommands converts raw bootstrap data to provisioning commands.
func (s *Service) RawBootstrapDataToProvisioningCommands(config []byte) ([]commands.Cmd, error) {
	// validate cloudConfigScript is a valid yaml, as required by the cloud config specification
	if err := yaml.Unmarshal(config, &map[string]interface{}{}); err != nil {
		return nil, errors.Wrapf(err, "cloud-config is not valid yaml")
	}

	// parse the cloud config yaml into a slice of cloud config actions.
	actions, err := getActions(s.SSHClient, config)
	if err != nil {
		return nil, err
	}

	commands := []commands.Cmd{}
	for _, action := range actions {
		cmds, err := action.Commands()
		if err != nil {
			return commands, err
		}
		commands = append(commands, cmds...)
	}

	return commands, nil
}
