# 在线安装 Kubernetes 和 KubeSphere

本节介绍如何在可访问 Internet 的环境下安装 Kubernetes 和 KubeSphere。

安装过程中将使用开源工具 KubeKey 的 v4.x 版本。有关 KubeKey 的更多信息，请访问 [GitHub KubeKey 仓库](https://github.com/kubesphere/kubekey)。

## 核心概念

阅读后续步骤前，建议先了解以下基本概念：

- **控制平面节点（Master）**：负责集群的调度与管理，通常不运行业务负载。
- **工作节点（Worker）**：运行实际业务容器的工作负载节点。
- **etcd**：分布式键值存储，保存集群的所有关键状态数据。
- **容器运行时**：负责创建和运行容器的底层软件。KubeKey 支持自动安装 Docker 和 containerd 两种容器运行时。如需使用 CRI-O 或 iSula，需提前手动安装，但其与 KubeSphere 的兼容性尚未经过充分验证。
- **CNI（容器网络插件）**：为集群中的 Pod 提供网络连通能力，常用插件包括 Calico、Cilium、Flannel 等。

## 前提条件

- 至少需要 1 台 Linux 服务器作为集群节点。在生产环境中，为确保集群具备高可用性，建议准备至少 5 台 Linux 服务器，其中 3 台作为控制平面节点，另外 2 台作为工作节点。如果您在多台 Linux 服务器上安装 KubeSphere，请确保所有服务器属于同一子网。
- 集群节点的操作系统和版本须为 Ubuntu 18.04、Ubuntu 20.04、Ubuntu 22.04、Ubuntu 24.04、Debian 10、Debian 11、CentOS 8、AlmaLinux 9.0 或 Kylin v10。多台服务器的操作系统可以不同。关于其它操作系统和版本支持，请咨询青云科技官方解决方案专家或交付服务专家。
- 在生产环境中，为确保集群具有足够的计算和存储资源，建议每台集群节点配置至少 8 个 CPU 核心、16 GB 内存和 200 GB 磁盘空间。除此之外，建议在每台集群节点的 `/var/lib/docker`（对于 Docker）或 `/var/lib/containerd`（对于 containerd）目录额外挂载至少 200 GB 磁盘空间，用于存储容器运行时数据。
- 在生产环境中，建议提前为 KubeSphere 集群配置高可用性，以避免单个控制平面节点出现故障时集群服务中断。有关更多信息，请参阅[配置高可用性](https://docs.kubesphere.com.cn/v4.2.1/03-installation-and-upgrade/01-preparations/02-configure-high-availability/02-configure-k8s-high-availability/)。

  > **说明**：如果您规划了多个控制平面节点，请务必提前为集群配置高可用性。

- 默认情况下，KubeSphere 使用集群节点的本地磁盘空间作为持久化存储。在生产环境中，建议提前配置外部存储系统作为持久化存储。有关更多信息，请参阅[配置外部持久化存储](https://docs.kubesphere.com.cn/v4.2.1/03-installation-and-upgrade/01-preparations/04-configure-external-persistent-storage/)。
- 如果集群节点未安装容器运行时，安装工具 KubeKey 将在安装过程中自动为每个集群节点安装容器运行时。KubeKey 默认安装 containerd，您也可以在配置文件中指定 Docker 作为容器运行时。

  > **说明**：如需使用 CRI-O 或 iSula，需提前手动安装，但其与 KubeSphere 的兼容性尚未经过充分验证，可能存在未知问题。

- 请确保所有集群节点上 `/etc/resolv.conf` 文件中配置的 DNS 服务器地址可用。否则，KubeSphere 集群可能会出现域名解析问题。
- 请确保在所有集群节点上都可以使用 `sudo`、`tar`、`curl` 和 `openssl` 命令。
- 请确保所有集群节点时间同步。

> **说明**：KubeKey 依赖 `tar` 工具完成软件包的压缩与解压，请务必确认已安装。

## 配置防火墙规则

KubeSphere 需要特定端口和协议用于服务之间的通信。如果您的基础设施环境已启用防火墙，您需要在防火墙设置中放行所需的端口和协议。如果您的基础设施环境未启用防火墙，您可以跳过此步骤。

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

## 选择安装方式

以下两种方法均可完成 Kubernetes 和 KubeSphere 的安装，**安装结果一致，您只需选择其中一种操作入口即可，无需同时执行**：

- **方法 1：命令行安装**：适用于熟悉命令行操作、需要精细化配置集群参数的场景。
- **方法 2：Web Installer 安装**：适用于希望通过图形化界面完成节点添加、参数配置和安装校验的场景。

> **说明**：两种方法任选其一，请勿重复执行。

### 方法 1：命令行安装

#### 1. 下载 KubeKey

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

#### 3. 创建安装配置文件

执行以下命令创建安装配置文件 `config.yaml`：

```shell
./kk create config --with-kubernetes <Kubernetes version> -o .
```

将 `<Kubernetes version>` 替换为实际需要的版本，例如 `v1.34.3`。KubeSphere 默认支持 Kubernetes `v1.23` ~ `v1.34`。

命令执行完毕后将生成安装配置文件 `config-<Kubernetes version>.yaml`。

#### 4. 配置集群参数

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

#### 5. 安装 Kubernetes

执行以下命令安装 Kubernetes：

```shell
./kk create cluster -i inventory.yaml -c config-<Kubernetes version>.yaml
```

如果已将安装配置文件重命名为 `config.yaml`，则可使用以下命令：

```shell
./kk create cluster -i inventory.yaml -c config.yaml
```

#### 6. 安装 KubeSphere

KubeKey v4.x 将 Kubernetes 和 KubeSphere 解耦安装，安装完 Kubernetes 后需手动执行以下 Helm 命令安装 KubeSphere。

> **说明**：Web Installer 安装方式会自动完成 KubeSphere 的安装，无需手动执行此步骤。

执行以下命令安装 KubeSphere：

```shell
chart=oci://hub.kubesphere.com.cn/kse/ks-core
version=1.2.5
helm upgrade --install -n kubesphere-system --create-namespace ks-core $chart \
  --debug --wait --version $version --reset-values --take-ownership \
  --set global.imageRegistry=hub.kubesphere.com.cn,extension.imageRegistry=hub.kubesphere.com.cn
```

> **说明**：
> - `--take-ownership` 用于接管集群中已存在的同名资源，避免安装冲突。该参数需要 Helm >= 3.17.0。

如果显示如下信息，则表示 KubeSphere 安装成功：

```text
NOTES:
Thank you for choosing KubeSphere Helm Chart.

Please be patient and wait for several seconds for the KubeSphere deployment to complete.

1. Wait for Deployment Completion

    Confirm that all KubeSphere components are running by executing the following command:

    kubectl get pods -n kubesphere-system

2. Access the KubeSphere Console

    Once the deployment is complete, you can access the KubeSphere console using the following URL:

    http://<节点 IP 地址>:30880

3. Login to KubeSphere Console

    Use the following credentials to log in:

    Account: admin
    Password: P@88w0rd

NOTE: It is highly recommended to change the default password immediately after the first login.
```

### 方法 2：Web Installer 安装

#### 1. 下载 KubeKey 和 Web Installer

如果您访问 GitHub / Google APIs 受限，请设置如下环境变量：

```shell
export KKZONE=cn
```

执行以下命令下载 KubeKey 最新版本（含 Web Installer）：

```shell
curl -sfL https://get-kk.kubesphere.io | SKIP_PACKAGE=true sh -
```

> **说明**：`SKIP_PACKAGE=true` 表示跳过离线打包脚本（`package.sh`）的下载，在线安装场景无需此脚本。

执行完成后，会在当前目录生成以下文件：

| 原文件 | 解压后文件 |
|---|---|
| `kubekey-v4.x.x-linux-amd64.tar.gz` | `kk`：KubeKey 二进制文件 |
| `web-installer.tgz` | `dist`：Web 页面资源；`host-check.yaml`、`kubernetes`、`kubesphere`：任务模板文件；`schema`：配置表单定义；`README.md`：安装说明文档 |

#### 2. 启动 Web Installer

解压 Web Installer 安装包

```shell
tar -zxvf web-installer.tgz
```

执行以下命令启动 Web Installer 页面：

```shell
./kk web --port 8080 --schema-path web-installer/schema --ui-path web-installer/dist
```

如果显示如下信息，表示 Web Installer 启动成功：

```text
Web server started successfully on port 8080
```

请勿关闭命令终端。

#### 3. 通过 Web Installer 部署 Kubernetes 和 KubeSphere

在浏览器中访问 `http://<启动节点 IP 地址>:8080`，打开 KubeKey 的 Web Installer 页面。

在页面中点击 **开始安装**，进入安装流程。

##### 3.1. 添加集群节点

在 **基本信息** 页面，添加集群节点。每个节点需要指定角色，支持以下三种：

- **Master**：控制平面节点，负责集群调度与管理。Master 节点上会自动安装 etcd。
- **Worker**：工作节点，运行实际业务容器。
- **Image**：镜像仓库节点，用于自动部署私有镜像仓库。在线安装时通常无需配置此角色。

Web Installer 支持以下三种节点添加方式：

- **手动添加**：适用于添加单个节点。您需要填写主机名、IP 地址、SSH 地址、SSH 认证等信息。
- **文件上传**：适用于批量添加节点。请根据模板填写节点信息后上传文件。
- **节点扫描**：适用于自动发现节点。您可以通过 IP CIDR 扫描节点，并根据扫描结果选择需要添加的节点。

> **注意**：如果只添加一个节点，节点角色必须为 `Master & Worker`。


##### 3.2. 修改配置参数

在安装配置页面，填写部署 Kubernetes 和 KubeSphere 所需的参数。

Kubernetes 和 KubeSphere Core 标签页均支持 **表单模式** 和 **YAML 模式**。您可以在表单中填写配置信息，也可以直接编辑 YAML 文件。如需了解更多配置参数，请参阅[配置示例](https://github.com/kubesphere/kubekey/blob/main/docs/zh/reference/config.md)。

> **说明**：如果您访问 GitHub / Google APIs 受限，请切换到 YAML 模式并添加 `zone: cn` 参数，KubeKey 将从国内地址下载二进制文件。同时建议将镜像仓库地址设置为 `hub.kubesphere.com.cn`。切换为国内源后，仅支持有限的 Kubernetes 版本，具体支持版本请参考 [get-images.kubesphere.io](https://get-images.kubesphere.io) 页面的 Kubernetes 标签页。

**安装镜像仓库（可选）**

如果需要在安装集群时同时部署私有镜像仓库，请在添加节点时添加 Image 角色的节点，并在 Kubernetes 配置参数中将镜像仓库类型设置为 `harbor` 或 `docker-registry`：

- **单节点镜像仓库**：添加 1 个 Image 角色节点，在 Kubernetes 配置参数中设置镜像仓库类型。
- **高可用镜像仓库**：添加多个 Image 角色节点（建议 ≥ 3 个），在 Kubernetes 配置参数中设置镜像仓库类型，并配置镜像仓库高可用虚拟 IP 作为访问入口。


配置完成后，点击 **下一步**。

##### 3.3. 预览安装配置

在 **安装预览** 页面，确认版本、节点、网络、存储等信息无误后，点击 **下一步：执行安装** 开始安装。

如需修改配置，可返回上一步重新编辑配置参数。

##### 3.4. 执行安装

等待安装完成。安装完成后，系统会自动进入安装校验流程。

如果安装过程中出现异常，可点击 **查看日志** 查看日志详情，并根据日志信息排查问题。必要时，可强制退出当前安装流程或初始化后重新安装。

> **说明**：如需在 Web Installer 页面重新安装、修改配置参数或清除配置，可点击左侧的 **初始化** 按钮。初始化操作会重置 Kubernetes 节点上的所有任务，并回到基本信息页面。该操作不可逆，请谨慎执行。

##### 3.5. 安装校验

1. 在 **安装校验** 页面，点击 **开始检测**，系统将自动运行检测脚本验证系统可用性。
2. 如果系统检测通过，点击 **完成**，即可查看 KubeSphere 的访问地址、管理员用户名和默认密码。
3. 在浏览器中输入访问地址，登录 KubeSphere Web 控制台，即可开始使用 KubeSphere。

> **说明**：取决于您的网络环境，您可能需要配置流量转发规则，并在防火墙中放行 `30880` 端口。

## 访问 KubeSphere Web 控制台

安装完成后，在浏览器中访问安装结果中显示的 KubeSphere 控制台地址。

使用安装结果中显示的管理员用户名和默认密码登录 KubeSphere Web 控制台。首次登录后，建议立即修改默认密码。

## 常见问题

**Q：下载时访问 GitHub 超时、速度缓慢？**
A：请先执行 `export KKZONE=cn` 使用国内源后再下载。切换为国内源后，仅支持有限的 Kubernetes 版本，请将 Kubernetes 版本修改为 [get-images.kubesphere.io](https://get-images.kubesphere.io) 页面 **Kubernetes 标签页** 中列出的版本之一（通过 `./kk create config --with-kubernetes <版本> -o .` 指定）。

**Q：单节点部署后无法访问控制台？**
A：单节点部署时节点角色必须为 `Master & Worker`；同时请确认防火墙已放行 `30880` 端口。

**Q：如何重新执行安装？**
A：Web Installer 中点击 **初始化** 按钮（该操作不可逆）；命令行方式请在`kk delete cluster --all --with-data`清理节点后重新执行 `kk create cluster`。

**Q：如何保留 CA 证书用于后续添加节点？**
A：如果使用 KubeKey 默认证书，请在安装完成后保留 `<工作目录>/kubekey/pki/root.crt` 文件（默认工作目录为 `/root/kubekey`）。后续添加节点时可能需要该证书。

**Q：生产环境有哪些额外建议？**
A：建议至少准备 5 台节点并提前配置高可用性，同时配置外部持久化存储，避免使用节点本地磁盘作为持久化存储。

**Q：如何在无网络环境中安装？**
A：请参考[离线安装 Kubernetes 和 KubeSphere](https://docs.kubesphere.com.cn/v4.2.1/03-installation-and-upgrade/02-install-kubesphere/02-offline-install-kubernetes-and-kubesphere/)。
