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
)

// EtcdEnv defines the template of etcd's env.
var EtcdEnv = template.Must(template.New("etcd.env").Parse(
	dedent.Dedent(`# Environment file for etcd {{ .Tag }}
{{- if .DataDir }}
ETCD_DATA_DIR={{ .DataDir }}
{{- else }}
ETCD_DATA_DIR=/var/lib/etcd
{{- end }}
ETCD_ADVERTISE_CLIENT_URLS=https://{{ .Ip }}:{{ .Port }}
ETCD_INITIAL_ADVERTISE_PEER_URLS=https://{{ .Ip }}:{{ .PeerPort }}
ETCD_INITIAL_CLUSTER_STATE={{ .State }}
ETCD_METRICS=basic
ETCD_LISTEN_CLIENT_URLS=https://{{ .Ip }}:{{ .Port }},https://127.0.0.1:{{ .Port }}
ETCD_INITIAL_CLUSTER_TOKEN=k8s_etcd
ETCD_LISTEN_PEER_URLS=https://{{ .Ip }}:{{ .PeerPort }}
ETCD_NAME={{ .Name }}
ETCD_PROXY=off
ETCD_ENABLE_V2=true
ETCD_INITIAL_CLUSTER={{ .PeerAddresses }}
{{- if .ElectionTimeout }}
ETCD_ELECTION_TIMEOUT={{ .ElectionTimeout }}
{{- else }}
ETCD_ELECTION_TIMEOUT=5000
{{- end }}
{{- if .HeartbeatInterval }}
ETCD_HEARTBEAT_INTERVAL={{ .HeartbeatInterval }}
{{- else }}
ETCD_HEARTBEAT_INTERVAL=250
{{- end }}
{{- if .CompactionRetention  }}
ETCD_AUTO_COMPACTION_RETENTION={{ .CompactionRetention }}
{{- else }}
ETCD_AUTO_COMPACTION_RETENTION=8
{{- end }}
{{- if .SnapshotCount }}
ETCD_SNAPSHOT_COUNT={{ .SnapshotCount }}
{{- else }}
ETCD_SNAPSHOT_COUNT=10000
{{- end }}
{{- if .Metrics }}
ETCD_METRICS={{ .Metrics }}
{{- end }}
{{- if .QuotaBackendBytes }}
ETCD_QUOTA_BACKEND_BYTES={{ .QuotaBackendBytes }}
{{- end }}
{{- if .MaxRequestBytes }}
ETCD_MAX_REQUEST_BYTES={{ .MaxRequestBytes }}
{{- end }}
{{- if .MaxSnapshots }}
ETCD_MAX_SNAPSHOTS={{ .MaxSnapshots }}
{{- end }}
{{- if .MaxWals }}
ETCD_MAX_WALS={{ .MaxWals }}
{{- end }}
{{- if .LogLevel }}
ETCD_LOG_LEVEL={{ .LogLevel }}
{{- end }}
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
ETCD_PEER_CLIENT_CERT_AUTH=true

# CLI settings
ETCDCTL_ENDPOINTS=https://127.0.0.1:{{ .Port }}
ETCDCTL_CACERT=/etc/ssl/etcd/ssl/ca.pem
ETCDCTL_KEY=/etc/ssl/etcd/ssl/admin-{{ .Hostname }}-key.pem
ETCDCTL_CERT=/etc/ssl/etcd/ssl/admin-{{ .Hostname }}.pem
    `)))
