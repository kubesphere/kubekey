# 离线安装 Kubernetes

本节介绍如何在不能访问 Internet 的环境下使用离线安装包部署 Kubernetes。

安装过程中将使用开源工具 KubeKey 的 v4.x 版本。有关 KubeKey 的更多信息，请访问 [GitHub KubeKey 仓库](https://github.com/kubesphere/kubekey)。

> **说明**：安装过程依赖 `tar` 工具完成软件包的压缩和解压，请提前确认系统环境已预装该命令。若 `config.yaml` 中配置了 charts 参数，请确保打包节点已提前安装 `Helm`。

## 概述

离线安装需要先在可访问 Internet 的机器上将所需的组件与镜像打包为离线安装包，再将其传输至目标环境进行安装。整体流程如下：

1. **构建离线安装包**（在联网机器上）：下载组件与镜像，并打包为 `artifact.tgz`。
2. **传输离线包**：将 `artifact.tgz` 拷贝至目标环境（例如通过存储介质或内网传输）。
3. **安装集群**（在目标环境）：解压离线包，将镜像推送至私有镜像仓库（若单独安装私有镜像仓库，即方式一；若由 KubeKey 在创建集群时同步安装，镜像推送由安装流程自动完成），然后安装 Kubernetes。

## 角色说明

离线安装涉及以下三类角色：

| 角色 | 职责 | 最低配置（每节点） | 网络要求 |
|---|---|---|---|
| 打包节点 | 从 Internet 下载所需软件包与镜像，构建离线安装包 | CPU：1 核，内存：1 GB，硬盘：150 GB | 需能访问互联网 |
| 私有镜像仓库节点 | 存放集群所需的容器镜像 | CPU：8 核，内存：16 GB，硬盘：100 GB | 与 Kubernetes 节点网络互通 |
| Kubernetes 节点 | 运行集群工作负载（无需提前安装 Kubernetes） | CPU：2 核，内存：4 GB，硬盘：40 GB | 节点间网络互通 |

> **说明**：
> - 同一台主机可同时承担多个角色，例如同时作为打包节点与私有镜像仓库节点，或同时作为打包节点与 Kubernetes 节点。
> - 若打包节点不承担集群角色，需另准备一台主机作为 Kubernetes 节点。

## 前提条件

> **说明**：以下为 Kubernetes 节点需满足的前提条件。

- 您需要准备至少 1 台 Linux 服务器作为集群节点。在生产环境中，为确保集群具备高可用性，建议准备至少 5 台 Linux 服务器，其中 3 台作为控制平面节点，另外 2 台作为工作节点。如果您在多台 Linux 服务器上安装 Kubernetes，请确保所有服务器属于同一子网。
- 集群节点的操作系统和版本须为 Ubuntu 18.04、Ubuntu 20.04、Ubuntu 22.04、Ubuntu 24.04、Debian 10、Debian 11、CentOS 8、AlmaLinux 9.0 或 Kylin v10。多台服务器的操作系统可以不同。关于其它操作系统和版本支持，请咨询青云科技官方解决方案专家或交付服务专家。
- 在生产环境中，为确保集群具有足够的计算和存储资源，建议每台集群节点配置至少 8 个 CPU 核心、16 GB 内存和 200 GB 磁盘空间。除此之外，建议在每台集群节点的 `/var/lib/docker`（对于 Docker）或 `/var/lib/containerd`（对于 containerd）目录额外挂载至少 200 GB 磁盘空间，用于存储容器运行时数据。
- 请确保所有集群节点上 `/etc/resolv.conf` 文件中配置的 DNS 服务器地址可用。否则，集群可能会出现域名解析问题。
- 请确保在所有集群节点上都可以使用 `sudo`、`tar`、`curl` 和 `openssl` 命令。
- 请确保所有集群节点时间同步。

## 配置防火墙规则

Kubernetes 需要特定端口和协议用于服务之间的通信。如果您的基础设施环境已启用防火墙，您需要在防火墙设置中放行所需的端口和协议。如果您的基础设施环境未启用防火墙，您可以跳过此步骤。

下表列出需要在防火墙中放行的端口和协议。

| 服务 | 协议 | 行为 | 起始端口 | 结束端口 | 备注 |
|---|---|---|---|---|---|
| ssh | TCP | allow | 22 | N/A | — |
| etcd | TCP | allow | 2379 | 2380 | — |
| apiserver | TCP | allow | 6443 | N/A | — |
| calico | TCP | allow | 9099 | 9100 | — |
| bgp | TCP | allow | 179 | N/A | — |
| nodeport | TCP | allow | 30000 | 32767 | — |
| master | TCP | allow | 10250 | 10258 | — |
| worker | TCP | allow | 10250 | N/A | — |
| dns | TCP | allow | 53 | N/A | — |
| dns | UDP | allow | 53 | N/A | — |
| local-registry | TCP | allow | 5000 | N/A | 离线环境需要 |
| local-apt | TCP | allow | 5080 | N/A | 离线环境需要 |
| rpcbind | TCP | allow | 111 | N/A | 使用 NFS 作为持久化存储时需要 |
| ipip | IPENCAP / IPIP | allow | N/A | N/A | 使用 Calico 时需要 |

## 构建离线安装包

### 创建配置文件

**提示**：可通过 [https://get-images.kubesphere.io](https://get-images.kubesphere.io) 页面可视化选择所需的组件并自动生成 `config.yaml`，也可参照以下示例手动创建。

登录打包节点，在打包节点上创建 `config.yaml` 文件：

```yaml
apiVersion: kubekey.kubesphere.io/v1
kind: Config
spec:
  zone: "cn"
  download:
    arch:
      - amd64
      - arm64
    images:
      policy: warn
      registry: hub.kubesphere.com.cn
    kubernetes:
      kube_version:
        - v1.36.4
        # 也可列出 v1.23~v1.36 中的其他版本
    cni:
      type:
        - calico
        - cilium
        - flannel
        - kubeovn
        - hybridnet
      # multi_cni:          # 可选，多 CNI 管理组件，如 multus
      #   - multus
    cri:
      container_manager:
        - containerd
        - docker
    storage_class:
      local:
        enabled: true
      nfs:
        enabled: true
    image_registry:
      type:
        - harbor
        - docker-registry
    iso:
      - "ubuntu-22.04-debs"
      - "centos-8-rpms"
      # 按需增删
```

**字段说明：**

| 字段 | 说明 |
|---|---|
| `apiVersion` | 配置文件的 API 版本，固定值为 `kubekey.kubesphere.io/v1` |
| `kind` | 资源类型，固定值为 `Config` |
| `spec.zone` | 软件包下载的区域，`cn` 表示使用国内源 |
| `spec.download.arch` | 指定需要下载的 CPU 架构，支持 `amd64` 和 `arm64` |
| `spec.download.images.policy` | 镜像下载策略，`warn` 表示如果镜像缺少某些 CPU 架构或操作系统，仅记录警告；`strict` 表示拉取的镜像必须包含所有选定的 CPU 架构和操作系统，否则报错 |
| `spec.download.images.registry` | 镜像仓库地址 |
| `spec.download.kubernetes.kube_version` | 需要包含的 Kubernetes 版本列表 |
| `spec.download.cni.type` | 需要包含的 CNI 插件类型 |
| `spec.download.cni.multi_cni` | 需要包含的多 CNI 管理组件 |
| `spec.download.cri.container_manager` | 容器运行时类型，支持 `containerd` 和 `docker` |
| `spec.download.storage_class` | 需要包含的存储类，支持 `local`、`nfs` |
| `spec.download.image_registry.type` | 镜像仓库类型，支持 `harbor` 和 `docker-registry` |
| `spec.download.iso` | 制作 ISO 依赖包的操作系统列表，用于安装系统依赖 |

### 获取 KubeKey

如果访问 GitHub 或 Google APIs 受限，请设置如下环境变量：

```bash
export KKZONE=cn
```

执行以下命令下载 KubeKey：

```bash
curl -sfL https://get-kk.kubesphere.io | sh -
```

执行完成后，会在当前目录生成以下文件：

| 原文件 | 解压后文件 |
|---|---|
| `kubekey-v4.x.x-linux-amd64.tar.gz` | `kk`：KubeKey 二进制文件 |
| `package.sh` | 离线安装包的构建脚本（由下载命令自动生成，内部调用 `kk artifact export` 完成下载与打包） |

### 制作离线安装包

执行构建脚本：

```bash
./package.sh config.yaml
```

当输出以下信息时，表示制作成功：

```text
Offline package artifact.tgz has been created successfully.
```

离线包为 `artifact.tgz`，包含以下内容：

```text
artifact/
├── kubekey-artifact.tgz    # 完整的离线资源包
└── tools/                  # 不同架构的工具包
    ├── amd64/
    │   ├── kubekey-v4.x.x-linux-amd64.tar.gz
    │   ├── nerdctl-2.2.1-linux-amd64.tar.gz
    │   └── oras_1.3.0_linux_amd64.tar.gz
    └── arm64/
        ├── kubekey-v4.x.x-linux-arm64.tar.gz
        ├── nerdctl-2.2.1-linux-arm64.tar.gz
        └── oras_1.3.0_linux_arm64.tar.gz
```

将 `artifact.tgz` 传输至目标环境后，后续步骤均在目标环境中执行。

## 使用离线包安装集群

安装集群前，需要指定私有镜像仓库地址。有以下两种方式：

- **方式一**：单独安装私有镜像仓库，请参考[镜像仓库安装](https://github.com/kubesphere/kubekey/blob/main/docs/zh/image-registry/README.md)。
- **方式二**：在创建集群时同时安装镜像仓库。详情请参考下述安装步骤。

### 解压离线包

```bash
tar -zxvf artifact.tgz
```

### 安装集群

#### 1. 进入离线包目录并解压工具

KubeKey 工具位于 `tools/{arch}/` 目录下，请根据安装机器的架构解压对应工具。

查看机器架构：

```bash
uname -m
```

进入离线包目录：

```bash
cd artifact/
```

解压 KubeKey 到离线包目录：

```bash
tar -zxvf tools/$(uname -m)/kubekey-v4.x.x-linux-$(uname -m).tar.gz -C .
```

#### 2. 推送镜像到私有镜像仓库（仅方式一：单独安装私有镜像仓库）

> **说明**：仅当采用 **方式一（单独安装私有镜像仓库）** 时，需要手动执行本步骤，将离线包中的镜像推送到已部署的私有镜像仓库。如果采用 **方式二（创建集群时同步安装镜像仓库）**，KubeKey 会在执行 `kk create cluster` 时自动部署镜像仓库并推送镜像，**请跳过本步骤**。

执行以下命令将离线包中的镜像推送到已部署的私有镜像仓库：

```bash
./kk artifact images --push -c config.yaml -a kubekey-artifact.tgz
```

> **注意**：执行前请确保推送所用的 `config.yaml` 中已正确配置私有镜像仓库地址（即 `spec.image_registry.auth` 字段）。该 `config.yaml` 与上文「构建离线安装包」步骤中用于打包的配置文件为同一文件。

#### 3. 创建节点配置文件

执行以下命令创建节点配置文件 `inventory.yaml`：

```bash
./kk create inventory -o .
```

命令执行完毕后，将生成节点配置文件 `inventory.yaml`。`inventory.yaml` 主要用于设置集群中各节点的连接信息，示例如下：

```yaml
apiVersion: kubekey.kubesphere.io/v1
kind: Inventory
metadata:
  name: default
spec:
  hosts: {}
  groups:
    k8s_cluster:
      groups:
        - kube_control_plane
        - kube_worker
    kube_control_plane:
      hosts:
        - localhost
    kube_worker:
      hosts:
        - localhost
    etcd:
      hosts:
        - localhost
```

`spec.hosts` 中的节点连接参数：

| 参数 | 描述 |
|---|---|
| `<key>` | 节点名称 |
| `<key>.connector.type` | 节点连接类型。支持 `local`（本地连接）和 `ssh`（远程连接）。KubeKey 会根据节点名称或 IP 自动识别连接类型 |
| `<key>.connector.host` | 使用 SSH 连接节点时的地址 |
| `<key>.connector.port` | 使用 SSH 连接节点时的端口。默认值：`22` |
| `<key>.connector.user` | 使用 SSH 连接节点时的用户名。默认值：`root` |
| `<key>.connector.password` | 连接节点时的密码。`local` 连接时对应 sudo 密码，`ssh` 连接时对应 SSH 密码 |
| `<key>.connector.private_key` | SSH 连接节点时的私钥文件路径。密码和密钥任选其一 |
| `<key>.connector.private_key_content` | SSH 连接节点时的私钥文件内容。可使用密钥内容替代密钥文件路径 |
| `<key>.internal_ipv4` | 节点在集群中通信时使用的 IPv4 地址 |
| `<key>.internal_ipv6` | 节点在集群中通信时使用的 IPv6 地址 |

`spec.groups` 中的节点角色参数：

| 参数 | 描述 |
|---|---|
| `k8s_cluster` | Kubernetes 集群节点分组。包含 `kube_control_plane` 和 `kube_worker`，无需额外配置 |
| `kube_control_plane` | Kubernetes 集群中的控制平面节点。在 `kube_control_plane.hosts` 中配置 `spec.hosts` 中定义的节点名称 |
| `kube_worker` | Kubernetes 集群中的工作节点。在 `kube_worker.hosts` 中配置 `spec.hosts` 中定义的节点名称 |
| `etcd` | Kubernetes 集群中的 etcd 节点。在 `etcd.hosts` 中配置 `spec.hosts` 中定义的节点名称 |
| `image_registry` | 用于创建私有镜像仓库的节点。离线安装时通常需要配置 |

如果选择在创建集群时同时安装镜像仓库，需要在 `inventory.yaml` 中额外添加 `image_registry` 节点和分组。示例如下：

```yaml
spec:
  hosts:
    harbor1:
      connector:
        type: ssh
        host: 172.16.0.1
        port: 22
        user: root
        password: 123456
      internal_ipv4: 172.16.0.1
  groups:
    image_registry:
      hosts:
        - harbor1
```

#### 4. 创建安装配置文件

> **注意**：此步骤生成的配置文件用于**安装集群**，与前面「构建离线安装包」步骤中用于**打包**的 `config.yaml` 不是同一个文件。

执行以下命令创建安装配置文件。以下示例使用 `v1.36.4`，该版本已包含在前文 `config.yaml` 示例的离线资源列表中：

```bash
./kk create config --with-kubernetes v1.36.4 -o .
```

将 `v1.36.4` 替换为实际需要的 Kubernetes 版本。请确保替换后的版本已包含在构建离线包时使用的 `spec.download.kubernetes.kube_version` 列表中。

命令执行完毕后，将生成安装配置文件 `config-v1.36.4.yaml`。

如果选择在创建集群时同时安装镜像仓库，需要在 `config-v1.36.4.yaml` 中补充镜像仓库配置：

```yaml
apiVersion: kubekey.kubesphere.io/v1
kind: Config
spec:
  download:
    arch:
      - amd64
      - arm64
    images:
      policy: warn
  image_registry:
    auth:
      registry: dockerhub.kubekey.local  # 替换为您的实际私有镜像仓库地址
      username: admin
      password: Harbor12345
      skip_tls_verify: false
      plain_http: false
```

#### 5. 安装集群

```bash
./kk create cluster -a kubekey-artifact.tgz -i inventory.yaml -c config-v1.36.4.yaml
```

安装完成后，可通过 `kubectl get nodes` 查看集群节点状态：

```shell
kubectl get nodes
```

## 常见问题

**Q：打包时选择的版本与安装时使用的版本不一致？**
A：安装使用的 Kubernetes 版本必须包含在构建离线包时 `spec.download.kubernetes.kube_version` 列表中，否则离线包中不包含该版本的镜像。

**Q：镜像推送失败？**
A：请确认 `config.yaml` 中仓库地址、用户名和密码正确，且网络可访问 `5000` 端口。

**Q：为何推送镜像时会去在线拉取 hub.kubesphere.com.cn 镜像？**
A：在推送镜像到私有镜像仓库之前，KubeKey 会先在线检查本地镜像文件的完整性（校验镜像清单与仓库中的镜像是否一致）。如无需在线检查，可在 `config.yaml` 中将 `spec.download.fetch` 设置为 `false` 来跳过该在线检查，从而完全离线完成镜像推送。

**Q：单台机器如何添加？**
A：节点角色必须为 `Master & Worker`。

**Q：如何重新执行安装？**
A：命令行方式请在清理节点后重新执行 `kk create cluster`。

**Q：如何保留 CA 证书用于后续添加节点？**
A：如果使用 KubeKey 默认证书，请在安装完成后保留 `<工作目录>/kubekey/pki/root.crt` 文件（默认工作目录为 `/root/kubekey`）。后续添加节点时可能需要该证书。

**Q：如何在有网络环境中安装？**
A：请参考[在线安装 Kubernetes](./online/)。
