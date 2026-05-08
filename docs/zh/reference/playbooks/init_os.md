# 初始化操作系统 (init_os.yaml)

`init_os.yaml` 用于初始化集群节点的操作系统环境，包括系统配置、软件包安装和证书准备，适用于集群安装前的节点准备工作。

## 执行流程

1. **全局初始化**
   - 在所有节点上执行 `native/root` 角色。

2. **加载默认变量**
   - 在所有节点上加载 `defaults` 角色。

3. **证书与资源准备**
   - 在 `localhost` 上执行：
     - `certs/init`：初始化并生成集群所需的证书。
     - `download`：下载 Kubernetes、容器运行时等所需软件包。

4. **节点系统初始化**
   - 针对 `etcd`、`k8s_cluster`、`image_registry`、`nfs` 组中的节点执行 `native` 角色，安装必要的系统软件包并完成基础环境配置。

## 说明

- 此 playbook 本身不会实际部署 Kubernetes 或 CRI；其主要目的是完成节点的操作系统级准备。
- 可在正式执行 `create_cluster.yaml` 前，使用此 playbook 批量初始化节点环境。
