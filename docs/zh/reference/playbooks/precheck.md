# 集群预检查 (precheck.yaml)

`precheck.yaml` 用于在集群安装或扩容前对节点进行环境和条件检查，确保满足操作系统、Kubernetes、网络、etcd、容器运行时及存储等各项要求。

## 执行流程

1. **全局初始化**
   - 在所有节点上执行 `native/root` 角色（标签：`always`）。

2. **加载默认变量与检查**
   - 在所有节点上加载 `defaults` 角色（标签：`always`）。
   - 执行 `precheck` 角色，完成以下分项检查：
     - **操作系统检查**：主机名规范、支持的发行版、系统架构、内存、内核版本、CPU 是否支持 Calico 所需的 x86-64-v2 微架构级别。
     - **Kubernetes 检查**：IP 地址配置、KubeVIP 有效性、Kubernetes 版本兼容性、已安装的 Kubernetes 版本匹配。
     - **网络检查**：网络接口、CIDR 格式、双栈支持、网络插件合法性、可用地址空间。
     - **etcd 检查**：部署类型校验、磁盘 IO 性能、已安装 etcd 检测。
     - **容器运行时检查**：容器管理器支持、containerd 最低版本、containerd v2 与节点 glibc / 内核的兼容性（glibc 低于 2.35 的节点需设置 `cri.containerd.static_binary=true`）、runc 版本与内核的兼容性（runc ≥ v1.3.6 遮掩路径时挂载 `nr_blocks=1,nr_inodes=1`，要求内核 ≥ 5.15）。
     - **NFS 检查**：NFS 服务器节点唯一性。
     - **镜像仓库检查**：必要软件（Docker、Docker Compose）是否已配置。

## 说明

- 任何一项检查未通过，playbook 将停止执行，并返回对应的错误信息。
- 建议在正式安装或添加节点前始终先运行此 playbook，以避免中途失败。
