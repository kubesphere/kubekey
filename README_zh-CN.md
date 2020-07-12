# KubeKey

[![CI](https://github.com/kubesphere/kubekey/workflows/CI/badge.svg?branch=master&event=push)](https://github.com/kubesphere/kubekey/actions?query=event%3Apush+branch%3Amaster+workflow%3ACI+)

> [English](README.md) | 中文

从 v3.0.0 开始，[KubeSphere](https://kubesphere.io) 将基于 ansible 的安装程序更改为使用 Go 语言开发的名为 KubeKey 的新安装程序。使用 KubeKey，您可以轻松、高效、灵活地单独或整体安装 Kubernetes 和 KubeSphere。

有三种情况可以使用 KubeKey。

* 仅安装 Kubernetes
* 用一个命令中安装 Kubernetes 和 KubeSphere
* 首先安装 Kubernetes，然后使用 [ks-installer](https://github.com/kubesphere/ks-installer) 在其上部署 KubeSphere

> 重要提示：Kubekey 将会帮您安装 Kubernetes，若已有 Kubernetes 集群请参考 [在 Kubernetes 之上安装 KubeSphere](https://github.com/kubesphere/ks-installer/)。

## 优势

* 基于 Ansible 的安装程序具有大量软件依赖性，例如 Python。KubeKey 是使用 Go 语言开发的，可以消除在各种环境中出现的问题，从而提高安装成功率。
* KubeKey 使用 Kubeadm 在节点上尽可能多地并行安装 K8s 集群，以降低安装复杂性并提高效率。与较早的安装程序相比，它将大大节省安装时间。
* KubeKey 支持将群集从 all-in-one 扩展到多节点群集甚至 HA 集群。
* KubeKey 旨在将群集当作一个对象操作，即 CaaO。

## 支持的环境

### Linux 发行版

* **Ubuntu**  *16.04, 18.04*
* **Debian**  *Buster, Stretch*
* **CentOS/RHEL**  *7*

### <span id = "KubernetesVersions">Kubernetes 版本</span>

* **v1.15**: &ensp; *v1.15.12*
* **v1.16**: &ensp; *v1.16.12*
* **v1.17**: &ensp; *v1.17.8* (默认)
* **v1.18**: &ensp; *v1.18.5*

## 要求和建议

* 最低资源要求（仅对于最小安装 KubeSphere）：
  * 2 核虚拟 CPU
  * 4 GB 内存
  * 20 GB 储存空间

> /var/lib/docker 主要用于存储容器数据，在使用和操作过程中会逐渐增大。对于生产环境，建议 /var/lib/docker 单独挂盘。

* 操作系统要求：
  * `SSH` 可以访问所有节点。
  * 所有节点的时间同步。
  * `sudo`/`curl`/`openssl` 应在所有节点使用。
  * `ebtables`/`socat`/`ipset`/`conntrack` 应安装在所有节点。
  * `docker` 可以自己安装，也可以通过 KubeKey 安装。

> * 建议您的操作系统环境足够干净 (不安装任何其他软件)，否则可能会发生冲突。
> * 如果在从 dockerhub.io 下载镜像时遇到问题，建议准备一个容器镜像仓库 (加速器)。[为 Docker 守护程序配置镜像加速](https://docs.docker.com/registry/recipes/mirror/#configure-the-docker-daemon)。
> * 默认情况下，KubeKey 将安装 [OpenEBS](https://openebs.io/) 来为开发和测试环境配置 LocalPV，这对新用户来说非常方便。对于生产，请使用 NFS/Ceph/GlusterFS 或商业化存储作为持久化存储，并在所有节点中安装[相关的客户端](./docs/storage-client.md) 。

* 网络和 DNS 要求：
  * 确保 `/etc/resolv.conf` 中的 DNS 地址可用。否则，可能会导致群集中出现某些 DNS 问题。
  * 如果您的网络配置使用防火墙或安全组，则必须确保基础结构组件可以通过特定端口相互通信。建议您关闭防火墙或遵循链接配置：[网络访问](docs/network-access.md)。

## 用法

### 获取安装程序可执行文件

* 下载二进制文件

    ```shell script
    curl -O -k https://kubernetes.pek3b.qingstor.com/tools/kubekey/kk
    chmod +x kk
    ```

或者

* 从源代码生成二进制文件

    ```shell script
    git clone https://github.com/kubesphere/kubekey.git
    cd kubekey
    ./build.sh
    ```

> 注意：
>
> * 在构建之前，需要先安装 Docker。
> * 如果无法访问 `https://proxy.golang.org/`，比如在大陆，请执行 `build.sh -p`。

### 创建集群

#### 快速开始

快速入门使用 `all-in-one` 安装，这是熟悉 KubeSphere 的良好开始。

##### 命令

```shell script
./kk create cluster [--with-kubernetes version] [--with-kubesphere version]
```

##### 例子

* 使用默认版本创建一个纯 Kubernetes 集群

    ```shell script
    ./kk create cluster
    ```

* 创建指定一个（[支持的版本](#KubernetesVersions)）的 Kubernetes 集群

    ```shell script
    ./kk create cluster --with-kubernetes v1.17.8
    ```

* 创建一个部署了 KubeSphere 的 Kubernetes 集群 （例如 `--with-kubesphere v3.0.0`）

    ```shell script
    ./kk create cluster --with-kubesphere [version]
    ```

#### 高级用法

您可以使用高级安装来控制自定义参数或创建多节点群集。具体来说，通过指定配置文件来创建集群。

1. 首先，创建一个示例配置文件

    ```shell script
    ./kk create config [--with-kubernetes version] [--with-storage plugins] [--with-kubesphere version] [(-f | --file) path]
    ```

   **例子：**

   * 使用默认配置创建一个示例配置文件。您也可以指定文件名称或文件所在的文件夹。

        ```shell script
        ./kk create config [-f ~/myfolder/config-sample.yaml]
        ```

   * 同时安装存储插件 (支持：localVolume、nfsClient、rbd、glusterfs)。您可以指定多个插件并用逗号分隔。请注意，您添加的第一个将是默认存储类。

        ```shell script
        ./kk create config --with-storage localVolume
        ```

   * 同时安装 KubeSphere

        ```shell script
        ./kk create config --with-kubesphere
        ```

2. 根据您的环境修改配置文件 config-sample.yaml
3. 使用配置文件创建集群。

      ```shell script
      ./kk create cluster -f ~/myfolder/config-sample.yaml
      ```

### 启用多集群管理

默认情况下，Kubekey 将仅安装一个 Solo 模式的单集群，即未开启 Kubernetes 多集群联邦。如果您希望将 KubeSphere 作为一个支持多集群集中管理的中央面板，您需要在 [config-example.yaml](docs/config-example.md) 中设置 `ClusterRole`。关于多集群的使用文档，请参考 [如何启用多集群](https://github.com/kubesphere/community/blob/master/sig-multicluster/how-to-setup-multicluster-on-kubesphere/README_zh.md)。

### 开启可插拔功能组件

KubeSphere 从 2.1.0 版本开始对 Installer 的各功能组件进行了解耦，快速安装将默认仅开启最小化安装（Minimal Installation），Installer 支持在安装前或安装后自定义可插拔的功能组件的安装。使最小化安装更快速轻量且资源占用更少，也方便不同用户按需选择安装不同的功能组件。

KubeSphere 有多个可插拔功能组件，功能组件的介绍可参考 [配置示例](docs/config-example.md)。您可以根据需求，选择开启安装 KubeSphere 的可插拔功能组件。我们非常建议您开启这些功能组件来体验 KubeSphere 完整的功能以及端到端的解决方案。请在安装前确保您的机器有足够的 CPU 与内存资源。开启可插拔功能组件可参考 [开启可选功能组件](https://github.com/kubesphere/ks-installer/blob/master/README_zh.md#%E5%AE%89%E8%A3%85%E5%8A%9F%E8%83%BD%E7%BB%84%E4%BB%B6)。

### 添加节点

将新节点的信息添加到群集配置文件，然后应用更改。

```shell script
./kk scale -f config-sample.yaml
```

### 删除集群

您可以通过以下命令删除集群：

* 如果您以快速入门（all-in-one）开始：

```shell script
./kk delete cluster
```

* 如果从高级安装开始（使用配置文件创建的集群）：

```shell script
./kk delete cluster [-f config-sample.yaml]
```

### 启用 kubectl 自动补全

KubeKey 不会启用 kubectl 自动补全功能。请参阅下面的指南并将其打开：

**先决条件**：确保已安装 `bash-autocompletion` 并可以正常工作。

```shell script
# 安装 bash-completion
apt-get install bash-completion

# 将 completion 脚本添加到你的 ~/.bashrc 文件
echo 'source <(kubectl completion bash)' >>~/.bashrc

# 将 completion 脚本添加到 /etc/bash_completion.d 目录
kubectl completion bash >/etc/bash_completion.d/kubectl
```

更详细的参考可以在[这里](https://kubernetes.io/docs/tasks/tools/install-kubectl/#enabling-shell-autocompletion)找到。

## 相关文档

* [配置示例](docs/config-example.md)
* [网络访问](docs/network-access.md)
* [存储客户端](docs/storage-client.md)
* [路线图](docs/roadmap.md)
