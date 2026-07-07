# 安装 Kubernetes

本文档介绍如何使用 KubeKey 安装 Kubernetes 集群。

- 支持部署环境：Linux 发行版
- 支持的 Kubernetes 版本：v1.23.x ~ v1.34.x

## 系统要求

- 一台或多台运行兼容 deb/rpm 的 Linux 操作系统的计算机，例如 Ubuntu 或 CentOS。
- 每台机器 2 GB 以上的内存，内存不足时应用会受限制。
- 用作控制平面节点的计算机上至少有 2 个 CPU。
- 集群中所有计算机之间具有完全的网络连接。你可以使用公共网络或专用网络。

### 系统依赖

Kubernetes 要求操作系统预装以下依赖：

`socat` `conntrack` `ipset` `ebtables` `chrony` `ipvsadm`

KubeKey 已为部分 Linux 发行版制作了预编译的依赖包，可在 [iso-latest](https://github.com/kubesphere/kubekey/releases/tag/iso-latest) 获取。
支持的具体发行版与构建方式详见 [依赖包管理](../dependency-packages/README.md)。

## 定义节点信息

KubeKey 使用 `Inventory` 资源来定义节点的连接信息。
可使用 `kk create inventory` 来获取默认的 `inventory.yaml` 资源。默认的 `inventory.yaml` 配置如下：

```yaml
apiVersion: kubekey.kubesphere.io/v1
kind: Inventory
metadata:
  name: default
spec:
  hosts: # 你可以在此设置所有节点，或在特定组中设置节点
#    node1:
#      connector:
#        type: ssh
#        host: node1
#        port: 22
#        user: root
#        password: 123456
  groups:
    # 所有 Kubernetes 节点
    k8s_cluster:
      groups:
        - kube_control_plane
        - kube_worker
    # 控制平面节点
    kube_control_plane:
      hosts:
        - localhost
    # 工作节点
    kube_worker:
      hosts:
        - localhost
    # 当 etcd_deployment_type 为 external 时的 etcd 节点
    etcd:
      hosts:
        - localhost
#    image_registry:
#      hosts:
#        - localhost
    # 用于镜像仓库存储和 Kubernetes NFS 存储的 NFS 节点
#    nfs:
#      hosts:
#        - localhost
```
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `spec.hosts` | Object | 是 | 主机列表，key 为主机名称，value 为主机配置 |
| `spec.hosts.<name>.connector` | Object | 是 | 主机连接配置 |
| `spec.hosts.<name>.connector.host` | String | 是 | SSH 目标主机 IP 或域名 |
| `spec.hosts.<name>.connector.private_key` | String | 否 | SSH 私钥路径，默认使用系统默认密钥 |
| `spec.hosts.<name>.internal_ipv4` | String | 否 | 主机内部 IPv4 地址，会应用于/etc/hosts域名解析 |
| `spec.groups` | Object | 是 | 节点分组配置 |
| `spec.groups.k8s_cluster` | Object | 是 | Kubernetes 集群。包含两个子 group：`kube_control_plane`、`kube_worker` |
| `spec.groups.kube_control_plane` | Object | 是 | Kubernetes 集群中的控制平面节点组 |
| `spec.groups.kube_worker` | Object | 是 | Kubernetes 集群中的工作节点组 |
| `spec.groups.etcd` | Object | 否 | 安装 etcd 集群的节点组 |
| `spec.groups.image_registry` | Object | 是 | 镜像仓库节点组，指定哪些主机用于部署镜像仓库 |
| `spec.groups.nfs` | Object | 否 | 安装 NFS 的节点组 |
| `spec.groups.<group name>.hosts` | Array | 是 | 对应组的节点名称列表 |

节点磁盘格式化与 multipath 配置见 [存储与 Multipath 配置](../reference/storage.md)。

## 定义关键配置信息

KubeKey 使用 `Config` 资源来定义集群的关键配置信息。
可使用 `kk create config --with-kubernetes v1.33.1` 来获取默认的 `config.yaml` 资源。

针对不同的 Kubernetes 版本，给出了不同的默认 config 配置作为参考：

- [安装 v1.23.x 版本的 Kubernetes 配置](../../../builtin/core/defaults/config/v1.23.yaml)
- [安装 v1.24.x 版本的 Kubernetes 配置](../../../builtin/core/defaults/config/v1.24.yaml)
- [安装 v1.25.x 版本的 Kubernetes 配置](../../../builtin/core/defaults/config/v1.25.yaml)
- [安装 v1.26.x 版本的 Kubernetes 配置](../../../builtin/core/defaults/config/v1.26.yaml)
- [安装 v1.27.x 版本的 Kubernetes 配置](../../../builtin/core/defaults/config/v1.27.yaml)
- [安装 v1.28.x 版本的 Kubernetes 配置](../../../builtin/core/defaults/config/v1.28.yaml)
- [安装 v1.29.x 版本的 Kubernetes 配置](../../../builtin/core/defaults/config/v1.29.yaml)
- [安装 v1.30.x 版本的 Kubernetes 配置](../../../builtin/core/defaults/config/v1.30.yaml)
- [安装 v1.31.x 版本的 Kubernetes 配置](../../../builtin/core/defaults/config/v1.31.yaml)
- [安装 v1.32.x 版本的 Kubernetes 配置](../../../builtin/core/defaults/config/v1.32.yaml)
- [安装 v1.33.x 版本的 Kubernetes 配置](../../../builtin/core/defaults/config/v1.33.yaml)
- [安装 v1.34.x 版本的 Kubernetes 配置](../../../builtin/core/defaults/config/v1.34.yaml)

完整配置参考[配置参考](../reference/config.md)。

## 安装集群

KubeKey 支持**在线安装**和**离线安装**两种方式。

### 方式一：在线安装

在线安装时，KubeKey 会自动从互联网下载所需的 Kubernetes 组件和镜像。更详细的步骤请参考 [在线安装](online.md)。

### 方式二：离线安装

离线安装适用于无外网环境，需提前准备制品包（artifact）和系统依赖。详细步骤请参考 [离线安装](offline.md)。

## 集群节点管理

集群创建完成后，可根据业务需求进行节点扩缩容。

> **注意**：当前 Web Installer 暂不支持添加和删除集群节点，请通过命令行操作。

- **添加节点**：向已有 Kubernetes 集群添加新的控制平面节点、工作节点或 etcd 节点。详细步骤请参考 [添加集群节点](add-nodes.md)。
- **删除节点**：从 Kubernetes 集群中安全移除指定节点。详细步骤请参考 [删除集群节点](delete-nodes.md)。

## 启用 kubectl 自动补全

KubeKey 默认不启用 kubectl 自动补全。参考以下指南开启：

**前置条件**：确保已安装 bash-autocompletion 并正常工作。

```shell script
# 安装 bash-completion
apt-get install bash-completion

# 将补全脚本添加到 ~/.bashrc
echo 'source <(kubectl completion bash)' >>~/.bashrc

# 将补全脚本添加到 /etc/bash_completion.d 目录
kubectl completion bash >/etc/bash_completion.d/kubectl
```

更多详情请参考[官方文档](https://kubernetes.io/docs/tasks/tools/install-kubectl/#enabling-shell-autocompletion)。
