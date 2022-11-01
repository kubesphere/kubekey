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
)

const (
	controlPlaneJoinCloudInit = `{{.Header}}
{{template "files" .WriteFiles}}
-   path: /run/cluster-api/placeholder
    owner: root:root
    permissions: '0640'
    content: "This placeholder file is used to create the /run/cluster-api sub directory in a way that is compatible with both Linux and Windows (mkdir -p /run/cluster-api does not work with Windows)"
runcmd:
{{- template "commands" .PreK3sCommands }}
  - "INSTALL_K3S_SKIP_DOWNLOAD=true INSTALL_K3S_EXEC='server' /usr/local/bin/k3s-install.sh"
{{- template "commands" .PostK3sCommands }}
`
)

// NewJoinControlPlane returns the cloudinit string to be used on joining a control plane instance.
func NewJoinControlPlane(input *ControlPlaneInput) ([]byte, error) {
	input.prepare()
	userData, err := generate("JoinControlplane", controlPlaneJoinCloudInit, input)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to generate user data for machine joining control plane")
	}

	return userData, err
}
