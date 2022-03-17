# NAME
**kk**: Kubernetes/KubeSphere Deploy Tool

# DESCRIPTION
KubeKey is an open-source lightweight tool for deploying Kubernetes clusters. It provides a flexible, rapid, and convenient way to install Kubernetes/K3s only, both Kubernetes/K3s and KubeSphere, and related cloud-native add-ons. It is also an efficient tool to scale and upgrade your cluster.

In addition, KubeKey also supports customized Air-Gap package, which is convenient for users to quickly deploy clusters in offline environments.

Use KubeKey in the following three scenarios.

* Install Kubernetes/K3s only
* Install Kubernetes/K3s and KubeSphere together in one command
* Install Kubernetes/K3s first, then deploy KubeSphere on it using [ks-installer](https://github.com/kubesphere/ks-installer)

# COMMANDS
| Command | Description |
| - | - |
| [kk add](./kk-add.md) | Add nodes to kubernetes cluster. |
| [kk artifact](./kk-artifact.md)| Manage a KubeKey offline installation package. |
| [kk certs](./kk-certs.md) | Manage cluster certs. |
| [kk completion](./kk-completion.md) | Generate shell completion scripts. |
| [kk create](./kk-create.md) | Create a cluster, a cluster configuration file or an offline installation package configuration file. |
| [kk delete](./kk-delete.md) | Delete node or cluster. |
| [kk init](./kk-init.md) | Initializes the installation environment. |
| [kk plugin](./kk-plugin.md) | Provides utilities for interacting with plugins. |
| [kk upgrade](./kk-upgrade.md) | Upgrade your cluster smoothly to a newer version with this command. |
| [kk version](./kk-version.md) | Print the client version information. |