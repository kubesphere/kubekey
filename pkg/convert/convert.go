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

// Package convert converts a KubeKey v3 (v1alpha2) cluster configuration into
// the KubeKey v4 inventory.yaml and config.yaml files.
package convert

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/cockroachdb/errors"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// v4 inventory group names.
const (
	groupCluster      = "k8s_cluster"
	groupControlPlane = "kube_control_plane"
	groupWorker       = "kube_worker"
	groupEtcd         = "etcd"
	groupRegistry     = "image_registry"
)

// v3 etcd deployment types.
const (
	v3EtcdKubeKey  = "kubekey"
	v3EtcdKubeadm  = "kubeadm"
	v3EtcdExternal = "external"
)

// Result holds the conversion output.
type Result struct {
	// Inventory is the converted v4 Inventory object.
	Inventory *kkcorev1.Inventory
	// Config is the converted v4 Config spec as an unstructured map.
	Config map[string]any
	// Warnings lists fields that could not be converted automatically and
	// require manual review.
	Warnings []string
}

func (r *Result) warnf(format string, args ...any) {
	r.Warnings = append(r.Warnings, fmt.Sprintf(format, args...))
}

// Convert transforms a parsed v3 cluster configuration into v4 Inventory and
// Config objects. Fields that have no v4 equivalent or whose semantics differ
// are reported in Result.Warnings instead of being silently dropped.
func Convert(cluster *Cluster) (*Result, error) {
	r := &Result{}

	roleGroups, err := expandRoleGroups(cluster.Spec.RoleGroups, hostNames(cluster.Spec.Hosts))
	if err != nil {
		return nil, err
	}

	r.convertInventory(cluster, roleGroups)
	r.convertConfig(cluster, roleGroups)

	return r, nil
}

func hostNames(hosts []HostCfg) []string {
	names := make([]string, 0, len(hosts))
	for _, h := range hosts {
		names = append(names, h.Name)
	}
	return names
}

// convertInventory builds the v4 Inventory from v3 hosts and roleGroups.
func (r *Result) convertInventory(cluster *Cluster, roleGroups map[string][]string) {
	name := cluster.Metadata.Name
	if name == "" {
		name = "default"
	}

	hosts := make(map[string]runtime.RawExtension, len(cluster.Spec.Hosts))
	for _, h := range cluster.Spec.Hosts {
		connector := map[string]any{
			"type": "ssh",
			"host": h.Address,
			"port": h.Port,
			"user": h.User,
		}
		if h.Port == 0 {
			connector["port"] = 22
		}
		if h.User == "" {
			connector["user"] = "root"
		}
		if h.Password != "" {
			connector["password"] = h.Password
		}
		if h.PrivateKeyPath != "" {
			connector["private_key"] = h.PrivateKeyPath
		}
		if h.PrivateKey != "" {
			connector["private_key_content"] = h.PrivateKey
		}

		vars := map[string]any{"connector": connector}
		if h.InternalAddress != "" {
			ipv4, ipv6 := splitInternalAddress(h.InternalAddress)
			if ipv4 != "" {
				vars["internal_ipv4"] = ipv4
			}
			if ipv6 != "" {
				vars["internal_ipv6"] = ipv6
			}
		}
		if h.Arch != "" {
			vars["arch"] = h.Arch
		}
		if len(h.Labels) > 0 {
			vars["labels"] = h.Labels
		}

		raw, err := toRawExtension(vars)
		if err != nil {
			r.warnf("failed to serialize vars for host %q: %v", h.Name, err)
			continue
		}
		hosts[h.Name] = raw
	}

	groups := map[string]kkcorev1.InventoryGroup{}
	appendGroup := func(group string, names []string) {
		if len(names) == 0 {
			return
		}
		groups[group] = kkcorev1.InventoryGroup{Hosts: names}
	}
	appendGroup(groupControlPlane, mergeUnique(roleGroups[v3RoleMaster], roleGroups[v3RoleControlPlane], roleGroups[v3RoleControlPlane2]))
	appendGroup(groupWorker, roleGroups[v3RoleWorker])
	appendGroup(groupEtcd, roleGroups[v3RoleEtcd])
	appendGroup(groupRegistry, mergeUnique(roleGroups[v3RoleRegistry]))

	for role := range roleGroups {
		switch role {
		case v3RoleMaster, v3RoleControlPlane, v3RoleControlPlane2, v3RoleWorker, v3RoleEtcd, v3RoleRegistry:
		default:
			r.warnf("roleGroups/%s has no v4 inventory group equivalent, its hosts are not converted", role)
		}
	}
	if _, ok := groups[groupEtcd]; !ok && cluster.Spec.Etcd.Type != v3EtcdExternal {
		r.warnf("no etcd roleGroup defined; etcd hosts must be added to the %q group manually", groupEtcd)
	}

	// k8s_cluster aggregates control plane and worker groups.
	clusterGroups := []string{groupControlPlane}
	if _, ok := groups[groupWorker]; ok {
		clusterGroups = append(clusterGroups, groupWorker)
	}
	groups[groupCluster] = kkcorev1.InventoryGroup{Groups: clusterGroups}

	r.Inventory = &kkcorev1.Inventory{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kubekey.kubesphere.io/v1",
			Kind:       "Inventory",
		},
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: kkcorev1.InventorySpec{
			Hosts:  hosts,
			Groups: groups,
		},
	}
}

// convertConfig builds the v4 Config spec from v3 fields. Only fields that are
// explicitly set in the v3 configuration are written; the rest fall back to v4
// defaults at install time.
func (r *Result) convertConfig(cluster *Cluster, roleGroups map[string][]string) {
	spec := map[string]any{}
	spec["kubernetes"] = r.convertKubernetes(cluster)
	if cni := r.convertCNI(cluster); cni != nil {
		spec["cni"] = cni
	}
	if dns := r.convertDNS(cluster); dns != nil {
		spec["dns"] = dns
	}
	if etcd := r.convertEtcd(cluster); etcd != nil {
		spec["etcd"] = etcd
	}
	if cri := r.convertCRI(cluster); cri != nil {
		spec["cri"] = cri
	}
	if registry := r.convertImageRegistry(cluster); registry != nil {
		spec["image_registry"] = registry
	}
	if storage := r.convertStorage(cluster); storage != nil {
		spec["storage_class"] = storage
	}
	if native := r.convertNative(cluster); native != nil {
		spec["native"] = native
	}
	if audit := cluster.Spec.Kubernetes.Audit; len(audit) > 0 {
		if enabled, ok := audit["enabled"].(bool); ok && enabled {
			spec["audit"] = true
		}
	}
	r.Config = spec
}

func (r *Result) convertKubernetes(cluster *Cluster) map[string]any {
	k3 := cluster.Spec.Kubernetes
	k := map[string]any{}

	if k3.Version != "" {
		k["kube_version"] = k3.Version
	}
	if k3.ClusterName != "" {
		k["cluster_name"] = k3.ClusterName
	}

	// control_plane_endpoint
	cpe := map[string]any{}
	if cluster.Spec.ControlPlaneEndpoint.Domain != "" {
		cpe["host"] = cluster.Spec.ControlPlaneEndpoint.Domain
	}
	if cluster.Spec.ControlPlaneEndpoint.Port != 0 {
		cpe["port"] = cluster.Spec.ControlPlaneEndpoint.Port
	}
	switch cluster.Spec.ControlPlaneEndpoint.InternalLoadbalancer {
	case "", "none":
		cpe["type"] = "local"
		if cluster.Spec.ControlPlaneEndpoint.Address != "" {
			cpe["local"] = map[string]any{"address": cluster.Spec.ControlPlaneEndpoint.Address}
		}
	case "kubevip":
		cpe["type"] = "kube-vip"
		kv := map[string]any{}
		if cluster.Spec.ControlPlaneEndpoint.Address != "" {
			kv["address"] = cluster.Spec.ControlPlaneEndpoint.Address
		}
		if cluster.Spec.ControlPlaneEndpoint.KubeVip.Mode != "" {
			kv["mode"] = cluster.Spec.ControlPlaneEndpoint.KubeVip.Mode
		}
		if len(kv) > 0 {
			cpe["kube_vip"] = kv
		}
	case "haproxy":
		cpe["type"] = "haproxy"
		if cluster.Spec.ControlPlaneEndpoint.Address != "" {
			r.warnf("controlPlaneEndpoint.address %q has no direct equivalent when type is haproxy (v4 haproxy listens on 127.0.0.1); review kubernetes.control_plane_endpoint", cluster.Spec.ControlPlaneEndpoint.Address)
		}
	default:
		r.warnf("unsupported controlPlaneEndpoint.internalLoadbalancer %q, converted to type local", cluster.Spec.ControlPlaneEndpoint.InternalLoadbalancer)
		cpe["type"] = "local"
	}
	k["control_plane_endpoint"] = cpe

	if k3.AutoRenewCerts != nil {
		setNested(k, *k3.AutoRenewCerts, "certs", "renew")
	}

	// apiserver
	api := map[string]any{}
	if len(k3.ApiserverCertExtraSans) > 0 {
		api["certSANs"] = k3.ApiserverCertExtraSans
	}
	if args := argsToMap(k3.ApiServerArgs); len(args) > 0 {
		api["extra_args"] = args
	}
	if fg := featureGatesArg(k3.FeatureGates); fg != "" {
		extra, _ := api["extra_args"].(map[string]any)
		if extra == nil {
			extra = map[string]any{}
		}
		if existing, ok := extra["feature-gates"].(string); ok && existing != "" {
			extra["feature-gates"] = existing + "," + fg
		} else {
			extra["feature-gates"] = fg
		}
		api["extra_args"] = extra
	}
	if len(api) > 0 {
		k["apiserver"] = api
	}

	if args := argsToMap(k3.ControllerManagerArgs); len(args) > 0 {
		setNested(k, args, "controller_manager", "extra_args")
	}
	if args := argsToMap(k3.SchedulerArgs); len(args) > 0 {
		setNested(k, args, "scheduler", "extra_args")
	}

	// kube_proxy
	kp := map[string]any{}
	if k3.ProxyMode != "" {
		kp["mode"] = k3.ProxyMode
	}
	if k3.MasqueradeAll {
		setNested(kp, true, "config", "iptables", "masqueradeAll")
	}
	if k3.DisableKubeProxy {
		setNested(kp, false, "manage", "enabled")
		r.warnf("kubernetes.disableKubeProxy is mapped to kube_proxy.manage.enabled=false; verify kube-proxy deployment ownership")
	}
	if len(kp) > 0 {
		k["kube_proxy"] = kp
	}
	if len(k3.KubeProxyArgs) > 0 {
		r.warnf("kubernetes.kubeProxyArgs has no v4 equivalent and is dropped")
	}
	if len(k3.KubeProxyConfiguration) > 0 {
		r.warnf("kubernetes.kubeProxyConfiguration has no direct v4 equivalent; migrate entries to kubernetes.kube_proxy.config manually")
	}

	// kubelet
	kubelet := map[string]any{}
	if k3.MaxPods != 0 {
		kubelet["max_pods"] = k3.MaxPods
	}
	if k3.PodPidsLimit != 0 {
		kubelet["pod_pids_limit"] = k3.PodPidsLimit
	}
	if len(kubelet) > 0 {
		k["kubelet"] = kubelet
	}
	if len(k3.KubeletArgs) > 0 {
		r.warnf("kubernetes.kubeletArgs has no v4 equivalent and is dropped")
	}
	if len(k3.KubeletConfiguration) > 0 {
		r.warnf("kubernetes.kubeletConfiguration has no direct v4 equivalent; migrate entries to kubernetes.kubelet manually")
	}

	if k3.ContainerRuntimeEndpoint != "" {
		r.warnf("kubernetes.containerRuntimeEndpoint has no v4 equivalent and is dropped")
	}
	if len(k3.NodeFeatureDiscovery) > 0 {
		r.warnf("kubernetes.nodeFeatureDiscovery has no v4 equivalent and is dropped")
	}
	if len(k3.Kata) > 0 {
		r.warnf("kubernetes.kata has no v4 equivalent and is dropped")
	}
	if k3.NvidiaRuntime != nil && *k3.NvidiaRuntime {
		r.warnf("kubernetes.nvidiaRuntime has no v4 equivalent and is dropped")
	}
	if k3.Type != "" {
		r.warnf("kubernetes.type %q is ignored by v4", k3.Type)
	}

	return k
}

func (r *Result) convertCNI(cluster *Cluster) map[string]any {
	n := cluster.Spec.Network
	cni := map[string]any{}

	switch n.Plugin {
	case "":
		return nil
	case "calico", "cilium", "flannel", "kubeovn", "hybridnet":
		cni["type"] = n.Plugin
	default:
		r.warnf("network.plugin %q has no v4 cni.type equivalent, use \"other\" and configure it manually", n.Plugin)
		cni["type"] = "other"
	}
	if n.KubePodsCIDR != "" {
		cni["pod_cidr"] = n.KubePodsCIDR
	}
	if n.KubeServiceCIDR != "" {
		cni["service_cidr"] = n.KubeServiceCIDR
	}
	if cluster.Spec.Kubernetes.NodeCidrMaskSize != 0 {
		cni["ipv4_mask_size"] = cluster.Spec.Kubernetes.NodeCidrMaskSize
	}
	if cluster.Spec.Kubernetes.NodeCidrMaskSizeIPv6 != 0 {
		cni["ipv6_mask_size"] = cluster.Spec.Kubernetes.NodeCidrMaskSizeIPv6
	}

	// Per-plugin details are largely not exposed by v4 defaults.
	if !calicoEmpty(n.Calico) {
		c := n.Calico
		if c.IPIPMode != "" && c.IPIPMode != "Always" {
			r.warnf("network.calico.ipipMode %q has no v4 equivalent; configure calico values manually", c.IPIPMode)
		}
		if c.VXLANMode != "" && c.VXLANMode != "Never" {
			r.warnf("network.calico.vxlanMode %q has no v4 equivalent; configure calico values manually", c.VXLANMode)
		}
		if c.VethMTU != 0 {
			r.warnf("network.calico.vethMTU %d has no v4 equivalent; configure calico values manually", c.VethMTU)
		}
		if c.IPAutoDetectionMethod != "" {
			r.warnf("network.calico.ipAutoDetectionMethod %q has no v4 equivalent; configure calico values manually", c.IPAutoDetectionMethod)
		}
		if c.Ipv4NatOutgoing != nil && !*c.Ipv4NatOutgoing {
			r.warnf("network.calico.ipv4NatOutgoing=false has no v4 equivalent; configure calico values manually")
		}
		if len(c.Typha) > 0 || len(c.Controller) > 0 {
			r.warnf("network.calico.typha/controller have no v4 equivalent; configure calico values manually")
		}
	}
	for _, p := range []struct {
		name string
		cfg  map[string]any
	}{{"flannel", n.Flannel}, {"kubeovn", n.Kubeovn}, {"hybridnet", n.Hybridnet}} {
		if len(p.cfg) > 0 {
			r.warnf("network.%s has no direct v4 equivalent; review and configure it manually", p.name)
		}
	}
	if v, ok := n.MultusCNI["enabled"].(bool); ok && v {
		cni["multi_cni"] = "multus"
		r.warnf("network.multusCNI.enabled is mapped to cni.multi_cni=multus; verify multus settings")
	}

	return cni
}

// convertDNS builds the top-level v4 dns section from the v3 kubernetes
// dnsDomain/nodelocaldns fields and the v3 dns block.
func (r *Result) convertDNS(cluster *Cluster) map[string]any {
	dns := map[string]any{}

	if k3 := cluster.Spec.Kubernetes; k3.DNSDomain != "" {
		dns["domain"] = k3.DNSDomain
	}
	if k3 := cluster.Spec.Kubernetes; k3.Nodelocaldns != nil {
		setNested(dns, *k3.Nodelocaldns, "nodelocaldns", "enabled")
	}

	d3 := cluster.Spec.DNS
	if len(d3.CoreDNS) > 0 {
		r.warnf("dns.coredns advanced settings (additionalConfigs/externalZones/rewriteBlock/upstreamDNSServers) have no direct v4 equivalent; migrate them to dns.coredns.zone_configs manually")
	}
	if len(d3.NodeLocalDNS) > 0 {
		r.warnf("dns.nodelocaldns.externalZones have no direct v4 equivalent; migrate them manually")
	}
	if d3.DNSEtcHosts != "" || d3.NodeEtcHosts != "" {
		r.warnf("dns.dnsEtcHosts/nodeEtcHosts have no direct v4 equivalent; review dns.coredns.dns_etc_hosts and node /etc/hosts handling manually")
	}

	if len(dns) == 0 {
		return nil
	}
	return dns
}

func (r *Result) convertEtcd(cluster *Cluster) map[string]any {
	e3 := cluster.Spec.Etcd
	etcd := map[string]any{}

	switch e3.Type {
	case "", v3EtcdKubeKey, v3EtcdKubeadm:
		etcd["deployment_type"] = "internal"
	case v3EtcdExternal:
		etcd["deployment_type"] = "external"
		if len(e3.External) > 0 {
			r.warnf("etcd.external endpoints/certs have no direct v4 equivalent; configure the etcd host group and certificates manually")
		}
	default:
		r.warnf("unsupported etcd.type %q, converted to deployment_type internal", e3.Type)
		etcd["deployment_type"] = "internal"
	}
	if e3.Port != nil {
		etcd["port"] = *e3.Port
	}
	if e3.PeerPort != nil {
		etcd["peer_port"] = *e3.PeerPort
	}

	env := map[string]any{}
	if e3.DataDir != nil && *e3.DataDir != "" {
		env["data_dir"] = *e3.DataDir
	}
	if e3.HeartbeatInterval != nil {
		env["heartbeat_interval"] = *e3.HeartbeatInterval
	}
	if e3.ElectionTimeout != nil {
		env["election_timeout"] = *e3.ElectionTimeout
	}
	if e3.SnapshotCount != nil {
		env["snapshot_count"] = *e3.SnapshotCount
	}
	if e3.AutoCompactionRetention != nil {
		env["compaction_retention"] = *e3.AutoCompactionRetention
	}
	if e3.Metrics != nil && *e3.Metrics != "" {
		env["metrics"] = *e3.Metrics
	}
	if e3.QuotaBackendBytes != nil {
		env["quota_backend_bytes"] = *e3.QuotaBackendBytes
	}
	if e3.MaxRequestBytes != nil {
		env["max_request_bytes"] = *e3.MaxRequestBytes
	}
	if e3.MaxSnapshots != nil {
		env["max_snapshots"] = *e3.MaxSnapshots
	}
	if e3.MaxWals != nil {
		env["max_wals"] = *e3.MaxWals
	}
	if e3.LogLevel != nil && *e3.LogLevel != "" {
		env["log_level"] = *e3.LogLevel
	}
	if len(env) > 0 {
		etcd["env"] = env
	}

	backup := map[string]any{}
	if e3.BackupDir != "" {
		backup["backup_dir"] = e3.BackupDir
	}
	if e3.KeepBackupNumber != 0 {
		backup["keep_backup_number"] = e3.KeepBackupNumber
	}
	if len(backup) > 0 {
		etcd["backup"] = backup
	}
	if e3.BackupPeriod != 0 {
		r.warnf("etcd.backupPeriod %d uses a different scheduling model in v4 (etcd.backup.on_calendar); review it manually", e3.BackupPeriod)
	}
	if e3.BackupScriptDir != "" {
		r.warnf("etcd.backupScript %q has no direct v4 equivalent (etcd.backup.etcd_backup_script); review it manually", e3.BackupScriptDir)
	}
	if len(e3.ExtraArgs) > 0 {
		r.warnf("etcd.extraArgs has no v4 equivalent and is dropped")
	}

	return etcd
}

func (r *Result) convertCRI(cluster *Cluster) map[string]any {
	cri := map[string]any{}

	if cm := cluster.Spec.Kubernetes.ContainerManager; cm != "" {
		cri["container_manager"] = cm
	}

	reg := cluster.Spec.Registry
	criRegistry := map[string]any{}
	if len(reg.RegistryMirrors) > 0 {
		criRegistry["mirrors"] = reg.RegistryMirrors
	}
	if len(reg.InsecureRegistries) > 0 {
		criRegistry["insecure_registries"] = reg.InsecureRegistries
	}
	if len(reg.Auths) > 0 {
		auths := convertRegistryAuths(reg.Auths, r)
		if len(auths) > 0 {
			criRegistry["auths"] = auths
		}
	}
	if len(criRegistry) > 0 {
		cri["registry"] = criRegistry
	}

	if reg.ContainerdDataDir != "" {
		setNested(cri, reg.ContainerdDataDir, "containerd", "data_root")
	}
	if reg.DockerDataDir != "" {
		setNested(cri, reg.DockerDataDir, "docker", "daemon", "data-root")
	}
	if reg.BridgeIP != "" {
		r.warnf("registry.bridgeIP has no v4 equivalent and is dropped")
	}
	if reg.NamespaceOverride != "" {
		r.warnf("registry.namespaceOverride %q has no direct v4 equivalent; review image naming manually", reg.NamespaceOverride)
	}
	if len(reg.NamespaceRewrite) > 0 {
		r.warnf("registry.namespaceRewrite has no v4 equivalent and is dropped")
	}
	if len(reg.RemoteMirrors) > 0 {
		r.warnf("registry.remoteMirrors has no v4 equivalent and is dropped")
	}
	if reg.Type != "" {
		r.warnf("registry.type %q is ignored by v4; use image_registry.type instead", reg.Type)
	}

	if len(cri) == 0 {
		return nil
	}
	return cri
}

// convertRegistryAuths converts the v3 auths map (keyed by registry) into the
// v4 auths list ({registry, username, password, skip_tls_verify, ...}).
func convertRegistryAuths(auths map[string]any, r *Result) []map[string]any {
	keys := make([]string, 0, len(auths))
	for k := range auths {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	result := make([]map[string]any, 0, len(auths))
	for _, registry := range keys {
		entry, ok := auths[registry].(map[string]any)
		if !ok {
			r.warnf("registry.auths[%q] has an unexpected format and is dropped", registry)
			continue
		}
		item := map[string]any{"registry": registry}
		for src, dst := range map[string]string{
			"username":      "username",
			"password":      "password",
			"skipTLSVerify": "skip_tls_verify",
		} {
			if v, ok := entry[src]; ok {
				item[dst] = v
			}
		}
		if v, ok := entry["plainHTTP"].(bool); ok && v {
			r.warnf("registry.auths[%q].plainHTTP has no v4 auths equivalent; use image_registry.auth.plain_http", registry)
		}
		if v, ok := entry["certsPath"].(string); ok && v != "" {
			r.warnf("registry.auths[%q].certsPath has no direct v4 equivalent; use auths ca_cert/cert_file/key_file", registry)
		}
		result = append(result, item)
	}
	return result
}

func (r *Result) convertImageRegistry(cluster *Cluster) map[string]any {
	reg := cluster.Spec.Registry
	ir := map[string]any{}

	if reg.PrivateRegistry != "" {
		ir["auth"] = map[string]any{"registry": reg.PrivateRegistry}
		r.warnf("registry.privateRegistry is mapped to image_registry.auth.registry; v4 uses it as the pull/push registry, verify the value %q", reg.PrivateRegistry)
	}
	if reg.RegistryDataDir != "" {
		r.warnf("registry.registryDataDir has no v4 equivalent and is dropped")
	}
	if len(ir) == 0 {
		return nil
	}
	return ir
}

func (r *Result) convertStorage(cluster *Cluster) map[string]any {
	basePath := cluster.Spec.Storage.OpenEBS.BasePath
	if basePath == "" {
		return nil
	}
	return map[string]any{
		"local": map[string]any{
			"enabled": true,
			"path":    basePath,
		},
	}
}

func (r *Result) convertNative(cluster *Cluster) map[string]any {
	sys := cluster.Spec.System
	native := map[string]any{}

	if len(sys.NtpServers) > 0 {
		setNested(native, sys.NtpServers, "ntp", "servers")
	}
	if sys.Timezone != "" {
		native["timezone"] = sys.Timezone
	}
	if len(sys.Rpms) > 0 || len(sys.Debs) > 0 {
		r.warnf("system.rpms/debs have no v4 equivalent; install extra packages via hook playbooks")
	}
	if len(sys.PreInstall) > 0 || len(sys.PostClusterInstall) > 0 || len(sys.PostInstall) > 0 {
		r.warnf("system.preInstall/postClusterInstall/postInstall scripts have no v4 equivalent; migrate them to hook playbooks")
	}
	if sys.SkipConfigureOS {
		r.warnf("system.skipConfigureOS has no v4 equivalent and is dropped")
	}

	if len(native) == 0 {
		return nil
	}
	return native
}

// calicoEmpty reports whether no calico field is set.
func calicoEmpty(c CalicoCfg) bool {
	return c.IPIPMode == "" && c.IPAutoDetectionMethod == "" && c.VXLANMode == "" && c.VethMTU == 0 &&
		c.Ipv4NatOutgoing == nil && c.Ipv6NatOutgoing == nil && c.DefaultIPPOOL == nil &&
		len(c.Typha) == 0 && len(c.Controller) == 0
}

// argsToMap converts v3 component args ("K=V" strings) into the v4 extra_args map.
func argsToMap(args []string) map[string]any {
	if len(args) == 0 {
		return nil
	}
	m := make(map[string]any, len(args))
	for _, arg := range args {
		k, v, found := strings.Cut(arg, "=")
		if !found {
			m[k] = ""
			continue
		}
		m[k] = v
	}
	return m
}

// featureGatesArg joins feature gates into the "A=true,B=false" argument value.
func featureGatesArg(gates map[string]bool) string {
	if len(gates) == 0 {
		return ""
	}
	keys := make([]string, 0, len(gates))
	for k := range gates {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%t", k, gates[k]))
	}
	return strings.Join(parts, ",")
}

func mergeUnique(lists ...[]string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0)
	for _, list := range lists {
		for _, item := range list {
			if _, ok := seen[item]; ok {
				continue
			}
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

// setNested sets a nested value in a map[string]any tree, creating
// intermediate maps as needed. It panics-free overwrites existing values.
func setNested(obj map[string]any, value any, fields ...string) {
	for i, field := range fields {
		if i == len(fields)-1 {
			obj[field] = value
			return
		}
		next, ok := obj[field].(map[string]any)
		if !ok {
			next = map[string]any{}
			obj[field] = next
		}
		obj = next
	}
}

// ConvertInventoryAndConfig converts v3 configuration YAML bytes into v4
// inventory and config YAML bytes.
func ConvertInventoryAndConfig(data []byte) (inventoryYAML, configYAML []byte, warnings []string, err error) {
	cluster, err := ParseCluster(data)
	if err != nil {
		return nil, nil, nil, err
	}
	result, err := Convert(cluster)
	if err != nil {
		return nil, nil, nil, err
	}

	inventoryYAML, err = marshalYAML(inventoryOutputMap(result.Inventory))
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to marshal inventory")
	}
	configYAML, err = marshalYAML(map[string]any{
		"apiVersion": "kubekey.kubesphere.io/v1",
		"kind":       "Config",
		"spec":       result.Config,
	})
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to marshal config")
	}
	return inventoryYAML, configYAML, result.Warnings, nil
}

// inventoryOutputMap renders the Inventory as a clean output document without
// CRD zero values (empty status, null vars, ...).
func inventoryOutputMap(inv *kkcorev1.Inventory) map[string]any {
	hosts := make(map[string]any, len(inv.Spec.Hosts))
	for name, raw := range inv.Spec.Hosts {
		vars := map[string]any{}
		if len(raw.Raw) > 0 {
			if err := json.Unmarshal(raw.Raw, &vars); err != nil {
				vars = map[string]any{}
			}
		}
		hosts[name] = vars
	}

	groups := make(map[string]any, len(inv.Spec.Groups))
	for name, group := range inv.Spec.Groups {
		g := map[string]any{}
		if len(group.Groups) > 0 {
			g["groups"] = group.Groups
		}
		if len(group.Hosts) > 0 {
			g["hosts"] = group.Hosts
		}
		groups[name] = g
	}

	return map[string]any{
		"apiVersion": "kubekey.kubesphere.io/v1",
		"kind":       "Inventory",
		"metadata":   map[string]any{"name": inv.Name},
		"spec": map[string]any{
			"hosts":  hosts,
			"groups": groups,
		},
	}
}
