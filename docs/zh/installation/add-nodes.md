# 添加集群节点

本节介绍如何使用 KubeKey 向已有的 Kubernetes 集群添加新节点，支持添加控制平面节点、工作节点和 etcd 节点。

## 前置条件

- 已有一个使用 KubeKey 部署的 Kubernetes 集群。
- 已准备待加入集群的新节点，并满足系统要求（参考 [安装 Kubernetes](README.md) 中的系统要求）。

> **注意**：当前 Web Installer 暂不支持添加集群节点，请通过命令行操作。

## 获取当前集群配置文件

如果集群是通过 **Web Installer** 安装的，可通过以下方式获取当前集群的配置文件。

### 获取 inventory.yaml

```shell
cp kubekey/runtime/kubekey.kubesphere.io/v1/inventories/default/default.yaml kkv4-inventory.yaml
```

### 获取 config.yaml

```shell
cat schema/config.json | jq '{spec: .["kubernetes.json"]}' > kkv4-config.json
```

## 方式一：通过 inventory.yaml 分组添加

此方式要求待添加的节点已预先在 `inventory.yaml` 中定义好连接信息，并分配到对应的分组（如 `kube_control_plane`、`kube_worker`、`etcd`）。

1. 确认 `inventory.yaml` 中已包含新节点的连接信息和分组配置。

   示例如下：

   ```yaml
   spec:
     hosts:
       node1:
         connector:
           type: ssh
           host: 192.168.1.101
           port: 22
           user: root
           password: 123456
     groups:
       kube_control_plane:
         hosts:
           - localhost
           - node1
       kube_worker:
         hosts:
           - localhost
           - node1
       etcd:
         hosts:
           - localhost
   ```

2. 执行以下命令添加节点：

   ```shell
   ./kk add nodes -i inventory.yaml -c config.yaml
   ```

   KubeKey 会自动识别 `inventory.yaml` 中已定义但尚未加入集群的节点，并将其安装为对应分组的角色。

## 方式二：通过命令行参数指定分组

此方式只需在 `inventory.yaml` 中定义节点的连接信息，无需事先分配到分组。通过命令行参数指定节点角色，并可使用 `--override` 自动更新 `inventory.yaml`。

1. 确认 `inventory.yaml` 中已定义待添加节点的连接信息。

   示例如下：

   ```yaml
   spec:
     hosts:
       node1:
         connector:
           type: ssh
           host: 192.168.1.101
           port: 22
           user: root
           password: 123456
       node2:
         connector:
           type: ssh
           host: 192.168.1.102
           port: 22
           user: root
           password: 123456
   ```

2. 执行以下命令添加节点并指定角色：

   ```shell
   ./kk add nodes --control-plane node1 --worker node2 -i inventory.yaml -c config.yaml --override
   ```

   - `--control-plane`：指定作为控制平面节点的主机名列表，多个节点用逗号分隔。
   - `--worker`：指定作为工作节点的主机名列表，多个节点用逗号分隔。
   - `--etcd`：指定作为 etcd 节点的主机名列表，多个节点用逗号分隔。
   - `--override`：执行成功后，自动将节点加入对应分组并更新 `inventory.yaml` 文件。

## 参数说明

| 参数 | 描述 |
|------|------|
| `-i, --inventory` | Inventory 文件路径，定义节点连接信息 |
| `-c, --config` | Config 文件路径，定义集群关键配置 |
| `--with-kubernetes` | 指定 Kubernetes 版本，默认使用集群当前版本 |
| `--control-plane` | 指定要添加为控制平面节点的节点列表 |
| `--worker` | 指定要添加为工作节点的节点列表 |
| `--etcd` | 指定要添加为 etcd 节点的节点列表 |
| `--override` | 执行成功后覆盖更新 inventory.yaml 文件 |
| `-a, --artifact` | 离线包路径，离线环境时使用 |
