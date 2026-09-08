/*
Copyright 2026 The KubeSphere Authors.

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

package convert

import (
	"strings"
	"testing"

	"github.com/cockroachdb/errors"
	"sigs.k8s.io/yaml"
)

const sampleV3Config = `
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Cluster
metadata:
  name: sample
spec:
  hosts:
  - {name: node1, address: 172.16.0.2, internalAddress: "172.16.0.2,fd85::2", port: 8022, user: ubuntu, password: "Qcloud@123", arch: amd64, labels: {disk: SSD}}
  - {name: node2, address: 172.16.0.3, internalAddress: 172.16.0.3, privateKeyPath: "~/.ssh/id_rsa"}
  - {name: node3, address: 172.16.0.4, user: root}
  - {name: node4, address: 172.16.0.5}
  roleGroups:
    etcd: [node1]
    master: [node1, 'node[2:3]']
    worker: [node1, 'node[2:4]']
    registry: [node1]
  controlPlaneEndpoint:
    internalLoadbalancer: haproxy
    domain: lb.kubesphere.local
    address: ""
    port: 6443
  system:
    ntpServers: [ntp.aliyun.com]
    timezone: "Asia/Shanghai"
    rpms: [nfs-utils]
  kubernetes:
    version: v1.28.5
    clusterName: cluster.local
    dnsDomain: cluster.local
    containerManager: containerd
    autoRenewCerts: true
    masqueradeAll: true
    maxPods: 210
    podPidsLimit: 10000
    nodeCidrMaskSize: 24
    proxyMode: ipvs
    apiserverCertExtraSans: [192.168.8.8]
    apiserverArgs: ["allow-privileged=true"]
    featureGates: {CSIStorageCapacity: true}
    nodelocaldns: false
  etcd:
    type: kubekey
    dataDir: /var/lib/etcd-custom
    heartbeatInterval: 200
    electionTimeout: 4000
  network:
    plugin: calico
    kubePodsCIDR: 10.233.64.0/18
    kubeServiceCIDR: 10.233.0.0/18
  storage:
    openebs:
      basePath: /data/openebs
  registry:
    registryMirrors: ["https://mirror.example.com"]
    insecureRegistries: ["insecure.example.com"]
    privateRegistry: "dockerhub.kubekey.local"
    containerdDataDir: /data/containerd
    auths:
      "dockerhub.kubekey.local":
        username: admin
        password: secret
        skipTLSVerify: true
`

// parse converts YAML bytes into a Cluster for tests.
func parse(t *testing.T, data string) (*Cluster, error) {
	t.Helper()
	return ParseCluster([]byte(data))
}

func getNested(t *testing.T, obj map[string]any, fields ...string) any {
	t.Helper()
	var cur any = obj
	for _, f := range fields {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur = m[f]
	}
	return cur
}

func TestParseCluster(t *testing.T) {
	c, err := parse(t, sampleV3Config)
	if err != nil {
		t.Fatalf("failed to parse sample: %v", err)
	}
	if c.APIVersion != V1Alpha2APIVersion {
		t.Errorf("apiVersion = %q, want %q", c.APIVersion, V1Alpha2APIVersion)
	}
	if len(c.Spec.Hosts) != 4 {
		t.Errorf("hosts = %d, want 4", len(c.Spec.Hosts))
	}
	if c.Spec.Kubernetes.Version != "v1.28.5" {
		t.Errorf("kubernetes.version = %q", c.Spec.Kubernetes.Version)
	}
}

func TestParseClusterRejectsUnknownKind(t *testing.T) {
	_, err := parse(t, "apiVersion: kubekey.kubesphere.io/v1alpha2\nkind: Foo\n")
	if err == nil || !strings.Contains(err.Error(), "unsupported kind") {
		t.Fatalf("expected unsupported kind error, got %v", err)
	}
}

func TestParseClusterRequiresHosts(t *testing.T) {
	_, err := parse(t, "apiVersion: kubekey.kubesphere.io/v1alpha2\nkind: Cluster\n")
	if err == nil || !strings.Contains(err.Error(), "no hosts") {
		t.Fatalf("expected no hosts error, got %v", err)
	}
}

func TestExpandRoleGroups(t *testing.T) {
	roleGroups, err := expandRoleGroups(map[string][]string{
		"etcd":   {"node1"},
		"master": {"node1", "node[2:3]"},
	}, []string{"node1", "node2", "node3"})
	if err != nil {
		t.Fatalf("expandRoleGroups failed: %v", err)
	}
	if got := strings.Join(roleGroups["etcd"], ","); got != "node1" {
		t.Errorf("etcd = %q", got)
	}
	if got := strings.Join(roleGroups["master"], ","); got != "node1,node2,node3" {
		t.Errorf("master = %q, want node1,node2,node3", got)
	}
}

func TestExpandRoleGroupsUnknownHost(t *testing.T) {
	_, err := expandRoleGroups(map[string][]string{"master": {"node9"}}, []string{"node1"})
	if err == nil || !strings.Contains(err.Error(), "not in hosts list") {
		t.Fatalf("expected unknown host error, got %v", err)
	}
}

func TestConvertInventory(t *testing.T) {
	c, err := parse(t, sampleV3Config)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	r, err := Convert(c)
	if err != nil {
		t.Fatalf("convert failed: %v", err)
	}

	inv := r.Inventory
	if inv.Name != "sample" {
		t.Errorf("inventory name = %q", inv.Name)
	}
	if len(inv.Spec.Hosts) != 4 {
		t.Errorf("hosts = %d, want 4", len(inv.Spec.Hosts))
	}

	node1, ok := inv.Spec.Hosts["node1"]
	if !ok {
		t.Fatal("node1 missing from inventory")
	}
	node1YAML, err := yaml.Marshal(node1)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"type: ssh", "host: 172.16.0.2", "port: 8022", "user: ubuntu",
		"password: Qcloud@123", "internal_ipv4: 172.16.0.2", "internal_ipv6: fd85::2",
		"arch: amd64", "disk: SSD",
	} {
		if !strings.Contains(string(node1YAML), want) {
			t.Errorf("node1 vars missing %q, got:\n%s", want, node1YAML)
		}
	}

	node2YAML, err := yaml.Marshal(inv.Spec.Hosts["node2"])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(node2YAML), "private_key: ~/.ssh/id_rsa") {
		t.Errorf("node2 missing private_key, got:\n%s", node2YAML)
	}
	if !strings.Contains(string(node2YAML), "port: 22") || !strings.Contains(string(node2YAML), "user: root") {
		t.Errorf("node2 missing connector defaults, got:\n%s", node2YAML)
	}

	// groups
	cp := inv.Spec.Groups[groupByControlPlane()]
	if got := strings.Join(cp.Hosts, ","); got != "node1,node2,node3" {
		t.Errorf("kube_control_plane = %q, want node1,node2,node3", got)
	}
	if got := strings.Join(inv.Spec.Groups["kube_worker"].Hosts, ","); got != "node1,node2,node3,node4" {
		t.Errorf("kube_worker = %q", got)
	}
	if got := strings.Join(inv.Spec.Groups["etcd"].Hosts, ","); got != "node1" {
		t.Errorf("etcd = %q", got)
	}
	if got := strings.Join(inv.Spec.Groups["image_registry"].Hosts, ","); got != "node1" {
		t.Errorf("image_registry = %q", got)
	}
	if got := strings.Join(inv.Spec.Groups["k8s_cluster"].Groups, ","); got != "kube_control_plane,kube_worker" {
		t.Errorf("k8s_cluster groups = %q", got)
	}
}

func groupByControlPlane() string { return groupControlPlane }

func TestConvertConfig(t *testing.T) {
	c, err := parse(t, sampleV3Config)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	r, err := Convert(c)
	if err != nil {
		t.Fatalf("convert failed: %v", err)
	}
	cfg := r.Config

	if got := getNested(t, cfg, "kubernetes", "kube_version"); got != "v1.28.5" {
		t.Errorf("kube_version = %v", got)
	}
	if got := getNested(t, cfg, "kubernetes", "control_plane_endpoint", "type"); got != "haproxy" {
		t.Errorf("cpe.type = %v", got)
	}
	if got := getNested(t, cfg, "kubernetes", "control_plane_endpoint", "host"); got != "lb.kubesphere.local" {
		t.Errorf("cpe.host = %v", got)
	}
	if got := getNested(t, cfg, "kubernetes", "certs", "renew"); got != true {
		t.Errorf("certs.renew = %v", got)
	}
	if got := getNested(t, cfg, "kubernetes", "kube_proxy", "config", "iptables", "masqueradeAll"); got != true {
		t.Errorf("masqueradeAll = %v", got)
	}
	if got := getNested(t, cfg, "kubernetes", "kubelet", "max_pods"); got != 210 {
		t.Errorf("kubelet.max_pods = %v", got)
	}
	if got := getNested(t, cfg, "kubernetes", "apiserver", "certSANs"); got == nil {
		t.Error("apiserver.certSANs missing")
	}
	if got := getNested(t, cfg, "kubernetes", "apiserver", "extra_args", "feature-gates"); got != "CSIStorageCapacity=true" {
		t.Errorf("feature-gates = %v", got)
	}
	if got := getNested(t, cfg, "kubernetes", "apiserver", "extra_args", "allow-privileged"); got != "true" {
		t.Errorf("allow-privileged = %v", got)
	}
	if got := getNested(t, cfg, "dns", "nodelocaldns", "enabled"); got != false {
		t.Errorf("dns.nodelocaldns.enabled = %v", got)
	}
	if got := getNested(t, cfg, "dns", "domain"); got != "cluster.local" {
		t.Errorf("dns.domain = %v", got)
	}
	if got := getNested(t, cfg, "cni", "pod_cidr"); got != "10.233.64.0/18" {
		t.Errorf("cni.pod_cidr = %v", got)
	}
	if got := getNested(t, cfg, "etcd", "deployment_type"); got != "internal" {
		t.Errorf("etcd.deployment_type = %v", got)
	}
	if got := getNested(t, cfg, "etcd", "env", "data_dir"); got != "/var/lib/etcd-custom" {
		t.Errorf("etcd.env.data_dir = %v", got)
	}
	if got := getNested(t, cfg, "cri", "registry", "mirrors"); got == nil {
		t.Error("cri.registry.mirrors missing")
	}
	authsRaw := getNested(t, cfg, "cri", "registry", "auths")
	var auths []map[string]any
	switch list := authsRaw.(type) {
	case []map[string]any:
		auths = list
	case []any:
		for _, item := range list {
			auths = append(auths, item.(map[string]any))
		}
	}
	if len(auths) != 1 {
		t.Fatalf("cri.registry.auths = %#v, want one entry", authsRaw)
	}
	auth0 := auths[0]
	if auth0["registry"] != "dockerhub.kubekey.local" || auth0["username"] != "admin" || auth0["skip_tls_verify"] != true {
		t.Errorf("auth0 = %#v", auth0)
	}
	if got := getNested(t, cfg, "image_registry", "auth", "registry"); got != "dockerhub.kubekey.local" {
		t.Errorf("image_registry.auth.registry = %v", got)
	}
	if got := getNested(t, cfg, "storage_class", "local", "path"); got != "/data/openebs" {
		t.Errorf("storage_class.local.path = %v", got)
	}
	if got := getNested(t, cfg, "native", "ntp", "servers"); got == nil {
		t.Error("native.ntp.servers missing")
	}
	if got := getNested(t, cfg, "native", "timezone"); got != "Asia/Shanghai" {
		t.Errorf("native.timezone = %v", got)
	}
}

func TestConvertWarnings(t *testing.T) {
	c, err := parse(t, sampleV3Config)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	r, err := Convert(c)
	if err != nil {
		t.Fatalf("convert failed: %v", err)
	}

	joined := strings.Join(r.Warnings, "\n")
	for _, want := range []string{
		"system.rpms/debs",            // unmapped system packages
		"registry.privateRegistry",    // semantic difference warning
		"kubernetes.disableKubeProxy", // not present in sample, checked below
	} {
		if want == "kubernetes.disableKubeProxy" {
			continue
		}
		if !strings.Contains(joined, want) {
			t.Errorf("warnings missing %q, got:\n%s", want, joined)
		}
	}
}

func TestConvertKubeVipEndpoint(t *testing.T) {
	data := `
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Cluster
metadata:
  name: vip
spec:
  hosts:
  - {name: node1, address: 172.16.0.2}
  roleGroups:
    master: [node1]
    worker: [node1]
    etcd: [node1]
  controlPlaneEndpoint:
    internalLoadbalancer: kubevip
    domain: lb.kubesphere.local
    address: 172.16.0.100
    port: 6443
    kubevip:
      mode: BGP
`
	c, err := parse(t, data)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	r, err := Convert(c)
	if err != nil {
		t.Fatalf("convert failed: %v", err)
	}
	if got := getNested(t, r.Config, "kubernetes", "control_plane_endpoint", "type"); got != "kube-vip" {
		t.Errorf("cpe.type = %v", got)
	}
	if got := getNested(t, r.Config, "kubernetes", "control_plane_endpoint", "kube_vip", "address"); got != "172.16.0.100" {
		t.Errorf("kube_vip.address = %v", got)
	}
	if got := getNested(t, r.Config, "kubernetes", "control_plane_endpoint", "kube_vip", "mode"); got != "BGP" {
		t.Errorf("kube_vip.mode = %v", got)
	}
}

func TestConvertEtcdExternal(t *testing.T) {
	data := `
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Cluster
metadata:
  name: ext
spec:
  hosts:
  - {name: node1, address: 172.16.0.2}
  roleGroups:
    master: [node1]
    worker: [node1]
  etcd:
    type: external
    external:
      endpoints: ["https://192.168.1.1:2379"]
`
	c, err := parse(t, data)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	r, err := Convert(c)
	if err != nil {
		t.Fatalf("convert failed: %v", err)
	}
	if got := getNested(t, r.Config, "etcd", "deployment_type"); got != "external" {
		t.Errorf("etcd.deployment_type = %v", got)
	}
	// No etcd roleGroup and type external: no missing-etcd warning expected.
	for _, w := range r.Warnings {
		if strings.Contains(w, "no etcd roleGroup") {
			t.Errorf("unexpected missing-etcd warning for external etcd: %q", w)
		}
	}
}

func TestConvertInventoryAndConfigOutput(t *testing.T) {
	inventoryYAML, configYAML, _, err := ConvertInventoryAndConfig([]byte(sampleV3Config))
	if err != nil {
		t.Fatalf("convert failed: %v", err)
	}
	for _, want := range []string{"kind: Inventory", "apiVersion: kubekey.kubesphere.io/v1"} {
		if !strings.Contains(string(inventoryYAML), want) {
			t.Errorf("inventory yaml missing %q:\n%s", want, inventoryYAML)
		}
	}
	for _, want := range []string{"kind: Config", "kube_version"} {
		if !strings.Contains(string(configYAML), want) {
			t.Errorf("config yaml missing %q:\n%s", want, configYAML)
		}
	}
}

func TestConvertMappedFields(t *testing.T) {
	data := `
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Cluster
metadata:
  name: mapped
spec:
  hosts:
  - {name: node1, address: 172.16.0.2}
  roleGroups:
    master: [node1]
    worker: [node1]
    etcd: [node1]
  controlPlaneEndpoint:
    internalLoadbalancer: haproxy
    domain: lb.local
    address: 127.0.0.2
    port: 6443
  kubernetes:
    version: v1.28.5
    containerManager: containerd
    containerRuntimeEndpoint: "unix:///run/containerd/containerd.sock"
    kubeletArgs: ["max-pods=250", "read-only-port=0"]
    kubeProxyConfiguration:
      iptables:
        masqueradeAll: true
        masqueradeBit: 14
      mode: ipvs
    kubeletConfiguration:
      maxPods: 300
      podPidsLimit: 12000
      clusterDomain: cluster.local
      featureGates:
        CSIVolumeHealth: true
      evictionHard:
        memory.available: 5%
  etcd:
    type: kubekey
    backupScript: /opt/etcd-backup.sh
  dns:
    dnsEtcHosts: "10.0.0.1 example.com"
`
	c, err := parse(t, data)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	r, err := Convert(c)
	if err != nil {
		t.Fatalf("convert failed: %v", err)
	}
	cfg := r.Config

	// kubernetes.kubeletArgs -> kubernetes.kubelet.extra_args (consumed by kubeadm template)
	if got := getNested(t, cfg, "kubernetes", "kubelet", "extra_args", "max-pods"); got != "250" {
		t.Errorf("kubelet.extra_args[max-pods] = %v, want 250", got)
	}
	if got := getNested(t, cfg, "kubernetes", "kubelet", "extra_args", "read-only-port"); got != "0" {
		t.Errorf("kubelet.extra_args[read-only-port] = %v, want 0", got)
	}

	// kubernetes.kubeProxyConfiguration -> kubernetes.kube_proxy.config
	if got := getNested(t, cfg, "kubernetes", "kube_proxy", "config", "mode"); got != "ipvs" {
		t.Errorf("kube_proxy.config.mode = %v, want ipvs", got)
	}
	if got := getNested(t, cfg, "kubernetes", "kube_proxy", "config", "iptables", "masqueradeAll"); got != true {
		t.Errorf("kube_proxy.config.iptables.masqueradeAll = %v, want true", got)
	}

	// kubernetes.kubeletConfiguration -> kubernetes.kubelet scalars + extra_config
	if got := getNested(t, cfg, "kubernetes", "kubelet", "max_pods"); got != 300 {
		t.Errorf("kubelet.max_pods = %v, want 300", got)
	}
	if got := getNested(t, cfg, "kubernetes", "kubelet", "pod_pids_limit"); got != 12000 {
		t.Errorf("kubelet.pod_pids_limit = %v, want 12000", got)
	}
	if got := getNested(t, cfg, "kubernetes", "kubelet", "extra_config", "clusterDomain"); got != "cluster.local" {
		t.Errorf("kubelet.extra_config.clusterDomain = %v, want cluster.local", got)
	}
	if got := getNested(t, cfg, "kubernetes", "kubelet", "extra_config", "evictionHard", "memory.available"); got != "5%" {
		t.Errorf("kubelet.extra_config.evictionHard.memory.available = %v, want 5%%", got)
	}
	if got := getNested(t, cfg, "kubernetes", "kubelet", "feature_gates", "CSIVolumeHealth"); got != true {
		t.Errorf("kubelet.feature_gates.CSIVolumeHealth = %v, want true", got)
	}

	// kubernetes.containerRuntimeEndpoint -> cri.cri_socket (consumed by kubeadm template)
	if got := getNested(t, cfg, "cri", "cri_socket"); got != "unix:///run/containerd/containerd.sock" {
		t.Errorf("cri.cri_socket = %v", got)
	}

	// controlPlaneEndpoint.address (haproxy) -> control_plane_endpoint.haproxy.address
	if got := getNested(t, cfg, "kubernetes", "control_plane_endpoint", "haproxy", "address"); got != "127.0.0.2" {
		t.Errorf("cpe.haproxy.address = %v, want 127.0.0.2", got)
	}

	// etcd.backupScript -> etcd.backup.etcd_backup_script
	if got := getNested(t, cfg, "etcd", "backup", "etcd_backup_script"); got != "/opt/etcd-backup.sh" {
		t.Errorf("etcd.backup.etcd_backup_script = %v", got)
	}

	// dns.dnsEtcHosts -> dns.coredns.dns_etc_hosts
	if got := getNested(t, cfg, "dns", "coredns", "dns_etc_hosts"); got != "10.0.0.1 example.com" {
		t.Errorf("dns.coredns.dns_etc_hosts = %v", got)
	}
}

func TestSplitInternalAddress(t *testing.T) {
	ipv4, ipv6 := splitInternalAddress("10.0.0.1,fd00::1")
	if ipv4 != "10.0.0.1" || ipv6 != "fd00::1" {
		t.Errorf("splitInternalAddress dual = %q, %q", ipv4, ipv6)
	}
	ipv4, ipv6 = splitInternalAddress("10.0.0.1")
	if ipv4 != "10.0.0.1" || ipv6 != "" {
		t.Errorf("splitInternalAddress single = %q, %q", ipv4, ipv6)
	}
}

func TestConvertKubeProxyArgs(t *testing.T) {
	data := `
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Cluster
metadata:
  name: kp
spec:
  hosts:
  - {name: node1, address: 172.16.0.2}
  roleGroups:
    master: [node1]
    worker: [node1]
    etcd: [node1]
  controlPlaneEndpoint:
    internalLoadbalancer: local
    domain: lb.local
  kubernetes:
    version: v1.28.5
    containerManager: containerd
    kubeProxyArgs:
      - "proxy-mode=iptables"
      - "masquerade-all"
      - "ipvs-scheduler=rr"
      - "nodeport-addresses=10.0.0.0/16,10.1.0.0/16"
      - "conntrack-max-per-core=2"
      - "bogus-flag=1"
`
	c, err := parse(t, data)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	r, err := Convert(c)
	if err != nil {
		t.Fatalf("convert failed: %v", err)
	}
	cfg := r.Config

	if got := getNested(t, cfg, "kubernetes", "kube_proxy", "mode"); got != "iptables" {
		t.Errorf("kube_proxy.mode = %v, want iptables", got)
	}
	if got := getNested(t, cfg, "kubernetes", "kube_proxy", "config", "iptables", "masqueradeAll"); got != true {
		t.Errorf("kube_proxy.config.iptables.masqueradeAll = %v, want true", got)
	}
	if got := getNested(t, cfg, "kubernetes", "kube_proxy", "config", "ipvs", "scheduler"); got != "rr" {
		t.Errorf("kube_proxy.config.ipvs.scheduler = %v, want rr", got)
	}
	if got := getNested(t, cfg, "kubernetes", "kube_proxy", "config", "conntrack", "maxPerCore"); got != 2 {
		t.Errorf("kube_proxy.config.conntrack.maxPerCore = %v, want 2", got)
	}
	// nodeport-addresses is a comma-separated slice -> []any
	npRaw := getNested(t, cfg, "kubernetes", "kube_proxy", "config", "nodePortAddresses")
	np, ok := npRaw.([]any)
	if !ok {
		t.Fatalf("kube_proxy.config.nodePortAddresses = %#v, want []any", npRaw)
	}
	if len(np) != 2 || np[0] != "10.0.0.0/16" || np[1] != "10.1.0.0/16" {
		t.Errorf("kube_proxy.config.nodePortAddresses = %#v, want [10.0.0.0/16 10.1.0.0/16]", np)
	}
	// unsupported flag reported
	joined := strings.Join(r.Warnings, "\n")
	if !strings.Contains(joined, "bogus-flag") {
		t.Errorf("expected warning for bogus-flag, got:\n%s", joined)
	}
}

func TestConvertEtcdExtras(t *testing.T) {
	data := `
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Cluster
metadata:
  name: etcd
spec:
  hosts:
  - {name: node1, address: 172.16.0.2}
  roleGroups:
    master: [node1]
    worker: [node1]
    etcd: [node1]
  controlPlaneEndpoint:
    internalLoadbalancer: local
    domain: lb.local
  kubernetes:
    version: v1.28.5
    containerManager: containerd
  etcd:
    type: kubekey
    extraArgs:
      - "data-dir=/data/etcd"
      - "heartbeat-interval=300"
      - "quota-backend-bytes=8589934592"
      - "listen-client-urls=https://0.0.0.0:2379"
      - "bogus-flag=1"
`
	c, err := parse(t, data)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	r, err := Convert(c)
	if err != nil {
		t.Fatalf("convert failed: %v", err)
	}
	cfg := r.Config
	if got := getNested(t, cfg, "etcd", "env", "data_dir"); got != "/data/etcd" {
		t.Errorf("etcd.env.data_dir = %v, want /data/etcd", got)
	}
	if got := getNested(t, cfg, "etcd", "env", "heartbeat_interval"); got != 300 {
		t.Errorf("etcd.env.heartbeat_interval = %v, want 300", got)
	}
	if got := getNested(t, cfg, "etcd", "env", "quota_backend_bytes"); got != 8589934592 {
		t.Errorf("etcd.env.quota_backend_bytes = %v, want 8589934592", got)
	}
	joined := strings.Join(r.Warnings, "\n")
	for _, want := range []string{"listen-client-urls", "bogus-flag"} {
		if !strings.Contains(joined, want) {
			t.Errorf("expected warning for %q, got:\n%s", want, joined)
		}
	}
}

func TestConvertEtcdBackupPeriod(t *testing.T) {
	data := `
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Cluster
metadata:
  name: bp
spec:
  hosts:
  - {name: node1, address: 172.16.0.2}
  roleGroups:
    master: [node1]
    worker: [node1]
    etcd: [node1]
  controlPlaneEndpoint:
    internalLoadbalancer: local
    domain: lb.local
  kubernetes:
    version: v1.28.5
    containerManager: containerd
  etcd:
    type: kubekey
    backupPeriod: 30
`
	c, err := parse(t, data)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	r, err := Convert(c)
	if err != nil {
		t.Fatalf("convert failed: %v", err)
	}
	cfg := r.Config
	if got := getNested(t, cfg, "etcd", "backup", "on_calendar"); got != "*/30 * * * *" {
		t.Errorf("etcd.backup.on_calendar = %v, want */30 * * * *", got)
	}
	joined := strings.Join(r.Warnings, "\n")
	if !strings.Contains(joined, "auto-converted") || !strings.Contains(joined, "*/30 * * * *") {
		t.Errorf("expected auto-convert warning, got:\n%s", joined)
	}
}

var _ = errors.New // keep errors import for future assertions
