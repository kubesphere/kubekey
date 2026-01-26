<div align=center><img src="docs/images/kubekey-logo.svg?raw=true"></div>

[![CI](https://github.com/kubesphere/kubekey/workflows/GolangCILint/badge.svg?branch=main&event=push)](https://github.com/kubesphere/kubekey/actions/workflows/golangci-lint.yaml?query=event%3Apush+branch%3Amain+workflow%3ACI)

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

# Install kubekey

## Install in Kubernetes
Install kubekey via helm.
```shell
helm upgrade --install --create-namespace -n kubekey-system kubekey config/kubekey
```

## Binary
Get the corresponding binary files from the [release](https://github.com/kubesphere/kubekey/releases) page.

## Download Binary with UI

**UI only support after v4.0.0**

```shell
VERSION=v4.0.0 WEB_INSTALLER_VERSION=v1.0.0 hack/downloadKubekey.sh
# run with UI
kk web --schema-path schema --ui-path dist
```
> If there is a config.yaml file in the current directory, running `./package.sh config.yaml` to build an offline package.

# Deploy Kubernetes

- Supported deployment environments: Linux distributions
    - almaLinux: 9.0 (not fully tested)
    - centOS: 8
    - debian: 10, 11
    - kylin: V10SP3 (not fully tested)
    - ubuntu: 18.04, 20.04, 22.04, 24.04.

- Supported Kubernetes versions: v1.23.x ~ v1.33.x

## Requirements

- One or more computers running Linux operating systems compatible with deb/rpm; for example: Ubuntu or CentOS.
- Each machine should have more than 2 GB of memory; applications will be limited if memory is insufficient.
- Control plane nodes should have at least 2 CPUs.
- Full network connectivity among all machines in the cluster. You can use public or private networks.

## Define node information

kubekey uses the `inventory` resource to define node connection information.    
You can use `kk create inventory` to get the default inventory.yaml resource. The default `inventory.yaml` configuration is as follows:    
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
The inventory contains the following built-in groups:
1. k8s_cluster: Kubernetes cluster. Contains two subgroups: kube_control_plane, kube_worker
2. kube_control_plane: control_plane node group in the Kubernetes cluster
3. kube_worker: worker node group in the Kubernetes cluster.
4. etcd: node group for installing etcd cluster.
5. image_registry: node group for installing image registry (including harbor, registry)
6. nfs: node group for installing nfs.

## Define key configuration information

kubekey uses the `config` resource to define node connection information.    
You can use `kk create config --with-kubernetes v1.33.1` to get the default inventory.yaml resource. The default `config.yaml` configuration is as follows:    

Default config configurations are provided as references for different Kubernetes versions:
- [Config for installing Kubernetes v1.23.x](builtin/core/defaults/config/v1.23.yaml)
- [Config for installing Kubernetes v1.24.x](builtin/core/defaults/config/v1.24.yaml)  
- [Config for installing Kubernetes v1.25.x](builtin/core/defaults/config/v1.25.yaml)
- [Config for installing Kubernetes v1.26.x](builtin/core/defaults/config/v1.26.yaml)
- [Config for installing Kubernetes v1.27.x](builtin/core/defaults/config/v1.27.yaml)
- [Config for installing Kubernetes v1.28.x](builtin/core/defaults/config/v1.28.yaml)
- [Config for installing Kubernetes v1.29.x](builtin/core/defaults/config/v1.29.yaml)
- [Config for installing Kubernetes v1.30.x](builtin/core/defaults/config/v1.30.yaml)
- [Config for installing Kubernetes v1.31.x](builtin/core/defaults/config/v1.31.yaml)
- [Config for installing Kubernetes v1.32.x](builtin/core/defaults/config/v1.32.yaml)
- [Config for installing Kubernetes v1.33.x](builtin/core/defaults/config/v1.33.yaml)

## Install cluster
```shell
kk create cluster -i inventory.yaml -c config.yaml
```
If `-i inventory.yaml` is not provided, the default inventory.yaml is used. Kubernetes will only be installed on the executing machine.
If `-c config.yaml` is not provided, the default config.yaml is used. Installs Kubernetes version v1.33.1.

# Documentation
**[Custom Playbook](docs/en/custom/README.md)**    
**[Kubernetes Playbook](docs/en/core/README.md)**
