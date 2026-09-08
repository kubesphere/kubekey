# Migrate a v3 (v1alpha2) Configuration to v4

KubeKey v4 defines a cluster with two files:

- `inventory.yaml` — the `Inventory` resource (hosts and groups).
- `config.yaml` — the `Config` resource (cluster spec; `spec` is an unstructured map).

KubeKey v3 (the `3.x` branch) uses a single `Cluster` manifest of API version
`kubekey.kubesphere.io/v1alpha2`. To reduce the manual rewrite when adopting v4,
`kk create convert` parses a v3 config and emits the v4 `inventory.yaml` and
`config.yaml`.

```bash
# Print both documents to stdout, separated by `---`
kk create convert --input config-v3.yaml

# Or write them to a directory
kk create convert --input config-v3.yaml -o ./out
# ./out/inventory.yaml
# ./out/config.yaml
```

Fields that have no clean 1:1 v4 equivalent (for example KubeSphere enablement,
advanced CoreDNS, or `system.rpms`) are **not silently dropped** — they are
reported as warnings on stderr. Review the warnings and adjust the generated
`config.yaml` by hand where needed.

## How the converter reads v3

`pkg/convert/v1alpha2.go` is a **minimal mirror** of the v3 schema: it declares
only the fields needed for conversion. The YAML decoder ignores unknown fields,
and composite helper types (`CustomScripts`, `MirrorConfig`, `NamespaceRewrite`)
are preserved as raw values, so parsing never loses data. `metadata` is reduced
to `name` only.

## Field mapping

All paths below are relative to the generated file. `inventory.yaml` paths are
under `spec`; `config.yaml` paths are under `spec`.

### Hosts → inventory.yaml

| v3 `spec.hosts[]` | v4 `inventory.yaml` | Note |
|---|---|---|
| `name` | `spec.hosts.<name>` (map key) | |
| `address` | `spec.hosts.<name>.connector.host` | |
| `port` | `connector.port` | `0` → default `22` |
| `user` | `connector.user` | `""` → default `root` |
| `password` | `connector.password` | |
| `privateKeyPath` | `connector.private_key` | |
| `privateKey` | `connector.private_key_content` | |
| `internalAddress` | `internal_ipv4` / `internal_ipv6` | comma-split dual-stack |
| `arch` | `arch` | |
| `labels` | `labels` | |

### roleGroups → inventory.yaml groups

v3 `HostCfg` carries no role/taint info; roles come only from `roleGroups`,
including the `node[i:j]` range shorthand (e.g. `node[1:3]` → `node1`,`node2`,`node3`).

| v3 `roleGroups` key | v4 group |
|---|---|
| `master` / `control-plane` / `controlplane` | `kube_control_plane` (merged) |
| `worker` | `kube_worker` |
| `etcd` | `etcd` |
| `registry` | `image_registry` |
| any other key | no v4 group → **warning** |
| aggregate | `k8s_cluster.groups = [kube_control_plane, kube_worker]` |

### controlPlaneEndpoint → config.yaml kubernetes.control_plane_endpoint

| v3 `controlPlaneEndpoint` | v4 `kubernetes.control_plane_endpoint` | Note |
|---|---|---|
| `domain` | `host` | |
| `port` | `port` | |
| `internalLoadbalancer` = `""` / `none` | `type: local` | `address` → `local.address` |
| `internalLoadbalancer` = `kubevip` | `type: kube-vip` | `address` → `kube_vip.address`; `kubevip.mode` → `kube_vip.mode` |
| `internalLoadbalancer` = `haproxy` | `type: haproxy` | `address` → **warning** (v4 haproxy listens on `127.0.0.1`) |
| `internalLoadbalancer` = other | `type: local` | **warning** |

### kubernetes → config.yaml kubernetes

| v3 `kubernetes` | v4 `kubernetes` | Note |
|---|---|---|
| `version` | `kube_version` | |
| `clusterName` | `cluster_name` | |
| `autoRenewCerts` | `certs.renew` | |
| `apiserverCertExtraSans` | `apiserver.certSANs` | |
| `apiserverArgs` (`K=V`) | `apiserver.extra_args` | map form |
| `featureGates` | `apiserver.extra_args.feature-gates` | joined as `K=true,K2=false` |
| `controllerManagerArgs` | `controller_manager.extra_args` | |
| `schedulerArgs` | `scheduler.extra_args` | |
| `proxyMode` | `kube_proxy.mode` | |
| `masqueradeAll` | `kube_proxy.config.iptables.masqueradeAll` | |
| `disableKubeProxy` | `kube_proxy.manage.enabled = false` | **warning** |
| `maxPods` | `kubelet.max_pods` | |
| `podPidsLimit` | `kubelet.pod_pids_limit` | |
| `nodeCidrMaskSize` | `cni.ipv4_mask_size` | see CNI |
| `nodeCidrMaskSizeIPv6` | `cni.ipv6_mask_size` | see CNI |
| `dnsDomain` | `dns.domain` | see DNS |
| `nodelocaldns` | `dns.nodelocaldns.enabled` | see DNS |

### network → config.yaml cni

| v3 `network` | v4 `cni` | Note |
|---|---|---|
| `plugin` | `type` | `calico`/`cilium`/`flannel`/`kubeovn`/`hybridnet`; unknown → `other` + **warning** |
| `kubePodsCIDR` | `pod_cidr` | |
| `kubeServiceCIDR` | `service_cidr` | |
| `multusCNI.enabled` | `multi_cni = "multus"` | **warning** |

### dns → config.yaml dns (top-level)

| v3 | v4 `dns` | Note |
|---|---|---|
| `kubernetes.dnsDomain` | `domain` | |
| `kubernetes.nodelocaldns` | `nodelocaldns.enabled` | |

### etcd → config.yaml etcd

| v3 `etcd` | v4 `etcd` | Note |
|---|---|---|
| `type` = `kubekey`/`kubeadm`/`""` | `deployment_type: internal` | |
| `type` = `external` | `deployment_type: external` | **warning** on external endpoints/certs |
| `port` | `port` | |
| `peerPort` | `peer_port` | |
| `dataDir` | `env.data_dir` | |
| `heartbeatInterval` | `env.heartbeat_interval` | |
| `electionTimeout` | `env.election_timeout` | |
| `snapshotCount` | `env.snapshot_count` | |
| `autoCompactionRetention` | `env.compaction_retention` | |
| `metrics` | `env.metrics` | |
| `quotaBackendBytes` | `env.quota_backend_bytes` | |
| `maxRequestBytes` | `env.max_request_bytes` | |
| `maxSnapshots` | `env.max_snapshots` | |
| `maxWals` | `env.max_wals` | |
| `logLevel` | `env.log_level` | |
| `backupDir` | `backup.backup_dir` | |
| `keepBackupNumber` | `backup.keep_backup_number` | |

### registry → config.yaml cri.registry + image_registry

| v3 `registry` | v4 | Note |
|---|---|---|
| `registryMirrors` | `cri.registry.mirrors` | |
| `insecureRegistries` | `cri.registry.insecure_registries` | |
| `auths` (map keyed by registry) | `cri.registry.auths` | list of `{registry, username, password, skip_tls_verify}` |
| `containerdDataDir` | `cri.containerd.data_root` | |
| `dockerDataDir` | `cri.docker.daemon.data-root` | |
| `privateRegistry` | `image_registry.auth.registry` | **warning**: v4 uses it as the pull/push registry |

### system → config.yaml native

| v3 `system` | v4 `native` | Note |
|---|---|---|
| `ntpServers` | `ntp.servers` | |
| `timezone` | `timezone` | |

### storage → config.yaml storage_class

| v3 `storage.openebs.basePath` | v4 `storage_class` | Note |
|---|---|---|
| `basePath` | `local.enabled = true`, `local.path = <basePath>` | |

## Fields with no direct v4 equivalent (reported as warnings)

These v3 fields are either dropped or require manual migration. The converter
prints a warning for each so you can adjust `config.yaml` by hand.

| v3 field | Outcome | Suggestion |
|---|---|---|
| `kubernetes.kubeProxyArgs` | dropped | — |
| `kubernetes.kubeProxyConfiguration` | manual | migrate to `kubernetes.kube_proxy.config` |
| `kubernetes.kubeletArgs` | dropped | — |
| `kubernetes.kubeletConfiguration` | manual | migrate to `kubernetes.kubelet` |
| `kubernetes.containerRuntimeEndpoint` | dropped | — |
| `kubernetes.nodeFeatureDiscovery` | dropped | — |
| `kubernetes.kata` | dropped | — |
| `kubernetes.nvidiaRuntime` | dropped | — |
| `kubernetes.type` | ignored | v4 has no cluster type |
| `network.calico.ipipMode` (non-Always) | manual | configure calico values |
| `network.calico.vxlanMode` (non-Never) | manual | configure calico values |
| `network.calico.vethMTU` | manual | configure calico values |
| `network.calico.ipAutoDetectionMethod` | manual | configure calico values |
| `network.calico.ipv4NatOutgoing=false` | manual | configure calico values |
| `network.calico.typha` / `controller` | manual | configure calico values |
| `network.flannel` / `kubeovn` / `hybridnet` | manual | v4 does not expose per-plugin details |
| `dns.coredns` | manual | migrate to `dns.coredns.zone_configs` |
| `dns.nodelocaldns.externalZones` | manual | — |
| `dns.dnsEtcHosts` / `dns.nodeEtcHosts` | manual | review `dns.coredns.dns_etc_hosts` |
| `etcd.backupPeriod` | manual | v4 uses `etcd.backup.on_calendar` |
| `etcd.backupScript` | manual | use `etcd.backup.etcd_backup_script` |
| `etcd.extraArgs` | dropped | — |
| `etcd.external` (endpoints/certs) | manual | configure the etcd group + certs |
| `registry.bridgeIP` | dropped | — |
| `registry.namespaceOverride` | manual | review image naming |
| `registry.namespaceRewrite` | dropped | — |
| `registry.remoteMirrors` | dropped | — |
| `registry.type` | ignored | use `image_registry.type` |
| `system.rpms` / `system.debs` | manual | install via hook playbooks |
| `system.preInstall` / `postClusterInstall` / `postInstall` | manual | migrate to hook playbooks |
| `system.skipConfigureOS` | dropped | — |
| `controlPlaneEndpoint.address` (haproxy) | warning | v4 haproxy listens on `127.0.0.1` |

## Parsed but not yet converted

The following v3 fields are captured by the parser but have no `config.yaml`
equivalent in the converter yet, so they are currently ignored (no warning):

- `spec.kubesphere` (`enabled` / `version` / `configurations`) — v4 does not
  enable KubeSphere through `config.yaml`; deploy KubeSphere separately.
- `spec.addons` — v4 has no direct `config.yaml` addons mapping.

Plan to add explicit handling (and warnings) for these in a later release.
