# 在线安装 Kubernetes

本节介绍如何在可访问 Internet 的环境下安装 Kubernetes。

安装过程中将使用开源工具 KubeKey 的 v4.x 版本。有关 KubeKey 的更多信息，请访问 [GitHub KubeKey 仓库](https://github.com/kubesphere/kubekey)。

## 核心概念

阅读后续步骤前，建议先了解以下基本概念：

- **控制平面节点（Master）**：负责集群的调度与管理，通常不运行业务负载。
- **工作节点（Worker）**：运行实际业务容器的工作负载节点。
- **etcd**：分布式键值存储，保存集群的所有关键状态数据。
- **容器运行时**：负责创建和运行容器的底层软件。KubeKey 支持自动安装 Docker 和 containerd 两种容器运行时。
- **CNI（容器网络插件）**：为集群中的 Pod 提供网络连通能力，常用插件包括 Calico、Cilium、Flannel 等。

## 前提条件

- 至少需要 1 台 Linux 服务器作为集群节点。在生产环境中，为确保集群具备高可用性，建议准备至少 5 台 Linux 服务器，其中 3 台作为控制平面节点，另外 2 台作为工作节点。如果您在多台 Linux 服务器上安装 Kubernetes，请确保所有服务器属于同一子网。
- 集群节点的操作系统和版本须为 Ubuntu 18.04、Ubuntu 20.04、Ubuntu 22.04、Ubuntu 24.04、Debian 10、Debian 11、CentOS 8、AlmaLinux 9.0 或 Kylin v10。多台服务器的操作系统可以不同。关于其它操作系统和版本支持，请咨询青云科技官方解决方案专家或交付服务专家。
- 在生产环境中，为确保集群具有足够的计算和存储资源，建议每台集群节点配置至少 8 个 CPU 核心、16 GB 内存和 200 GB 磁盘空间。除此之外，建议在每台集群节点的 `/var/lib/docker`（对于 Docker）或 `/var/lib/containerd`（对于 containerd）目录额外挂载至少 200 GB 磁盘空间，用于存储容器运行时数据。
- 如果集群节点未安装容器运行时，安装工具 KubeKey 将在安装过程中自动为每个集群节点安装容器运行时。KubeKey 默认安装 containerd，您也可以在配置文件中指定 Docker 作为容器运行时。
- 请确保所有集群节点上 `/etc/resolv.conf` 文件中配置的 DNS 服务器地址可用。否则，集群可能会出现域名解析问题。
- 请确保在所有集群节点上都可以使用 `sudo`、`tar`、`curl` 和 `openssl` 命令。
- 请确保所有集群节点时间同步。

> **说明**：KubeKey 依赖 `tar` 工具完成软件包的压缩与解压，请务必确认已安装。

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
| rpcbind | TCP | allow | 111 | N/A | 使用 NFS 作为持久化存储时需要 |
| ipip | IPENCAP / IPIP | allow | N/A | N/A | 使用 Calico 时需要 |

如需在测试环境中临时关闭防火墙，可执行 `systemctl stop firewalld`；生产环境请按上表精确放行所需端口。

## 安装 Kubernetes

当前仅支持命令行安装方式。

### 1. 下载 KubeKey

如果您访问 GitHub / Google APIs 受限，请设置如下环境变量：

```shell
export KKZONE=cn
```

执行以下命令下载 KubeKey 最新版本：

```shell
curl -sfL https://get-kk.kubesphere.io | SKIP_WEB_INSTALLER=true SKIP_PACKAGE=true sh -
```

> **说明**：`SKIP_WEB_INSTALLER=true` 跳过 Web Installer 资源下载，`SKIP_PACKAGE=true` 跳过离线打包脚本下载。命令行安装仅需 `kk` 二进制文件。

执行完成后，会在当前目录生成 `kk` 二进制文件。

### 2. 创建节点配置文件

执行以下命令创建节点配置文件 `inventory.yaml`：

```shell
./kk create inventory -o .
```

`inventory.yaml` 主要设置集群中各节点的连接信息。命令执行完毕后，将生成节点配置文件，示例如下：

```yaml
apiVersion: kubekey.kubesphere.io/v1
kind: Inventory
metadata:
  name: default
spec:
  hosts: # your can set all nodes here. or set nodes on special groups.
    # localhost:
    #   connector:
    #     password: 123456
    # node1:
    #   connector:
    #     type: ssh
    #     host: node1
    #     port: 22
    #     user: root
    #     password: 123456
    #   internal_ipv4: 1.1.1.1
  groups:
    # all kubernetes nodes.
    k8s_cluster:
      groups:
        - kube_control_plane
        - kube_worker
    # control_plane nodes
    kube_control_plane:
      hosts:
        - localhost
    # worker nodes
    kube_worker:
      hosts:
        - localhost
    # etcd nodes when etcd_deployment_type is external
    etcd:
      hosts:
        - localhost
#    image_registry:
#      hosts:
#        - localhost
    # nfs nodes for registry storage. and kubernetes nfs storage
#    nfs:
#      hosts:
#        - localhost

```

`spec.hosts` 用于配置各节点的连接参数。每个节点以节点名称作为键，例如 `localhost` 或 `node1`。

| 参数 | 描述 |
|---|---|
| `<key>` | 节点名称 |
| `<key>.connector.type` | 节点连接类型。支持 `local`（本地连接）和 `ssh`（远程连接）。会根据节点名称或 IP 自动识别 |
| `<key>.connector.host` | 使用 SSH 连接节点时的地址 |
| `<key>.connector.port` | 使用 SSH 连接节点时的端口。默认值：`22` |
| `<key>.connector.user` | 使用 SSH 连接节点时的用户名。默认值：`root` |
| `<key>.connector.password` | 连接节点时的密码。`local` 连接时对应 sudo 密码，`ssh` 连接时对应 SSH 密码 |
| `<key>.connector.private_key` | SSH 连接节点时的私钥文件路径。密码和密钥任选其一 |
| `<key>.connector.private_key_content` | SSH 连接节点时的私钥文件内容。可使用密钥内容替代密钥文件路径 |
| `<key>.internal_ipv4` | 节点在集群中通信时的 IPv4 地址 |
| `<key>.internal_ipv6` | 节点在集群中通信时的 IPv6 地址 |

`spec.groups` 用于配置节点的角色信息。

| 参数 | 描述 |
|---|---|
| `k8s_cluster` | Kubernetes 集群组织节点。包含 `kube_control_plane` 和 `kube_worker`，无需额外配置 |
| `kube_control_plane` | Kubernetes 集群中的控制平面节点。在 `kube_control_plane.hosts` 中配置 `spec.hosts` 中定义的节点名称 |
| `kube_worker` | Kubernetes 集群中的工作节点。在 `kube_worker.hosts` 中配置 `spec.hosts` 中定义的节点名称 |
| `etcd` | Kubernetes 集群中的 etcd 节点。在 `etcd.hosts` 中配置 `spec.hosts` 中定义的节点名称 |
| `image_registry` | 用于创建私有镜像仓库的节点。在线安装时通常无需配置 |

### 3. 创建安装配置文件

执行以下命令创建安装配置文件 `config.yaml`：

```shell
./kk create config --with-kubernetes <Kubernetes version> -o .
```

将 `<Kubernetes version>` 替换为实际需要的版本，例如 `v1.36.4`。Kubernetes 默认支持 `v1.23` ~ `v1.36`。

命令执行完毕后将生成安装配置文件 `config-<Kubernetes version>.yaml`。

### 4. 配置集群参数

在 `config-<Kubernetes version>.yaml` 中配置 Kubernetes 集群的信息。常用参数如下：

| 参数 | 描述 |
|---|---|
| `zone` | 文件及镜像的下载区域。如果您访问 GitHub / Google APIs 受限，请将该值设置为 `cn` |
| `kubernetes` | Kubernetes 集群配置，包括版本、控制平面端点、网络 CIDR 等 |
| `etcd` | etcd 集群配置，包括安装方式（内置或外部）、数据目录等 |
| `image_registry` | 私有镜像仓库配置。在线安装时通常使用默认值 |
| `cri` | 容器运行时配置，可选择 `docker` 或 `containerd` |
| `cni` | 网络插件配置，可选择 `calico`、`cilium`、`flannel` 等 |
| `storage_class` | 存储插件配置，支持 `local`、`nfs` 等 |
| `dns` | 域名解析配置 |
| `image_manifests` | 需要额外下载的镜像列表 |

> **注意**：各参数的完整字段说明请参考[配置参考](https://github.com/kubesphere/kubekey/blob/main/docs/zh/reference/config.md)。

### 5. 安装 Kubernetes

执行以下命令安装 Kubernetes：

```shell
./kk create cluster -i inventory.yaml -c config-<Kubernetes version>.yaml
```

如果已将安装配置文件重命名为 `config.yaml`，则可使用以下命令：

```shell
./kk create cluster -i inventory.yaml -c config.yaml
```

安装完成后，可通过 `kubectl get nodes` 查看集群节点状态：

```shell
kubectl get nodes
```

## 常见问题

**Q：下载时访问 GitHub 超时、速度缓慢？**
A：请先执行 `export KKZONE=cn` 使用国内源后再下载。切换为国内源后，仅支持有限的 Kubernetes 版本，请将 Kubernetes 版本修改为 [get-images.kubesphere.io](https://get-images.kubesphere.io) 页面 **Kubernetes 标签页** 中列出的版本之一（通过 `./kk create config --with-kubernetes <版本> -o .` 指定）。

**Q：如何重新执行安装？**
A：请先执行 `kk delete cluster --all --with-data` 清理节点，再重新执行 `kk create cluster`。

**Q：如何保留 CA 证书用于后续添加节点？**
A：如果使用 KubeKey 默认证书，请在安装完成后保留 `<工作目录>/kubekey/pki/root.crt` 文件（默认工作目录为 `/root/kubekey`）。后续添加节点时可能需要该证书。

**Q：生产环境有哪些额外建议？**
A：建议至少准备 5 台节点并提前配置高可用性，同时配置外部持久化存储，避免使用节点本地磁盘作为持久化存储。
