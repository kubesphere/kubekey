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
	kubekeyv1alpha2 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha2"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/lithammer/dedent"
	"text/template"
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
  operationSystems:
  {{- range $i, $v := .Options.OperationSystems }}
  - arch: {{ $v.Arch }}
    type: {{ $v.Type }}
    id: {{ $v.Id }}
    version: {{ $v.Version }}
    osImage: {{ $v.OsImage }}
    repository:
      iso:
        localPath: {{ $v.Repository.Iso.LocalPath }}
        url: {{ $v.Repository.Iso.Url }}
  {{- end }}
  kubernetesDistribution:
    type: {{ .Options.KubernetesDistribution.Type }}
    version: {{ .Options.KubernetesDistribution.Version }}
  components:
    helm: 
      version: {{ .Options.Components.Helm.Version }}
    cni: 
      version: {{ .Options.Components.CNI.Version }}
    etcd: 
      version: {{ .Options.Components.ETCD.Version }}
    # For now, if your cluster container runtime is containerd, KubeKey will add a docker 20.10.8 container runtime in the below list.
    # The reason is KubeKey creates a cluster with containerd by installing a docker first and making kubelet connect the socket file of containerd which docker contained.
    containerRuntimes:
      {{- range $i, $v := .Options.Components.ContainerRuntimes }}
      - type: {{ $v.Type }}
        version: {{ $v.Version }}
      {{- end}}
    crictl: 
      version: {{ .Options.Components.Crictl.Version }}
  images:
  {{- range .Options.Images }}
  - {{ . }}
  {{- end }}

    `)))

type Options struct {
	Name                   string
	Arches                 []string
	OperationSystems       []kubekeyv1alpha2.OperationSystem
	KubernetesDistribution kubekeyv1alpha2.KubernetesDistribution
	Components             kubekeyv1alpha2.Components
	Images                 []string
}

func RenderManifest(opt *Options) (string, error) {
	return util.Render(Manifest, util.Data{
		"Options": opt,
	})
}
