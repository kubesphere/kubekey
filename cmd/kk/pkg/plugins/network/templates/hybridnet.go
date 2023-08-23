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
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/utils"
	"github.com/lithammer/dedent"
	"text/template"
)

var HybridnetNetworks = template.Must(template.New("hybridnet-networks.yaml").Funcs(utils.FuncMap).Parse(
	dedent.Dedent(`
{{- range $index, $network := .Networks }}
---
apiVersion: networking.alibaba.com/v1 
kind: Network 
metadata: 
  name: {{ $network.Name }}
spec: 
{{- if $network.NetID }}
  netID: {{ $network.NetID }}
{{- end }}
  type: {{ $network.Type }}
{{- if $network.Mode }}
  mode: {{ $network.Mode }}
{{- end }}
{{- if $network.NodeSelector }}
  nodeSelector:
{{ toYaml $network.NodeSelector | indent 4 }}
{{- end }}

{{- range $network.Subnets }}
---
apiVersion: networking.alibaba.com/v1 
kind: Subnet             
metadata: 
  name: {{ .Name }}                  
spec: 
  network: {{ $network.Name }}
{{- if .NetID }}
  netID: {{ .NetID }}
{{- end }}
  range: 
    version: "4"
    cidr: "{{ .CIDR }}"
{{- if .Gateway }}
    gateway: "{{ .Gateway }}"
{{- end }}
{{- if .Start}}
    start: "{{ .Start }}"
{{- end}}
{{- if .End}}
    end: "{{ .End }}"
{{- end }}
{{- if .ReservedIPs }}
    reservedIPs:
{{ toYaml .ReservedIPs | indent 4 }}
{{- end }}
{{- if .ExcludeIPs }}
    excludeIPs:
{{ toYaml .ReservedIPs | indent 4 }}
{{- end }}
{{- end }}
{{- end }}
    `)))
