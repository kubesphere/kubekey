# 集群预检查 (precheck.yaml)

`precheck.yaml` 用于在集群安装或扩容前对节点进行环境和条件检查，确保满足操作系统、Kubernetes、网络、etcd、容器运行时及存储等各项要求。

## 执行流程

1. **全局初始化**
   - 在所有节点上执行 `native/root` 角色（标签：`always`）。

2. **加载默认变量与检查**
   - 在所有节点上加载 `defaults` 角色（标签：`always`）。
   - 执行 `precheck` 角色，完成以下分项检查：
     - **操作系统检查**：主机名规范、支持的发行版、系统架构、内存、内核版本。
     - **Kubernetes 检查**：IP 地址配置、KubeVIP 有效性、Kubernetes 版本兼容性、已安装的 Kubernetes 版本匹配。
     - **网络检查**：网络接口、CIDR 格式、双栈支持、网络插件合法性、可用地址空间。
     - **etcd 检查**：部署类型校验、磁盘 IO 性能、已安装 etcd 检测。
     - **容器运行时检查**：容器管理器支持、containerd 最低版本。
     - **NFS 检查**：NFS 服务器节点唯一性。
     - **镜像仓库检查**：必要软件（Docker、Docker Compose）是否已配置。

## 按分类筛选检查

precheck 支持按分类（tag）对检查项进行筛选，便于只运行或跳过特定类别的检查：

- **白名单（只跑指定的类）**：`kk precheck --tags cri,os` 仅执行容器运行时（cri）与操作系统（os）检查，其余类别全部跳过；也可用位置参数：`kk precheck cri os`。
- **黑名单（跳过指定的类）**：`kk precheck --skip-tags cri` 执行除容器运行时以外的所有检查。`--skip-tags` 是“关闭某一类校验、其余照跑”的推荐用法。

不指定任何筛选时，`kk precheck` 会运行全部检查类别。标签为 `always` 的角色（如 `native/root`、`defaults`）始终执行，不受筛选影响。

可用的分类标签：`artifact`、`cri`、`cni`、`storageclass`、`os`、`network`、`storage`、`kubernetes`、`etcd`、`nfs`。

## 说明

- 任何一项检查未通过，playbook 将停止执行，并返回对应的错误信息。
- 建议在正式安装或添加节点前始终先运行此 playbook，以避免中途失败。
