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

// K8eKillallScript defines the template of k3s-killall script.
var K8eKillallScript = template.Must(template.New("k8e-killall.sh").Parse(
	dedent.Dedent(`#!/bin/sh
[ $(id -u) -eq 0 ] || exec sudo $0 $@

for service in /etc/systemd/system/k8e*.service; do
    [ -s $service ] && systemctl stop $(basename $service)
done
pschildren() {
    ps -e -o ppid= -o pid= | \
    sed -e 's/^\s*//g; s/\s\s*/\t/g;' | \
    grep -w "^$1" | \
    cut -f2
}
pstree() {
    for pid in $@; do
        echo $pid
        for child in $(pschildren $pid); do
            pstree $child
        done
    done
}
killtree() {
    kill -9 $(
        { set +x; } 2>/dev/null;
        pstree $@;
        set -x;
    ) 2>/dev/null
}
getshims() {
    ps -e -o pid= -o args= | sed -e 's/^ *//; s/\s\s*/\t/;' | grep -w 'k8e/data/[^/]*/bin/containerd-shim' | cut -f1
}
killtree $({ set +x; } 2>/dev/null; getshims; set -x)
do_unmount_and_remove() {
    set +x
    while read -r _ path _; do
        case "$path" in $1*) echo "$path" ;; esac
    done < /proc/self/mounts | sort -r | xargs -r -t -n 1 sh -c 'umount "$0" && rm -rf "$0"'
    set -x
}
do_unmount_and_remove '/run/k8e'
do_unmount_and_remove '/var/lib/k8e'
do_unmount_and_remove '/var/lib/kubelet/pods'
do_unmount_and_remove '/var/lib/kubelet/plugins'
do_unmount_and_remove '/run/netns/cni-'
# Remove CNI namespaces
ip netns show 2>/dev/null | grep cni- | xargs -r -t -n 1 ip netns delete
rm -rf /var/lib/cni/
    `)))
