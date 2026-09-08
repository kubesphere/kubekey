# 将 v3（v1alpha2）配置迁移到 v4

KubeKey v4 使用两个文件来定义一个集群：

- `inventory.yaml` —— `Inventory` 资源（节点与主机组）。
- `config.yaml` —— `Config` 资源（集群规格；`spec` 为 unstructured 结构）。

KubeKey v3（`3.x` 分支）使用单一的 `Cluster` 清单，API 版本为
`kubekey.kubesphere.io/v1alpha2`。为了降低升级到 v4 时手写改写的成本，
`kk create convert` 会解析一份 v3 配置并生成 v4 的 `inventory.yaml` 与
`config.yaml`。

```bash
# 将两个文档打印到 stdout，以 `---` 分隔
kk create convert --input config-v3.yaml

# 或写入到指定目录
kk create convert --input config-v3.yaml -o ./out
# ./out/inventory.yaml
# ./out/config.yaml
```

没有清晰 1:1 v4 对应项的字段（例如 KubeSphere 启用、高级 CoreDNS、
`system.rpms` 等）**不会被静默丢弃** —— 它们会以告警形式输出到 stderr。
请审阅这些告警，并在需要时手工调整生成的 `config.yaml`。

## 转换器如何读取 v3

`pkg/convert/v1alpha2.go` 是 v3 schema 的**最小镜像**：只声明转换所需的字段。
YAML 解码器会忽略未知字段，复合辅助类型（`CustomScripts`、`MirrorConfig`、
`NamespaceRewrite`）以原始值保留，因此解析过程不会丢失数据。`metadata` 仅保留
`name`。

## 字段映射

以下所有路径均相对于生成文件。`inventory.yaml` 的路径在 `spec` 之下；
`config.yaml` 的路径在 `spec` 之下。

### Hosts → inventory.yaml

| v3 `spec.hosts[]` | v4 `inventory.yaml` | 说明 |
|---|---|---|
| `name` | `spec.hosts.<name>`（map 键） | |
| `address` | `spec.hosts.<name>.connector.host` | |
| `port` | `connector.port` | `0` → 默认 `22` |
| `user` | `connector.user` | `""` → 默认 `root` |
| `password` | `connector.password` | |
| `privateKeyPath` | `connector.private_key` | |
| `privateKey` | `connector.private_key_content` | |
| `internalAddress` | `internal_ipv4` / `internal_ipv6` | 逗号分隔的双栈地址拆分 |
| `arch` | `arch` | |
| `labels` | `labels` | |

### roleGroups → inventory.yaml 主机组

v3 的 `HostCfg` 本身不带 role/taint 信息；角色仅来自 `roleGroups`，
并支持 `node[i:j]` 范围简写（例如 `node[1:3]` → `node1`、`node2`、`node3`）。

| v3 `roleGroups` 键 | v4 主机组 |
|---|---|
| `master` / `control-plane` / `controlplane` | `kube_control_plane`（合并） |
| `worker` | `kube_worker` |
| `etcd` | `etcd` |
| `registry` | `image_registry` |
| 其他任意键 | 无对应 v4 组 → **告警** |
| 聚合 | `k8s_cluster.groups = [kube_control_plane, kube_worker]` |

### controlPlaneEndpoint → config.yaml kubernetes.control_plane_endpoint

| v3 `controlPlaneEndpoint` | v4 `kubernetes.control_plane_endpoint` | 说明 |
|---|---|---|
| `domain` | `host` | |
| `port` | `port` | |
| `internalLoadbalancer` = `""` / `none` | `type: local` | `address` → `local.address` |
| `internalLoadbalancer` = `kubevip` | `type: kube-vip` | `address` → `kube_vip.address`；`kubevip.mode` → `kube_vip.mode` |
| `internalLoadbalancer` = `haproxy` | `type: haproxy` | `address` → **告警**（v4 haproxy 监听 `127.0.0.1`） |
| `internalLoadbalancer` = 其他 | `type: local` | **告警** |

### kubernetes → config.yaml kubernetes

| v3 `kubernetes` | v4 `kubernetes` | 说明 |
|---|---|---|
| `version` | `kube_version` | |
| `clusterName` | `cluster_name` | |
| `autoRenewCerts` | `certs.renew` | |
| `apiserverCertExtraSans` | `apiserver.certSANs` | |
| `apiserverArgs`（`K=V`） | `apiserver.extra_args` | map 形式 |
| `featureGates` | `apiserver.extra_args.feature-gates` | 拼接为 `K=true,K2=false` |
| `controllerManagerArgs` | `controller_manager.extra_args` | |
| `schedulerArgs` | `scheduler.extra_args` | |
| `proxyMode` | `kube_proxy.mode` | |
| `masqueradeAll` | `kube_proxy.config.iptables.masqueradeAll` | |
| `disableKubeProxy` | `kube_proxy.manage.enabled = false` | **告警** |
| `maxPods` | `kubelet.max_pods` | |
| `podPidsLimit` | `kubelet.pod_pids_limit` | |
| `nodeCidrMaskSize` | `cni.ipv4_mask_size` | 见 CNI |
| `nodeCidrMaskSizeIPv6` | `cni.ipv6_mask_size` | 见 CNI |
| `dnsDomain` | `dns.domain` | 见 DNS |
| `nodelocaldns` | `dns.nodelocaldns.enabled` | 见 DNS |

### network → config.yaml cni

| v3 `network` | v4 `cni` | 说明 |
|---|---|---|
| `plugin` | `type` | `calico`/`cilium`/`flannel`/`kubeovn`/`hybridnet`；未知 → `other` + **告警** |
| `kubePodsCIDR` | `pod_cidr` | |
| `kubeServiceCIDR` | `service_cidr` | |
| `multusCNI.enabled` | `multi_cni = "multus"` | **告警** |

### dns → config.yaml dns（顶层）

| v3 | v4 `dns` | 说明 |
|---|---|---|
| `kubernetes.dnsDomain` | `domain` | |
| `kubernetes.nodelocaldns` | `nodelocaldns.enabled` | |

### etcd → config.yaml etcd

| v3 `etcd` | v4 `etcd` | 说明 |
|---|---|---|
| `type` = `kubekey`/`kubeadm`/`""` | `deployment_type: internal` | |
| `type` = `external` | `deployment_type: external` | 外部端点/证书 → **告警** |
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

| v3 `registry` | v4 | 说明 |
|---|---|---|
| `registryMirrors` | `cri.registry.mirrors` | |
| `insecureRegistries` | `cri.registry.insecure_registries` | |
| `auths`（以 registry 为键的 map） | `cri.registry.auths` | 列表，元素为 `{registry, username, password, skip_tls_verify}` |
| `containerdDataDir` | `cri.containerd.data_root` | |
| `dockerDataDir` | `cri.docker.daemon.data-root` | |
| `privateRegistry` | `image_registry.auth.registry` | **告警**：v4 将其用作拉取/推送镜像的仓库 |

### system → config.yaml native

| v3 `system` | v4 `native` | 说明 |
|---|---|---|
| `ntpServers` | `ntp.servers` | |
| `timezone` | `timezone` | |

### storage → config.yaml storage_class

| v3 `storage.openebs.basePath` | v4 `storage_class` | 说明 |
|---|---|---|
| `basePath` | `local.enabled = true`、`local.path = <basePath>` | |

## 无直接 v4 对应项的字段（以告警形式报告）

以下 v3 字段要么被丢弃，要么需要手工迁移。转换器会针对每一项打印告警，
方便你手工调整 `config.yaml`。

| v3 字段 | 处理结果 | 建议 |
|---|---|---|
| `kubernetes.kubeProxyArgs` | 丢弃 | — |
| `kubernetes.kubeProxyConfiguration` | 手工 | 迁移到 `kubernetes.kube_proxy.config` |
| `kubernetes.kubeletArgs` | 丢弃 | — |
| `kubernetes.kubeletConfiguration` | 手工 | 迁移到 `kubernetes.kubelet` |
| `kubernetes.containerRuntimeEndpoint` | 丢弃 | — |
| `kubernetes.nodeFeatureDiscovery` | 丢弃 | — |
| `kubernetes.kata` | 丢弃 | — |
| `kubernetes.nvidiaRuntime` | 丢弃 | — |
| `kubernetes.type` | 忽略 | v4 无集群类型概念 |
| `network.calico.ipipMode`（非 Always） | 手工 | 配置 calico values |
| `network.calico.vxlanMode`（非 Never） | 手工 | 配置 calico values |
| `network.calico.vethMTU` | 手工 | 配置 calico values |
| `network.calico.ipAutoDetectionMethod` | 手工 | 配置 calico values |
| `network.calico.ipv4NatOutgoing=false` | 手工 | 配置 calico values |
| `network.calico.typha` / `controller` | 手工 | 配置 calico values |
| `network.flannel` / `kubeovn` / `hybridnet` | 手工 | v4 未暴露各插件细节 |
| `dns.coredns` | 手工 | 迁移到 `dns.coredns.zone_configs` |
| `dns.nodelocaldns.externalZones` | 手工 | — |
| `dns.dnsEtcHosts` / `dns.nodeEtcHosts` | 手工 | 参考 `dns.coredns.dns_etc_hosts` |
| `etcd.backupPeriod` | 手工 | v4 使用 `etcd.backup.on_calendar` |
| `etcd.backupScript` | 手工 | 使用 `etcd.backup.etcd_backup_script` |
| `etcd.extraArgs` | 丢弃 | — |
| `etcd.external`（端点/证书） | 手工 | 配置 etcd 主机组与证书 |
| `registry.bridgeIP` | 丢弃 | — |
| `registry.namespaceOverride` | 手工 | 复核镜像命名 |
| `registry.namespaceRewrite` | 丢弃 | — |
| `registry.remoteMirrors` | 丢弃 | — |
| `registry.type` | 忽略 | 使用 `image_registry.type` |
| `system.rpms` / `system.debs` | 手工 | 通过 hook playbook 安装 |
| `system.preInstall` / `postClusterInstall` / `postInstall` | 手工 | 迁移到 hook playbook |
| `system.skipConfigureOS` | 丢弃 | — |

## 已解析但尚未转换

以下 v3 字段会被解析器捕获，但转换器中尚无对应的 `config.yaml` 等价项，
因此当前被忽略（不打印告警）：

- `spec.kubesphere`（`enabled` / `version` / `configurations`）—— v4 不通过
  `config.yaml` 启用 KubeSphere，请单独部署 KubeSphere。
- `spec.addons` —— v4 的 `config.yaml` 无直接的 addons 映射。

后续版本计划为这些字段补充显式处理（及告警）。
