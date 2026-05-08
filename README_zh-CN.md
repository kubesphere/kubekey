<div align=center><img src="docs/images/kubekey-logo.svg?raw=true"></div>

[![CI](https://github.com/kubesphere/kubekey/actions/workflows/golangci-lint.yaml/badge.svg?branch=main)](https://github.com/kubesphere/kubekey/actions/workflows/golangci-lint.yaml)

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

# 获取 KubeKey

## 方式一：Release 页面

在 [Release](https://github.com/kubesphere/kubekey/releases) 页面获取对应的二进制文件。

## 方式二：执行脚本

```shell
curl -sfL https://get-kk.kubesphere.io | sh -
```
| 原文件 | 解压后文件 |
| -------- | -------- |
| kubekey-v4.x.x-linux-amd64.tar.gz     | kk：为 KubeKey 的二进制文件   |
| web-installer.tgz | dist：Web 页面资源。<br>host-check.yaml，kubernetes，kubesphere：任务模板文件。<br>schema：配置文件。<br>README.md：安装说明文档。 |
| package.sh | 离线安装包的构建脚本。 |

# 快速开始

## 方式一：命令行

```shell
kk create cluster
```

## 方式二：Web 页面
**UI 页面仅在 v4.0.0 及以上版本提供支持**

```shell
kk web --schema-path schema --ui-path dist
```

# 文档导航

- **[安装 Kubernetes](docs/zh/installation/README.md)**
  - [安装组件说明](docs/zh/installation/components.md)
  - [在线安装](docs/zh/installation/online.md)
  - [离线安装](docs/zh/installation/offline.md)
- **[配置参考](docs/zh/reference/config.md)**
- **任务模板**
  - [创建 Kubernetes 集群](docs/zh/reference/playbooks/create_cluster.md)
  - [删除 Kubernetes 集群](docs/zh/reference/playbooks/delete_cluster.md)
  - [添加节点](docs/zh/reference/playbooks/add_nodes.md)
  - [删除节点](docs/zh/reference/playbooks/delete_nodes.md)
  - [证书续期](docs/zh/reference/playbooks/certs_renew.md)
  - [制作离线包](docs/zh/reference/playbooks/artifact_export.md)
  - [安装前检查](docs/zh/reference/playbooks/precheck.md)
  - [初始化操作系统](docs/zh/reference/playbooks/init_os.md)
- **[镜像仓库安装](docs/zh/image-registry/README.md)**
- **[依赖包管理](docs/zh/dependency-packages/README.md)**
- **[任务执行框架](docs/zh/framework/README.md)**
  - [项目](docs/zh/framework/001-project.md)
  - [流程](docs/zh/framework/002-playbook.md)
  - [角色](docs/zh/framework/003-role.md)
  - [任务](docs/zh/framework/004-task.md)
  - [模板语法](docs/zh/framework/101-syntax.md)
  - [变量](docs/zh/framework/201-variable.md)