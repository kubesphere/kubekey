/*
 Copyright 2022 The KubeSphere Authors.

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

// K8eUninstallScript defines the template of k3s-killall script.
var K8eUninstallScript = template.Must(template.New("k8e-uninstall.sh").Parse(
	dedent.Dedent(`#!/bin/sh
set -x
[ $(id -u) -eq 0 ] || exec sudo $0 $@

/usr/local/bin/k8e-killall.sh

if which systemctl; then
    systemctl disable k8e
    systemctl reset-failed k8e
    systemctl daemon-reload
fi
if which rc-update; then
    rc-update delete k8e default
fi

rm -f /etc/systemd/system/k8e.service
rm -rf /etc/systemd/system/k8e.service.d
rm -f /etc/systemd/system/k8e.service.env

remove_uninstall() {
    rm -f /usr/local/bin/k8e-uninstall.sh
}
trap remove_uninstall EXIT

if (ls /etc/systemd/system/k8e*.service || ls /etc/init.d/k8e*) >/dev/null 2>&1; then
    set +x; echo 'Additional k8e services installed, skipping uninstall of k8e'; set -x
    exit
fi

for cmd in kubectl crictl ctr; do
    if [ -L /usr/local/bin/$cmd ]; then
        rm -f /usr/local/bin/$cmd
    fi
done

rm -rf /etc/rancher/k8e
rm -rf /run/k8e
rm -rf /var/lib/k8e
rm -rf /var/lib/kubelet
rm -f /usr/local/bin/k8e
rm -f /usr/local/bin/k8e-killall.sh
    `)))
