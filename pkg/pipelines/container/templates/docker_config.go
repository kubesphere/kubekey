package templates

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"github.com/lithammer/dedent"
	"strings"
	"text/template"
)

var DockerConfig = template.Must(template.New("daemon.json").Parse(
	dedent.Dedent(`{
  "log-opts": {
    "max-size": "5m",
    "max-file":"3"
  },
  {{- if .Mirrors }}
  "registry-mirrors": [{{ .Mirrors }}],
  {{- end}}
  {{- if .InsecureRegistries }}
  "insecure-registries": [{{ .InsecureRegistries }}],
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
