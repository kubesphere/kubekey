# Install Kubernetes

This document explains how to install a Kubernetes cluster using KubeKey.

- Supported deployment environments: Linux distributions
- Supported Kubernetes versions: v1.23.x ~ v1.34.x

> **Documentation Navigation**
> - [Installation Architecture](architecture.md)：pre_hook → precheck → init → install → post_hook
> - [Component Versions](components.md)：etcd, CNI, storage, DNS and other component version compatibility
> - [Online Installation](online.md)：Install in an Internet-accessible environment
> - [Offline Installation](offline.md)：Install using offline packages in an air-gapped environment
> - [Configuration Reference](../reference/config-reference.md)：Complete `config.yaml` configuration guide
> - [Image Registry Installation](../image-registry/README.md)：Deploy Harbor or docker-registry, including HA solutions
> - [Dependency Packages](../dependency-packages/README.md)：System dependency package management

## System Requirements

- One or more computers running Linux operating systems compatible with deb/rpm, such as Ubuntu or CentOS.
- Each machine should have more than 2 GB of memory; applications will be limited if memory is insufficient.
- Control plane nodes should have at least 2 CPUs.
- Full network connectivity among all machines in the cluster. You can use public or private networks.

### System Dependencies

Kubernetes requires the following OS dependencies to be pre-installed:

`socat` `conntrack` `ipset` `ebtables` `chrony` `ipvsadm`

KubeKey provides pre-compiled dependency packages for some Linux distributions, available at [iso-latest](https://github.com/kubesphere/kubekey/releases/tag/iso-latest).
For supported distributions and build methods, see [Dependency Packages](../dependency-packages/README.md).

## Define Node Information

KubeKey uses the `Inventory` resource to define node connection information.
You can use `kk create inventory` to get the default `inventory.yaml` resource. The default `inventory.yaml` configuration is as follows:

```yaml
apiVersion: kubekey.kubesphere.io/v1
kind: Inventory
metadata:
  name: default
spec:
  hosts: # You can set all nodes here, or in specific groups.
#    node1:
#      connector:
#        type: ssh
#        host: node1
#        port: 22
#        user: root
#        password: 123456
  groups:
    # All Kubernetes nodes
    k8s_cluster:
      groups:
        - kube_control_plane
        - kube_worker
    # Control plane nodes
    kube_control_plane:
      hosts:
        - localhost
    # Worker nodes
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
    # NFS nodes for registry storage and Kubernetes NFS storage
#    nfs:
#      hosts:
#        - localhost
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `spec.hosts` | Object | Yes | Host list, key is the host name, value is the host configuration |
| `spec.hosts.<name>.connector` | Object | Yes | Host connection configuration |
| `spec.hosts.<name>.connector.host` | String | Yes | SSH target host IP or domain name |
| `spec.hosts.<name>.connector.private_key` | String | No | SSH private key path, uses system default key if not specified |
| `spec.hosts.<name>.internal_ipv4` | String | No | Internal IPv4 address of the host, used for /etc/hosts domain resolution |
| `spec.groups` | Object | Yes | Node group configuration |
| `spec.groups.k8s_cluster` | Object | Yes | Kubernetes cluster. Contains two sub-groups: `kube_control_plane`, `kube_worker` |
| `spec.groups.kube_control_plane` | Object | Yes | Control plane node group in the Kubernetes cluster |
| `spec.groups.kube_worker` | Object | Yes | Worker node group in the Kubernetes cluster |
| `spec.groups.etcd` | Object | No | Node group for installing the etcd cluster |
| `spec.groups.image_registry` | Object | Yes | Image registry node group, specifies which hosts are used to deploy the image registry |
| `spec.groups.nfs` | Object | No | NFS node group |
| `spec.groups.<group name>.hosts` | Array | Yes | Node name list for the corresponding group |

## Define Key Configuration

KubeKey uses the `Config` resource to define key cluster configuration.
You can use `kk create config --with-kubernetes v1.33.1` to get the default `config.yaml` resource.

Default config references for different Kubernetes versions:

- [Config for installing Kubernetes v1.23.x](../../builtin/core/defaults/config/v1.23.yaml)
- [Config for installing Kubernetes v1.24.x](../../builtin/core/defaults/config/v1.24.yaml)
- [Config for installing Kubernetes v1.25.x](../../builtin/core/defaults/config/v1.25.yaml)
- [Config for installing Kubernetes v1.26.x](../../builtin/core/defaults/config/v1.26.yaml)
- [Config for installing Kubernetes v1.27.x](../../builtin/core/defaults/config/v1.27.yaml)
- [Config for installing Kubernetes v1.28.x](../../builtin/core/defaults/config/v1.28.yaml)
- [Config for installing Kubernetes v1.29.x](../../builtin/core/defaults/config/v1.29.yaml)
- [Config for installing Kubernetes v1.30.x](../../builtin/core/defaults/config/v1.30.yaml)
- [Config for installing Kubernetes v1.31.x](../../builtin/core/defaults/config/v1.31.yaml)
- [Config for installing Kubernetes v1.32.x](../../builtin/core/defaults/config/v1.32.yaml)
- [Config for installing Kubernetes v1.33.x](../../builtin/core/defaults/config/v1.33.yaml)
- [Config for installing Kubernetes v1.34.x](../../builtin/core/defaults/config/v1.34.yaml)

For the complete configuration reference, see [Configuration Reference](../reference/config-reference.md).

## Install Cluster

KubeKey supports both **online installation** and **offline installation**.

### Method 1: Online Installation

When installing online, KubeKey automatically downloads required Kubernetes components and images from the Internet. For detailed steps, see [Online Installation](online.md).

#### Command Line

After preparing `inventory.yaml` and `config.yaml`, run the built-in `kk create cluster` subcommand, which automatically runs `playbooks/create_cluster.yaml` without requiring you to specify the playbook path.

```shell
kk create cluster -i inventory.yaml -c config.yaml
```

If `-i inventory.yaml` is not provided, the default `inventory.yaml` is used. Kubernetes will only be installed on the executing machine.
If `-c config.yaml` is not provided, the default `config.yaml` is used, which installs Kubernetes v1.34.1.

Common parameters:

- `--workdir`: KubeKey working directory (default: `<current-dir>/kubekey`).
- `--with-kubernetes`: Kubernetes version (e.g. `v1.33.1`), used when not specified in `config.yaml`.
- `--set`: Override configuration items, e.g. `--set download.fetch=false`. Specific nested keys follow KubeKey's `--set` syntax.
- `-n` / `--namespace`: Namespace for local runtime resources tied to the playbook.

#### Web UI

Requires **KubeKey v4.0.0 or newer** with the web installer bundle. Download method is described in the README under **Get KubeKey — Method 2: Run script** (web installer bundle is included by default). Start it like this:

```shell
kk web --schema-path schema --ui-path dist
```

In the browser, maintain nodes and configuration, and follow the UI flow to create the cluster (corresponds to playbook `playbooks/create_cluster.yaml`).

### Method 2: Offline Installation

Offline installation is suitable for air-gapped environments, requiring an artifact package and system dependencies in advance. For detailed steps, see [Offline Installation](offline.md).

External references:

- [KubeSphere Offline Installation Documentation](https://docs.kubesphere.com.cn/v4.2.1/03-installation-and-upgrade/02-install-kubesphere/02-offline-install-kubernetes-and-kubesphere/)
- [KubeKey Offline Installation Supplement](https://hackmd.io/@YZ8ZcD08T4O18Xzx7O5yNw/rJTGXCrpZe)

After preparing the offline artifact package, specify the artifact path via `-a` / `--artifact`. KubeKey will disable online pulling and use local resources:

```shell
kk create cluster -i inventory.yaml -c config.yaml --artifact /path/to/kubekey-artifact.tgz
```
