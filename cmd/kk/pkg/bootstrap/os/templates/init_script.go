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
	"fmt"
	"strings"
	"text/template"

	"github.com/lithammer/dedent"

	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/bootstrap/registry"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/connector"
)

var InitOsScriptTmpl = template.Must(template.New("initOS.sh").Parse(
	dedent.Dedent(`#!/usr/bin/env bash

# Copyright 2020 The KubeSphere Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

DRYRUN=false
for arg in "$@"; do
    case "$arg" in
        --dry-run) DRYRUN=true ;;
    esac
done

# Update /etc/hosts in both dry-run and normal mode
sed -i ':a;$!{N;ba};s@# kubekey hosts BEGIN.*# kubekey hosts END@@' /etc/hosts
sed -i '/^$/N;/\n$/N;//D' /etc/hosts

cat >>/etc/hosts<<EOF
# kubekey hosts BEGIN
{{- range .Hosts }}
{{ . }}
{{- end }}
# kubekey hosts END
EOF

check_sysctl_conf() {
    local key="$1"
    local expected="$2"
    local current
    current=$(sysctl -n "$key" 2>/dev/null)
    if [ -z "$current" ]; then
        echo "${key} expect ${expected} but not set"
    elif [ "$current" != "$expected" ]; then
        echo "${key} expect ${expected} but got ${current}"
    fi
}

check_limits_conf() {
    local domain="$1"
    local type="$2"
    local item="$3"
    local expected="$4"
    local current
    current=$(awk -v d="$domain" -v t="$type" -v i="$item" '$1 == d && $2 == t && $3 == i {print $4; exit}' /etc/security/limits.conf 2>/dev/null)
    if [ -z "$current" ]; then
        echo "${domain} ${type} ${item} expect ${expected} but not set"
    elif [ "$current" != "$expected" ]; then
        echo "${domain} ${type} ${item} expect ${expected} but got ${current}"
    fi
}

if [ "$DRYRUN" = true ]; then
    echo "check begin"

    # swap
    if grep -q '^/' /proc/swaps 2>/dev/null; then
        echo "swap expect disabled but is enabled"
    fi
    if grep -q '^[^#]*swap' /etc/fstab 2>/dev/null; then
        echo "/etc/fstab swap entries expect commented but are active"
    fi

    # selinux
    if [ -f /etc/selinux/config ]; then
        current=$(grep '^SELINUX=' /etc/selinux/config 2>/dev/null | sed 's/^SELINUX=//')
        if [ -n "$current" ] && [ "$current" != "disabled" ]; then
            echo "SELINUX expect disabled but got ${current} in /etc/selinux/config"
        fi
    fi
    if command -v getenforce > /dev/null 2>&1; then
        current=$(getenforce 2>/dev/null)
        if [ -n "$current" ] && [ "$current" != "Disabled" ] && [ "$current" != "Permissive" ]; then
            echo "SELinux expect Disabled or Permissive but got ${current}"
        fi
    fi

    # sysctl
    check_sysctl_conf "net.ipv4.ip_forward" "1"
    check_sysctl_conf "net.bridge.bridge-nf-call-arptables" "1"
    check_sysctl_conf "net.bridge.bridge-nf-call-ip6tables" "1"
    check_sysctl_conf "net.bridge.bridge-nf-call-iptables" "1"
    check_sysctl_conf "net.ipv4.ip_local_reserved_ports" "30000-32767"
    check_sysctl_conf "net.core.netdev_max_backlog" "65535"
    check_sysctl_conf "net.core.rmem_max" "33554432"
    check_sysctl_conf "net.core.wmem_max" "33554432"
    check_sysctl_conf "net.core.somaxconn" "32768"
    check_sysctl_conf "net.ipv4.tcp_max_syn_backlog" "1048576"
    check_sysctl_conf "net.ipv4.neigh.default.gc_thresh1" "512"
    check_sysctl_conf "net.ipv4.neigh.default.gc_thresh2" "2048"
    check_sysctl_conf "net.ipv4.neigh.default.gc_thresh3" "4096"
    check_sysctl_conf "net.ipv4.tcp_retries2" "15"
    check_sysctl_conf "net.ipv4.tcp_max_tw_buckets" "1048576"
    check_sysctl_conf "net.ipv4.tcp_max_orphans" "65535"
    check_sysctl_conf "net.ipv4.udp_rmem_min" "131072"
    check_sysctl_conf "net.ipv4.udp_wmem_min" "131072"
    check_sysctl_conf "net.ipv4.conf.all.rp_filter" "1"
    check_sysctl_conf "net.ipv4.conf.default.rp_filter" "1"
    check_sysctl_conf "net.ipv4.conf.all.arp_accept" "1"
    check_sysctl_conf "net.ipv4.conf.default.arp_accept" "1"
    check_sysctl_conf "net.ipv4.conf.all.arp_ignore" "1"
    check_sysctl_conf "net.ipv4.conf.default.arp_ignore" "1"
    check_sysctl_conf "vm.max_map_count" "262144"
    check_sysctl_conf "vm.swappiness" "0"
    check_sysctl_conf "vm.overcommit_memory" "1"
    check_sysctl_conf "fs.inotify.max_user_instances" "524288"
    check_sysctl_conf "fs.inotify.max_user_watches" "10240001"
    check_sysctl_conf "fs.pipe-max-size" "4194304"
    check_sysctl_conf "fs.aio-max-nr" "262144"
    check_sysctl_conf "kernel.pid_max" "4194304"
    check_sysctl_conf "kernel.watchdog_thresh" "5"
    check_sysctl_conf "kernel.hung_task_timeout_secs" "5"
    # net.ipv4.tcp_tw_recycle is removed in Linux 4.12+, only check if it exists
    if sysctl -n net.ipv4.tcp_tw_recycle 2>/dev/null; then
        check_sysctl_conf "net.ipv4.tcp_tw_recycle" "0"
    fi
    check_sysctl_conf "net.ipv4.tcp_tw_reuse" "0"
    check_sysctl_conf "net.ipv4.conf.eth0.arp_accept" "1"
{{- if .IPv6Support }}
    check_sysctl_conf "net.ipv6.conf.all.disable_ipv6" "0"
    check_sysctl_conf "net.ipv6.conf.default.disable_ipv6" "0"
    check_sysctl_conf "net.ipv6.conf.lo.disable_ipv6" "0"
    check_sysctl_conf "net.ipv6.conf.all.forwarding" "1"
    check_sysctl_conf "net.ipv6.conf.default.accept_dad" "0"
    check_sysctl_conf "net.ipv6.route.max_size" "65536"
    check_sysctl_conf "net.ipv6.neigh.default.retrans_time_ms" "1000"
{{- end}}

    # limits
    check_limits_conf "*" "soft" "nofile" "1048576"
    check_limits_conf "*" "hard" "nofile" "1048576"
    check_limits_conf "*" "soft" "nproc" "65536"
    check_limits_conf "*" "hard" "nproc" "65536"
    check_limits_conf "*" "soft" "memlock" "unlimited"
    check_limits_conf "*" "hard" "memlock" "unlimited"

    # firewall
    if systemctl list-unit-files firewalld.service 2>/dev/null | grep -q 'firewalld.service'; then
        if systemctl is-active firewalld 2>/dev/null | grep -q 'active'; then
            echo "firewalld expect disabled but is active"
        fi
    fi
    if systemctl list-unit-files ufw.service 2>/dev/null | grep -q 'ufw.service'; then
        if systemctl is-active ufw 2>/dev/null | grep -q 'active'; then
            echo "ufw expect disabled but is active"
        fi
    fi

    # modules
    modinfo br_netfilter > /dev/null 2>&1
    if [ $? -eq 0 ]; then
        if ! lsmod 2>/dev/null | grep -q 'br_netfilter'; then
            echo "module br_netfilter expect loaded but is not"
        fi
    fi
    modinfo overlay > /dev/null 2>&1
    if [ $? -eq 0 ]; then
        if ! lsmod 2>/dev/null | grep -q 'overlay'; then
            echo "module overlay expect loaded but is not"
        fi
    fi
    for mod in ip_vs ip_vs_rr ip_vs_wrr ip_vs_sh; do
        if ! lsmod 2>/dev/null | grep -q "$mod"; then
            echo "module ${mod} expect loaded but is not"
        fi
    done
    if ! lsmod 2>/dev/null | grep -q 'nf_conntrack_ipv4' && ! lsmod 2>/dev/null | grep -q 'nf_conntrack'; then
        echo "module nf_conntrack_ipv4 or nf_conntrack expect loaded but is not"
    fi

    echo "check end"
    exit 0
fi

swapoff -a
sed -i /^[^#]*swap*/s/^/\#/g /etc/fstab

# See https://github.com/kubernetes/website/issues/14457
if [ -f /etc/selinux/config ]; then 
  sed -ri 's/SELINUX=enforcing/SELINUX=disabled/' /etc/selinux/config
fi
# for ubuntu: sudo apt install selinux-utils
# for centos: yum install selinux-policy
if command -v setenforce &> /dev/null
then
  setenforce 0
  getenforce
fi

# Helper function to safely set sysctl config
# Adds the config if not exists, or updates it if exists (including commented ones)
set_sysctl_config() {
    local key="$1"
    local value="$2"
    # Check if key exists (including commented lines)
    if grep -qE "^[#[:space:]]*${key}\s*=" /etc/sysctl.conf 2>/dev/null; then
        # Key exists, update it (remove comment if present and set value)
        sed -r -i "s@^[#[:space:]]*${key}\s*=.*@${key} = ${value}@g" /etc/sysctl.conf
    else
        # Key doesn't exist, append it
        echo "${key} = ${value}" >> /etc/sysctl.conf
    fi
}

set_sysctl_config "net.ipv4.ip_forward" "1"
set_sysctl_config "net.bridge.bridge-nf-call-arptables" "1"
set_sysctl_config "net.bridge.bridge-nf-call-ip6tables" "1"
set_sysctl_config "net.bridge.bridge-nf-call-iptables" "1"
set_sysctl_config "net.ipv4.ip_local_reserved_ports" "30000-32767"
set_sysctl_config "net.core.netdev_max_backlog" "65535"
set_sysctl_config "net.core.rmem_max" "33554432"
set_sysctl_config "net.core.wmem_max" "33554432"
set_sysctl_config "net.core.somaxconn" "32768"
set_sysctl_config "net.ipv4.tcp_max_syn_backlog" "1048576"
set_sysctl_config "net.ipv4.neigh.default.gc_thresh1" "512"
set_sysctl_config "net.ipv4.neigh.default.gc_thresh2" "2048"
set_sysctl_config "net.ipv4.neigh.default.gc_thresh3" "4096"
set_sysctl_config "net.ipv4.tcp_retries2" "15"
set_sysctl_config "net.ipv4.tcp_max_tw_buckets" "1048576"
set_sysctl_config "net.ipv4.tcp_max_orphans" "65535"
set_sysctl_config "net.ipv4.udp_rmem_min" "131072"
set_sysctl_config "net.ipv4.udp_wmem_min" "131072"
set_sysctl_config "net.ipv4.conf.all.rp_filter" "1"
set_sysctl_config "net.ipv4.conf.default.rp_filter" "1"
set_sysctl_config "net.ipv4.conf.all.arp_accept" "1"
set_sysctl_config "net.ipv4.conf.default.arp_accept" "1"
set_sysctl_config "net.ipv4.conf.all.arp_ignore" "1"
set_sysctl_config "net.ipv4.conf.default.arp_ignore" "1"
set_sysctl_config "vm.max_map_count" "262144"
set_sysctl_config "vm.swappiness" "0"
set_sysctl_config "vm.overcommit_memory" "1"
set_sysctl_config "fs.inotify.max_user_instances" "524288"
set_sysctl_config "fs.inotify.max_user_watches" "10240001"
set_sysctl_config "fs.pipe-max-size" "4194304"
set_sysctl_config "fs.aio-max-nr" "262144"
set_sysctl_config "kernel.pid_max" "4194304"
set_sysctl_config "kernel.watchdog_thresh" "5"
set_sysctl_config "kernel.hung_task_timeout_secs" "5"

{{- if .IPv6Support }}
#add for ipv6
set_sysctl_config "net.ipv6.conf.all.disable_ipv6" "0"
set_sysctl_config "net.ipv6.conf.default.disable_ipv6" "0"
set_sysctl_config "net.ipv6.conf.lo.disable_ipv6" "0"
set_sysctl_config "net.ipv6.conf.all.forwarding" "1"
set_sysctl_config "net.ipv6.conf.default.accept_dad" "0"
set_sysctl_config "net.ipv6.route.max_size" "65536"
set_sysctl_config "net.ipv6.neigh.default.retrans_time_ms" "1000"
{{- end}}

#See https://help.aliyun.com/document_detail/118806.html#uicontrol-e50-ddj-w0y
# net.ipv4.tcp_tw_recycle is removed in Linux 4.12+, only set if it exists
if sysctl -n net.ipv4.tcp_tw_recycle 2>/dev/null; then
    set_sysctl_config "net.ipv4.tcp_tw_recycle" "0"
fi
set_sysctl_config "net.ipv4.tcp_tw_reuse" "0"


tmpfile="$$.tmp"
awk ' !x[$0]++{print > "'$tmpfile'"}' /etc/sysctl.conf
mv $tmpfile /etc/sysctl.conf

# ulimit
# Helper function to safely set limits config
# Adds the config if not exists, or updates it if exists (including commented ones)
set_limits_config() {
    local domain="$1"
    local type="$2"
    local item="$3"
    local value="$4"
    # Check if entry exists (including commented lines)
    if grep -qE "^[#[:space:]]*${domain}\s+${type}\s+${item}\s+" /etc/security/limits.conf 2>/dev/null; then
        # Entry exists, update it (remove comment if present and set value)
        sed -r -i "s@^[#[:space:]]*${domain}\s+${type}\s+${item}\s+.*@${domain} ${type} ${item} ${value}@g" /etc/security/limits.conf
    else
        # Entry doesn't exist, append it
        echo "${domain} ${type} ${item} ${value}" >> /etc/security/limits.conf
    fi
}

set_limits_config "*" "soft" "nofile" "1048576"
set_limits_config "*" "hard" "nofile" "1048576"
set_limits_config "*" "soft" "nproc" "65536"
set_limits_config "*" "hard" "nproc" "65536"
set_limits_config "*" "soft" "memlock" "unlimited"
set_limits_config "*" "hard" "memlock" "unlimited"

tmpfile="$$.tmp"
awk ' !x[$0]++{print > "'$tmpfile'"}' /etc/security/limits.conf
mv $tmpfile /etc/security/limits.conf

systemctl stop firewalld 1>/dev/null 2>/dev/null
systemctl disable firewalld 1>/dev/null 2>/dev/null
systemctl stop ufw 1>/dev/null 2>/dev/null
systemctl disable ufw 1>/dev/null 2>/dev/null

modinfo br_netfilter > /dev/null 2>&1
if [ $? -eq 0 ]; then
   modprobe br_netfilter
   mkdir -p /etc/modules-load.d
   echo 'br_netfilter' > /etc/modules-load.d/kubekey-br_netfilter.conf
fi

modinfo overlay > /dev/null 2>&1
if [ $? -eq 0 ]; then
   modprobe overlay
   echo 'overlay' >> /etc/modules-load.d/kubekey-br_netfilter.conf
fi

modprobe ip_vs
modprobe ip_vs_rr
modprobe ip_vs_wrr
modprobe ip_vs_sh

cat > /etc/modules-load.d/kube_proxy-ipvs.conf << EOF
ip_vs
ip_vs_rr
ip_vs_wrr
ip_vs_sh
EOF

modprobe nf_conntrack_ipv4 1>/dev/null 2>/dev/null
if [ $? -eq 0 ]; then
   echo 'nf_conntrack_ipv4' > /etc/modules-load.d/kube_proxy-ipvs.conf
else
   modprobe nf_conntrack
   echo 'nf_conntrack' > /etc/modules-load.d/kube_proxy-ipvs.conf
fi
sysctl -p

sync
echo 3 > /proc/sys/vm/drop_caches

# Make sure the iptables utility doesn't use the nftables backend.
update-alternatives --set iptables /usr/sbin/iptables-legacy >/dev/null 2>&1 || true
update-alternatives --set ip6tables /usr/sbin/ip6tables-legacy >/dev/null 2>&1 || true
update-alternatives --set arptables /usr/sbin/arptables-legacy >/dev/null 2>&1 || true
update-alternatives --set ebtables /usr/sbin/ebtables-legacy >/dev/null 2>&1 || true

    `)))

func GenerateHosts(runtime connector.ModuleRuntime, kubeConf *common.KubeConf) []string {
	var lbHost string
	var hostsList []string

	if kubeConf.Cluster.ControlPlaneEndpoint.Address != "" {
		lbHost = fmt.Sprintf("%s  %s", kubeConf.Cluster.ControlPlaneEndpoint.Address, kubeConf.Cluster.ControlPlaneEndpoint.Domain)
	} else {
		lbHost = fmt.Sprintf("%s  %s", runtime.GetHostsByRole(common.Master)[0].GetInternalIPv4Address(), kubeConf.Cluster.ControlPlaneEndpoint.Domain)
	}

	for _, host := range runtime.GetAllHosts() {
		if host.GetName() != "" {
			if host.GetInternalIPv4Address() != "" {
				hostsList = append(hostsList, fmt.Sprintf("%s  %s.%s %s",
					host.GetInternalIPv4Address(),
					host.GetName(),
					kubeConf.Cluster.Kubernetes.ClusterName,
					host.GetName()))
			}
			if host.GetInternalIPv6Address() != "" {
				hostsList = append(hostsList, fmt.Sprintf("%s  %s.%s %s",
					host.GetInternalIPv6Address(),
					host.GetName(),
					kubeConf.Cluster.Kubernetes.ClusterName,
					host.GetName()))
			}
		}
	}

	if len(runtime.GetHostsByRole(common.Registry)) > 0 {
		if kubeConf.Cluster.Registry.PrivateRegistry != "" {
			hostsList = append(hostsList, fmt.Sprintf("%s  %s", runtime.GetHostsByRole(common.Registry)[0].GetInternalIPv4Address(), kubeConf.Cluster.Registry.PrivateRegistry))
			if runtime.GetHostsByRole(common.Registry)[0].GetInternalIPv6Address() != "" {
				hostsList = append(hostsList, fmt.Sprintf("%s  %s", runtime.GetHostsByRole(common.Registry)[0].GetInternalIPv6Address(), kubeConf.Cluster.Registry.PrivateRegistry))
			}

		} else {
			hostsList = append(hostsList, fmt.Sprintf("%s  %s", runtime.GetHostsByRole(common.Registry)[0].GetInternalIPv4Address(), registry.RegistryCertificateBaseName))
			if runtime.GetHostsByRole(common.Registry)[0].GetInternalIPv6Address() != "" {
				hostsList = append(hostsList, fmt.Sprintf("%s  %s", runtime.GetHostsByRole(common.Registry)[0].GetInternalIPv6Address(), registry.RegistryCertificateBaseName))
			}
		}

	}

	hostsList = append(hostsList, lbHost)
	return hostsList
}

func EnabledIPv6(kubeConf *common.KubeConf) bool {
	if len(strings.Split(kubeConf.Cluster.Network.KubePodsCIDR, ",")) == 2 {
		return true
	}
	return false
}
