# 添加节点 (add_nodes.yaml)

`add_nodes.yaml` 用于向已有的 Kubernetes 集群添加新节点，支持添加 etcd 节点、工作节点或控制平面节点。

## 执行流程

1. **全局初始化**
   - 在所有节点上执行 `native/root` 角色。

2. **Pre Install Hook**
   - 导入并执行 `hook/pre_install.yaml` 中的前置脚本。

3. **加载默认变量与预检查**
   - 在所有节点上加载默认配置（`defaults`）。
   - 执行 `precheck` 角色，检查新节点是否满足加入集群的条件。

4. **资源准备**
   - 在 `localhost` 上执行 `certs/init`，生成或更新证书。
   - 在 `localhost` 上执行 `download`，下载所需软件包和镜像。

5. **节点初始化**
   - 针对 `etcd`、`k8s_cluster`、`image_registry`、`nfs` 组中的所有节点执行 `native` 角色，安装基础软件包并配置系统环境。

6. **etcd 扩容（仅 external 模式）**
   - 针对 `etcd` 组中的节点，依次执行：
     - `etcd/prepare`
     - `etcd/backup`
     - `etcd/scaling_up/learner`
     - `etcd/install`
     - `etcd/scaling_up/promote`
     - `etcd/postprocess`
   - 以上步骤仅在 `etcd.deployment_type` 为 `external` 且节点属于 `need_installed_etcd` 列表时触发。

7. **同步 etcd 配置**
   - 在 `kube_control_plane` 节点上执行 `kubernetes/sync-etcd-config`，将 etcd 配置同步到控制平面。

8. **容器运行时与 Kubernetes 安装**
   - 针对 `k8s_cluster` 组中的节点执行：
     - `cri`：安装容器运行时。
     - `kubernetes/pre-kubernetes`：安装前置依赖。
     - `kubernetes/init-kubernetes`：初始化 Kubernetes。
     - `kubernetes/join-kubernetes`：将新节点加入集群（仅在节点尚未加载 Kubernetes 服务时触发）。
     - `kubernetes/certs`：分发或续期证书（仅在控制平面节点上且启用证书续期时触发）。
   - 上述角色会根据 `add_nodes` 列表进行过滤，仅对需要添加的节点生效。
