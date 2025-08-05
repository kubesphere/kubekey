# architecture

![architecture](../../images/architecture.png)

## pre_hook

pre_hook 允许用户在创建集群之前，在对应的节点上执行脚本。

执行流程：
1. 复制本地目录的脚本文件到远程节点的 `/etc/kubekey/scripts/pre_install_{{ .inventory_hostname }}.sh`
2. 设置脚本文件权限为 0755
3. 遍历每个远程节点上 `/etc/kubekey/scripts/` 目录下所有 `pre_install_*.sh` 文件，并执行该脚本文件
 
 > work_dir: 工作目录，默认当前命令执行目录。     
 > inventory_hostname: Inventory.yaml 文件中定义的host对应的名称。

 ## precheck

precheck 集群安装前，对集群节点进行检查是否满足集群安装条件。

**os_precheck**: 操作系统检查，包括：
- **主机名检查**: 验证主机名格式是否符合规范（仅包含小写字母数字字符、'.'或'-'，且必须以字母数字字符开头和结尾）
- **操作系统版本检查**: 验证当前操作系统是否在支持的操作系统发行版列表中，除非允许不支持的发行版设置
- **架构检查**: 验证系统架构是否为支持的架构（amd64或arm64）
- **内存检查**: 
  - 对于主节点：验证内存是否满足最小主节点内存要求
  - 对于工作节点：验证内存是否满足最小工作节点内存要求
- **内核版本检查**: 验证内核版本是否满足最低版本要求
**kubernetes_precheck**: Kubernetes 相关检查，包括：
- **IP地址检查**: 验证节点是否定义了 internal_ipv4 或 internal_ipv6，两者不能同时为空
- **KubeVIP检查**: 当使用 kube_vip 类型的控制平面端点时，验证 kube_vip 地址是否有效且未被使用
- **Kubernetes版本检查**: 验证 Kubernetes 版本是否满足最低版本要求
- **已安装Kubernetes检查**: 验证节点上是否已安装 Kubernetes，如果已安装则检查版本是否与配置的 kube_version 匹配
**network_precheck**: 网络连通性检查，包括：
- **网络接口检查**: 验证节点上是否存在配置的 IPv4 或 IPv6 网络接口
- **CIDR 配置检查**: 验证 Pod CIDR 和 Service CIDR 配置格式是否正确（支持双栈格式：ipv4_cidr/ipv6_cidr 或 ipv4_cidr,ipv6_cidr）
- **双栈网络支持检查**: 当配置双栈网络时，验证 Kubernetes 版本是否支持（v1.20.0+）
- **网络插件检查**: 验证配置的网络插件是否在支持列表中
- **网络地址空间检查**: 确保节点上可用的网络地址空间足够容纳配置的最大 Pod 数量
- **Hybridnet 版本检查**: 当使用 Hybridnet 网络插件时，验证 Kubernetes 版本是否满足要求（v1.16.0+）
**etcd_precheck**: etcd 集群检查，包括：
- **部署类型校验**：校验 etcd 的部署类型（internal 或 external），并在 external 模式下确保 etcd 组不为空且节点数量为奇数
- **磁盘 IO 性能检查**：通过 fio 工具对 etcd 数据盘进行写入延迟测试，确保磁盘同步延迟（如 WAL fsync）满足集群要求
- **已安装 etcd 检查**：检测当前主机是否已安装 etcd 服务
**cri_precheck**: 容器运行时检查，包括：
- **容器管理器检查**: 验证配置的容器管理器是否在支持的列表中（docker 或 containerd）
- **containerd 版本检查**: 当使用 containerd 作为容器管理器时，验证 containerd 版本是否满足最低版本要求
**nfs_precheck**: NFS 存储检查，包括：
- **NFS 服务器数量检查**: 验证集群中只能有一个 NFS 服务器节点，确保 NFS 服务部署的唯一性
**image_registry_precheck**: 镜像仓库检查，包括：
- **镜像仓库必要软件检查**: 需检查 `docker_version` 和 `dockercompose_version` 均已配置且不为空。镜像仓库通过 docker_compose 进行安装，缺少必要软件会导致安装失败。

## init

init 阶段负责准备和构建集群安装所需的所有资源，包括：
- **软件包下载**: 下载 Kubernetes、容器运行时、网络插件等核心组件的二进制文件，确保所有必需的软件包都已准备就绪
- **Helm Chart 准备**: 获取和验证所需的 Helm Chart 包，为后续的应用程序部署做准备
- **容器镜像拉取**: 下载集群组件所需的 Docker 镜像，包括核心组件镜像和依赖镜像
- **离线包构建**: 当配置离线安装时，将所有依赖资源（二进制文件、镜像、Chart 包等）打包成完整的离线安装包
- **证书管理**: 生成集群安装和组件间通信所需的各种证书，包括 CA 证书、服务证书等

## install

install 阶段是 KubeKey 的核心安装阶段，负责在集群节点上实际部署和配置 Kubernetes 集群，包括：

**install nfs**: 为 `nfs` 组中的节点安装nfs服务。  
**install image_registry**: 为 `image_registry` 组中的节点安装镜像仓库。目前支持两种类型的镜像仓库：harbor，registry。  
**install etcd**: 为 `etcd` 组中的节点安装etcd。  
**install cri**: 为 `k8s_cluster` 组中的节点安装cri。目前支持两种CRI：docker，containerd。  
**kubernetes_install**: 为 `k8s_cluster` 组中的节点安装kubernetes。  
**install helm**: 为已安装好的kubernetes集群安装额外的helm 应用。包含：CNI（calico，cilium，flannel，hybridnet，kubeovn，multus）


## post_hook

post_hook 阶段在集群安装完成后执行，负责集群的最终配置和验证：

执行流程：
1. 复制本地目录的脚本文件到远程节点的 `/etc/kubekey/scripts/post_install_{{ .inventory_hostname }}.sh`
2. 设置脚本文件权限为 0755
3. 遍历每个远程节点上 `/etc/kubekey/scripts/` 目录下所有 `post_install_*.sh` 文件，并执行该脚本文件
 
 > **work_dir**: 工作目录，默认当前命令执行目录。     
 > **inventory_hostname**: Inventory.yaml 文件中定义的host对应的名称。

