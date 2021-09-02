package templates

import (
	"github.com/lithammer/dedent"
	"text/template"
)

var (
	// ETCDService defines the template of etcd's service for systemd.
	ETCDService = template.Must(template.New("etcd.service").Parse(
		dedent.Dedent(`[Unit]
Description=etcd
After=network.target

[Service]
User=root
Type=notify
EnvironmentFile=/etc/etcd.env
ExecStart=/usr/local/bin/etcd
NotifyAccess=all
RestartSec=10s
LimitNOFILE=40000
Restart=always

[Install]
WantedBy=multi-user.target
    `)))
)
