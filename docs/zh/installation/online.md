# 在线安装 Kubernetes 和 KubeSphere

本节介绍如何在可访问 Internet 的环境下安装 Kubernetes 和 KubeSphere。

安装过程中将用到开源工具 KubeKey。有关 KubeKey 的更多信息，请访问 [GitHub KubeKey 仓库](https://github.com/kubesphere/kubekey)。

## 安装依赖项

安装过程中依赖 `tar` 工具实现软件包的压缩、解压处理，请提前确认系统环境已预装该命令。

## 安装 Kubernetes 和 KubeSphere

支持两种安装方式：**命令行安装** 和 **Web 页面安装**。

### 命令行安装

#### 1. 下载 KubeKey

如果您访问 GitHub/GoogleAPIs 受限，请设置如下环境变量：

```shell
export KKZONE=cn
```

执行以下命令下载 KubeKey 最新版本：

```shell
curl -sfL https://get-kk.kubesphere.io | SKIP_WEB_INSTALLER=true SKIP_PACKAGE=true sh -
```

执行完成后，会在当前目录生成以下文件：

| 原文件 | 解压后文件 |
|--------|-----------|
| kubekey-v4.x.x-linux-amd64.tar.gz | kk：KubeKey 二进制文件 |

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
| `image_registry` | 用于创建私有镜像仓库的节点。在线安装时通常无需配置 |

#### 3. 创建安装配置文件

执行以下命令创建安装配置文件 `config.yaml`：

```shell
./kk create config --with-kubernetes <Kubernetes version> -o .
```

将 `<Kubernetes version>` 替换为实际需要的版本，例如 `v1.27.4`。KubeSphere 默认支持 Kubernetes `v1.23~v1.34`。

命令执行完毕后将生成安装配置文件 `config-<Kubernetes version>.yaml`。

#### 4. 配置集群参数

在 `config-<Kubernetes version>.yaml` 中配置 Kubernetes 集群的信息：

| 参数 | 描述 |
|------|------|
| `zone` | 文件及镜像的下载区域。如果您访问 GitHub/GoogleAPIs 受限，请将该值设置为 `cn` |
| `kubernetes` | Kubernetes 相关配置 |
| `etcd` | etcd 相关配置 |
| `image_registry` | 私有镜像仓库相关配置 |
| `cri` | 容器运行时相关配置 |
| `cni` | 网络插件相关配置 |
| `storage_class` | 存储插件相关配置 |
| `dns` | 域名解析相关配置 |
| `image_manifests` | 需要下载的额外镜像 |

> **注意**：完整的配置参数说明请参考 [配置参考](../reference/config.md)。

#### 5. 安装 Kubernetes

执行以下命令安装 Kubernetes：

```shell
./kk create cluster -i inventory.yaml -c config.yaml
```

#### 6. 安装 KubeSphere

执行以下命令安装 KubeSphere：

```shell
chart=oci://hub.kubesphere.com.cn/kse/ks-core
version=1.2.4
helm upgrade --install -n kubesphere-system --create-namespace ks-core $chart \
  --debug --wait --version $version --reset-values --take-ownership \
  --set global.imageRegistry=hub.kubesphere.com.cn,extension.imageRegistry=hub.kubesphere.com.cn
```

> **注意**：Helm 版本需要 >= 3.17.0

如果显示如下信息，则表示 KubeSphere 安装成功：

```
NOTES:
Thank you for choosing KubeSphere Helm Chart.

Please be patient and wait for several seconds for the KubeSphere deployment to complete.

1. Wait for Deployment Completion

    Confirm that all KubeSphere components are running by executing the following command:

    kubectl get pods -n kubesphere-system

2. Access the KubeSphere Console

    Once the deployment is complete, you can access the KubeSphere console using the following URL:

    http://192.168.6.10:30880

3. Login to KubeSphere Console

    Use the following credentials to log in:

    Account: admin
    Password: P@88w0rd

NOTE: It is highly recommended to change the default password immediately after the first login.
```

### Web 页面安装

#### 1. 下载 KubeKey（含 Web Installer）

如果您访问 GitHub/GoogleAPIs 受限，请设置如下环境变量：

```shell
export KKZONE=cn
```

执行以下命令下载 KubeKey 最新版本（含 Web Installer）：

```shell
curl -sfL https://get-kk.kubesphere.io | SKIP_PACKAGE=true sh -
```

执行完成后，会在当前目录生成以下文件：

| 原文件 | 解压后文件 |
|--------|-----------|
| kubekey-v4.x.x-linux-amd64.tar.gz | kk：KubeKey 二进制文件 |
| web-installer.tgz | dist：Web 页面资源<br>host-check.yaml、kubernetes、kubesphere：任务模板文件<br>schema：配置文件<br>README.md：安装说明文档 |

#### 2. 启动 Web Installer

执行以下命令启动 Web Installer 页面：

```shell
./kk web --port 8080 --schema-path web-installer/schema --ui-path web-installer/dist
```

如果显示如下信息，表示 Web Installer 启动成功：

```
Web server started successfully on port 8080
```

请勿关闭命令终端，在浏览器中通过 `http://<启动节点 IP 地址>:8080` 打开 KubeKey 的 UI 页面。