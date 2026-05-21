# 删除集群 (delete_cluster.yaml)

`delete_cluster.yaml` 用于卸载整个 Kubernetes 集群及其相关组件，支持根据配置选择性清理资源。

## 执行流程

1. **全局初始化**
   - 在所有节点上执行 `native/root` 角色。
   - 在所有节点上加载 `defaults` 角色。

2. **卸载 Kubernetes 与容器运行时**
   - 针对 `k8s_cluster` 组中的节点执行：
     - `uninstall/kubernetes`：卸载 Kubernetes 组件。
     - `uninstall/cri`：卸载容器运行时（仅在 `delete.cri` 为 `true` 且当前节点不属于 `image_registry` 组时触发）。

3. **清理本地 DNS 配置**
   - 在需要卸载的节点上清理由 KubeKey 写入的本地 DNS（hosts）标记段。
   - 仅在 `delete.dns` 为 `true` 时触发。

4. **卸载 etcd（仅 external 模式）**
   - 针对 `etcd` 组中的节点执行 `etcd/scaling_down`。
   - 仅在 `delete.etcd` 为 `true` 且 `etcd.deployment_type` 为 `external` 时触发。

5. **卸载镜像仓库**
   - 针对 `image_registry` 组中的节点执行 `uninstall/image-registry`。
   - 仅在 `delete.image_registry` 为 `true` 时触发。

## 说明

- 删除集群前请确认已备份重要数据。
- 可通过 `delete` 配置项（如 `delete.cri`、`delete.etcd`、`delete.dns`、`delete.image_registry`）控制是否清理对应资源。
