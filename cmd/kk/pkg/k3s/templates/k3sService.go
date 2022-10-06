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
	"github.com/lithammer/dedent"
	"text/template"
)

var (
	// K3sService defines the template of kubelet service for systemd.
	K3sService = template.Must(template.New("k3s.service").Parse(
		dedent.Dedent(`[Unit]
Description=Lightweight Kubernetes
Documentation=https://k3s.io
Wants=network-online.target
After=network-online.target

[Install]
WantedBy=multi-user.target

[Service]
Type=notify
EnvironmentFile=/etc/systemd/system/k3s.service.env
{{ if .IsMaster }}
Environment="K3S_ARGS= {{ range .CertSANs }} --tls-san={{ . }}{{- end }} {{ range .ApiserverArgs }} --kube-apiserver-arg={{ . }}{{- end }} {{ range .ControllerManager }} --kube-controller-manager-arg={{ . }}{{- end }} {{ range .SchedulerArgs }} --kube-scheduler-arg={{ . }}{{- end }} --cluster-cidr={{ .PodSubnet }} --service-cidr={{ .ServiceSubnet }} --cluster-dns={{ .ClusterDns }} --flannel-backend=none --disable-network-policy --disable-cloud-controller --disable=servicelb,traefik,metrics-server,local-storage"
{{ end }}
Environment="K3S_EXTRA_ARGS=--node-name={{ .HostName }}  --node-ip={{ .NodeIP }}  --pause-image={{ .PauseImage }} {{ range .KubeletArgs }} --kubelet-arg={{ . }}{{- end }} {{ range .KubeProxyArgs }} --kube-proxy-arg={{ . }}{{- end }}"
Environment="K3S_ROLE={{ if .IsMaster }}server{{ else }}agent{{ end }}"
Environment="K3S_SERVER_ARGS={{ if .Server }}--server={{ .Server }}{{ end }}"
KillMode=process
Delegate=yes
LimitNOFILE=1048576
LimitNPROC=infinity
LimitCORE=infinity
TasksMax=infinity
TimeoutStartSec=0
Restart=always
RestartSec=5s
ExecStartPre=-/sbin/modprobe br_netfilter
ExecStartPre=-/sbin/modprobe overlay
ExecStart=/usr/local/bin/k3s $K3S_ROLE $K3S_ARGS $K3S_EXTRA_ARGS $K3S_SERVER_ARGS
    `)))
)
