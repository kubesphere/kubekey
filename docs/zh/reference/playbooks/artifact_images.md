# 离线镜像同步 (artifact_images.yaml)

`artifact_images.yaml` 用于拉取集群所需的容器镜像，并将其推送到私有镜像仓库，常用于离线环境的镜像准备。

## 执行流程

1. **全局初始化**
   - 在所有节点上执行 `native/root` 角色。

2. **加载默认变量**
   - 在 `localhost` 上加载 `defaults` 角色。

3. **镜像下载与推送**
   - 在 `localhost` 上依次执行：
     - `download`：下载所需资源。
     - `image-registry/pull`（带 `pull` 标签）：拉取容器镜像到本地。
     - `image-registry/push`（带 `push` 标签）：将拉取的镜像推送到配置的私有镜像仓库。

## 说明

- 该 playbook 通常在 `localhost` 上执行。
- 推送目标仓库由 `image_registry` 配置决定。
