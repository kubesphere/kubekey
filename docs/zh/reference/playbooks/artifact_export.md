# 离线包导出 (artifact_export.yaml)

`artifact_export.yaml` 用于导出完整的离线安装包（artifact），便于在无网络环境中部署集群。

## 执行流程

1. **全局初始化**
   - 在所有节点上执行 `native/root` 角色（带 `package` 标签）。

2. **加载默认变量**
   - 在所有节点上加载 `defaults` 角色。

3. **下载与打包**
   - 在 `localhost` 上执行：
     - `download`（带 `package` 标签）：下载所需二进制文件、镜像等资源。
     - `download/package`（带 `package` 标签）：将下载的资源打包为离线安装包。

## 说明

执行此 playbook 时，请确保已配置好所需下载的组件版本及镜像列表，以便打包的内容完整可用。
