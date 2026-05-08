# 离线安装 Kubernetes 和 KubeSphere

本节介绍如何在无法访问 Internet 的环境下，使用离线安装包部署 Kubernetes 和 KubeSphere。

> **前置依赖**：安装过程依赖 `tar` 工具完成软件包的压缩和解压，请提前确认系统环境已预装该命令。

---

## 准备工作

参考如下最低配置要求，准备 Linux 主机。

| 角色 | 主机数量 | 最低要求（每个节点） | 网络要求 |
|------|---------|---------------------|---------|
| 打包节点 | 1 | CPU：1 核，内存：1 GB，硬盘：150 GB |  |
| 部署节点（运行 Web Installer 服务） | 1 | CPU：1 核，内存：1 GB，硬盘：150 GB | 与 Kubernetes 节点网络互通 |
| 私有镜像仓库节点 | 1 | CPU：8 核，内存：16 GB，硬盘：100 GB | 与 Kubernetes 节点网络互通 |
| Kubernetes 节点 | ≥ 1 | CPU：2 核，内存：4 GB，硬盘：40 GB | 节点间网络互通 |

> **注意事项**
>
> - 同一台主机可同时承担多个角色，例如：同时作为部署节点和私有镜像仓库节点，或者同时作为部署节点和 Kubernetes 节点。
> - **私有镜像仓库节点与 Kubernetes 节点不能是同一台主机。**

**各角色说明：**

- **打包节点**：需要准备至少 1 台 Linux 服务器作为打包节点。该节点将从互联网下载所需软件包与镜像，需确保能够访问以下地址：`github.com`、`docker.io`、`quay.io`。
- **部署节点**（运行 Web Installer 服务）：安装过程中需要在该节点上执行 kk 命令以运行安装服务。该节点需与私有镜像仓库节点、Kubernetes 节点保持网络互通。
- **私有镜像仓库节点**：如果尚未部署任何私有镜像仓库，需准备至少 1 台 Linux 服务器。该节点需与 Kubernetes 各节点保持网络互通。
- **Kubernetes 节点**：需要准备至少 1 台 Linux 服务器作为集群节点（无需提前安装 Kubernetes）。

---

## 构建离线安装包

### 创建配置文件

> 可通过 https://get-images.kubesphere.io 页面生成

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
        - v1.23.17
        - v1.24.17
        - v1.25.16
        - v1.26.15
        - v1.27.16
        - v1.28.15
        - v1.29.15
        - v1.30.14
        - v1.31.12
        - v1.32.11
        - v1.33.7
        - v1.34.3
    cni:
      type:
        - calico
        - cilium
        - flannel
        - kubeovn
        - hybridnet
      multi_cni:
        - multus
        - spiderpool
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
      - "almalinux-9.0-rpms"
      - "kylin-v10SP3-rpms"
      - "ubuntu-22.04-debs"
      - "centos-8-rpms"
      - "kylin-v10SP2-rpms"
      - "ubuntu-24.04-debs"
      - "debian-10-debs"
      - "kylin-v10SP1-rpms"
      - "debian-11-debs"
      - "ubuntu-18.04-debs"
      - "kylin-v10SP3-2403-rpms"
      - "ubuntu-20.04-debs"
```

**字段说明：**

| 字段 | 说明 |
|------|------|
| `apiVersion` | 配置文件的 API 版本，固定值为 `kubekey.kubesphere.io/v1` |
| `kind` | 资源类型，固定值为 `Config` |
| `spec.zone` | 软件包下载的区域，`cn` 表示使用国内源 |
| `spec.download.arch` | 指定需要下载的 CPU 架构，支持 `amd64` 和 `arm64` |
| `spec.download.images.policy` | 镜像下载策略，`warn` 表示镜像不存在时仅警告 |
| `spec.download.images.registry` | 镜像仓库地址 |
| `spec.download.kubernetes.kube_version` | 需要包含的 Kubernetes 版本列表 |
| `spec.download.cni.type` | 需要包含的 CNI 插件类型 |
| `spec.download.cni.multi_cni` | 需要包含的多 CNI 管理组件 |
| `spec.download.cri.container_manager` | 容器运行时类型，支持 `containerd` 和 `docker` |
| `spec.download.storage_class` | 需要包含的存储类，支持 `local`、`nfs` |
| `spec.download.image_registry.type` | 镜像仓库类型，支持 `harbor` 和 `docker-registry` |
| `spec.download.iso` | 制作 ISO 依赖包的操作系统列表，用于安装系统依赖 |

### 获取 kk 与 Web Installer

如果您访问 GitHub/GoogleAPIs 受限，请设置如下环境变量：

```shell
export KKZONE=cn
```

执行以下命令下载 KubeKey 和 Web Installer：

```shell
curl -sfL https://get-kk.kubesphere.io | sh -
```

执行完成后，会在当前目录生成以下文件：

| 原文件 | 解压后文件 |
|--------|-----------|
| `kubekey-v4.x.x-linux-amd64.tar.gz` | kk：KubeKey 二进制文件 |
| `web-installer.tgz` | dist：Web 页面资源<br>host-check.yaml、kubernetes、kubesphere：任务模板文件<br>schema：配置文件<br>README.md：安装说明文档 |
| `package.sh` | 离线安装包的构建脚本 |

### 制作离线安装包

执行构建脚本：

```shell
./package.sh config.yaml
```

当输出 `Offline package artifact.tgz has been created successfully.` 时表示制作成功。离线包为 `artifact.tgz`。


离线包包含以下内容：

```text
artifact/
├── artifact/kubekey-artifact.tgz    # 完整的离线资源包
└── artifact/tools/                  # 不同架构的工具包
    ├── amd64/
    │   ├── kubekey-v4.x.x-linux-amd64.tar.gz
    │   ├── nerdctl-2.2.1-linux-amd64.tar.gz
    │   └── oras_1.3.0_linux_amd64.tar.gz
    └── arm64/
        ├── kubekey-v4.x.x-linux-arm64.tar.gz
        ├── nerdctl-2.2.1-linux-arm64.tar.gz
        └── oras_1.3.0_linux_arm64.tar.gz
```

---

## 使用离线包安装集群

安装集群前，需指定私有镜像仓库地址。有以下两种方式：

- **方式一**：单独安装私有镜像仓库，请参考 [镜像仓库安装](../image-registry/README.md)。
- **方式二**：在创建集群时同时安装镜像仓库，需要在 `inventory.yaml` 和 `config.yaml` 中添加对应配置信息（详见下文命令行安装步骤）。

### 解压离线包

```shell
tar -zxvf artifact.tgz
```

### 方法 1：Web Installer 安装

> **提示**：Web Installer 暂不支持安装私有镜像仓库，请提前参考 [镜像仓库安装](../image-registry/README.md) 单独安装。

#### 1. 进入离线包目录并解压工具

KubeKey 工具位于 `tools/{arch}/` 目录下，根据安装机器的架构解压对应的工具：

```shell
# 查看机器架构
uname -m
```

解压KubeKey 到离线包目录

```shell
tar -zxvf tools/{arch}/kubekey-v4.x.x-linux-{arch}.tar.gz .
```

#### 2. 启动 Web Installer

```shell
kk web --port 8080 --schema-path web-installer/schema --ui-path web-installer/dist
```

如果显示如下信息，表示 Web Installer 启动成功：

```
Web server started successfully on port 8080
```

请勿关闭命令终端，在浏览器中通过 `http://<启动节点 IP 地址>:8080` 打开 KubeKey 的 UI 页面。

### 方法 2：命令行安装
> **提示**：命令行支持 [镜像仓库安装](../image-registry/README.md) 单独安装。以及在创建集群过程中同步安装（修改inventory.yaml 以及 config.yaml 即可）。

#### 1. 进入离线包目录

KubeKey 工具位于 `tools/{arch}/` 目录下，根据安装机器的架构解压对应的工具：

```shell
# 查看机器架构
uname -m
```

解压KubeKey 到离线包目录

```shell
tar -zxvf tools/{arch}/kubekey-v4.x.x-linux-{arch}.tar.gz .
```

#### 2. 创建节点配置文件

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
  hosts:
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
    # etcd 节点（仅当 etcd_deployment_type 为 external 时）
    etcd:
      hosts:
        - localhost
#    image_registry:
#      hosts:
#        - localhost
    # NFS 节点（用于镜像仓库存储和 Kubernetes NFS 存储）
#    nfs:
#      hosts:
#        - localhost
```

**`spec:hosts` 中配置节点的连接参数：**

| 参数 | 描述 |
|------|------|
| `<key>` | 节点名称 |
| `<key>:connector` | 节点连接信息 |
| `<key>:connector:type` | 节点连接类型。支持 `local`（本地连接）和 `ssh`（远程连接）。会根据节点名称或 IP 自动识别 |
| `<key>:connector:host` | 使用 SSH 连接节点时的地址 |
| `<key>:connector:port` | 使用 SSH 连接节点时的端口。默认值：`22` |
| `<key>:connector:user` | 使用 SSH 连接节点时的用户名。默认值：`root` |
| `<key>:connector:password` | 连接节点时的密码。`local` 连接时对应 sudo 密码，`ssh` 连接时对应 SSH 密码 |
| `<key>:connector:private_key` | SSH 连接节点时的私钥文件路径。密码和密钥任选其一 |
| `<key>:connector:private_key_content` | SSH 连接节点时的私钥文件内容。可使用密钥内容替代密钥文件路径 |
| `<key>:internal_ipv4` | 节点在集群中通信时的 IPv4 地址 |
| `<key>:internal_ipv6` | 节点在集群中通信时的 IPv6 地址 |

**`spec:groups` 中配置节点的角色信息：**

| 参数 | 描述 |
|------|------|
| `k8s_cluster` | Kubernetes 集群组织节点。包含 `kube_control_plane` 和 `kube_worker`，无需额外配置 |
| `kube_control_plane` | Kubernetes 集群中的控制平面节点。在 `kube_control_plane:hosts` 中配置 `spec:hosts` 中定义的节点名称 |
| `kube_worker` | Kubernetes 集群中的工作节点。在 `kube_worker:hosts` 中配置 `spec:hosts` 中定义的节点名称 |
| `etcd` | Kubernetes 集群中的 etcd 节点。在 `etcd:hosts` 中配置 `spec:hosts` 中定义的节点名称 |
| `image_registry` | 用于创建私有镜像仓库的节点。离线安装时通常需要配置 |

>
> 如果选择在创建集群时同时安装镜像仓库，需要在 `inventory.yaml` 中额外添加 `image_registry` 节点和分组。示例：
>
> ```yaml
> spec:
>   hosts:
>     harbor1:
>       connector:
>         type: ssh
>         host: 172.16.0.1
>         port: 22
>         user: root
>         password: 123456
>       internal_ipv4: 172.16.0.1
>   groups:
>     image_registry:
>       hosts:
>         - harbor1
> ```

#### 3. 创建安装配置文件

执行以下命令创建安装配置文件 `config.yaml`：

```shell
./kk create config --with-kubernetes v1.32.13 -o .
```

将 `v1.32.13` 替换为实际需要的版本。KubeKey 默认支持 Kubernetes `v1.23~v1.34`。

命令执行完毕后将生成安装配置文件 `config-v1.32.13.yaml`。

>
> 如果选择在创建集群时同时安装镜像仓库，需要在 `config.yaml` 中补充镜像仓库配置：
>
> ```yaml
> spec:
>   image_registry:
>     # 镜像仓库类型。支持 harbor、docker-registry，留空则不安装
>     type: "harbor"
>     auth:
>       # 私有镜像仓库地址
>       registry: "dockerhub.kubekey.local"
> ```

#### 4. 配置集群参数

在 `config-v1.32.13.yaml` 中配置 Kubernetes 集群的信息：

| 参数 | 描述 |
|------|------|
| `zone` | 文件及镜像的下载区域。离线安装时无需联网，但此项在构建离线包时生效 |
| `kubernetes` | Kubernetes 相关配置 |
| `etcd` | etcd 相关配置 |
| `image_registry` | 私有镜像仓库相关配置 |
| `cri` | 容器运行时相关配置 |
| `cni` | 网络插件相关配置 |
| `storage_class` | 存储插件相关配置 |
| `dns` | 域名解析相关配置 |
| `image_manifests` | 需要下载的额外镜像 |

> **注意**：完整的配置参数说明请参考 [配置参考](../reference/config.md)。

#### 5. 安装集群

```shell
kk create cluster -a kubekey-artifact.tgz -i inventory.yaml -c config.yaml
```
