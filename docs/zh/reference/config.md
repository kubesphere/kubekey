# 配置参考

本文档汇总了 KubeKey 内置默认配置文件中所有可用的参数，这些默认值位于 `builtin/core/roles/defaults/defaults/main/` 目录下。您可以参考本文档来编写或修改自己的集群配置文件。

---

## 全局配置 (01-main.yaml)

### 默认配置

```yaml
# KubeKey 工作主目录
work_dir: /root/kubekey
# KubeKey 二进制文件存放目录
binary_dir: >-
  {{ .work_dir }}/kubekey
# 脚本存放目录
scripts_dir: >-
  {{ .work_dir }}/scripts
# 制品（artifact）存放目录
artifact_dir: >-
  {{ .work_dir }}/artifact
# 临时目录
tmp_dir: /tmp/kubekey

# 将常见的机器架构名称映射到标准形式
transform_architectures:
  amd64:
    - amd64
    - x86_64
  arm64:
    - arm64
    - aarch64

# 如果设置为 "cn"，在线下载时将尽可能使用国内可用源
zone: ""

# 启用增强安全特性，以满足更严格的集群安全要求
security_enhancement: false

# 启用 Kubernetes 审计日志
# 审计日志记录并跟踪集群内的关键操作，帮助管理员监控安全事件、排查问题并满足合规要求（如 SOC2、ISO 27001）
audit: false

delete:
# 移除节点时，是否同时卸载该节点的容器运行时（CRI），例如 Docker 或 containerd
# deleteCRI: true
  cri: false

# 移除节点时，是否同时卸载 etcd
# deleteETCD: true
  etcd: false

# 移除节点时，是否恢复该节点的 DNS 配置
# deleteDNS: true
  dns: false

# 移除节点时，是否同时卸载该节点上的私有镜像仓库（如 Harbor 或 registry）
# 通常与 inventory.groups.image_registry 中定义的节点配合使用
# deleteImageRegistry: false
  image_registry: false

# 移除节点时，是否同时删除数据目录（Harbor 数据、registry 数据等）
# 通常与 --with-data 标志或 delete.data: true 一起使用
# deleteData: false
  data: false

# 需要同步到私有镜像仓库的容器镜像列表
image_manifests: []
```

### 参数说明

| 参数 | 说明 |
|------|------|
| `work_dir` | KubeKey 安装及运行时所使用的根工作目录。 |
| `binary_dir` | KubeKey 二进制文件及相关工具存放目录，基于 `work_dir` 自动生成。 |
| `scripts_dir` | 安装过程中所需脚本的存放目录，基于 `work_dir` 自动生成。 |
| `artifact_dir` | 离线包（artifact）的存放目录，基于 `work_dir` 自动生成。 |
| `tmp_dir` | 安装过程中存放临时文件的目录。 |
| `transform_architectures` | 机器架构名称标准化映射，用于统一 `amd64`/`x86_64`、`arm64`/`aarch64` 等。 |
| `zone` | 区域设置，设置为 `"cn"` 时可优先使用国内下载加速源。 |
| `security_enhancement` | 是否启用集群增强安全特性。 |
| `audit` | 是否启用 Kubernetes 审计日志功能。 |
| `delete` | 节点删除时的各项资源清理开关。包含 `cri`、`etcd`、`dns`、`image_registry`、`data`。 |
| `image_manifests` | 用于向私有镜像仓库同步的自定义容器镜像列表。 |

---

## 集群要求 (01-cluster_require.yaml)

### 默认配置

```yaml
# 集群参数边界与兼容性要求
cluster_require:
  # etcd WAL fsync 99分位最大耗时（纳秒）
  etcd_disk_wal_fysnc_duration_seconds: 10000000
  # 是否允许在不支持的 Linux 发行版上安装
  allow_unsupported_distribution_setup: false
  # 支持的操作系统发行版
  supported_os_distributions:
    - ubuntu
    - '"ubuntu"'
    - centos
    - '"centos"'
    - kylin
    - '"kylin"'
    - rocky
    - '"rocky"'
  # 支持的网络插件
  require_network_plugin: ['calico', 'flannel', 'cilium', 'hybridnet', 'kube-ovn']
  # 最低支持的 Kubernetes 版本
  kube_version_min_required: v1.23.0
  # 每个控制平面节点的最低内存要求（MB）
  minimal_master_memory_mb: 10
  # 每个工作节点的最低内存要求（MB）
  minimal_node_memory_mb: 10
  # 支持的 etcd 部署类型
  require_etcd_deployment_type: ['internal', 'external']
  # 支持的容器运行时
  require_container_manager: ['docker', 'containerd']
  # 最低要求的 containerd 版本
  containerd_min_version_required: v1.6.0
  # 支持的 CPU 架构
  supported_architectures:
    - amd64
    - x86_64
    - arm64
    - aarch64
  # 最低要求的 Linux 内核版本
  min_kernel_version: 4.9.17

  # 各 Kubernetes 版本允许的 Calico 版本
  calico_allowed_versions:
    v3.25: ["v1.23", "v1.24", "v1.25", "v1.26", "v1.27", "v1.28"]
    v3.26: ["v1.23", "v1.24", "v1.25", "v1.26", "v1.27", "v1.28"]
    v3.27: ["v1.23", "v1.24", "v1.25", "v1.26", "v1.27", "v1.28", "v1.29"]
    v3.28: ["v1.23", "v1.24", "v1.25", "v1.26", "v1.27", "v1.28", "v1.29", "v1.30"]
    v3.29: ["v1.23", "v1.24", "v1.25", "v1.26", "v1.27", "v1.28", "v1.29", "v1.30", "v1.31", "v1.32"]
    v3.30: ["v1.23", "v1.24", "v1.25", "v1.26", "v1.27", "v1.28", "v1.29", "v1.30", "v1.31", "v1.32", "v1.33"]
    v3.31: ["v1.23", "v1.24", "v1.25", "v1.26", "v1.27", "v1.28", "v1.29", "v1.30", "v1.31", "v1.32", "v1.33", "v1.34"]

  # 各 Kubernetes 版本允许的 Cilium 版本
  cilium_allowed_versions:
    "1.14": ["v1.23", "v1.24", "v1.25", "v1.26", "v1.27"]
    "1.15": ["v1.23", "v1.24", "v1.25", "v1.26", "v1.27", "v1.28", "v1.29"]
    "1.16": ["v1.23", "v1.24", "v1.25", "v1.26", "v1.27", "v1.28", "v1.29", "v1.30"]
    "1.17": ["v1.23", "v1.24", "v1.25", "v1.26", "v1.27", "v1.28", "v1.29", "v1.30", "v1.31", "v1.32"]
    "1.18": ["v1.23", "v1.24", "v1.25", "v1.26", "v1.27", "v1.28", "v1.29", "v1.30", "v1.31", "v1.32", "v1.33"]
    "1.19": ["v1.23", "v1.24", "v1.25", "v1.26", "v1.27", "v1.28", "v1.29", "v1.30", "v1.31", "v1.32", "v1.33", "v1.34"]

  # 各 Kubernetes 版本允许的 kube-ovn 版本
  kubeovn_allowed_versions:
    v1.12: ["v1.23", "v1.24", "v1.25", "v1.26", "v1.27", "v1.28"]
    v1.13: ["v1.23", "v1.24", "v1.25", "v1.26", "v1.27", "v1.28"]
    v1.14: ["v1.23", "v1.24", "v1.25", "v1.26", "v1.27", "v1.28", "v1.29", "v1.30", "v1.31", "v1.32", "v1.33", "v1.34"]
    v1.15: ["v1.23", "v1.24", "v1.25", "v1.26", "v1.27", "v1.28", "v1.29", "v1.30", "v1.31", "v1.32", "v1.33", "v1.34"]

  # 各 Kubernetes 版本要求的 etcd 最低版本
  etcd_min_versions:
    v1.23: v3.2.18
    v1.24: v3.2.18
    v1.25: v3.2.18
    v1.26: v3.2.18
    v1.27: v3.2.18
    v1.28: v3.4.13-4
    v1.29: v3.4.13-4
    v1.30: v3.4.13-4
    v1.31: v3.5.11-0
    v1.31.14: v3.5.24-0
    v1.32: v3.5.11-0
    v1.32.10: v3.5.24-0
    v1.32.11: v3.5.24-0
    v1.33.0: v3.5.11-0
    v1.33.1: v3.5.11-0
    v1.33.2: v3.5.11-0
    v1.33.3: v3.5.11-0
    v1.33.4: v3.5.11-0
    v1.33.5: v3.5.11-0
    v1.33: v3.5.21-0
    v1.34.0: v3.5.21-0
    v1.34.1: v3.5.21-0
    v1.34: v3.5.24-0
```

### 参数说明

| 参数 | 说明 |
|------|------|
| `cluster_require.etcd_disk_wal_fysnc_duration_seconds` | etcd 磁盘 WAL fsync 在 99 分位的最大允许耗时（纳秒），用于性能边界检查。 |
| `cluster_require.allow_unsupported_distribution_setup` | 是否允许在未经明确支持的操作系统发行版上执行安装。 |
| `cluster_require.supported_os_distributions` | KubeKey 明确支持的操作系统发行版列表。 |
| `cluster_require.require_network_plugin` | 集群支持的网络插件列表，部署时会校验所选插件是否在此范围内。 |
| `cluster_require.kube_version_min_required` | 允许安装的最低 Kubernetes 版本。 |
| `cluster_require.minimal_master_memory_mb` | 每个控制平面节点所需的最低内存（MB）。 |
| `cluster_require.minimal_node_memory_mb` | 每个工作节点所需的最低内存（MB）。 |
| `cluster_require.require_etcd_deployment_type` | 支持的 etcd 部署方式：`internal`（集群内部署）或 `external`（外部已有集群）。 |
| `cluster_require.require_container_manager` | 支持的容器运行时列表：`docker`、`containerd`。 |
| `cluster_require.containerd_min_version_required` | 使用 containerd 时所需的最低版本号。 |
| `cluster_require.supported_architectures` | 支持的 CPU 架构列表。 |
| `cluster_require.min_kernel_version` | 节点所需的最低 Linux 内核版本。 |
| `cluster_require.calico_allowed_versions` | 按 Calico 版本列出的兼容 Kubernetes 版本矩阵。 |
| `cluster_require.cilium_allowed_versions` | 按 Cilium 版本列出的兼容 Kubernetes 版本矩阵。 |
| `cluster_require.kubeovn_allowed_versions` | 按 kube-ovn 版本列出的兼容 Kubernetes 版本矩阵。 |
| `cluster_require.etcd_min_versions` | 按 Kubernetes 版本列出的兼容 etcd 最低版本矩阵。 |

---

## 证书配置 (02-certs.yaml)

### 默认配置

```yaml
# 证书生成配置
# 将生成以下证书：
# - etcd 证书
# - Kubernetes 集群证书（替换 kubeadm 生成的 CA 证书，kubeadm 默认有效期仅 10 年）
# - 镜像仓库证书（用于 Harbor 等私有仓库）

# 证书链结构：
# CA (自签名或由用户提供)
# |- etcd.cert
# |- etcd.key
# |- etcd-client.cert
# |- etcd-client.key
# |
# |- image_registry.cert
# |- image_registry.key
# |- image-registry-client.cert
# |- image-registry-client.key
# |
# |- kubernetes.cert
# |- kubernetes.key
# |     |- kubeadm 使用此 CA 生成服务端证书（kube-apiserver 证书）
# |- front-proxy.cert
# |- front-proxy.key
# |
# |- image-registry.cert
# |- image-registry.key

certs:
  # CA 证书设置
  ca:
    # CA 证书有效期
    date: 87600h
    # 证书生成策略：
    # IfNotPresent：如果证书已存在则进行校验；仅在不存在时生成自签名证书
    gen_cert_policy: IfNotPresent
  kubernetes_ca:
    date: 87600h
    # 证书生成方式。支持值：IfNotPresent, Always
    gen_cert_policy: IfNotPresent
  front_proxy_ca:
    date: 87600h
    # 证书生成方式。支持值：IfNotPresent, Always
    gen_cert_policy: IfNotPresent
  # etcd 证书
  etcd:
    date: 87600h
    # 证书生成方式。支持值：IfNotPresent, Always
    gen_cert_policy: IfNotPresent
  # image_registry 证书
  image_registry:
    date: 87600h
    # 证书生成方式。支持值：IfNotPresent, Always
    gen_cert_policy: IfNotPresent
```

### 参数说明

| 参数 | 说明 |
|------|------|
| `certs` | 定义 KubeKey 需要生成或管理的各类证书。 |
| `certs.ca` | 集群根 CA 证书配置，影响 etcd、kubernetes 及镜像仓库等服务的 CA。 |
| `certs.ca.date` | CA 证书有效期，例如 `87600h` 代表 10 年。 |
| `certs.ca.gen_cert_policy` | CA 证书生成策略。`IfNotPresent` 表示仅在缺失时生成；`Always` 表示每次都重新生成。 |
| `certs.kubernetes_ca` | Kubernetes 集群 CA 证书配置。 |
| `certs.front_proxy_ca` | Kubernetes front-proxy CA 证书配置，用于聚合层（如 metrics-server）。 |
| `certs.etcd` | etcd 集群的 CA 及节点证书配置。 |
| `certs.image_registry` | 私有镜像仓库（如 Harbor）的 TLS 证书配置。 |

---

## 镜像仓库配置 (02-image_registry.yaml)

### 默认配置

```yaml
# 在线环境（image_registry.auth.registry 为空）下，镜像直接从原始仓库拉取到集群。
# 离线环境（image_registry.auth.registry 已设置）下，镜像先从源仓库拉取、缓存在本地，
# 然后推送到私有仓库（如 Harbor），最后由集群使用。

image_registry:
  # 指定要安装的镜像仓库类型。支持：harbor, docker-registry
  # 如果留空，则不会安装镜像仓库（假定已有可用仓库）
  type: ""
  # 镜像仓库的高可用虚拟 IP（VIP）
  ha_vip: ""
  harbor:
    http_port: 80
    https_port: 443
  # 镜像仓库认证设置
  auth:
    registry: >-
      {{- if .image_registry.type | empty | not -}}
        {{- if .image_registry.ha_vip | empty | not -}}
      {{ .image_registry.ha_vip }}
        {{- else if .groups.image_registry | default list | empty | not -}}
          {{- $internalIPv4 := index .hostvars (.groups.image_registry | default list | first) "internal_ipv4" | default "" -}}
          {{- $internalIPv6 := index .hostvars (.groups.image_registry | default list | first) "internal_ipv6" | default "" -}}
          {{- if $internalIPv4 | empty | not -}}
      {{ $internalIPv4 }}
          {{- else if $internalIPv6 | empty | not -}}
      {{ $internalIPv6 }}
          {{- end -}}
        {{- end -}}
      {{- else -}}
        {{- if .zone | eq "cn" -}}
      hub.kubesphere.com.cn
        {{- end -}}
      {{- end -}}
    username: >-
      {{- if .image_registry.type | empty | not -}}
      admin
      {{- end }}
    password: >-
      {{- if .image_registry.type | empty | not -}}
      Harbor12345
      {{- end }}
    skip_tls_verify: >-
      {{- .image_registry.type | empty -}}
    ca_file: >-
      {{- if .groups.image_registry | default list | empty | not -}}
      {{ .binary_dir }}/pki/root.crt
      {{- end -}}
    cert_file: >-
      {{- if .groups.image_registry | default list | empty | not -}}
      {{ .binary_dir }}/pki/image-registry-client.crt
      {{- end -}}
    key_file: >-
      {{- if .groups.image_registry | default list | empty | not -}}
      {{ .binary_dir }}/pki/image-registry-client.key
      {{- end -}}
  # docker.io 来源镜像所使用的镜像仓库端点
  dockerio_registry: >-
    {{- .image_registry.auth.registry | empty | ternary "docker.io" .image_registry.auth.registry -}}

  # quay.io 来源镜像所使用的镜像仓库端点
  quayio_registry: >-
    {{- .image_registry.auth.registry | empty | ternary "quay.io" .image_registry.auth.registry -}}

  # ghcr.io 来源镜像所使用的镜像仓库端点
  ghcrio_registry: >-
    {{- .image_registry.auth.registry | empty | ternary "ghcr.io" .image_registry.auth.registry -}}

  # registry.k8s.io 来源镜像所使用的镜像仓库端点
  k8sio_registry: >-
    {{- .image_registry.auth.registry | empty | ternary "registry.k8s.io" .image_registry.auth.registry -}}
```

### 参数说明

| 参数 | 说明 |
|------|------|
| `image_registry.type` | 要部署的镜像仓库类型：`harbor`、`docker-registry` 或 `""`（使用已有仓库）。 |
| `image_registry.ha_vip` | 部署 Harbor 等高可用仓库时使用的虚拟 IP。 |
| `image_registry.harbor.http_port` | Harbor HTTP 服务端口，默认为 `80`。 |
| `image_registry.harbor.https_port` | Harbor HTTPS 服务端口，默认为 `443`。 |
| `image_registry.auth.registry` | 集群实际使用的镜像仓库地址。若部署了仓库，会根据 `ha_vip` 或节点 IP 自动渲染；在线模式下为空；若 zone 为 `cn`，默认使用 `hub.kubesphere.com.cn`。 |
| `image_registry.auth.username` | 登录镜像仓库的用户名。部署 Harbor 时默认为 `admin`。 |
| `image_registry.auth.password` | 登录镜像仓库的密码。部署 Harbor 时默认为 `Harbor12345`。 |
| `image_registry.auth.skip_tls_verify` | 是否跳过 TLS 证书校验。部署 Harbor 时默认为 `false`。 |
| `image_registry.auth.ca_file` | 镜像仓库 CA 证书路径。 |
| `image_registry.auth.cert_file` | 客户端证书路径。 |
| `image_registry.auth.key_file` | 客户端私钥路径。 |
| `image_registry.dockerio_registry` | 替代 `docker.io` 的镜像仓库端点。若未配置私有仓库，则默认为 `docker.io`。 |
| `image_registry.quayio_registry` | 替代 `quay.io` 的镜像仓库端点。 |
| `image_registry.ghcrio_registry` | 替代 `ghcr.io` 的镜像仓库端点。 |
| `image_registry.k8sio_registry` | 替代 `registry.k8s.io` 的镜像仓库端点。 |

---

## 原生模式配置 (02-native.yaml)

### 默认配置

```yaml
# 基础操作系统配置设置
native:
  ntp:
    # NTP 服务器列表，用于系统时间同步
    servers:
      - "cn.pool.ntp.org"
    # 是否启用 NTP 服务
    enabled: true
  # 系统时区配置
  timezone: Asia/Shanghai

  # 为 inventory 中标记为 'nfs' 角色的节点配置 NFS 服务
  nfs:
    # 通过 NFS 共享的目录
    share_dir:
      - /share/
  # 是否将节点主机名设置为 inventory.hosts 中定义的值
  set_hostname: true
  # 需要在每个节点上更新的 DNS 配置文件列表
  # 这可以确保在安装集群期间，即使没有 DNS 服务，也能在本地解析关键主机名
  # 例如：
  #   [控制平面端点] -> 主节点 IP
  #   [当前安装节点的主机名] -> 对应节点 IP
  localDNS:
    - /etc/hosts
```

### 参数说明

| 参数 | 说明 |
|------|------|
| `native.ntp.servers` | 用于时间同步的 NTP 服务器地址列表。 |
| `native.ntp.enabled` | 是否启用 NTP 服务以保持节点间时间一致。 |
| `native.timezone` | 节点的系统时区，例如 `Asia/Shanghai`。 |
| `native.nfs.share_dir` | NFS 共享目录，供标记了 `nfs` 角色的节点使用。 |
| `native.set_hostname` | 安装时是否根据 inventory 中的定义自动设置节点主机名。 |
| `native.localDNS` | 本地 DNS 解析文件列表（如 `/etc/hosts`），用于在安装期间提供临时域名解析。 |

---

## Kubernetes 配置 (03-kubernetes.yaml)

### 默认配置

```yaml
kubernetes:
  # 要安装的集群名称
  cluster_name: kubekey

  # 内置 Kubernetes 镜像的仓库地址
  image_repository: >-
    {{ .image_registry.k8sio_registry }}{{ if .image_registry.auth.registry | empty | not }}/kubernetes{{ end }}

  # Pause/Sandbox 镜像配置
  sandbox_image:
    # Pause 镜像的仓库地址
    registry: >-
      {{ .image_registry.k8sio_registry }}
    # Pause 镜像的仓库路径
    repository: >-
      {{- .image_registry.auth.registry | empty | ternary "pause" "kubernetes/pause" -}}

  # Kubernetes 网络配置
  # kube-apiserver 参数
  apiserver:
    # kube-apiserver 监听端口
    port: 6443
    # 需要添加到 apiserver 证书中的额外 SAN 列表
    certSANs: []
    # kube-apiserver 额外启动参数
    extra_args:
      # 示例: feature-gates: ExpandCSIVolumes=true,CSIStorageCapacity=true,RotateKubeletServerCertificate=true

  # kube-controller-manager 参数
  controller_manager:
    # kube-controller-manager 额外启动参数
    extra_args:
      # 集群证书签名有效期
      cluster-signing-duration: 87600h
      # 示例: feature-gates: ExpandCSIVolumes=true,CSIStorageCapacity=true,RotateKubeletServerCertificate=true

  # kube-scheduler 参数
  scheduler:
    # kube-scheduler 额外启动参数
    extra_args:
      # 示例: feature-gates: ExpandCSIVolumes=true,CSIStorageCapacity=true,RotateKubeletServerCertificate=true

  # kube-proxy 参数
  kube_proxy:
    # 是否接管 kube-proxy 的部署
    manage:
      enabled: false
      # affinity:
      #   nodeAffinity:
      #     requiredDuringSchedulingIgnoredDuringExecution:
      #       nodeSelectorTerms:
      #         - matchExpressions:
      #             - key: kubernetes.io/os
      #               operator: In
      #               values:
      #                 - linux
    # kube-proxy 代理模式：ipvs, iptables
    mode: "ipvs"
    # kube-proxy 配置
    config:
      iptables:
        masqueradeAll: false
        masqueradeBit: 14
        minSyncPeriod: 0s
        syncPeriod: 30s

  # kubelet 服务参数
  kubelet:
    # 每个节点最大 Pod 数量
    max_pods: 110
    # 每个 Pod 的 PID 限制
    pod_pids_limit: 10000
    # feature_gates:
    # 容器日志文件的最大大小
    container_log_max_size: 5Mi
    # 保留的容器日志文件数量
    container_log_max_files: 3
    # extra_args:

  # 指定控制平面的稳定 IP 地址或 DNS 名称
  # 高可用场景下建议设置为 DNS 名称
  # 配置指引：
  # 1. 如果有可用的 DNS 名称：
  #    - 将 control_plane_endpoint 设置为该 DNS 名称，并确保它解析到所有控制平面节点 IP
  # 2. 如果没有可用的 DNS 名称：
  #    - 可以先设置一个 DNS 名称，稍后再添加解析
  #    - 在每个节点的本地 DNS 文件中添加解析，例如：
  #      {{ vip }} {{ control_plane_endpoint }}
  #    - 如果有 VIP（虚拟 IP）：
  #        在控制平面节点部署 kube-vip，将 VIP 映射到实际节点 IP
  #    - 如果没有 VIP：
  #        在工作节点部署 HAProxy，使用固定 IP（如 127.0.0.2）作为 VIP，并转发到所有控制平面节点 IP
  #
  # 非高可用场景（仅手动配置，不会自动安装）：
  # 可以将 VIP 设置为单个控制平面节点的 IP
  control_plane_endpoint:
    # 控制平面端点主机名或 IP
    host: lb.kubesphere.local
    # 控制平面端点端口，默认继承 apiserver 端口
    port: "{{ .kubernetes.apiserver.port }}"
    # 负载均衡类型：local, kube-vip, haproxy
    # 当 type 为 local 时，配置如下：
    #   - 控制平面节点：127.0.0.1 {{ .kubernetes.control_plane_endpoint.host }}
    #   - 工作节点：{{ .init_kubernetes_node }} {{ .kubernetes.control_plane_endpoint.host }}
    type: local
    local:
      # 使用 local 负载均衡时，可在此处指定外部负载均衡器地址
      # 注意：您必须自行搭建实际的负载均衡；此设置仅用于 DNS 解析
      address: ""
    kube_vip:
      # 节点网卡的 IP 地址（例如 "eth0"）
      address: ""
      # 支持的模式：ARP, BGP
      mode: ARP
      image:
        # kube-vip 镜像仓库
        registry: >-
          {{ .image_registry.dockerio_registry }}
        # kube-vip 镜像路径
        repository: plndr/kube-vip
        # kube-vip 镜像标签
        tag: v0.7.2
    haproxy:
      # 节点 "lo"（回环）接口上的 IP 地址
      address: 127.0.0.1
      # HAProxy 健康检查端口
      health_port: 8081
      image:
        # HAProxy 镜像仓库
        registry: >-
          {{ .image_registry.dockerio_registry }}
        # HAProxy 镜像路径
        repository: library/haproxy
        # HAProxy 镜像标签
        tag: 2.9.6-alpine

  # 是否自动续期 Kubernetes 证书
  certs:
    # 提供 Kubernetes CA 文件有三种方式：
    # 1. kubeadm：将 ca_cert 和 ca_key 留空，kubeadm 会自动生成。有效期 10 年，不可更改。
    # 2. kubekey：将 ca_cert 设置为 {{ .binary_dir }}/pki/ca.cert，ca_key 设置为 {{ .binary_dir }}/pki/ca.key。
    #    这些证书由 kubekey 生成，有效期 10 年，可通过 cert.ca_date 更新。
    # 3. 自定义：手动指定 ca_cert 和 ca_key 的绝对路径以使用您自己的 CA 文件。
    #
    # 若要使用自定义 CA 文件，请在下方填写绝对路径。
    # 如果留空，则使用默认行为（kubeadm 或 kubekey）。
    ca_cert: ""
    ca_key: ""
    # 以下字段用于 Kubernetes front-proxy CA 证书和私钥。
    # 若要使用自定义 front-proxy CA 文件，请在下方填写绝对路径。
    # 如果留空，则使用默认行为。
    front_proxy_cert: ""
    front_proxy_key: ""
    # 自动续期服务证书（注意：CA 证书不支持自动续期）
    renew: true

  # 通过补丁文件对 Kubernetes 组件进行自定义配置
  patches: []
    # 补丁通过包含补丁文件的目录应用。
    # - name: kube-apiserver0+merge.yaml
    #   path: /etc/kubernetes/kube-apiserver-patch.yaml
    #   content: |
    #     apiVersion: v1
    #     kind: Pod
    #     spec:
    #       containers:
    #         - name: kube-apiserver
    #           command:
    #             - kube-apiserver
    #             - --service-account-issuer=https://kubernetes.default.svc.cluster.local
    #             - --service-account-jwks-uri=https://kubernetes.default.svc.cluster.local/openid/v1/jwks
    #   # 目录中包含名为 "target[suffix][+patchtype].extension" 的文件。
    #   # "target" 可以是：kube-apiserver, kube-controller-manager, kube-scheduler,
    #   #                       etcd, kubeletconfiguration, corednsdeployment
    #   # "patchtype" 可以是：strategic（默认）, merge, json
    #   # "extension" 可以是：yaml 或 json
    #   # "suffix"（可选）决定应用顺序（按字母数字排序）。
    #   # 示例：
    #   #   kube-apiserver+merge.yaml          # kube-apiserver 的 merge 补丁
    #   #   kube-apiserver001+strategic.yaml   # 带排序后缀的 strategic 补丁
    #   #   kube-controller-manager+merge.yaml
    #   #   kube-scheduler+json.yaml
    #   #   kubeletconfiguration+merge.yaml

  # kubeadm init 要跳过的阶段
  skip_phases: []
    # - addon/kube-proxy
```

### 参数说明

| 参数 | 说明 |
|------|------|
| `kubernetes.cluster_name` | Kubernetes 集群的名称。 |
| `kubernetes.image_repository` | 拉取 Kubernetes 核心组件镜像的仓库前缀，默认通过 `k8sio_registry` 计算。 |
| `kubernetes.sandbox_image` | pause（sandbox）容器镜像的完整配置，包含 registry 和 repository。 |
| `kubernetes.apiserver.port` | kube-apiserver 的 HTTPS 监听端口，默认 `6443`。 |
| `kubernetes.apiserver.certSANs` | 需要加入到 kube-apiserver 证书 Subject Alternative Names 中的额外地址列表。 |
| `kubernetes.apiserver.extra_args` | 传递给 kube-apiserver 的额外命令行参数。 |
| `kubernetes.controller_manager.extra_args` | 传递给 kube-controller-manager 的额外命令行参数。 |
| `kubernetes.scheduler.extra_args` | 传递给 kube-scheduler 的额外命令行参数。 |
| `kubernetes.kube_proxy.manage.enabled` | 是否由 KubeKey 接管 kube-proxy 的部署（而非 kubeadm 默认部署）。 |
| `kubernetes.kube_proxy.mode` | kube-proxy 的工作模式，`ipvs` 或 `iptables`。 |
| `kubernetes.kube_proxy.config.iptables` | iptables 模式下的详细配置项。 |
| `kubernetes.kubelet.max_pods` | 单个节点允许调度的最大 Pod 数。 |
| `kubernetes.kubelet.pod_pids_limit` | 每个 Pod 可以使用的最大 PID 数量。 |
| `kubernetes.kubelet.container_log_max_size` | 单个容器日志文件在轮转前的最大大小。 |
| `kubernetes.kubelet.container_log_max_files` | 保留的旧容器日志文件数量。 |
| `kubernetes.control_plane_endpoint.host` | 控制平面的稳定访问地址（IP 或 DNS）。 |
| `kubernetes.control_plane_endpoint.port` | 控制平面端点端口。 |
| `kubernetes.control_plane_endpoint.type` | 负载均衡实现方式：`local`（本地解析）、`kube-vip`（基于 VIP）、`haproxy`。 |
| `kubernetes.control_plane_endpoint.local.address` | 使用 `local` 模式时，可指定外部负载均衡器地址仅用于解析。 |
| `kubernetes.control_plane_endpoint.kube_vip.address` | kube-vip 绑定的网卡名称或 IP。 |
| `kubernetes.control_plane_endpoint.kube_vip.mode` | kube-vip 的工作模式：`ARP` 或 `BGP`。 |
| `kubernetes.control_plane_endpoint.kube_vip.image` | kube-vip 容器镜像配置。 |
| `kubernetes.control_plane_endpoint.haproxy.address` | HAProxy 在本机回环接口上监听的地址。 |
| `kubernetes.control_plane_endpoint.haproxy.health_port` | HAProxy 健康检查端口。 |
| `kubernetes.control_plane_endpoint.haproxy.image` | HAProxy 容器镜像配置。 |
| `kubernetes.certs.ca_cert` | 自定义 Kubernetes CA 证书路径（留空则使用 kubeadm/kubekey 生成）。 |
| `kubernetes.certs.ca_key` | 自定义 Kubernetes CA 私钥路径。 |
| `kubernetes.certs.front_proxy_cert` | 自定义 front-proxy CA 证书路径。 |
| `kubernetes.certs.front_proxy_key` | 自定义 front-proxy CA 私钥路径。 |
| `kubernetes.certs.renew` | 是否自动续期集群中的服务证书（CA 本身不会自动续期）。 |
| `kubernetes.patches` | 以文件或内联内容方式对 Kubernetes 静态 Pod 或组件配置打补丁。 |
| `kubernetes.skip_phases` | `kubeadm init` 执行过程中要显式跳过的阶段列表。 |

---

## CNI 网络插件配置 (04-cni.yaml)

### 默认配置

```yaml
cni:
  # 要使用的 CNI 插件类型
  # 指定要为集群安装的网络插件。支持：calico, cilium, flannel, hybridnet, kubeovn, other
  type: calico
  # 集群 Pod 的完整 IP 地址池。支持 IPv4、IPv6 及双栈
  pod_cidr: 10.233.64.0/18
  # 每个节点上 Pod 分配的 IPv4 子网掩码长度，决定每个节点可分配的 Pod IP 数量
  ipv4_mask_size: 24
  # 每个节点上 Pod 分配的 IPv6 子网掩码长度
  ipv6_mask_size: 64
  # 集群 Service 的完整 IP 地址池。支持 IPv4、IPv6 及双栈
  service_cidr: 10.233.0.0/18

  # 多 CNI 类型配置。支持：multus, none
  multi_cni: "none"
  # 为 Pod 提供多网络接口的网络增强插件（Multus）
  multus:
    image:
      # Multus 镜像仓库
      registry: >-
        {{ .image_registry.ghcrio_registry }}
      # Multus 镜像路径
      repository: k8snetworkplumbingwg/multus-cni
      # tag: v4.3.0
```

### 参数说明

| 参数 | 说明 |
|------|------|
| `cni.type` | 集群网络插件类型，可选 `calico`、`cilium`、`flannel`、`hybridnet`、`kubeovn`、`other`。 |
| `cni.pod_cidr` | 整个集群 Pod 网络的 CIDR 网段。 |
| `cni.ipv4_mask_size` | 为每个节点划分的 Pod IPv4 子网掩码长度。例如在 `/18` 大网段中使用 `/24` 掩码，每个节点可获得约 256 个 Pod IP。 |
| `cni.ipv6_mask_size` | 为每个节点划分的 Pod IPv6 子网掩码长度。 |
| `cni.service_cidr` | 整个集群 Service 网络的 CIDR 网段。 |
| `cni.multi_cni` | 是否启用多 CNI 支持。`multus` 表示启用 Multus，`none` 表示不启用。 |
| `cni.multus.image` | Multus CNI 容器镜像的配置（registry、repository、tag）。 |

---

## 容器运行时 (CRI) 配置 (04-cri.yaml)

### 默认配置

```yaml
cri:
  # 容器运行时类型。支持：containerd, docker
  container_manager: containerd
  # 容器运行时的 Cgroup 驱动。支持：systemd, cgroupfs
  cgroup_driver: systemd
    # tag: "3.9"
  # 所选容器运行时的 CRI socket 端点
  cri_socket: >-
    {{- if .cri.container_manager | eq "containerd" -}}
    unix:///var/run/containerd/containerd.sock
    {{- else if and (.cri.container_manager | eq "docker") (.kubernetes.kube_version | semverCompare ">=v1.24.0") -}}
    unix:///var/run/cri-dockerd.sock
    {{- end -}}

  # CRI 的镜像仓库配置，包括镜像加速、非安全仓库及认证信息
  registry:
    # 镜像仓库加速地址列表
    mirrors: ["https://registry-1.docker.io"]
    # 不安全的镜像仓库列表（允许 HTTP 访问）
    insecure_registries: []
    # 私有镜像仓库认证信息列表
    auths: []
    # 配置示例：
    # auths:
    #   - registry: docker.io
    #     username: MyDockerAccount
    #     password: my_password
    #     skip_tls_verify: true
    #     ca_cert: /etc/docker/certs.d/docker.io/ca.crt
    #     cert_file: /etc/docker/certs.d/docker.io/cert.crt
    #     key_file: /etc/docker/certs.d/docker.io/key.crt

  # Docker 配置
  docker:
    # Docker daemon 配置
    daemon:
      # Docker 数据根目录
      data-root: "{{ .cri.docker.data_root | default \"/var/lib/docker\" }}"
      # 容器日志配置
      log-opts:
        # 单个日志文件最大大小
        max-size: "{{ .kubernetes.kubelet.container_log_max_size | default \"5Mi\" | toLowerByteUnit }}"
        # 保留的日志文件数量
        max-file: "{{ .kubernetes.kubelet.container_log_max_files | default 3  | toString | toJson }}"
      # 是否启用 live-restore
      live-restore: true
      # 容器 exec 选项
      exec-opts:
        - "native.cgroupdriver={{ .cri.cgroup_driver | default \"systemd\" }}"

  # containerd 配置
  containerd:
    config:
      # containerd 数据根目录
      root: "{{ .cri.containerd.data_root | default \"/var/lib/containerd\" }}"
      # containerd 配置文件版本
      version: 2
      # containerd 运行状态目录
      state: "/run/containerd"
      grpc:
        address: "/run/containerd/containerd.sock"
        uid: 0
        gid: 0
        max_recv_message_size: 16777216
        max_send_message_size: 16777216
      ttrpc:
        address: ""
        uid: 0
        gid: 0
      debug:
        address: ""
        uid: 0
        gid: 0
        level: ""
      metrics:
        address: ""
        grpc_histogram: false
      cgroup:
        path: ""
      timeouts:
        "io.containerd.timeout.shim.cleanup": "5s"
        "io.containerd.timeout.shim.load": "5s"
        "io.containerd.timeout.shim.shutdown": "3s"
        "io.containerd.timeout.task.state": "2s"
      plugins:
        "io.containerd.grpc.v1.cri":
          containerd:
            runtimes:
              runc:
                runtime_type: "io.containerd.runc.v2"
                options:
                  SystemdCgroup: "{{ .cri.cgroup_driver | eq \"systemd\" }}"
          cni:
            bin_dir: "/opt/cni/bin"
            conf_dir: "/etc/cni/net.d"
            max_conf_num: 1
            conf_template: ""

```

### 参数说明

| 参数 | 说明 |
|------|------|
| `cri.container_manager` | 容器运行时管理器，可选 `containerd` 或 `docker`。 |
| `cri.cgroup_driver` | 容器运行时使用的 Cgroup 驱动，推荐 `systemd`（与大多数现代操作系统 init 系统兼容）。 |
| `cri.cri_socket` | 当前容器运行时对应的 CRI socket 路径，会根据 `container_manager` 和 Kubernetes 版本自动选择。 |
| `cri.registry.mirrors` | 镜像加速地址，可配置国内镜像源以提高拉取速度。 |
| `cri.registry.insecure_registries` | 允许使用 HTTP（非 HTTPS）访问的镜像仓库地址列表。 |
| `cri.registry.auths` | 私有镜像仓库的认证信息列表，包含用户名、密码及可选的 TLS 证书配置。 |
| `cri.docker.daemon` | Docker daemon 配置项，映射为 `/etc/docker/daemon.json`。 |
| `cri.docker.daemon.data-root` | Docker 数据根目录。 |
| `cri.docker.daemon.log-opts.max-size` | 单个容器日志文件的最大大小。 |
| `cri.docker.daemon.log-opts.max-file` | 保留的旧容器日志文件数量。 |
| `cri.docker.daemon.live-restore` | 是否启用 Docker live-restore。 |
| `cri.docker.daemon.exec-opts` | Docker exec 选项列表，例如 cgroup 驱动。 |
| `cri.containerd.config` | containerd 配置，映射为 `/etc/containerd/config.toml`。 |
| `cri.containerd.config.root` | containerd 数据持久化根目录。 |
| `cri.containerd.config.version` | containerd 配置文件版本。 |
| `cri.containerd.config.state` | containerd 运行状态目录。 |
| `cri.containerd.config.grpc.address` | containerd gRPC socket 地址。 |
| `cri.containerd.config.grpc.uid` | containerd gRPC socket 属主 UID。 |
| `cri.containerd.config.grpc.gid` | containerd gRPC socket 属主 GID。 |
| `cri.containerd.config.grpc.max_recv_message_size` | containerd gRPC 最大接收消息大小。 |
| `cri.containerd.config.grpc.max_send_message_size` | containerd gRPC 最大发送消息大小。 |
| `cri.containerd.config.ttrpc` | containerd TTRPC 配置。 |
| `cri.containerd.config.debug` | containerd 调试配置。 |
| `cri.containerd.config.metrics` | containerd metrics 配置。 |
| `cri.containerd.config.cgroup` | containerd cgroup 配置。 |
| `cri.containerd.config.timeouts` | containerd 各操作超时时间。 |
| `cri.containerd.config.plugins` | containerd CRI 插件配置，包含运行时与 CNI 设置。 |
| `cri.containerd.config.plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc.runtime_type` | runc 运行时类型。 |
| `cri.containerd.config.plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc.options.SystemdCgroup` | runc 是否使用 systemd cgroup。 |
| `cri.containerd.config.plugins."io.containerd.grpc.v1.cri".cni` | CNI 插件配置，包括二进制目录、配置目录等。 |

---

## etcd 配置 (04-etcd.yaml)

### 默认配置

```yaml
# etcd 服务配置
etcd:
  # etcd 部署类型：
  # - external：使用外部 etcd 集群
  # - internal：在集群内部署 etcd 静态 Pod
  deployment_type: external
  image:
    # etcd 镜像仓库
    registry: >-
      {{ .image_registry.dockerio_registry }}
    # etcd 镜像路径
    repository: kubesphere/etcd
    # etcd 镜像标签
    tag: "{{ .etcd.etcd_version }}"
  # etcd 客户端端口
  port: 2379
  # etcd 节点间通信端口
  peer_port: 2380
  # etcd 服务环境变量
  env:
    # Leader 选举超时时间（毫秒）
    election_timeout: 5000
    # 心跳间隔（毫秒）
    heartbeat_interval: 250
    # 数据压缩保留时长（小时）
    compaction_retention: 8
    # 触发快照的事务数
    snapshot_count: 10000
    # etcd 数据目录
    data_dir: /var/lib/etcd
    # etcd 集群 token
    token: k8s_etcd
    # metrics: basic
    # quota_backend_bytes: 100
    # max_request_bytes: 100
    # max_snapshots: 100
    # max_wals: 5
    # log_level: info
    # unsupported_arch: arm64
  # etcd 备份配置
  backup:
    # 备份文件存放目录
    backup_dir: /var/lib/etcd-backup
    # 保留的备份文件数量
    keep_backup_number: 5
    # 备份脚本名称
    etcd_backup_script: "backup.sh"
    # 定时备份周期（systemd OnCalendar 格式）
    on_calendar: "*-*-* *:00/30:00"
  # 是否启用 etcd 性能调优
  performance: false
  # 是否启用 etcd 流量优先级控制
  traffic_priority: false
  # CA 证书路径
  ca_file: >-
    {{ .binary_dir }}/pki/root.crt
  # 服务端证书路径
  server_cert_file: >-
    {{ .binary_dir }}/pki/etcd-{{ "{{ . }}" }}.crt
  # 服务端私钥路径
  server_key_file: >-
    {{ .binary_dir }}/pki/etcd-{{ "{{ . }}" }}.key
  # 客户端证书路径
  client_cert_file: >-
    {{ .binary_dir }}/pki/etcd-client.crt
  # 客户端私钥路径
  client_key_file: >-
    {{ .binary_dir }}/pki/etcd-client.key
```

### 参数说明

| 参数 | 说明 |
|------|------|
| `etcd.deployment_type` | etcd 部署模式。`external` 使用外部已有 etcd；`internal` 在 Kubernetes 集群中以静态 Pod 形式部署。 |
| `etcd.image` | etcd 容器镜像配置（registry、repository、tag）。 |
| `etcd.port` | etcd 客户端请求端口，默认 `2379`。 |
| `etcd.peer_port` | etcd 节点间对等通信端口，默认 `2380`。 |
| `etcd.env.election_timeout` | Leader 选举超时时间，单位毫秒。 |
| `etcd.env.heartbeat_interval` | 节点间心跳间隔，单位毫秒。 |
| `etcd.env.compaction_retention` | 自动数据压缩保留的数据历史时长，单位小时。 |
| `etcd.env.snapshot_count` | 触发一次快照所需的事务数。 |
| `etcd.env.data_dir` | etcd 数据持久化目录。 |
| `etcd.env.token` | 集群初始化时的共享 token，用于成员发现。 |
| `etcd.backup.backup_dir` | etcd 备份文件存放目录。 |
| `etcd.backup.keep_backup_number` | 本地保留的备份副本数。 |
| `etcd.backup.etcd_backup_script` | 执行的备份脚本名称。 |
| `etcd.backup.on_calendar` | 基于 systemd timer 的定时备份周期格式，例如每 30 分钟执行一次。 |
| `etcd.performance` | 是否启用 etcd 性能调优参数。 |
| `etcd.traffic_priority` | 是否启用 etcd 网络流量优先级控制。 |
| `etcd.ca_file` | etcd CA 证书文件路径。 |
| `etcd.server_cert_file` | etcd 服务端证书路径。 |
| `etcd.server_key_file` | etcd 服务端私钥路径。 |
| `etcd.client_cert_file` | etcd 客户端证书路径。 |
| `etcd.client_key_file` | etcd 客户端私钥路径。 |

---

## DNS 配置 (05-dns.yaml)

### 默认配置

```yaml
dns:
  # ====== 集群内 DNS 服务配置 ======
  # 集群内服务和 Pod 使用的 DNS 域后缀
  domain: cluster.local

  # NodeLocalDNS Pod 配置
  nodelocaldns:
    # 是否启用 NodeLocalDNS
    enabled: true
    # NodeLocalDNS 在每个节点上绑定的 IP 地址
    ip: 169.254.25.10
    # NodeLocalDNS 镜像配置
    image:
      # NodeLocalDNS 镜像仓库
      registry: >-
        {{ .image_registry.k8sio_registry }}
      # NodeLocalDNS 镜像路径
      repository: >-
        dns/k8s-dns-node-cache
      # tag: 1.24.0

  # CoreDNS Pod 配置
  coredns:
    # 集群 DNS 服务的 IP 地址
    ip: >-
      {{ index (.cni.service_cidr | ipInCIDR) 2 }}
    # CoreDNS 镜像配置
    image:
      # CoreDNS 镜像仓库
      registry: >-
        {{ .image_registry.k8sio_registry }}
      # CoreDNS 镜像路径
      repository: >-
        coredns
      # tag: v1.11.1
    # 自定义 hosts 条目
    dns_etc_hosts: []
    # DNS 区域匹配配置
    zone_configs:
      # 每个条目定义要匹配的 DNS 区域，默认端口为 53
      # ".": 匹配所有 DNS 区域
      # "example.com": 使用端口 53 的 DNS 服务器匹配 *.example.com
      # "example.com:54": 使用端口 54 的 DNS 服务器匹配 *.example.com
      - zones: [".:53"]
        additional_configs:
          - errors
          - ready
          - prometheus :9153
          - loop
          - reload
          - loadbalance
        cache: 30
        kubernetes:
          zones:
            - "{{ .dns.domain }}"
        # 如需内部 DNS 消息重写，可在此配置
        # rewrite:
        #   - rule: continue
        #     field: name
        #     type: exact
        #     value: "example.com example2.com"
        #     options: ""
        forward:
          # DNS 查询转发规则
          - from: "."
            # 转发目标端点。'to' 语法允许指定协议
            to: ["/etc/resolv.conf"]
            # 要从转发中排除的域名
            except: []
            # 即使原始请求是 UDP，也使用 TCP 进行转发
            force_tcp: false
            # 优先使用 UDP 转发；如果响应被截断则回退到 TCP
            prefer_udp: false
            # 连续健康检查失败次数上限，超过则将上游标记为不可用
            # max_fails: 2
            # 缓存连接过期时间
            # expire: 10s
            # 安全连接的 TLS 属性可在此处设置
            # tls:
            #   cert_file: ""
            #   key_file: ""
            #   ca_file: ""
            # tls_servername: ""
            # 选择上游服务器的策略：random（默认）、round_robin、sequential
            # policy: "random"
            # 上游服务器健康检查配置
            # health_check: ""
            # 允许的最大并发 DNS 查询数
            max_concurrent: 1000
```

### 参数说明

| 参数 | 说明 |
|------|------|
| `dns.domain` | 集群的默认 DNS 域后缀（如 `cluster.local`）。 |
| `dns.nodelocaldns.enabled` | 是否启用 NodeLocalDNS，提升集群 DNS 解析性能并降低 CoreDNS 负载。 |
| `dns.nodelocaldns.ip` | NodeLocalDNS DaemonSet 在每个节点上绑定的本地链路 IP（默认 `169.254.25.10`）。 |
| `dns.nodelocaldns.image` | NodeLocalDNS 容器镜像配置。 |
| `dns.coredns.ip` | CoreDNS 集群服务 IP，通常取自 Service CIDR 的第 3 个地址。 |
| `dns.coredns.image` | CoreDNS 容器镜像配置。 |
| `dns.coredns.dns_etc_hosts` | 向 CoreDNS 注入的自定义 `/etc/hosts` 格式条目。 |
| `dns.coredns.zone_configs` | CoreDNS Corefile 的区域配置列表，可定义匹配的域、缓存、重写、转发等规则。 |
| `dns.coredns.zone_configs[].zones` | 该区域规则匹配的 DNS 域及端口列表。 |
| `dns.coredns.zone_configs[].additional_configs` | 附加的 CoreDNS 插件指令列表（如 `errors`、`ready`、`prometheus`、`loop`、`reload`、`loadbalance`）。 |
| `dns.coredns.zone_configs[].cache` | DNS 记录缓存时间（秒）。 |
| `dns.coredns.zone_configs[].kubernetes.zones` | 由 CoreDNS Kubernetes 插件提供解析的集群 DNS 域。 |
| `dns.coredns.zone_configs[].forward` | 无法本地解析的查询的转发规则列表。 |
| `dns.coredns.zone_configs[].forward[].from` | 需要转发解析的源域。 |
| `dns.coredns.zone_configs[].forward[].to` | 上游 DNS 服务器或解析文件地址列表。 |
| `dns.coredns.zone_configs[].forward[].except` | 不转发给上游的例外域名列表。 |
| `dns.coredns.zone_configs[].forward[].force_tcp` | 是否强制使用 TCP 向上游转发查询。 |
| `dns.coredns.zone_configs[].forward[].prefer_udp` | 是否优先使用 UDP 向上游转发。 |
| `dns.coredns.zone_configs[].forward[].max_concurrent` | 该转发规则允许的最大并发查询数。 |

---

## 存储类配置 (05-storage_class.yaml)

### 默认配置

```yaml
# Kubernetes 持久存储集成的存储类配置
storage_class:
  # 本地存储类配置
  local:
    enabled: true    # 是否启用本地存储类
    default: true    # 是否设为默认存储类
    path: /var/openebs/local  # 本地存储卷的主机路径

  # NFS 存储类配置
  nfs:
    # 确保 k8s_cluster 组中的每个节点都已安装 nfs-utils
    enabled: false   # 是否启用 NFS 存储类
    default: false   # 是否设为默认存储类
    # NFS 服务器地址
    server: >-
      {{ .groups.nfs | default list | first }}
    path: /share/kubernetes  # NFS 导出路径，用于持久卷
```

### 参数说明

| 参数 | 说明 |
|------|------|
| `storage_class.local.enabled` | 是否创建并启用基于节点本地磁盘的 `local` StorageClass。 |
| `storage_class.local.default` | 是否将 `local` StorageClass 标记为集群默认存储类。 |
| `storage_class.local.path` | 本地存储卷在节点上的实际主机路径。 |
| `storage_class.nfs.enabled` | 是否创建并启用基于 NFS 的 StorageClass。 |
| `storage_class.nfs.default` | 是否将 NFS StorageClass 标记为集群默认存储类。 |
| `storage_class.nfs.server` | NFS 服务端地址，默认取 inventory 中 `nfs` 组的第一个节点。 |
| `storage_class.nfs.path` | NFS 服务器上导出的共享目录路径。 |

---

## 下载配置 (10-download.yaml)

### 默认配置

```yaml
download:
  # 下载超时时间
  timeout: 300s
  # 中国区文件存储默认主机
  cn_host: kubekey.pek3b.qingstor.com
  # 目标操作系统
  os: linux
  # 目标 CPU 架构列表
  arch: [ "amd64" ]
  # KubeKey 离线制品包文件路径
  artifact_file: ""
  # 制品包的 MD5 校验文件
  artifact_md5: ""
  # 是否在线下载软件包、Helm Chart、容器镜像等
  # 如果所有必需的镜像和包都已在本地可用，且不需要与远程仓库校验，则设为 false
  fetch: true
  # 各组件的下载 URL 模板
  artifact_url:
    # etcd 二进制包
    etcd: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      github.com/etcd-io/etcd/releases/download/{{ "{{ .version }}" }}/etcd-{{ "{{ .version }}" }}-linux-{{ "{{ .arch }}" }}.tar.gz
    # kubelet 二进制
    kubelet: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      dl.k8s.io/release/{{ "{{ .version }}" }}/bin/linux/{{ "{{ .arch }}" }}/kubelet
    # kubeadm 二进制
    kubeadm: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      dl.k8s.io/release/{{ "{{ .version }}" }}/bin/linux/{{ "{{ .arch }}" }}/kubeadm
    # kubectl 二进制
    kubectl: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      dl.k8s.io/release/{{ "{{ .version }}" }}/bin/linux/{{ "{{ .arch }}" }}/kubectl
    # CNI 插件包
    cni_plugins: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      github.com/containernetworking/plugins/releases/download/{{ "{{ .version }}" }}/cni-plugins-linux-{{ "{{ .arch }}" }}-{{ "{{ .version }}" }}.tgz
    # Helm 二进制包
    helm: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      get.helm.sh/helm-{{ "{{ .version }}" }}-linux-{{ "{{ .arch }}" }}.tar.gz
    # crictl 工具包
    crictl: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      github.com/kubernetes-sigs/cri-tools/releases/download/{{ "{{ .version }}" }}/crictl-{{ "{{ .version }}" }}-linux-{{ "{{ .arch }}" }}.tar.gz
    # Docker 二进制包
    docker: >-
      https://mirrors.aliyun.com/docker-ce/linux/static/stable/
      {{- "{{ if eq .arch \"amd64\" }}x86_64{{ else if eq .arch \"arm64\" }}aarch64{{ end }}" -}}
      /docker-{{ "{{ .version }}" }}.tgz
    # cri-dockerd 包
    cridockerd: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      github.com/Mirantis/cri-dockerd/releases/download/{{ "{{ .version }}" }}/cri-dockerd-{{ "{{ .version | default \"\" | trimPrefix \"v\" }}" }}.{{ "{{ .arch }}" }}.tgz
    # containerd 二进制包
    containerd: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      github.com/containerd/containerd/releases/download/{{ "{{ .version }}" }}/containerd-{{ "{{ .version | default \"\" | trimPrefix \"v\" }}" }}-linux-{{ "{{ .arch }}" }}.tar.gz
    # runc 二进制
    runc: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      github.com/opencontainers/runc/releases/download/{{ "{{ .version }}" }}/runc.{{ "{{ .arch }}" }}
    # calicoctl 二进制
    calicoctl: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      github.com/projectcalico/calico/releases/download/{{ "{{ .version }}" }}/calicoctl-linux-{{ "{{ .arch }}" }}
    # docker-registry 包
    docker_registry: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      docker.io/registry/{{ "{{ .version }}" }}/docker-registry-{{ "{{ .version }}" }}-linux-{{ "{{ .arch }}" }}.tgz
    # docker-compose 二进制
    docker_compose: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      github.com/docker/compose/releases/download/{{ "{{ .version }}" }}/docker-compose-linux-
      {{- "{{ if eq .arch \"amd64\" }}x86_64{{ else if eq .arch \"arm64\" }}aarch64{{ end }}" -}}
    # Harbor 离线安装包
    harbor: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      github.com/
      {{- "{{ if eq .arch \"amd64\" }}goharbor/harbor{{ else if eq .arch \"arm64\" }}kubesphere/kubekey{{ end }}" -}}
      /releases/download/
      {{- "{{ if eq .arch \"amd64\" }}{{ .version }}{{ else if eq .arch \"arm64\" }}iso-latest{{ end }}" -}}
      /harbor-offline-installer-{{ "{{ .version }}" }}.tgz
    # keepalived 包
    keepalived: >-
      https://{{ .download.cn_host}}/osixia/keepalived/{{ "{{ .version }}" }}/keepalived-{{ "{{ .version }}" }}-linux-{{ "{{ .arch }}" }}.tgz
    # Helm Chart 包：Calico
    calico: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      github.com/projectcalico/calico/releases/download/{{ "{{ .version }}" }}/tigera-operator-{{ "{{ .version }}" }}.tgz
    # Helm Chart 包：Cilium
    cilium: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      helm.cilium.io/cilium-{{ "{{ .version }}" }}.tgz
    # Helm Chart 包：Flannel
    flannel: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      github.com/flannel-io/flannel/releases/download/{{ "{{ .version }}" }}/flannel.tgz
    # Helm Chart 包：Kube-OVN
    kubeovn: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      kubeovn.github.io/kube-ovn/kube-ovn-{{ "{{ .version }}" }}.tgz
    # Helm Chart 包：Hybridnet
    hybridnet: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      github.com/alibaba/hybridnet/releases/download/helm-chart-{{ "{{ .version }}" }}/hybridnet-{{ "{{ .version }}" }}.tgz
    # Helm Chart 包：OpenEBS LocalPV Provisioner
    localpv_provisioner: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      openebs.github.io/dynamic-localpv-provisioner/localpv-provisioner-{{ "{{ .version }}" }}.tgz
    # Helm Chart 包：NFS Subdir External Provisioner
    nfs_subdir_external_provisioner: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      github.com/kubernetes-sigs/nfs-subdir-external-provisioner/releases/download/nfs-subdir-external-provisioner-{{ "{{ .version }}" }}/nfs-subdir-external-provisioner-{{ "{{ .version }}" }}.tgz
    # Helm Chart 包：Spiderpool
    spiderpool: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      github.com/spidernet-io/spiderpool/releases/download/{{ "{{ .version }}" }}/spiderpool-{{ "{{ .version | default \"\" | trimPrefix \"v\" }}" }}.tgz
  # 额外需要打包的工具
  tools:
    # ORAS 工具包
    oras: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      github.com/oras-project/oras/releases/download/v1.3.0/oras_1.3.0_linux_{{ "{{ \"{{ .arch }}\" }}" }}.tar.gz
    # nerdctl 工具包
    nerdctl: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      github.com/containerd/nerdctl/releases/download/v2.2.1/nerdctl-2.2.1-linux-{{ "{{ \"{{ .arch }}\" }}" }}.tar.gz
  # 额外的 Helm Chart 列表
  charts: []
  # charts:
  #   # 仓库 Chart
  #   - url: flannel@https://flannel-io.github.io/flannel/
  #     version: 0.28.1
  #   # OCI Chart
  #   - url: oci://ghcr.io/flannel-io/flannel-chart
  #     version: 0.28.1
  # 操作系统 ISO/软件包列表
  iso:
    - "almalinux-9.0-rpms"
    - "centos-8-rpms"
    - "debian-10-debs"
    - "debian-11-debs"
    - "kylin-v10SP3-2403-rpms"
    - "kylin-v10SP3-rpms"
    - "kylin-v10SP2-rpms"
    - "kylin-v10SP1-rpms"
    - "ubuntu-18.04-debs"
    - "ubuntu-20.04-debs"
    - "ubuntu-22.04-debs"
    - "ubuntu-24.04-debs"
  # CNI 类型列表（用于下载）
  cni:
    type: []
  # 下载配置中的存储类开关
  storage_class:
    local:
      enabled: true
    nfs:
      enabled: false
  # 根据 K8s 版本动态决定的容器运行时列表（用于下载）
  cri:
    container_manager: >
      {{- $container_manager := list }}
      {{- range .download.kubernetes.kube_version }}
        {{- if and (. | semverCompare "<v1.24.0") ($container_manager | has "docker" | not) }}
          {{- $container_manager = append $container_manager "docker" }}
        {{- else if and (. | semverCompare ">=v1.24.0") ($container_manager | has "containerd" | not) }}
          {{- $container_manager = append $container_manager "containerd" }}
        {{- end }}
      {{- end }}
      {{- $container_manager | toJson }}

  # 各组件所需的容器镜像列表
  images:
    manifests: []
    # 下载镜像时使用的默认仓库地址
    registry: >-
      {{- if .zone | eq "cn" }}
      hub.kubesphere.com.cn
      {{- end }}
    # 镜像拉取策略，支持 strict, warn
    policy: "strict"
    # Kubernetes 相关镜像列表（按 Helm Chart 及版本组织）
    openebs-localpv/localpv-provisioner:
      "4.4.0":
        - docker.io/openebs/linux-utils:4.3.0
        - docker.io/openebs/provisioner-localpv:4.4.0
    nfs-subdir-external-provisioner/nfs-subdir-external-provisioner:
      "4.0.18":
        - registry.k8s.io/sig-storage/nfs-subdir-external-provisioner:v4.0.2
    projectcalico/tigera-operator:
      v3.25.2:
        - quay.io/tigera/operator:v1.29.6
        - docker.io/calico/apiserver:v3.25.2
        - docker.io/calico/cni:v3.25.2
        - docker.io/calico/csi:v3.25.2
        - docker.io/calico/ctl:v3.25.2
        - docker.io/calico/kube-controllers:v3.25.2
        - docker.io/calico/node-driver-registrar:v3.25.2
        - docker.io/calico/typha:v3.25.2
        - docker.io/calico/node:v3.25.2
        - docker.io/calico/pod2daemon-flexvol:v3.25.2
        - docker.io/flannel/flannel:v0.24.4
      v3.26.5:
        - quay.io/tigera/operator:v1.30.11
        - docker.io/calico/ctl:v3.26.5
        - docker.io/calico/typha:v3.26.5
        - quay.io/calico/node:v3.26.5
        - docker.io/calico/cni:v3.26.5
        - docker.io/calico/csi:v3.26.5
        - docker.io/calico/apiserver:v3.26.5
        - docker.io/calico/kube-controllers:v3.26.5
        - docker.io/calico/flannel-migration-controller:v3.26.5
        - docker.io/flannel/flannel:v0.24.4
        - docker.io/calico/dikastes:v3.26.5
        - docker.io/calico/node-driver-registrar:v3.26.5
        - quay.io/calico/pod2daemon-flexvol:v3.26.5
      v3.28.5:
        - quay.io/tigera/operator:v1.34.13
        - docker.io/calico/ctl:v3.28.5
        - docker.io/calico/typha:v3.28.5
        - quay.io/calico/node:v3.28.5
        # - docker.io/calico/node-windows:v3.28.5
        - docker.io/calico/cni:v3.28.5
        # - docker.io/calico/cni-windows:v3.28.5
        - docker.io/calico/csi:v3.28.5
        - docker.io/calico/apiserver:v3.28.5
        - docker.io/calico/kube-controllers:v3.28.5
        - docker.io/calico/flannel-migration-controller:v3.28.5
        - docker.io/flannel/flannel:v0.24.4
        - docker.io/calico/dikastes:v3.28.5
        - docker.io/calico/node-driver-registrar:v3.28.5
        - quay.io/calico/pod2daemon-flexvol:v3.28.5
      v3.29.7:
        - quay.io/tigera/operator:v1.36.16
        - docker.io/calico/ctl:v3.29.7
        - docker.io/calico/typha:v3.29.7
        - quay.io/calico/node:v3.29.7
        # - docker.io/calico/node-windows:v3.29.7
        - docker.io/calico/cni:v3.29.7
        # - docker.io/calico/cni-windows:v3.29.7
        - docker.io/calico/csi:v3.29.7
        - docker.io/calico/apiserver:v3.29.7
        - docker.io/calico/kube-controllers:v3.29.7
        - docker.io/calico/flannel-migration-controller:v3.29.7
        # - docker.io/calico/windows:v3.29.7
        - docker.io/flannel/flannel:v0.24.4
        - docker.io/calico/dikastes:v3.29.7
        - docker.io/calico/node-driver-registrar:v3.29.7
        - quay.io/calico/pod2daemon-flexvol:v3.29.7
      v3.30.5:
        - quay.io/tigera/operator:v1.38.9
        - docker.io/calico/ctl:v3.30.5
        - docker.io/calico/typha:v3.30.5
        - quay.io/calico/node:v3.30.5
        # - docker.io/calico/node-windows:v3.30.5
        - docker.io/calico/cni:v3.30.5
        # - docker.io/calico/cni-windows:v3.30.5
        - docker.io/calico/csi:v3.30.5
        - docker.io/calico/apiserver:v3.30.5
        - docker.io/calico/kube-controllers:v3.30.5
        - docker.io/calico/envoy-gateway:v3.30.5
        - docker.io/calico/envoy-proxy:v3.30.5
        - docker.io/calico/envoy-ratelimit:v3.30.5
        - docker.io/calico/flannel-migration-controller:v3.30.5
        - docker.io/flannel/flannel:v0.24.4
        - docker.io/calico/dikastes:v3.30.5
        - docker.io/calico/node-driver-registrar:v3.30.5
        - quay.io/calico/pod2daemon-flexvol:v3.30.5
        - docker.io/calico/csi:v3.30.5
        - docker.io/calico/key-cert-provisioner:v3.30.5
        - docker.io/calico/goldmane:v3.30.5
        - docker.io/calico/whisker:v3.30.5
        - docker.io/calico/whisker-backend:v3.30.5
      v3.31.3:
        - quay.io/tigera/operator:v1.40.3
        - quay.io/calico/ctl:v3.31.3
        - docker.io/calico/typha:v3.31.3
        - quay.io/calico/node:v3.31.3
        # - docker.io/calico/node-windows:v3.31.3
        - docker.io/calico/cni:v3.31.3
        # - docker.io/calico/cni-windows:v3.31.3
        - docker.io/calico/csi:v3.31.3
        - docker.io/calico/apiserver:v3.31.3
        - docker.io/calico/kube-controllers:v3.31.3
        - docker.io/calico/envoy-gateway:v3.31.3
        - docker.io/calico/envoy-proxy:v3.31.3
        - docker.io/calico/envoy-ratelimit:v3.31.3
        - docker.io/calico/flannel-migration-controller:v3.31.3
        - docker.io/flannel/flannel:v0.24.4
        - docker.io/calico/dikastes:v3.31.3
        - docker.io/calico/node-driver-registrar:v3.31.3
        - quay.io/calico/pod2daemon-flexvol:v3.31.3
        - docker.io/calico/csi:v3.31.3
        - docker.io/calico/key-cert-provisioner:v3.31.3
        - docker.io/calico/goldmane:v3.31.3
        - docker.io/calico/whisker:v3.31.3
        - docker.io/calico/whisker-backend:v3.31.3
    cilium/cilium:
      "1.14.19":
        - quay.io/cilium/cilium:v1.14.19
        - quay.io/cilium/certgen:v0.1.16
        - quay.io/cilium/hubble-relay:v1.14.19
        - quay.io/cilium/hubble-ui-backend:v0.13.1
        - quay.io/cilium/hubble-ui:v0.13.1
        - quay.io/cilium/cilium-envoy:v1.30.9-1734953328-6db0e437ba7ed2169f032ceec25922dd06e0b12b
        # - quay.io/cilium/cilium-etcd-operator:v2.0.7
        - quay.io/cilium/operator:v1.14.19
        - quay.io/cilium/startup-script:c54c7edeab7fde4da68e59acd319ab24af242c3f
        - quay.io/cilium/clustermesh-apiserver:v1.14.19
        - quay.io/coreos/etcd:v3.5.4
        - quay.io/cilium/kvstoremesh:v1.14.19
        - ghcr.io/spiffe/spire-agent:1.6.3
        - ghcr.io/spiffe/spire-server:1.6.3
      "1.15.19":
        - quay.io/cilium/cilium:v1.15.19
        - quay.io/cilium/certgen:v0.1.19
        - quay.io/cilium/hubble-relay:v1.15.19
        - quay.io/cilium/hubble-ui-backend:v0.13.2
        - quay.io/cilium/hubble-ui:v0.13.2
        - quay.io/cilium/cilium-envoy:v1.33.4-1752151664-7c2edb0b44cf95f326d628b837fcdd845102ba68
        # - quay.io/cilium/cilium-etcd-operator:v2.0.7
        - quay.io/cilium/operator:v1.15.19
        - quay.io/cilium/startup-script:c54c7edeab7fde4da68e59acd319ab24af242c3f
        - quay.io/cilium/clustermesh-apiserver:v1.15.19
        - docker.io/library/busybox:1.36.1
        - ghcr.io/spiffe/spire-agent:1.8.5
        - ghcr.io/spiffe/spire-server:1.8.5
      "1.16.19":
        - quay.io/cilium/cilium:v1.16.19
        - quay.io/cilium/certgen:v0.3.1
        - quay.io/cilium/hubble-relay:v1.16.19
        - quay.io/cilium/hubble-ui-backend:v0.13.3
        - quay.io/cilium/hubble-ui:v0.13.3
        - quay.io/cilium/cilium-envoy:v1.34.12-1767177245-7935d4d711cb6f8020385a50c996b90896e16a71
        - quay.io/cilium/operator:v1.16.19
        - quay.io/cilium/startup-script:1755531540-60ee83e
        - quay.io/cilium/clustermesh-apiserver:v1.16.19
        - docker.io/library/busybox:1.36.1
        - ghcr.io/spiffe/spire-agent:1.9.6
        - ghcr.io/spiffe/spire-server:1.9.6
      "1.17.15":
        - quay.io/cilium/cilium:v1.17.15
        - quay.io/cilium/certgen:v0.4.1
        - quay.io/cilium/hubble-relay:v1.17.15
        - quay.io/cilium/hubble-ui-backend:v0.13.3
        - quay.io/cilium/hubble-ui:v0.13.3
        - quay.io/cilium/cilium-envoy:v1.36.6-1776000132-2437d2edeaf4d9b56ef279bd0d71127440c067aa
        - quay.io/cilium/operator:v1.17.15
        - quay.io/cilium/startup-script:1755531540-60ee83e
        - quay.io/cilium/clustermesh-apiserver:v1.17.15
        - docker.io/library/busybox:1.37.0
        - ghcr.io/spiffe/spire-agent:1.9.6
        - ghcr.io/spiffe/spire-server:1.9.6
      "1.18.9":
        - quay.io/cilium/cilium:v1.18.9
        - quay.io/cilium/certgen:v0.4.1
        - quay.io/cilium/hubble-relay:v1.18.9
        - quay.io/cilium/hubble-ui-backend:v0.13.3
        - quay.io/cilium/hubble-ui:v0.13.3
        - quay.io/cilium/cilium-envoy:v1.36.6-1776000132-2437d2edeaf4d9b56ef279bd0d71127440c067aa
        - quay.io/cilium/operator:v1.18.9
        - quay.io/cilium/startup-script:1755531540-60ee83e
        - quay.io/cilium/clustermesh-apiserver:v1.18.9
        - docker.io/library/busybox:1.37.0
        - ghcr.io/spiffe/spire-agent:1.12.4
        - ghcr.io/spiffe/spire-server:1.12.4
      "1.19.3":
        - quay.io/cilium/cilium:v1.19.3
        - docker.io/istio/ztunnel:1.28.0-distroless
        - quay.io/cilium/certgen:v0.4.1
        - quay.io/cilium/hubble-relay:v1.19.3
        - quay.io/cilium/hubble-ui-backend:v0.13.3
        - quay.io/cilium/hubble-ui:v0.13.3
        - quay.io/cilium/cilium-envoy:v1.36.6-1776000132-2437d2edeaf4d9b56ef279bd0d71127440c067aa
        - quay.io/cilium/operator:v1.19.3
        - quay.io/cilium/startup-script:1763560095-8f36c34
        - quay.io/cilium/clustermesh-apiserver:v1.19.3
        - docker.io/library/busybox:1.37.0
        - ghcr.io/spiffe/spire-agent:1.9.6
        - ghcr.io/spiffe/spire-server:1.9.6
    flannel/flannel:
      v0.27.4:
        - ghcr.io/flannel-io/flannel-cni-plugin:v1.8.0-flannel1
        - ghcr.io/flannel-io/flannel:v0.27.4
    hybridnet/hybridnet:
      0.6.8:
        - docker.io/hybridnetdev/hybridnet:v0.8.8
    kubeovn/kube-ovn:
      v1.13.15:
        - docker.io/kubeovn/kube-ovn:v1.13.15
        - docker.io/kubeovn/vpc-nat-gateway:v1.13.15
      v1.15.0:
        - docker.io/kubeovn/kube-ovn:v1.15.0
        - docker.io/kubeovn/vpc-nat-gateway:v1.15.0
    spiderpool/spiderpool:
      v1.1.1:
        - ghcr.io/spidernet-io/spiderpool/spiderpool-plugins:27c4f118b1cec3773f2679b772e7583fc77e5686
        - ghcr.io/k8snetworkplumbingwg/multus-cni:v4.1.4
        - ghcr.io/spidernet-io/spiderpool/spiderpool-agent:v1.1.1
        - ghcr.io/spidernet-io/spiderpool/spiderpool-controller:v1.1.1
```

### 参数说明

| 参数 | 说明 |
|------|------|
| `download.timeout` | 下载二进制包、镜像等资源的超时时间。 |
| `download.cn_host` | 当 `zone` 设置为 `cn` 时，作为默认下载加速域名。 |
| `download.os` | 下载资源所针对的目标操作系统，默认 `linux`。 |
| `download.arch` | 下载资源所针对的目标 CPU 架构列表，默认 `["amd64"]`。 |
| `download.artifact_file` | 离线制品包（artifact）文件的本地路径，用于离线安装。 |
| `download.artifact_md5` | 离线制品包对应的 MD5 校验文件路径。 |
| `download.fetch` | 是否执行在线下载。若所有资源已预先准备到本地，可设为 `false`。 |
| `download.artifact_url` | 各组件二进制包及 Helm Chart 的下载 URL 模板，支持根据 `zone` 自动切换国内源。 |
| `download.tools` | 额外需要下载并打包的工具，例如 `oras`、`nerdctl`。 |
| `download.charts` | 除默认组件外，额外需要拉取的 Helm Chart 列表（支持仓库或 OCI 格式）。 |
| `download.iso` | 制作离线包时包含的操作系统 RPM/DEB 软件包列表。 |
| `download.cni.type` | 下载阶段需要准备哪些 CNI 插件类型。 |
| `download.storage_class` | 下载阶段预制存储类相关包/镜像的开关。 |
| `download.cri.container_manager` | 根据目标 Kubernetes 版本动态计算需要下载哪些容器运行时。 |
| `download.images` | 定义各组件（按 Helm Chart 分类）所需的容器镜像列表。 |
| `download.images.manifests` | 额外需要下载并推送到私有仓库的自定义镜像清单。 |
| `download.images.registry` | 下载镜像时使用的默认仓库地址。 |
| `download.images.policy` | 镜像下载/校验策略：`strict`（严格校验）或 `warn`（仅警告）。 |
| `download.images.<chart_name>` | 以 Helm Chart 名称为键的镜像映射；值为版本号到所需镜像列表的映射。 |

---
