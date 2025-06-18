# kubernetes cluster manager

内建的playbook实现了kubernetes的一整套生命周期管理，包含创建集群，删除集群，添加节点，删除节点，升级集群等。

## requirement

- 一台或多台运行兼容 deb/rpm 的 Linux 操作系统的计算机；例如：Ubuntu 或 CentOS。
- 每台机器 2 GB 以上的内存，内存不足时应用会受限制。
- 用作控制平面节点的计算机上至少有 2 个 CPU。
- 集群中所有计算机之间具有完全的网络连接。你可以使用公共网络或专用网络

## 构建inventory
默认的inventory配置如下：
```yaml
apiVersion: kubekey.kubesphere.io/v1
kind: Inventory
metadata:
  name: default
spec:
  hosts: # your can set all nodes here. or set nodes on special groups.
#    node1:
#      connector:
#        type: ssh
#        host: node1
#        port: 22
#        user: root
#        password: 123456
  groups:
    # all kubernetes nodes.
    k8s_cluster:
      groups:
        - kube_control_plane
        - kube_worker
    # control_plane nodes
    kube_control_plane:
      hosts:
        - localhost
    # worker nodes
    kube_worker:
      hosts:
        - localhost
    # etcd nodes when etcd_deployment_type is external
    etcd:
      hosts:
        - localhost
#    image_registry:
#      hosts:
#        - localhost
    # nfs nodes for registry storage. and kubernetes nfs storage
#    nfs:
#      hosts:
#        - localhost

```
inventory包含如下几个内置的group:
1. k8s_cluster: kubernetes集群。包含两个子group: kube_control_plane, kube_worker
2. kube_control_plane: kubernetes集群中的control_plane节点组
3. kube_worker: kubernetes集群中的worker节点组。
4. etcd: 安装etcd集群的节点组。
5. image_registry: 安装镜像仓库的节点组。（包含harbor，registry）
6. nfs: 安装nfs的节点组。

## 构建config
默认的 config 配置如下：


针对不同的kubernetes版本，给出了不同默认config配置作为参考:
- [安装 v1.23.x 版本的kubernetes 配置](../../../builtin/core/defaults/config/v1.23.yaml)
- [安装 v1.24.x 版本的kubernetes 配置](../../../builtin/core/defaults/config/v1.24.yaml)  
- [安装 v1.25.x 版本的kubernetes 配置](../../../builtin/core/defaults/config/v1.25.yaml)
- [安装 v1.26.x 版本的kubernetes 配置](../../../builtin/core/defaults/config/v1.26.yaml)
- [安装 v1.27.x 版本的kubernetes 配置](../../../builtin/core/defaults/config/v1.27.yaml)
- [安装 v1.28.x 版本的kubernetes 配置](../../../builtin/core/defaults/config/v1.28.yaml)
- [安装 v1.29.x 版本的kubernetes 配置](../../../builtin/core/defaults/config/v1.29.yaml)
- [安装 v1.30.x 版本的kubernetes 配置](../../../builtin/core/defaults/config/v1.30.yaml)
- [安装 v1.31.x 版本的kubernetes 配置](../../../builtin/core/defaults/config/v1.31.yaml)
- [安装 v1.32.x 版本的kubernetes 配置](../../../builtin/core/defaults/config/v1.32.yaml)
- [安装 v1.33.x 版本的kubernetes 配置](../../../builtin/core/defaults/config/v1.33.yaml)

## 安装集群
```shell
kk create cluster -i inventory.yaml -c config.yaml
```
`-i inventory.yaml`不传时，使用默认的inventory.yaml. 只会在执行的机器上安装kubernetes.
`-c config.yaml`不传时，使用默认的config.yaml. 安装 v1.33.1 版本的kubernetes