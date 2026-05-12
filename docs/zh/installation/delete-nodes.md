# 删除集群节点

本节介绍如何使用 KubeKey 从 Kubernetes 集群中安全移除指定节点，包括从 Kubernetes 中驱逐和删除节点、清理 etcd 成员、卸载容器运行时及清理 DNS 配置。

## 前置条件

- 已有一个使用 KubeKey 部署的 Kubernetes 集群。
- 确保待删除节点上的业务已迁移或备份。
- 删除控制平面节点时，需确保集群中至少保留一个可用的控制平面节点。

> **注意**：当前 Web Installer 暂不支持删除集群节点，请通过命令行操作。

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

## 删除节点

执行以下命令删除指定节点：

```shell
./kk delete nodes node1 node2 -i inventory.yaml -c config.yaml
```

其中 `node1 node2` 为待删除节点的名称，需与 `inventory.yaml` 中定义的 host 名称一致。

KubeKey 会依次执行以下操作：

1. 若节点属于 etcd 集群且配置了 `--all` 或相关删除选项，先将该 etcd 成员下线。
2. 对 Kubernetes 节点执行 `cordon` 禁止新 Pod 调度，`drain` 驱逐现有的工作负载。
3. 若使用 Calico，执行 `calicoctl delete node` 清理网络资源。
4. 从 Kubernetes 集群中删除该节点。
5. 卸载 Kubernetes 组件和容器运行时（根据配置）。
6. 清理本地 DNS hosts 配置（根据配置）。

## 参数说明

| 参数 | 描述 |
|------|------|
| `-i, --inventory` | Inventory 文件路径，定义节点连接信息 |
| `-c, --config` | Config 文件路径，定义集群关键配置 |
| `--with-kubernetes` | 指定 Kubernetes 版本 |
| `--all` | 删除所有集群组件，包括 cri、etcd、dns 和 image_registry |
| `--with-data` | 同时删除数据目录（如 harbor 数据、registry 数据等），请谨慎使用 |
| `--override` | 执行成功后从 inventory.yaml 中移除已删除的节点 |
| `-a, --artifact` | 离线包路径，离线环境时使用 |

## 注意事项

- **业务迁移**：删除节点前，请确保该节点上的业务已迁移或备份，避免数据丢失。
- **控制平面节点**：删除控制平面节点会触发额外的安全校验，防止误删导致集群失控。请务必保证删除后集群中仍有可用的控制平面节点。
- **etcd 节点缩容**：当使用 external etcd 模式时，删除 etcd 节点需确保剩余的 etcd 成员数量为奇数，以维持集群高可用。
