# 升级集群 (upgrade_cluster.yaml)

`upgrade_cluster.yaml` 用于对已有 Kubernetes 集群执行滚动升级。默认情况下只升级 Kubernetes 控制面和工作节点二进制，其他组件需要通过 `upgrade` 开关显式开启。

## 升级开关

在 `config.yaml` 中可以通过 `upgrade` 字段控制是否同步升级可选组件：

```yaml
upgrade:
  cri: false          # 是否同步升级容器运行时（docker/containerd）
  etcd: false         # 是否同步升级外部 etcd 集群
  dns: false          # 是否同步升级 CoreDNS / NodeLocalDNS
  image_registry: false
  cni: false          # 是否同步升级网络插件
  storage_class: false
  nfs: false
```

也可以在命令行通过 `--all` 或 `--set upgrade.xxx=true` 覆盖。

## pre_hook

`pre_hook` 允许用户在升级之前，在对应节点上执行脚本。

执行流程：
1. 复制本地脚本到远程节点的 `/etc/kubekey/scripts/pre_install_{{ .inventory_hostname }}.sh`
2. 设置脚本文件权限为 `0755`
3. 遍历每个远程节点上 `/etc/kubekey/scripts/` 目录下所有 `pre_install_*.sh` 文件并执行

> **work_dir**: 工作目录，默认当前命令执行目录。
> **inventory_hostname**: `inventory.yaml` 文件中定义的主机名称。

## precheck

`precheck` 阶段在升级前检查集群是否满足升级条件。

**os_precheck**: 操作系统检查，包括：
- **主机名检查**: 验证主机名格式是否合规
- **操作系统版本检查**: 验证操作系统是否在支持列表中
- **架构检查**: 验证系统架构是否为 amd64 或 arm64
- **内存检查**: 验证控制平面节点和工作节点内存是否满足最小要求
- **内核版本检查**: 验证内核版本是否满足最低要求

**kubernetes_precheck**: Kubernetes 相关检查，包括：
- **KubeVIP 检查**: 当使用 `kube-vip` 类型控制平面端点时，验证地址有效且未被占用
- **Kubernetes 版本检查**: 验证目标 Kubernetes 版本是否满足最低要求
- **升级路径检查**: 升级场景下，要求已安装版本必须低于目标版本

**etcd_precheck**: etcd 集群检查，包括：
- **部署类型校验**: 校验 `internal` 或 `external`
- **版本校验**: 当 `upgrade.etcd=true` 时，目标版本不能低于已安装版本，且需满足目标 Kubernetes 版本的最小 etcd 版本要求
- **磁盘 IO 性能检查**: 使用 `fio` 测试 etcd 数据盘 WAL fsync 延迟

**cri_precheck**: 容器运行时检查，包括：
- **容器管理器检查**（localhost 执行）: 验证配置的容器管理器是否在支持列表中（docker 或 containerd）
- **containerd 版本检查**（localhost 执行）: 升级 CRI 时验证目标 containerd 版本满足最低要求
- **已安装 containerd 版本检查**（k8s_cluster 节点执行）: 不升级 CRI 时，验证节点上已安装 containerd 版本满足目标 Kubernetes 要求
- **Docker live-restore 检查**（k8s_cluster 节点执行）: 当 `upgrade.cri=true` 且使用 Docker 时，检查 `/etc/docker/daemon.json` 中是否启用了 `live-restore`；未启用时打印警告

**network_precheck**: 网络连通性检查，包括：
- 网络接口、CIDR 格式、双栈支持、网络插件、地址空间等检查

## init

`init` 阶段在 `localhost` 上准备升级所需资源，包括：
- 根据目标 Kubernetes 版本加载版本特定的默认变量
- 生成或更新证书
- 下载 Kubernetes、etcd、CRI、CNI 等二进制包和镜像清单

## upgrade

`upgrade` 阶段按顺序在各节点组上执行实际升级。

### native

为 `etcd`、`k8s_cluster`、`image_registry`、`nfs` 组中的节点执行 OS 级初始化，包括仓库、NTP、DNS、主机名等。

### etcd

仅当 `upgrade.etcd=true` 且 `etcd.deployment_type=external` 时执行：
- 在 etcd leader 上备份数据
- 分发新版 etcd 二进制
- 逐个节点（`serial: 1`）重启 etcd 并等待健康

### cri

为 `k8s_cluster` 组中的节点升级容器运行时：
- 当 `upgrade.cri=true` 或节点尚未安装 CRI 时，执行完整升级
- 备份原配置和二进制
- 同步新版本二进制、镜像仓库证书、systemd 服务文件
- 重启容器运行时服务

### kubernetes

1. **pre-kubernetes**: 同步 K8s 二进制、创建目录、同步 CA / etcd / front-proxy 证书、应用 kubeadm patches。
2. **control plane**（`serial: 1`）: 逐个升级控制平面节点。
   - 第一个控制平面节点执行 `kubeadm upgrade apply`
   - 其他控制平面节点执行 `kubeadm upgrade node`
3. **worker**: 并行升级所有工作节点，执行 `kubeadm upgrade node` 并重启 kubelet。

### cni / storage_class

仅在随机选中的一个控制平面节点上执行：
- 当 `upgrade.cni=true` 时升级网络插件
- 当 `upgrade.storage_class=true` 时升级 StorageClass provisioner

## post_hook

`post_hook` 阶段在升级完成后执行：
1. 如果启用安全增强（`.security_enhancement=true`），执行 `security` 角色
2. 复制并执行 `/etc/kubekey/scripts/post_install_*.sh` 脚本

> **work_dir**: 工作目录，默认当前命令执行目录。
> **inventory_hostname**: `inventory.yaml` 文件中定义的主机名称。

## 升级风险说明

### Worker 节点并行升级

`upgrade_cluster.yaml` 中 Worker 节点的升级 play **没有 `serial: 1`**，因此默认会**并行**升级所有工作节点。playbook 本身也不会自动执行 `kubectl drain` 或 `cordon`。

这意味着：
- 多个 Worker 可能同时进入 `NotReady` 状态；
- 如果工作负载副本不足或集中在部分节点上，业务可能短暂不可用；
- 如需滚动升级，请提前手动 `drain` 节点，或自行控制并发。

### 同步升级 CRI（`upgrade.cri=true`）

当 `upgrade.cri=true` 时，CRI 服务会被重启：

- **containerd**: 重启 containerd daemon 通常不会导致已有容器退出（containerd 重启后会重新 attach 到 shim），但节点会短暂 `NotReady`。
- **Docker**: 是否中断容器取决于是否启用了 `live-restore`：
  - 启用 `live-restore` 时，重启 dockerd 不会中断已有容器。
  - 未启用 `live-restore` 时，重启 dockerd 可能导致正在运行的容器停止。

因此，当使用 Docker 且 `upgrade.cri=true` 时，precheck 会检查每个 `k8s_cluster` 节点的 `/etc/docker/daemon.json`，如果未启用 `live-restore` 则打印警告。建议在升级前启用 `live-restore`，或确保关键业务副本充分分布在多个 Worker 上。
