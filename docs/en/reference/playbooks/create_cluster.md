# Create Cluster (create_cluster.yaml)

![architecture](../../images/architecture.png)

## pre_hook

The `pre_hook` allows users to execute scripts on corresponding nodes before creating the cluster.

Execution flow:
1. Copy local scripts to remote nodes at `/etc/kubekey/scripts/pre_install_{{ .inventory_hostname }}.sh`
2. Set script file permissions to 0755
3. Iterate over all `pre_install_*.sh` files in `/etc/kubekey/scripts/` on each remote node and execute them

> **work_dir**: working directory, defaults to the current command execution directory.  
> **inventory_hostname**: the host name defined in the `inventory.yaml` file.

## precheck

The `precheck` phase verifies that cluster nodes meet the installation requirements.

**os_precheck**: OS checks, including:
- **Hostname check**: Verify that the hostname format is valid (contains only lowercase letters, digits, '.', or '-', and must start and end with a letter or digit)
- **OS version check**: Verify that the current OS is in the supported OS distribution list, unless unsupported distributions are allowed
- **Architecture check**: Verify that the system architecture is supported (amd64 or arm64)
- **Memory check**:
  - Master nodes: verify memory meets the minimum master node requirement
  - Worker nodes: verify memory meets the minimum worker node requirement
- **Kernel version check**: Verify that the kernel version meets the minimum requirement

**kubernetes_precheck**: Kubernetes-related checks, including:
- **IP address check**: Verify that the node defines either `internal_ipv4` or `internal_ipv6` (both cannot be empty)
- **KubeVIP check**: When using `kube_vip` type for the control plane endpoint, verify that the kube_vip address is valid and not already in use
- **Kubernetes version check**: Verify that the Kubernetes version meets the minimum version requirement
- **Existing Kubernetes check**: Verify whether Kubernetes is already installed on the node; if so, check whether the version matches the configured `kube_version`

**network_precheck**: Network connectivity checks, including:
- **Network interface check**: Verify that the configured IPv4 or IPv6 network interface exists on the node
- **CIDR configuration check**: Verify that Pod CIDR and Service CIDR formats are correct (supports dual-stack format: `ipv4_cidr/ipv6_cidr` or `ipv4_cidr,ipv6_cidr`)
- **Dual-stack support check**: When dual-stack networking is configured, verify that the Kubernetes version supports it (v1.20.0+)
- **Network plugin check**: Verify that the configured network plugin is in the supported list
- **Network address space check**: Ensure that available network address space on the node is sufficient to accommodate the configured maximum Pod count
- **Hybridnet version check**: When using the Hybridnet network plugin, verify that the Kubernetes version meets the requirement (v1.16.0+)

**etcd_precheck**: etcd cluster checks, including:
- **Deployment type validation**: Validate the etcd deployment type (`internal` or `external`); in `external` mode, ensure that the etcd group is not empty and that the node count is odd
- **Disk IO performance check**: Use the `fio` tool to test write latency on the etcd data disk, ensuring that disk sync latency (e.g., WAL fsync) meets cluster requirements
- **Existing etcd check**: Detect whether etcd is already installed on the current host

**cri_precheck**: Container runtime checks, including:
- **Container manager check**: Verify that the configured container manager is in the supported list (`docker` or `containerd`)
- **containerd version check**: When using `containerd` as the container manager, verify that the containerd version meets the minimum version requirement

**nfs_precheck**: NFS storage checks, including:
- **NFS server count check**: Verify that there is only one NFS server node in the cluster, ensuring uniqueness of the NFS service deployment

**image_registry_precheck**: Image registry checks, including:
- **Required software check**: Verify that both `docker_version` and `dockercompose_version` are configured and not empty. The image registry is installed via docker-compose; missing required software will cause installation failure.

## init

The `init` phase is responsible for preparing and building all resources required for cluster installation, including:
- **Software package download**: Download binary files for Kubernetes, container runtime, network plugins, and other core components to ensure all required packages are ready
- **Helm Chart preparation**: Acquire and validate required Helm Chart packages for subsequent application deployment
- **Container image pull**: Download Docker images required for cluster components, including core component images and dependency images
- **Offline package build**: When offline installation is configured, package all dependency resources (binary files, images, Chart packages, etc.) into a complete offline installation package
- **Certificate management**: Generate various certificates required for cluster installation and inter-component communication, including CA certificates and service certificates

## install

The `install` phase is KubeKey's core installation phase, responsible for actually deploying and configuring the Kubernetes cluster on cluster nodes, including:

**install nfs**: Install the NFS service for nodes in the `nfs` group.

**install image_registry**: Install an image registry for nodes in the `image_registry` group. Currently supports two types of image registries: `harbor` and `registry`.

**install etcd**: Install etcd for nodes in the `etcd` group.

**install cri**: Install the container runtime (CRI) for nodes in the `k8s_cluster` group. Currently supports two CRIs: `docker` and `containerd`.

**kubernetes_install**: Install Kubernetes for nodes in the `k8s_cluster` group.

**install helm**: Install additional Helm applications for the installed Kubernetes cluster, including: CNI (`calico`, `cilium`, `flannel`, `hybridnet`, `kubeovn`, `multus`)

## post_hook

The `post_hook` phase executes after cluster installation completes, responsible for final cluster configuration and validation:

Execution flow:
1. Copy local scripts to remote nodes at `/etc/kubekey/scripts/post_install_{{ .inventory_hostname }}.sh`
2. Set script file permissions to 0755
3. Iterate over all `post_install_*.sh` files in `/etc/kubekey/scripts/` on each remote node and execute them

> **work_dir**: working directory, defaults to the current command execution directory.  
> **inventory_hostname**: the host name defined in the `inventory.yaml` file.
