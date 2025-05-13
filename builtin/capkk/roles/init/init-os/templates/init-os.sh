#!/usr/bin/env bash

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

# ------------------------ 1. Disable Swap and SELinux -----------------------
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

# ------------------------ 2. Network Settings (Sysctl) ------------------------

echo 'net.core.netdev_max_backlog = 65535' >> /etc/sysctl.conf
echo 'net.core.rmem_max = 33554432' >> /etc/sysctl.conf
echo 'net.core.wmem_max = 33554432' >> /etc/sysctl.conf
echo 'net.core.somaxconn = 32768' >> /etc/sysctl.conf
echo 'net.bridge.bridge-nf-call-arptables = 1' >> /etc/sysctl.conf
echo 'vm.max_map_count = 262144' >> /etc/sysctl.conf
echo 'vm.swappiness = 0' >> /etc/sysctl.conf
echo 'vm.overcommit_memory = 1' >> /etc/sysctl.conf
echo 'fs.inotify.max_user_instances = 524288' >> /etc/sysctl.conf
echo 'fs.inotify.max_user_watches = 10240001' >> /etc/sysctl.conf
echo 'fs.pipe-max-size = 4194304' >> /etc/sysctl.conf
echo 'fs.aio-max-nr = 262144' >> /etc/sysctl.conf
echo 'kernel.pid_max = 65535' >> /etc/sysctl.conf
echo 'kernel.watchdog_thresh = 5' >> /etc/sysctl.conf
echo 'kernel.hung_task_timeout_secs = 5' >> /etc/sysctl.conf
{{- if and .internal_ipv4 (.internal_ipv4 | ne "") }}
# add for ipv4
echo 'net.ipv4.ip_forward = 1' >> /etc/sysctl.conf
echo 'net.bridge.bridge-nf-call-ip6tables = 1' >> /etc/sysctl.conf
echo 'net.ipv4.ip_local_reserved_ports = 30000-32767' >> /etc/sysctl.conf
echo 'net.ipv4.tcp_max_syn_backlog = 1048576' >> /etc/sysctl.conf
echo 'net.ipv4.neigh.default.gc_thresh1 = 512' >> /etc/sysctl.conf
echo 'net.ipv4.neigh.default.gc_thresh2 = 2048' >> /etc/sysctl.conf
echo 'net.ipv4.neigh.default.gc_thresh3 = 4096' >> /etc/sysctl.conf
echo 'net.ipv4.tcp_retries2 = 15' >> /etc/sysctl.conf
echo 'net.ipv4.tcp_max_tw_buckets = 1048576' >> /etc/sysctl.conf
echo 'net.ipv4.tcp_max_orphans = 65535' >> /etc/sysctl.conf
echo 'net.ipv4.udp_rmem_min = 131072' >> /etc/sysctl.conf
echo 'net.ipv4.udp_wmem_min = 131072' >> /etc/sysctl.conf
echo 'net.ipv4.conf.all.rp_filter = 1' >> /etc/sysctl.conf
echo 'net.ipv4.conf.default.rp_filter = 1' >> /etc/sysctl.conf
echo 'net.ipv4.conf.all.arp_accept = 1' >> /etc/sysctl.conf
echo 'net.ipv4.conf.default.arp_accept = 1' >> /etc/sysctl.conf
echo 'net.ipv4.conf.all.arp_ignore = 1' >> /etc/sysctl.conf
echo 'net.ipv4.conf.default.arp_ignore = 1' >> /etc/sysctl.conf
{{- end }}
{{- if and .internal_ipv6 (.internal_ipv6 | ne "") }}
# add for ipv6
echo 'net.bridge.bridge-nf-call-iptables = 1' >> /etc/sysctl.conf
echo 'net.ipv6.conf.all.disable_ipv6 = 0' >> /etc/sysctl.conf
echo 'net.ipv6.conf.default.disable_ipv6 = 0' >> /etc/sysctl.conf
echo 'net.ipv6.conf.lo.disable_ipv6 = 0' >> /etc/sysctl.conf
echo 'net.ipv6.conf.all.forwarding=1' >> /etc/sysctl.conf
echo 'net.ipv6.conf.default.accept_dad=0' >> /etc/sysctl.conf
echo 'net.ipv6.route.max_size=65536' >> /etc/sysctl.conf
echo 'net.ipv6.neigh.default.retrans_time_ms=1000' >> /etc/sysctl.conf
{{- end }}

# ------------------------ 3. Tweaks for Specific Networking Configurations -----

#See https://help.aliyun.com/document_detail/118806.html#uicontrol-e50-ddj-w0y
sed -r -i  "s@#{0,}?net.bridge.bridge-nf-call-arptables ?= ?(0|1)@net.bridge.bridge-nf-call-arptables = 1@g" /etc/sysctl.conf
sed -r -i  "s@#{0,}?vm.max_map_count ?= ?([0-9]{1,})@vm.max_map_count = 262144@g" /etc/sysctl.conf
sed -r -i  "s@#{0,}?vm.swappiness ?= ?([0-9]{1,})@vm.swappiness = 0@g" /etc/sysctl.conf
sed -r -i  "s@#{0,}?fs.inotify.max_user_instances ?= ?([0-9]{1,})@fs.inotify.max_user_instances = 524288@g" /etc/sysctl.conf
sed -r -i  "s@#{0,}?kernel.pid_max ?= ?([0-9]{1,})@kernel.pid_max = 65535@g" /etc/sysctl.conf
sed -r -i "s@#{0,}?vm.overcommit_memory ?= ?(0|1|2)@vm.overcommit_memory = 0@g" /etc/sysctl.conf
sed -r -i  "s@#{0,}?fs.inotify.max_user_watches ?= ?([0-9]{1,})@fs.inotify.max_user_watches = 524288@g" /etc/sysctl.conf
sed -r -i  "s@#{0,}?fs.pipe-max-size ?= ?([0-9]{1,})@fs.pipe-max-size = 4194304@g" /etc/sysctl.conf
sed -r -i  "s@#{0,}?net.core.netdev_max_backlog ?= ?([0-9]{1,})@net.core.netdev_max_backlog = 65535@g" /etc/sysctl.conf
sed -r -i  "s@#{0,}?net.core.rmem_max ?= ?([0-9]{1,})@net.core.rmem_max = 33554432@g" /etc/sysctl.conf
sed -r -i  "s@#{0,}?net.core.wmem_max ?= ?([0-9]{1,})@net.core.wmem_max = 33554432@g" /etc/sysctl.conf
sed -r -i  "s@#{0,}?net.core.somaxconn ?= ?([0-9]{1,})@net.core.somaxconn = 32768@g" /etc/sysctl.conf
sed -r -i  "s@#{0,}?fs.aio-max-nr ?= ?([0-9]{1,})@fs.aio-max-nr = 262144@g" /etc/sysctl.conf
sed -r -i  "s@#{0,}?kernel.watchdog_thresh ?= ?([0-9]{1,})@kernel.watchdog_thresh = 5@g" /etc/sysctl.conf
sed -r -i  "s@#{0,}?kernel.hung_task_timeout_secs ?= ?([0-9]{1,})@kernel.hung_task_timeout_secs = 5@g" /etc/sysctl.conf
{{- if and .internal_ipv4 (.internal_ipv4 | ne "") }}
sed -r -i "s@#{0,}?net.ipv4.tcp_tw_recycle ?= ?(0|1|2)@net.ipv4.tcp_tw_recycle = 0@g" /etc/sysctl.conf
sed -r -i "s@#{0,}?net.ipv4.tcp_tw_reuse ?= ?(0|1)@net.ipv4.tcp_tw_reuse = 0@g" /etc/sysctl.conf
sed -r -i "s@#{0,}?net.ipv4.conf.all.rp_filter ?= ?(0|1|2)@net.ipv4.conf.all.rp_filter = 1@g" /etc/sysctl.conf
sed -r -i "s@#{0,}?net.ipv4.conf.default.rp_filter ?= ?(0|1|2)@net.ipv4.conf.default.rp_filter = 1@g" /etc/sysctl.conf
sed -r -i  "s@#{0,}?net.ipv4.ip_forward ?= ?(0|1)@net.ipv4.ip_forward = 1@g" /etc/sysctl.conf
sed -r -i  "s@#{0,}?net.bridge.bridge-nf-call-iptables ?= ?(0|1)@net.bridge.bridge-nf-call-iptables = 1@g" /etc/sysctl.conf
sed -r -i  "s@#{0,}?net.ipv4.ip_local_reserved_ports ?= ?([0-9]{1,}-{0,1},{0,1}){1,}@net.ipv4.ip_local_reserved_ports = 30000-32767@g" /etc/sysctl.conf
sed -r -i  "s@#{0,}?net.ipv4.tcp_max_syn_backlog ?= ?([0-9]{1,})@net.ipv4.tcp_max_syn_backlog = 1048576@g" /etc/sysctl.conf
sed -r -i  "s@#{0,}?net.ipv4.neigh.default.gc_thresh1 ?= ?([0-9]{1,})@net.ipv4.neigh.default.gc_thresh1 = 512@g" /etc/sysctl.conf
sed -r -i  "s@#{0,}?net.ipv4.neigh.default.gc_thresh2 ?= ?([0-9]{1,})@net.ipv4.neigh.default.gc_thresh2 = 2048@g" /etc/sysctl.conf
sed -r -i  "s@#{0,}?net.ipv4.neigh.default.gc_thresh3 ?= ?([0-9]{1,})@net.ipv4.neigh.default.gc_thresh3 = 4096@g" /etc/sysctl.conf
sed -r -i "s@#{0,}?net.ipv4.conf.eth0.arp_accept ?= ?(0|1)@net.ipv4.conf.eth0.arp_accept = 1@g" /etc/sysctl.conf
sed -r -i  "s@#{0,}?net.ipv4.tcp_retries2 ?= ?([0-9]{1,})@net.ipv4.tcp_retries2 = 15@g" /etc/sysctl.conf
sed -r -i  "s@#{0,}?net.ipv4.tcp_max_tw_buckets ?= ?([0-9]{1,})@net.ipv4.tcp_max_tw_buckets = 1048576@g" /etc/sysctl.conf
sed -r -i  "s@#{0,}?net.ipv4.tcp_max_orphans ?= ?([0-9]{1,})@net.ipv4.tcp_max_orphans = 65535@g" /etc/sysctl.conf
sed -r -i  "s@#{0,}?net.ipv4.udp_rmem_min ?= ?([0-9]{1,})@net.ipv4.udp_rmem_min = 131072@g" /etc/sysctl.conf
sed -r -i  "s@#{0,}?net.ipv4.udp_wmem_min ?= ?([0-9]{1,})@net.ipv4.udp_wmem_min = 131072@g" /etc/sysctl.conf
sed -r -i  "s@#{0,}?net.ipv4.conf.all.arp_ignore ?= ??(0|1|2)@net.ipv4.conf.all.arp_ignore = 1@g" /etc/sysctl.conf
sed -r -i  "s@#{0,}?net.ipv4.conf.default.arp_ignore ?= ??(0|1|2)@net.ipv4.conf.default.arp_ignore = 1@g" /etc/sysctl.conf
{{- end }}
{{- if and .internal_ipv6 (.internal_ipv6 | ne "") }}
#add for ipv6
sed -r -i  "s@#{0,}?net.bridge.bridge-nf-call-ip6tables ?= ?(0|1)@net.bridge.bridge-nf-call-ip6tables = 1@g" /etc/sysctl.conf
sed -r -i "s@#{0,}?net.ipv6.conf.all.disable_ipv6 ?= ?([0-9]{1,})@net.ipv6.conf.all.disable_ipv6 = 0@g" /etc/sysctl.conf
sed -r -i "s@#{0,}?net.ipv6.conf.default.disable_ipv6 ?= ?([0-9]{1,})@net.ipv6.conf.default.disable_ipv6 = 0@g" /etc/sysctl.conf
sed -r -i "s@#{0,}?net.ipv6.conf.lo.disable_ipv6 ?= ?([0-9]{1,})@net.ipv6.conf.lo.disable_ipv6 = 0@g" /etc/sysctl.conf
sed -r -i "s@#{0,}?net.ipv6.conf.all.forwarding ?= ?([0-9]{1,})@net.ipv6.conf.all.forwarding = 1@g" /etc/sysctl.conf
sed -r -i "s@#{0,}?net.ipv6.conf.default.accept_dad ?= ?([0-9]{1,})@net.ipv6.conf.default.accept_dad = 0@g" /etc/sysctl.conf
sed -r -i "s@#{0,}?net.ipv6.route.max_size ?= ?([0-9]{1,})@net.ipv6.route.max_size = 65536@g" /etc/sysctl.conf
sed -r -i "s@#{0,}?net.ipv6.neigh.default.retrans_time_ms ?= ?([0-9]{1,})@net.ipv6.neigh.default.retrans_time_ms = 1000@g" /etc/sysctl.conf
{{- end }}

tmpfile="$$.tmp"
awk ' !x[$0]++{print > "'$tmpfile'"}' /etc/sysctl.conf
mv $tmpfile /etc/sysctl.conf

# ------------------------ 4. Security Limit ------------------------------------

# ulimit
echo "* soft nofile 1048576" >> /etc/security/limits.conf
echo "* hard nofile 1048576" >> /etc/security/limits.conf
echo "* soft nproc 65536" >> /etc/security/limits.conf
echo "* hard nproc 65536" >> /etc/security/limits.conf
echo "* soft memlock unlimited" >> /etc/security/limits.conf
echo "* hard memlock unlimited" >> /etc/security/limits.conf

sed -r -i  "s@#{0,}?\* soft nofile ?([0-9]{1,})@\* soft nofile 1048576@g" /etc/security/limits.conf
sed -r -i  "s@#{0,}?\* hard nofile ?([0-9]{1,})@\* hard nofile 1048576@g" /etc/security/limits.conf
sed -r -i  "s@#{0,}?\* soft nproc ?([0-9]{1,})@\* soft nproc 65536@g" /etc/security/limits.conf
sed -r -i  "s@#{0,}?\* hard nproc ?([0-9]{1,})@\* hard nproc 65536@g" /etc/security/limits.conf
sed -r -i  "s@#{0,}?\* soft memlock ?([0-9]{1,}([TGKM]B){0,1}|unlimited)@\* soft memlock unlimited@g" /etc/security/limits.conf
sed -r -i  "s@#{0,}?\* hard memlock ?([0-9]{1,}([TGKM]B){0,1}|unlimited)@\* hard memlock unlimited@g" /etc/security/limits.conf

tmpfile="$$.tmp"
awk ' !x[$0]++{print > "'$tmpfile'"}' /etc/security/limits.conf
mv $tmpfile /etc/security/limits.conf

# ------------------------ 5. Firewall Configurations ---------------------------

if systemctl is-active firewalld --quiet; then
  systemctl stop firewalld 1>/dev/null 2>/dev/null
  systemctl disable firewalld 1>/dev/null 2>/dev/null
fi
if systemctl is-active ufw --quiet; then
  systemctl stop ufw 1>/dev/null 2>/dev/null
  systemctl disable ufw 1>/dev/null 2>/dev/null
fi

# ------------------------ 6. System Module Settings ----------------------------

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

# ------------------------ 7. IPTables and Connection Tracking -----------------

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

# ------------------------ 8. Local Host DNS Configuration ---------------------

sed -i ':a;$!{N;ba};s@# kubekey hosts BEGIN.*# kubekey hosts END@@' /etc/hosts
sed -i '/^$/N;/\n$/N;//D' /etc/hosts

cat >>/etc/hosts<<EOF
# kubekey hosts BEGIN
# kubernetes hosts
{{- range .groups.k8s_cluster | default list }}
  {{- $hostname := index $.hostvars . "hostname" -}}
  {{- $clusterName := $.kubernetes.cluster_name | default "kubekey" -}}
  {{- $dnsDomain := $.kubernetes.networking.dns_domain | default "cluster.local" -}}
  {{- if and (index $.hostvars . "internal_ipv4") (ne (index $.hostvars . "internal_ipv4") "") }}
{{ index $.hostvars . "internal_ipv4" }} {{ $hostname }} {{ printf "%s.%s" $hostname $clusterName }} {{ printf "%s.%s.%s" $hostname $clusterName $dnsDomain }}
  {{- end }}
  {{- if and (index $.hostvars . "internal_ipv6") (ne (index $.hostvars . "internal_ipv6") "") }}
{{ index $.hostvars . "internal_ipv6" }} {{ $hostname }} {{ printf "%s.%s" $hostname $clusterName }} {{ printf "%s.%s.%s" $hostname $clusterName $dnsDomain }}
  {{- end }}
{{- end }}
# etcd hosts
{{- range .groups.etcd | default list }}
  {{- if and (index $.hostvars . "internal_ipv4") (ne (index $.hostvars . "internal_ipv4") "") }}
{{ index $.hostvars . "internal_ipv4" }} {{ index $.hostvars . "hostname" }}
  {{- end }}
  {{- if and (index $.hostvars . "internal_ipv6") (ne (index $.hostvars . "internal_ipv6") "") }}
{{ index $.hostvars . "internal_ipv6" }} {{ index $.hostvars . "hostname" }}
  {{- end }}
{{- end }}
# image registry hosts
{{- range .groups.image_registry | default list }}
  {{- if and (index $.hostvars . "internal_ipv4") (ne (index $.hostvars . "internal_ipv4") "") }}
{{ index $.hostvars . "internal_ipv4" }} {{ index $.hostvars . "hostname" }}
  {{- end }}
  {{- if and (index $.hostvars . "internal_ipv6") (ne (index $.hostvars . "internal_ipv6") "") }}
{{ index $.hostvars . "internal_ipv6" }} {{ index $.hostvars . "hostname" }}
  {{- end }}
{{- end }}
# nfs hosts
{{- range .groups.nfs | default list }}
  {{- if and (index $.hostvars . "internal_ipv4") (ne (index $.hostvars . "internal_ipv4") "") }}
{{ index $.hostvars . "internal_ipv4" }} {{ index $.hostvars . "hostname" }}
  {{- end }}
  {{- if and (index $.hostvars . "internal_ipv6") (ne (index $.hostvars . "internal_ipv6") "") }}
{{ index $.hostvars . "internal_ipv4" }} {{ index $.hostvars . "hostname" }}
  {{- end }}
{{- end }}
# kubekey hosts END
EOF

sync
# echo 3 > /proc/sys/vm/drop_caches

# Make sure the iptables utility doesn't use the nftables backend.
{{- if and .internal_ipv4 (.internal_ipv4 | ne "") }}
update-alternatives --set iptables /usr/sbin/iptables-legacy >/dev/null 2>&1 || true
{{- end }}
{{- if and .internal_ipv6 (.internal_ipv6 | ne "") }}
update-alternatives --set ip6tables /usr/sbin/ip6tables-legacy >/dev/null 2>&1 || true
{{- end }}
update-alternatives --set arptables /usr/sbin/arptables-legacy >/dev/null 2>&1 || true
update-alternatives --set ebtables /usr/sbin/ebtables-legacy >/dev/null 2>&1 || true
