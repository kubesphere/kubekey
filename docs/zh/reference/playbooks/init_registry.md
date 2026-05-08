# 初始化镜像仓库 (init_registry.yaml)

`init_registry.yaml` 用于在指定节点上初始化并部署私有镜像仓库（如 Harbor），包括预检查、资源下载、证书生成及仓库安装。

## 执行流程

1. **全局初始化**
   - 在所有节点上执行 `native/root` 角色。

2. **加载默认变量与预检查**
   - 在所有节点上加载 `defaults` 角色。
   - 执行 `precheck/image-registry`，检查镜像仓库节点是否满足安装条件。

3. **证书与资源准备**
   - 在 `localhost` 上执行：
     - `certs/init`：生成证书。
     - `download`：下载镜像仓库相关软件包和镜像。

4. **安装镜像仓库**
   - 针对 `image_registry` 组中的节点依次执行：
     - `native/init`：初始化系统环境。
     - `native/dns`：配置本地 DNS 解析。
     - `image-registry`：安装并配置私有镜像仓库。

## 说明

- 目前支持的镜像仓库类型包括 Harbor 和 docker-registry，具体由 `image_registry.type` 决定。
- 若使用 Harbor，请确保已正确配置 `docker_version` 与 `dockercompose_version`。
