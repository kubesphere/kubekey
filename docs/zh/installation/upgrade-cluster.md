# 升级集群

本节介绍如何使用 KubeKey 升级已有的 Kubernetes 集群，包括 Kubernetes 控制平面和工作节点，以及可选的 etcd、容器运行时、CNI 插件和 StorageClass 存储插件。

## 前置条件

- 已有一个使用 KubeKey 部署的 Kubernetes 集群。
- 目标 Kubernetes 版本需高于集群当前安装的版本。
- 升级前请备份集群（尤其是 etcd 数据），以便升级失败时恢复。

> **注意**：当前 Web Installer 暂不支持升级集群，请通过命令行操作。

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

## 升级集群

默认情况下，`kk upgrade cluster` 只升级 Kubernetes 控制平面和工作节点二进制（`kubeadm` / `kubelet`），其他组件需显式开启。

### 仅升级 Kubernetes 版本

通过 `--with-kubernetes` 指定目标版本（或在 `config.yaml` 中设置 `kubernetes.kube_version`）：

```shell
./kk upgrade cluster -i inventory.yaml -c config.yaml --with-kubernetes v1.34.11
```

KubeKey 会进行滚动升级：第一个控制平面节点执行 `kubeadm upgrade apply`，其余控制平面节点执行 `kubeadm upgrade node`，所有工作节点并行升级。

### 同时升级相关组件

使用 `--all` 将 etcd、容器运行时、CNI 插件和 StorageClass 存储插件随 Kubernetes 一并升级：

```shell
./kk upgrade cluster -i inventory.yaml -c config.yaml --with-kubernetes v1.34.11 --all
```

也可用 `--set` 单独开启某个组件：

```shell
./kk upgrade cluster -i inventory.yaml -c config.yaml --with-kubernetes v1.34.11 --set upgrade.cni=true
```

### 仅升级单个组件

以下子命令只升级指定组件，不影响 Kubernetes 控制平面及其他组件：

```shell
./kk upgrade etcd -i inventory.yaml -c config.yaml
./kk upgrade cri -i inventory.yaml -c config.yaml
./kk upgrade cni -i inventory.yaml -c config.yaml
./kk upgrade storageclass -i inventory.yaml -c config.yaml
```

## 升级开关

`config.yaml` 中的 `upgrade` 段控制是否随集群一并升级可选组件：

```yaml
upgrade:
  etcd: false          # 是否升级 external etcd 集群
  cri: false           # 是否升级容器运行时（docker/containerd）
  cni: false           # 是否升级 CNI 插件
  storage_class: false # 是否升级 StorageClass 存储插件
```

以上开关也可通过命令行 `--all` 或 `--set upgrade.<component>=true` 覆盖。

> **注意**：CoreDNS / NodeLocalDNS 会随 Kubernetes 一并升级，无需单独开关；`image_registry` 和 `nfs` 不支持升级。

## 跨多个小版本升级

`kubeadm` 一次只允许升级一个小版本。当目标版本跨越多个小版本时，KubeKey 会自动逐个小版本升级，中间跳点使用 `cluster_require.kube_upgrade_path` 中定义的各小版本推荐补丁版本。可通过 `--set cluster_require.kube_upgrade_path.v1.24=v1.24.10` 覆盖某个跳点版本。

## 参数说明

| 参数 | 描述 |
|------|------|
| `-i, --inventory` | Inventory 文件路径，定义节点连接信息 |
| `-c, --config` | Config 文件路径，定义集群关键配置 |
| `--with-kubernetes` | 指定要升级到的目标 Kubernetes 版本，默认使用 config 中的 `kubernetes.kube_version` |
| `--all` | 升级所有相关组件，包括 etcd、cri、cni 和 storage_class |
| `--set` | 覆盖 config 中的值，格式 `--set key=val` 或 `--set k1=v1,k2=v2` |
| `-a, --artifact` | 离线包路径，离线环境时使用 |

## 注意事项

- **滚动升级**：KubeKey 对控制平面节点逐个升级，但工作节点是并行升级的，且不会自动执行 `cordon` 或 `drain`。若工作负载副本不足或集中于某些节点，服务可能短暂不可用。如需严格滚动升级，可先手动 `drain` 节点。
- **容器运行时**：当 `upgrade.cri=true` 时，容器运行时服务会重启。使用 containerd 时运行中的容器通常不受影响；使用 Docker 时，若未开启 `live-restore`，重启 dockerd 可能中断运行中的容器。
- **备份**：升级前请备份集群（尤其是 etcd 数据），以便升级失败时恢复。
