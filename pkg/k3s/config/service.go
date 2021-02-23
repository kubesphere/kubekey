/*
Copyright 2020 The KubeSphere Authors.

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

package config

import (
	"fmt"
	"strings"
	"text/template"

	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/lithammer/dedent"
)

var (
	// K3sServiceTempl defines the template of kubelete service for systemd.
	K3sServiceTempl = template.Must(template.New("k3sService").Parse(
		dedent.Dedent(`[Unit]
Description=Lightweight Kubernetes
Documentation=https://k3s.io
Wants=network-online.target
After=network-online.target

[Install]
WantedBy=multi-user.target

[Service]
Type=notify
# EnvironmentFile=/etc/systemd/system/k3s.service.env
KillMode=process
Delegate=yes
# Having non-zero Limit*s causes performance problems due to accounting overhead
# in the kernel. We recommend using cgroups to do container-local accounting.
LimitNOFILE=1048576
LimitNPROC=infinity
LimitCORE=infinity
TasksMax=infinity
TimeoutStartSec=0
Restart=always
RestartSec=5s
ExecStartPre=-/sbin/modprobe br_netfilter
ExecStartPre=-/sbin/modprobe overlay
ExecStart=/usr/local/bin/k3s server 
    `)))

	// K3sEnvTempl defines the template of kubelet's Env for the kubelet's systemd service.
	K3sEnvTempl = template.Must(template.New("k3sEnv").Parse(
		dedent.Dedent(`# Note: This dropin only works with k3s
[Service]
{{ if .IsMaster }}
Environment="K3S_ARGS=--datastore-endpoint={{ .DataStoreEndPoint }}  --datastore-cafile={{ .DataStoreCaFile }}  --datastore-certfile={{ .DataStoreCertFile }}  --datastore-keyfile={{ .DataStoreKeyFile }}"
{{ end }}
Environment="K3S_EXTRA_ARGS=--node-name={{ .HostName }}  --node-ip={{ .NodeIP }} {{ if .Server }}--server={{ .Server }}{{ end }} {{ if .Token }}--token={{ .Token }}{{ end }}"
Environment="K3S_ROLE={{ if .IsMaster }}server{{ else }}agent{{ end }}"
ExecStart=
ExecStart=/usr/local/bin/k3s $K3S_ROLE $K3S_ARGS $K3S_EXTRA_ARGS
    `)))
)

// GenerateK3sService is used to generate kubelet's service content for systemd.
func GenerateK3sService() (string, error) {
	return util.Render(K3sServiceTempl, util.Data{})
}

// GenerateK3sEnv is used to generate the env content of kubelet's service for systemd.
func GenerateK3sEnv(mgr *manager.Manager, node *kubekeyapiv1alpha1.HostCfg, token string) (string, error) {
	// var containerRuntime string

	// generate etcd configuration
	var externalEtcd kubekeyapiv1alpha1.ExternalEtcd
	var endpointsList []string
	var caFile, certFile, keyFile string

	for _, host := range mgr.EtcdNodes {
		endpoint := fmt.Sprintf("https://%s:%s", host.InternalAddress, kubekeyapiv1alpha1.DefaultEtcdPort)
		endpointsList = append(endpointsList, endpoint)
	}
	externalEtcd.Endpoints = endpointsList

	externalEtcdEndpoints := strings.Join(endpointsList, ",")
	caFile = "/etc/ssl/etcd/ssl/ca.pem"
	certFile = fmt.Sprintf("/etc/ssl/etcd/ssl/node-%s.pem", mgr.MasterNodes[0].Name)
	keyFile = fmt.Sprintf("/etc/ssl/etcd/ssl/node-%s-key.pem", mgr.MasterNodes[0].Name)

	externalEtcd.CaFile = caFile
	externalEtcd.CertFile = certFile
	externalEtcd.KeyFile = keyFile

	var server string
	if token != "" {
		server = fmt.Sprintf("https://%s:6443", mgr.MasterNodes[0].InternalAddress)
	} else {
		server = ""
	}

	return util.Render(K3sEnvTempl, util.Data{
		"DataStoreEndPoint": externalEtcdEndpoints,
		"DataStoreCaFile":   caFile,
		"DataStoreCertFile": certFile,
		"DataStoreKeyFile":  keyFile,
		"IsMaster":          node.IsMaster,
		"NodeIP":            node.InternalAddress,
		"HostName":          node.Name,
		"Token":             token,
		"Server":            server,
		// "ContainerRuntime": containerRuntime,
	})
}
