# 删除节点 (delete_nodes.yaml)

`delete_nodes.yaml` 用于安全地从集群中移除指定的节点，包括从 Kubernetes 中驱逐和删除节点、etcd 下线、CRI 卸载及 DNS 清理。

## 执行流程

1. **全局初始化**
   - 在所有节点上执行 `native/root` 角色。
   - 在所有节点上加载 `defaults` 角色。

2. **etcd 节点缩容（仅 external 模式）**
   - 针对 `etcd` 组中需要卸载的节点依次执行：
     - `etcd/prepare`
     - `etcd/scaling_down`
     - `etcd/postprocess`
   - 仅在 `delete.etcd` 为 `true` 且节点在 `need_uninstall_etcd` 列表中时触发。

3. **确保控制平面节点可用**
   - 在 `kube_control_plane` 节点上执行前置检查，确保删除后集群中至少保留一个控制平面节点。
   - 同时执行 `kubernetes/sync-etcd-config`，将 etcd 配置同步到剩余控制平面。

4. **从 Kubernetes 集群中移除节点**
   - 针对 `k8s_cluster` 组中待删除的节点执行前置任务：
     - `kubectl cordon`：禁止新 Pod 调度到该节点。
     - `kubectl drain`：驱逐节点上的工作负载。
     - 若使用 Calico，执行 `calicoctl delete node`。
     - `kubectl delete node`：从集群中删除该节点。
   - 随后执行：
     - `uninstall/kubernetes`：卸载 Kubernetes 组件。
     - `uninstall/cri`：卸载容器运行时（若配置 `delete.cri` 且节点不属于 `image_registry` 组）。

5. **清理本地 DNS 配置**
   - 清理 KubeKey 写入的本地 hosts 标记段。
   - 仅在 `delete.dns` 为 `true` 且节点在 `delete_nodes` 列表中时触发。

6. **卸载 etcd 与镜像仓库**
   - 在对应节点上执行 `etcd` 角色以卸载 etcd（若 `delete.etcd` 启用）。
   - 在 `image_registry` 节点上执行 `uninstall/image-registry`（若 `delete.image_registry` 启用）。

## 说明

- 删除节点前请确保该节点上的业务已迁移或备份。
- 控制平面节点删除操作会额外触发安全校验，防止误删导致集群失控。
