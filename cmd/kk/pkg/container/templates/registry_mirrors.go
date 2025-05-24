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

package templates

import (
	"text/template"

	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/util"
)

// RegistryMirrors defines the template for generating registry mirrors configuration files for containerd
var RegistryMirrors = template.Must(template.New("hosts.toml").Parse(`
[host]
{{- if .Endpoint }}
  capabilities = ["pull", "resolve"]
  skip_verify = true
  [host.tls]
    ca_file = ""
    cert_file = ""
    key_file = ""
  [host.header]
{{- if .Endpoint }}
  [host.header.hosts]
    "{{ .Registry }}" = ["{{ .Endpoint }}"]
{{- end }}
{{- end }}
`))

// RemoteMirrorConfig represents a mirror configuration for a registry
type RemoteMirrorConfig struct {
	Registry string
	Endpoint string
}

// NewRemoteMirrorConfig creates a new RemoteMirrorConfig
func NewRemoteMirrorConfig(registry, endpoint string) util.Data {
	return util.Data{
		"Registry": registry,
		"Endpoint": endpoint,
	}
}
