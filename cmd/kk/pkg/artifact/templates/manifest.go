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

	"github.com/lithammer/dedent"

	kubekeyv1alpha2 "github.com/kubesphere/kubekey/v3/cmd/kk/apis/kubekey/v1alpha2"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/util"
)

// Manifest defines the template of manifest file.
var Manifest = template.Must(template.New("Spec").Parse(
	dedent.Dedent(`
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Manifest
metadata:
  name: {{ .Options.Name }}
spec:
  arches:
  {{- range .Options.Arches }}
  - {{ . }}
  {{- end }}
  operatingSystems:
  {{- range $i, $v := .Options.OperatingSystems }}
  - arch: {{ $v.Arch }}
    type: {{ $v.Type }}
    id: {{ $v.Id }}
    version: "{{ $v.Version }}"
    osImage: {{ $v.OsImage }}
    repository:
      iso:
        localPath: 
        url: 
  {{- end }}
  kubernetesDistributions:
  {{- range $i, $v := .Options.KubernetesDistributions }}
  - type: {{ $v.Type }}
    version: {{ $v.Version }}
  {{- end}}
  components:
    helm: 
      version: {{ .Options.Components.Helm.Version }}
    cni: 
      version: {{ .Options.Components.CNI.Version }}
    etcd: 
      version: {{ .Options.Components.ETCD.Version }}
    containerRuntimes:
    {{- range $i, $v := .Options.Components.ContainerRuntimes }}
    - type: {{ $v.Type }}
      version: {{ $v.Version }}
    {{- end}}
    calicoctl:
      version: {{ .Options.Components.Calicoctl.Version }}
    crictl: 
      version: {{ .Options.Components.Crictl.Version }}
    {{ if .Options.Components.DockerRegistry.Version -}}
    docker-registry:
      version: "{{ .Options.Components.DockerRegistry.Version }}"
    harbor:
      version: {{ .Options.Components.Harbor.Version }}
    docker-compose:
      version: {{ .Options.Components.DockerCompose.Version }}
	{{- end }}
  images:
  {{- range .Options.Images }}
  - {{ . }}
  {{- end }}
  registry:
    auths: {}

    `)))

type Options struct {
	Name                    string
	Arches                  []string
	OperatingSystems        []kubekeyv1alpha2.OperatingSystem
	KubernetesDistributions []kubekeyv1alpha2.KubernetesDistribution
	Components              kubekeyv1alpha2.Components
	Images                  []string
}

func RenderManifest(opt *Options) (string, error) {
	return util.Render(Manifest, util.Data{
		"Options": opt,
	})
}
