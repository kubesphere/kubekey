package templates

import (
	"github.com/lithammer/dedent"
	"text/template"
)

var (
	// K8sCertsRenewService defines the template of k8s-certs-renew service for systemd.
	K8sCertsRenewService = template.Must(template.New("k8s-certs-renew.service").Parse(
		dedent.Dedent(`[Unit]
Description=Renew K8S control plane certificates
[Service]
Type=oneshot
ExecStart=/usr/local/bin/kube-scripts/k8s-certs-renew.sh
    `)))

	// K8sCertsRenewTimer defines the template of k8s-certs-renew timer for systemd.
	K8sCertsRenewTimer = template.Must(template.New("k8s-certs-renew.timer").Parse(
		dedent.Dedent(`[Unit]
Description=Timer to renew K8S control plane certificates
[Timer]
OnCalendar=Mon *-*-1,2,3,4,5,6,7 03:00:00
Unit=k8s-certs-renew.service
[Install]
WantedBy=multi-user.target
    `)))
)
