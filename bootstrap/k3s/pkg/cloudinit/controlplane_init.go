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
	"github.com/kubesphere/kubekey/util/secret"
)

const (
	controlPlaneCloudInit = `{{.Header}}
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

// ControlPlaneInput defines the context to generate a controlplane instance user data.
type ControlPlaneInput struct {
	BaseUserData
	secret.Certificates

	ServerConfiguration string
}

// NewInitControlPlane returns the clouding string to be used on initializing a controlplane instance.
func NewInitControlPlane(input *ControlPlaneInput) ([]byte, error) {
	input.Header = cloudConfigHeader
	input.WriteFiles = input.Certificates.AsFiles()
	input.WriteFiles = append(input.WriteFiles, input.AdditionalFiles...)
	input.WriteFiles = append(input.WriteFiles, input.ConfigFile)
	input.SentinelFileCommand = sentinelFileCommand
	userData, err := generate("InitControlplane", controlPlaneCloudInit, input)
	if err != nil {
		return nil, err
	}

	return userData, nil
}
