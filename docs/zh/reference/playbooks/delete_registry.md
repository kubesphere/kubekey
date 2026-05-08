# 删除镜像仓库 (delete_registry.yaml)

`delete_registry.yaml` 用于卸载已部署的私有镜像仓库（如 Harbor 或 docker-registry），并清理本地 DNS 解析配置。

## 执行流程

1. **全局初始化**
   - 在所有节点上执行 `native/root` 角色。
   - 在所有节点上加载 `defaults` 角色。

2. **卸载镜像仓库**
   - 针对 `image_registry` 组中的节点执行 `uninstall/image-registry`。

3. **清理本地 DNS 配置**
   - 在 `image_registry` 节点上清理 KubeKey 写入的本地 DNS（hosts）标记段。
   - 仅在 `delete.dns` 为 `true` 时触发。

## 说明

- 删除镜像仓库前，请确保该仓库中已无集群运行所依赖的镜像，或已做好镜像迁移。
- 此操作仅卸载镜像仓库服务，不会删除集群中的容器运行时或 Kubernetes 组件。
