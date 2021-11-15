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
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/lithammer/dedent"
	"text/template"
)

// Add is used in the template to implement the addition operation.
func Add(a int, b int) int {
	return a + b
}

var (
	funcMap = template.FuncMap{"Add": Add}

	// EtcdSslCfgTempl defines the template of openssl's configuration for etcd.
	ETCDOpenSSLConf = template.Must(template.New("openssl.conf").Funcs(funcMap).Parse(
		dedent.Dedent(`[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name

[req_distinguished_name]

[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names

[ ssl_client ]
extendedKeyUsage = clientAuth, serverAuth
basicConstraints = CA:FALSE
subjectKeyIdentifier=hash
authorityKeyIdentifier=keyid,issuer
subjectAltName = @alt_names

[ v3_ca ]
basicConstraints = CA:TRUE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names
authorityKeyIdentifier=keyid:always,issuer

[alt_names]
{{- range $i, $v := .Dns }}
DNS.{{ Add $i 1 }} = {{ $v }}
{{- end }}
{{- range $i, $v := .Ips }}
IP.{{ Add $i 1 }} = {{ $v }}
{{- end }}

    `)))
)

func DNSAndIp(kubeConf *common.KubeConf) (dns []string, ip []string) {
	dnsList := []string{"localhost", "etcd.kube-system.svc.cluster.local", "etcd.kube-system.svc", "etcd.kube-system", "etcd"}
	ipList := []string{"127.0.0.1"}

	if kubeConf.Cluster.ControlPlaneEndpoint.Domain == "" {
		dnsList = append(dnsList, kubekeyapiv1alpha2.DefaultLBDomain)
	} else {
		dnsList = append(dnsList, kubeConf.Cluster.ControlPlaneEndpoint.Domain)
	}

	for _, host := range kubeConf.Cluster.Hosts {
		dnsList = append(dnsList, host.Name)
		ipList = append(ipList, host.InternalAddress)
	}
	return dnsList, ipList
}
