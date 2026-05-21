<div align=center><img src="docs/images/kubekey-logo.svg?raw=true"></div>

[![CI](https://github.com/kubesphere/kubekey/actions/workflows/golangci-lint.yaml/badge.svg?branch=main)](https://github.com/kubesphere/kubekey/actions/workflows/golangci-lint.yaml)

> English | [中文](README_zh-CN.md)

**👋 Welcome to KubeKey!**

KubeKey is an open-source lightweight task flow execution tool. It provides a flexible and fast way to install Kubernetes.

> KubeKey has passed the [CNCF Kubernetes Conformance Certification](https://www.cncf.io/certification/software-conformance/)

# Comparison of new features in 3.x
1. Expanded from Kubernetes lifecycle management tool to task execution tool (flow design refers to [Ansible](https://github.com/ansible/ansible))
2. Supports multiple ways to manage task templates: git, local, etc.
3. Supports multiple node connection methods, including: local, ssh, kubernetes, prometheus.
4. Supports cloud-native automated batch task management
5. Advanced features: UI page (not yet open)

# Get KubeKey

## Method 1: Release Page

Get the corresponding binary files from the [Release](https://github.com/kubesphere/kubekey/releases) page.

## Method 2: Run Script

```shell
curl -sfL https://get-kk.kubesphere.io | sh -
```

| Original File | Extracted File |
|--------|--------|
| kubekey-v4.x.x-linux-amd64.tar.gz | kk: KubeKey binary |
| web-installer.tgz | dist: Web UI resources.<br>host-check.yaml, kubernetes, kubesphere: Task template files.<br>schema: Configuration files.<br>README.md: Installation documentation. |
| package.sh | Offline package build script. |

# Quick Start

## Method 1: Command Line

```shell
./kk create cluster
```

## Method 2: Web UI

**UI only supported after v4.0.0**

```shell
./kk web --schema-path web-installer/schema --ui-path web-installer/dist
```

# Documentation Navigation

- **[Install Kubernetes](docs/en/installation/README.md)**
  - [Component Versions](docs/en/installation/components.md)
  - [Online Installation](docs/en/installation/online.md)
  - [Offline Installation](docs/en/installation/offline.md)
  - [Add Cluster Nodes](docs/en/installation/add-nodes.md)
  - [Delete Cluster Nodes](docs/en/installation/delete-nodes.md)
- **[Configuration Reference](docs/en/reference/config.md)**
- **Playbooks**
  - [Create Kubernetes Cluster](docs/en/reference/playbooks/create_cluster.md)
  - [Delete Kubernetes Cluster](docs/en/reference/playbooks/delete_cluster.md)
  - [Add Nodes](docs/en/reference/playbooks/add_nodes.md)
  - [Delete Nodes](docs/en/reference/playbooks/delete_nodes.md)
  - [Renew Certificates](docs/en/reference/playbooks/certs_renew.md)
  - [Export Offline Artifact](docs/en/reference/playbooks/artifact_export.md)
  - [Pre-installation Check](docs/en/reference/playbooks/precheck.md)
  - [Initialize OS](docs/en/reference/playbooks/init_os.md)
- **[Image Registry Installation](docs/en/image-registry/README.md)**
- **[Dependency Packages](docs/en/dependency-packages/README.md)**
- **[Task Execution Framework](docs/en/framework/README.md)**
  - [Project](docs/en/framework/001-project.md)
  - [Playbook](docs/en/framework/002-playbook.md)
  - [Role](docs/en/framework/003-role.md)
  - [Task](docs/en/framework/004-task.md)
  - [Template Syntax](docs/en/framework/101-syntax.md)
  - [Variables](docs/en/framework/201-variable.md)
