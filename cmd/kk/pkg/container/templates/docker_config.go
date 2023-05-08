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
	"fmt"
	"net"
	"strings"
	"text/template"

	"github.com/lithammer/dedent"

	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/common"
)

var DockerConfig = template.Must(template.New("daemon.json").Parse(
	dedent.Dedent(`{
  "log-opts": {
    "max-size": "5m",
    "max-file":"3"
  },
  {{- if .DataRoot }}
  "data-root": {{ .DataRoot }},
  {{- end}}
  {{- if .Mirrors }}
  "registry-mirrors": [{{ .Mirrors }}],
  {{- end}}
  {{- if .InsecureRegistries }}
  "insecure-registries": [{{ .InsecureRegistries }}],
  {{- end}}
  {{- if .BridgeBip }}
  "bip": {{ .BridgeBip }},
  {{- end}}
  "exec-opts": ["native.cgroupdriver=systemd"]
}
    `)))

func Mirrors(kubeConf *common.KubeConf) string {
	var mirrors string
	if kubeConf.Cluster.Registry.RegistryMirrors != nil {
		var mirrorsArr []string
		for _, mirror := range kubeConf.Cluster.Registry.RegistryMirrors {
			mirrorsArr = append(mirrorsArr, fmt.Sprintf("\"%s\"", mirror))
		}
		mirrors = strings.Join(mirrorsArr, ", ")
	}
	return mirrors
}

func InsecureRegistries(kubeConf *common.KubeConf) string {
	var insecureRegistries string
	if kubeConf.Cluster.Registry.InsecureRegistries != nil {
		var registriesArr []string
		for _, repo := range kubeConf.Cluster.Registry.InsecureRegistries {
			registriesArr = append(registriesArr, fmt.Sprintf("\"%s\"", repo))
		}
		insecureRegistries = strings.Join(registriesArr, ", ")
	}
	return insecureRegistries
}

func DataRoot(kubeConf *common.KubeConf) string {
	var dataRoot string
	if kubeConf.Cluster.Registry.DataRoot != "" {
		dataRoot = fmt.Sprintf("\"%s\"", kubeConf.Cluster.Registry.DataRoot)
	}
	return dataRoot
}

func BridgeIP(kubeConf *common.KubeConf) string {
	var bip string
	if kubeConf.Cluster.Registry.BridgeIP != "" {
		bip = "172.17.0.1/16"
		_, cidr, err := net.ParseCIDR(kubeConf.Cluster.Registry.BridgeIP)
		if err != nil {
			return bip
		}
		ip4 := cidr.IP.To4()
		if ip4 == nil {
			return bip
		}
		bridge0 := net.IPv4(ip4[0], ip4[1], ip4[2], ip4[3]+1)
		if !cidr.Contains(bridge0) {
			return bip
		}
		ones, _ := cidr.Mask.Size()

		bip = fmt.Sprintf("\"%s/%d\"", bridge0.String(), ones)
	}

	return bip
}
