package templates

import (
	"github.com/lithammer/dedent"
	"text/template"
)

// K3sUninstallScript defines the template of k3s-killall script.
var K3sUninstallScript = template.Must(template.New("k3s-uninstall.sh").Parse(
	dedent.Dedent(`#!/bin/sh
set -x
[ $(id -u) -eq 0 ] || exec sudo $0 $@

/usr/local/bin/k3s-killall.sh

if which systemctl; then
    systemctl disable k3s
    systemctl reset-failed k3s
    systemctl daemon-reload
fi
if which rc-update; then
    rc-update delete k3s default
fi

rm -f /etc/systemd/system/k3s.service
rm -rf /etc/systemd/system/k3s.service.d
rm -f /etc/systemd/system/k3s.service.env

remove_uninstall() {
    rm -f /usr/local/bin/k3s-uninstall.sh
}
trap remove_uninstall EXIT

if (ls /etc/systemd/system/k3s*.service || ls /etc/init.d/k3s*) >/dev/null 2>&1; then
    set +x; echo 'Additional k3s services installed, skipping uninstall of k3s'; set -x
    exit
fi

for cmd in kubectl crictl ctr; do
    if [ -L /usr/local/bin/$cmd ]; then
        rm -f /usr/local/bin/$cmd
    fi
done

rm -rf /etc/rancher/k3s
rm -rf /run/k3s
rm -rf /run/flannel
rm -rf /var/lib/rancher/k3s
rm -rf /var/lib/kubelet
rm -f /usr/local/bin/k3s
rm -f /usr/local/bin/k3s-killall.sh

if type yum >/dev/null 2>&1; then
    yum remove -y k3s-selinux
    rm -f /etc/yum.repos.d/rancher-k3s-common*.repo
fi
    `)))
