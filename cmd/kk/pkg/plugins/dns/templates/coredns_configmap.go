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

var (
	CorednsConfigMap = template.Must(template.New("coredns-configmap.yaml").Funcs(utils.FuncMap).Parse(
		dedent.Dedent(`---
apiVersion: v1
kind: ConfigMap
metadata:
  name: coredns
  namespace: kube-system
  labels:
      addonmanager.kubernetes.io/mode: EnsureExists
data:
  Corefile: |
{{- if .ExternalZones }}
{{- range .ExternalZones }}
{{ range .Zones }}{{ . | indent 4 }} {{ end }}{
        log
        errors
{{- if .Rewrite }}
{{- range .Rewrite }}
        rewrite {{ . }}
{{- end }}
{{- end }}
        forward .{{ range .Nameservers }} {{ . }}{{ end}}
        loadbalance
        cache {{ .Cache }}
        reload
{{- if $.DNSEtcHosts }}
        hosts /etc/coredns/hosts {
          fallthrough
        }
{{- end }}
    }
{{- end }}
{{- end }}
    .:53 {
{{- if .AdditionalConfigs }}
{{  .AdditionalConfigs | indent 8 }}
{{- end }}
        errors
        health {
          lameduck 5s
        }
{{- if .RewriteBlock }}
{{ .RewriteBlock | indent 8 }}
{{- end }}
        ready
        kubernetes {{ .ClusterDomain }} in-addr.arpa ip6.arpa {
          pods insecure
          fallthrough in-addr.arpa ip6.arpa
          ttl 30
        }
        prometheus :9153
        forward . {{ if .UpstreamDNSServers }}{{ range .UpstreamDNSServers }}{{ . }} {{ end }}{{else}}/etc/resolv.conf{{ end }} {
          prefer_udp
          max_concurrent 1000
        }
        cache 30
        loop
        reload
        loadbalance
{{- if .DNSEtcHosts }}
        hosts /etc/coredns/hosts {
          fallthrough
        }
{{- end }}
    }
{{- if .DNSEtcHosts }}
  hosts: |
{{ .DNSEtcHosts | indent 4 }}
{{- end }}

    `)))
)
