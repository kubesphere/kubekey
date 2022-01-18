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
	kubekeyapiv1alpha2 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha2"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/lithammer/dedent"
	"text/template"
)

// Cluster defines the template of cluster configuration file default.
var Cluster = template.Must(template.New("Cluster").Parse(
	dedent.Dedent(`
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Cluster
metadata:
  name: {{ .Options.Name }}
spec:
  hosts:
  - {name: node1, address: 172.16.0.2, internalAddress: 172.16.0.2, user: ubuntu, password: "Qcloud@123"}
  - {name: node2, address: 172.16.0.3, internalAddress: 172.16.0.3, user: ubuntu, password: "Qcloud@123"}
  roleGroups:
    etcd:
    - node1
    control-plane: 
    - node1
    worker:
    - node1
    - node2
  controlPlaneEndpoint:
    ## Internal loadbalancer for apiservers 
    # internalLoadbalancer: haproxy

    domain: lb.kubesphere.local
    address: ""
    port: 6443
  kubernetes:
    version: {{ .Options.KubeVersion }}
    clusterName: cluster.local
  network:
    plugin: calico
    kubePodsCIDR: 10.233.64.0/18
    kubeServiceCIDR: 10.233.0.0/18
    ## multus support. https://github.com/k8snetworkplumbingwg/multus-cni
    enableMultusCNI: false
  registry:
    plainHTTP: false
    privateRegistry: ""
    registryMirrors: []
    insecureRegistries: []
  addons: []

{{ if .Options.KubeSphereEnabled }}
{{ .Options.KubeSphereConfigMap }}
{{ end }}
    `)))

// Options defines the parameters of cluster configuration.
type Options struct {
	Name                string
	KubeVersion         string
	KubeSphereEnabled   bool
	KubeSphereConfigMap string
}

// GenerateCluster is used to generate cluster configuration content.
func GenerateCluster(opt *Options) (string, error) {
	return util.Render(Cluster, util.Data{
		"KubeVersion": kubekeyapiv1alpha2.DefaultKubeVersion,
		"Options":     opt,
	})
}
