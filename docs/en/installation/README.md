# Install Kubernetes

This document explains how to install a Kubernetes cluster using KubeKey.

- Supported deployment environments: Linux distributions
- Supported Kubernetes versions: v1.23.x ~ v1.34.x

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

- [Config for installing Kubernetes v1.23.x](../../../builtin/core/defaults/config/v1.23.yaml)
- [Config for installing Kubernetes v1.24.x](../../../builtin/core/defaults/config/v1.24.yaml)
- [Config for installing Kubernetes v1.25.x](../../../builtin/core/defaults/config/v1.25.yaml)
- [Config for installing Kubernetes v1.26.x](../../../builtin/core/defaults/config/v1.26.yaml)
- [Config for installing Kubernetes v1.27.x](../../../builtin/core/defaults/config/v1.27.yaml)
- [Config for installing Kubernetes v1.28.x](../../../builtin/core/defaults/config/v1.28.yaml)
- [Config for installing Kubernetes v1.29.x](../../../builtin/core/defaults/config/v1.29.yaml)
- [Config for installing Kubernetes v1.30.x](../../../builtin/core/defaults/config/v1.30.yaml)
- [Config for installing Kubernetes v1.31.x](../../../builtin/core/defaults/config/v1.31.yaml)
- [Config for installing Kubernetes v1.32.x](../../../builtin/core/defaults/config/v1.32.yaml)
- [Config for installing Kubernetes v1.33.x](../../../builtin/core/defaults/config/v1.33.yaml)
- [Config for installing Kubernetes v1.34.x](../../../builtin/core/defaults/config/v1.34.yaml)

For the complete configuration reference, see [Configuration Reference](../reference/config.md).

## Install Cluster

KubeKey supports both **online installation** and **offline installation**.

### Method 1: Online Installation

When installing online, KubeKey automatically downloads required Kubernetes components and images from the Internet. For detailed steps, see [Online Installation](online.md).

### Method 2: Offline Installation

Offline installation is suitable for air-gapped environments, requiring an artifact package and system dependencies in advance. For detailed steps, see [Offline Installation](offline.md).

## Enable kubectl autocompletion

KubeKey doesn't enable kubectl autocompletion. Refer to the guide below and turn it on:

**Prerequisite**: make sure bash-autocompletion is installed and works.

```shell script
# Install bash-completion
apt-get install bash-completion

# Source the completion script in your ~/.bashrc file
echo 'source <(kubectl completion bash)' >>~/.bashrc

# Add the completion script to the /etc/bash_completion.d directory
kubectl completion bash >/etc/bash_completion.d/kubectl
```

More detail reference could be found [here](https://kubernetes.io/docs/tasks/tools/install-kubectl/#enabling-shell-autocompletion).
