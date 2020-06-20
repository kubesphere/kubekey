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

package tmpl

import (
	"fmt"
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/cluster/preinstall"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/lithammer/dedent"
	"strings"
	"text/template"
)

var (
	EtcdServiceTempl = template.Must(template.New("EtcdService").Parse(
		dedent.Dedent(`[Unit]
Description=etcd docker wrapper
Wants=docker.socket
After=docker.service

[Service]
User=root
PermissionsStartOnly=true
EnvironmentFile=-/etc/etcd.env
ExecStart=/usr/local/bin/etcd
ExecStartPre=-/usr/bin/docker rm -f {{ .Name }}
ExecStop=/usr/bin/docker stop {{ .Name }}
Restart=always
RestartSec=15s
TimeoutStartSec=30s

[Install]
WantedBy=multi-user.target
    `)))

	EtcdEnvTempl = template.Must(template.New("etcdEnv").Parse(
		dedent.Dedent(`# Environment file for etcd {{ .Tag }}
ETCD_DATA_DIR=/var/lib/etcd
ETCD_ADVERTISE_CLIENT_URLS=https://{{ .Ip }}:2379
ETCD_INITIAL_ADVERTISE_PEER_URLS=https://{{ .Ip }}:2380
ETCD_INITIAL_CLUSTER_STATE={{ .State }}
ETCD_METRICS=basic
ETCD_LISTEN_CLIENT_URLS=https://{{ .Ip }}:2379,https://127.0.0.1:2379
ETCD_ELECTION_TIMEOUT=5000
ETCD_HEARTBEAT_INTERVAL=250
ETCD_INITIAL_CLUSTER_TOKEN=k8s_etcd
ETCD_LISTEN_PEER_URLS=https://{{ .Ip }}:2380
ETCD_NAME={{ .Name }}
ETCD_PROXY=off
ETCD_INITIAL_CLUSTER={{ .peerAddresses }}
ETCD_AUTO_COMPACTION_RETENTION=8
ETCD_SNAPSHOT_COUNT=10000
{{- if .UnsupportedArch }}
ETCD_UNSUPPORTED_ARCH={{ .Arch }}
{{ end }}

# TLS settings
ETCD_TRUSTED_CA_FILE=/etc/ssl/etcd/ssl/ca.pem
ETCD_CERT_FILE=/etc/ssl/etcd/ssl/member-{{ .Hostname }}.pem
ETCD_KEY_FILE=/etc/ssl/etcd/ssl/member-{{ .Hostname }}-key.pem
ETCD_CLIENT_CERT_AUTH=true

ETCD_PEER_TRUSTED_CA_FILE=/etc/ssl/etcd/ssl/ca.pem
ETCD_PEER_CERT_FILE=/etc/ssl/etcd/ssl/member-{{ .Hostname }}.pem
ETCD_PEER_KEY_FILE=/etc/ssl/etcd/ssl/member-{{ .Hostname }}-key.pem
ETCD_PEER_CLIENT_CERT_AUTH=True

# CLI settings
ETCDCTL_ENDPOINTS=https://127.0.0.1:2379
ETCDCTL_CA_FILE=/etc/ssl/etcd/ssl/ca.pem
ETCDCTL_KEY_FILE=/etc/ssl/etcd/ssl/admin-{{ .Hostname }}-key.pem
ETCDCTL_CERT_FILE=/etc/ssl/etcd/ssl/admin-{{ .Hostname }}.pem
    `)))

	EtcdTempl = template.Must(template.New("etcd").Parse(
		dedent.Dedent(`#!/bin/bash
/usr/bin/docker run \
  --restart=on-failure:5 \
  --env-file=/etc/etcd.env \
  --net=host \
  -v /etc/ssl/certs:/etc/ssl/certs:ro \
  -v /etc/ssl/etcd/ssl:/etc/ssl/etcd/ssl:ro \
  -v /var/lib/etcd:/var/lib/etcd:rw \
  --memory=512M \
  --blkio-weight=1000 \
  --name={{ .Name }} \
  {{ .EtcdImage }} \
  /usr/local/bin/etcd \
  "$@"
    `)))
)

func GenerateEtcdBinary(mgr *manager.Manager, index int) (string, error) {
	return util.Render(EtcdTempl, util.Data{
		"Name":      fmt.Sprintf("etcd%d", index+1),
		"EtcdImage": preinstall.GetImage(mgr, "etcd").ImageName(),
	})
}

func GenerateEtcdService(index int) (string, error) {
	return util.Render(EtcdServiceTempl, util.Data{
		"Name": fmt.Sprintf("etcd%d", index+1),
	})
}

func GenerateEtcdEnv(node *kubekeyapi.HostCfg, index int, endpoints []string, state string) (string, error) {
	UnsupportedArch := false
	if node.Arch != "amd64" {
		UnsupportedArch = true
	}
	return util.Render(EtcdEnvTempl, util.Data{
		"Tag":             kubekeyapi.DefaultEtcdVersion,
		"Name":            fmt.Sprintf("etcd%d", index+1),
		"Ip":              node.InternalAddress,
		"Hostname":        node.Name,
		"State":           state,
		"peerAddresses":   strings.Join(endpoints, ","),
		"UnsupportedArch": UnsupportedArch,
		"Arch":            node.Arch,
	})
}
