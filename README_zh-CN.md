<div align=center><img src="docs/images/kubekey-logo.svg?raw=true"></div>

[![CI](https://github.com/kubesphere/kubekey/workflows/GolangCILint/badge.svg?branch=main&event=push)](https://github.com/kubesphere/kubekey/actions/workflows/golangci-lint.yaml?query=event%3Apush+branch%3Amain+workflow%3ACI)

> [English](README.md) | 中文

**👋 欢迎使用KubeKey!**

KubeKey 是一个开源的轻量的任务流程执行工具。提供了一种灵活、快速的方式来安装kubernetes。

> KubeKey 通过了 [CNCF kubernetes 一致性认证](https://www.cncf.io/certification/software-conformance/)

# 对比3.x新特性
1. 从kubernetes生命周期管理工具扩展为任务执行工具(流程设计参考[Ansible](https://github.com/ansible/ansible))
2. 支持多种方式管理任务模版：git，本地等。
3. 支持多种节点连接方式。包括：local、ssh、kubernetes、prometheus。
4. 支持云原生方式自动化批量任务管理
5. 高级特性：UI页面（暂未开放）

# 安装kubekey

## kubernetes中安装
通过helm安装kubekey。
```shell
helm upgrade --install --create-namespace -n kubekey-system kubekey config/kubekey
```

## 二进制
在 [release](https://github.com/kubesphere/kubekey/releases) 页面获取对应的二进制文件。

# 部署kubernetes

- 支持部署环境：Linux发行版
    - almaLinux: 9.0 (未充分测试)
    - centOS: 8
    - debian: 10, 11
    - kylin: V10SP3 (未充分测试)
    - ubuntu: 18.04, 20.04, 22.04, 24.04.

- 支持的Kubernetes版本：v1.23.x ~ v1.33.x

## requirement

- 一台或多台运行兼容 deb/rpm 的 Linux 操作系统的计算机；例如：Ubuntu 或 CentOS。
- 每台机器 2 GB 以上的内存，内存不足时应用会受限制。
- 用作控制平面节点的计算机上至少有 2 个 CPU。
- 集群中所有计算机之间具有完全的网络连接。你可以使用公共网络或专用网络

## 定义节点信息

kubekey使用 `inventory` 资源来定义节点的连接信息。    
可使用 `kk create inventory` 来获取默认的inventory.yaml 资源。默认的`inventory.yaml`配置如下：    
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

## 定义关键配置信息

kubekey使用 `config` 资源来定义节点的连接信息。    
可使用 `kk create config --with-kubernetes v1.33.1` 来获取默认的inventory.yaml 资源。默认的`config.yaml`配置如下：    

针对不同的kubernetes版本，给出了不同默认config配置作为参考:
- [安装 v1.23.x 版本的kubernetes 配置](builtin/core/defaults/config/v1.23.yaml)
- [安装 v1.24.x 版本的kubernetes 配置](builtin/core/defaults/config/v1.24.yaml)  
- [安装 v1.25.x 版本的kubernetes 配置](builtin/core/defaults/config/v1.25.yaml)
- [安装 v1.26.x 版本的kubernetes 配置](builtin/core/defaults/config/v1.26.yaml)
- [安装 v1.27.x 版本的kubernetes 配置](builtin/core/defaults/config/v1.27.yaml)
- [安装 v1.28.x 版本的kubernetes 配置](builtin/core/defaults/config/v1.28.yaml)
- [安装 v1.29.x 版本的kubernetes 配置](builtin/core/defaults/config/v1.29.yaml)
- [安装 v1.30.x 版本的kubernetes 配置](builtin/core/defaults/config/v1.30.yaml)
- [安装 v1.31.x 版本的kubernetes 配置](builtin/core/defaults/config/v1.31.yaml)
- [安装 v1.32.x 版本的kubernetes 配置](builtin/core/defaults/config/v1.32.yaml)
- [安装 v1.33.x 版本的kubernetes 配置](builtin/core/defaults/config/v1.33.yaml)

## 安装集群
```shell
kk create cluster -i inventory.yaml -c config.yaml
```
`-i inventory.yaml`不传时，使用默认的inventory.yaml. 只会在执行的机器上安装kubernetes.
`-c config.yaml`不传时，使用默认的config.yaml. 安装 v1.33.1 版本的kubernetes

# 文档
**[项目模版编写规范](docs/zh/001-project.md)**  
**[模板语法](docs/zh/101-syntax.md)**  
**[参数定义](docs/zh/201-variable.md)**    
**[集群管理](docs/zh/core/README.md)**    

