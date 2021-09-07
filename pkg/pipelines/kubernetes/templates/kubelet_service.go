package templates

import (
	"github.com/lithammer/dedent"
	"text/template"
)

// KubeletService defines the template of kubelete service for systemd.
var KubeletService = template.Must(template.New("kubelet.service").Parse(
	dedent.Dedent(`[Unit]
Description=kubelet: The Kubernetes Node Agent
Documentation=http://kubernetes.io/docs/

[Service]
ExecStart=/usr/local/bin/kubelet
Restart=always
StartLimitInterval=0
RestartSec=10

[Install]
WantedBy=multi-user.target
    `)))
