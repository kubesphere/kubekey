package tmpl

import (
	"github.com/lithammer/dedent"
	kubekeyapi "github.com/pixiake/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/pixiake/kubekey/pkg/util"
	"text/template"
)

var initOsScriptTmpl = template.Must(template.New("initOS").Parse(
	dedent.Dedent(`#!/bin/sh

swapoff -a
sed -i /^[^#]*swap*/s/^/\#/g /etc/fstab

echo 'net.ipv4.ip_forward = 1' >> /etc/sysctl.conf
echo 'net.bridge.bridge-nf-call-arptables = 1' >> /etc/sysctl.conf
echo 'net.bridge.bridge-nf-call-ip6tables = 1' >> /etc/sysctl.conf
echo 'net.bridge.bridge-nf-call-iptables = 1' >> /etc/sysctl.conf
echo 'net.ipv4.ip_local_reserved_ports = 30000-32767' >> /etc/sysctl.conf

sed -r -i  "s@#{0,}?net.ipv4.ip_forward ?= ?(0|1)@net.ipv4.ip_forward = 1@g" /etc/sysctl.conf
sed -r -i  "s@#{0,}?net.bridge.bridge-nf-call-arptables ?= ?(0|1)@net.bridge.bridge-nf-call-arptables = 1@g" /etc/sysctl.conf
sed -r -i  "s@#{0,}?net.bridge.bridge-nf-call-ip6tables ?= ?(0|1)@net.bridge.bridge-nf-call-ip6tables = 1@g" /etc/sysctl.conf
sed -r -i  "s@#{0,}?net.bridge.bridge-nf-call-iptables ?= ?(0|1)@net.bridge.bridge-nf-call-iptables = 1@g" /etc/sysctl.conf
sed -r -i  "s@#{0,}?net.ipv4.ip_local_reserved_ports ?= ?(0|1)@net.ipv4.ip_local_reserved_ports = 30000-32767@g" /etc/sysctl.conf

awk ' !x[$0]++{print > "/etc/sysctl.conf"}' /etc/sysctl.conf

systemctl stop firewald 1>/dev/null 2>/dev/null
systemctl disable firewald 1>/dev/null 2>/dev/null
systemctl stop ufw 1>/dev/null 2>/dev/null
systemctl disable ufw 1>/dev/null 2>/dev/null

modinfo br_netfilter > /dev/null 2>&1
if [ $? -eq 0 ]; then
   modprobe br_netfilter
   mkdir -p /etc/modules-load.d
   echo 'br_netfilter' > /etc/modules-load.d/kubekey-br_netfilter.conf
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

sed -i ':a;$!{N;ba};s@# kubekey hosts BEGIN.*# kubekey hosts END@@' /etc/hosts
sed -i '/^$/N;/\n$/N;//D' /etc/hosts

cat >>/etc/hosts<<EOF
# kubekey hosts BEGIN
{{- range .Hosts }}
{{ . }}
{{- end }}
# kubekey hosts END
EOF
    `)))

func InitOsScript(cfg *kubekeyapi.K2ClusterSpec) (string, error) {
	hostlist := cfg.GenerateHosts()
	return util.Render(initOsScriptTmpl, util.Data{
		"Hosts": hostlist,
	})
}
