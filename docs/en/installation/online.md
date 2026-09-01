# Online Installation of Kubernetes

This section describes how to install Kubernetes in an environment with Internet access.

The installation process uses the open-source tool KubeKey v4.x. For more information about KubeKey, visit the [GitHub KubeKey repository](https://github.com/kubesphere/kubekey).

## Core Concepts

Before proceeding, it is recommended to understand the following basic concepts:

- **Control plane node (Master)**: Responsible for cluster scheduling and management, and typically does not run business workloads.
- **Worker node**: A workload node that runs actual business containers.
- **etcd**: A distributed key-value store that holds all critical state data of the cluster.
- **Container runtime**: The underlying software responsible for creating and running containers. KubeKey supports automatically installing two container runtimes: Docker and containerd.
- **CNI (Container Network Interface)**: Provides network connectivity for Pods in the cluster. Common plugins include Calico, Cilium, Flannel, and so on.

## Prerequisites

- At least 1 Linux server is required as a cluster node. In production environments, to ensure high availability, it is recommended to prepare at least 5 Linux servers, 3 of which act as control plane nodes and another 2 as worker nodes. If you install Kubernetes on multiple Linux servers, make sure all servers belong to the same subnet.
- The operating system and version of the cluster nodes must be Ubuntu 18.04, Ubuntu 20.04, Ubuntu 22.04, Ubuntu 24.04, Debian 10, Debian 11, CentOS 8, AlmaLinux 9.0, or Kylin v10. The operating systems of multiple servers can be different. For support of other operating systems and versions, consult the official solution experts or delivery service experts of QingCloud.
- In production environments, to ensure the cluster has sufficient compute and storage resources, it is recommended that each cluster node be configured with at least 8 CPU cores, 16 GB of memory, and 200 GB of disk space. In addition, it is recommended to mount at least another 200 GB of disk space under `/var/lib/docker` (for Docker) or `/var/lib/containerd` (for containerd) on each cluster node to store container runtime data.
- If the cluster nodes do not have a container runtime installed, the installation tool KubeKey will automatically install a container runtime on each cluster node during the installation process. KubeKey installs containerd by default; you can also specify Docker as the container runtime in the configuration file.
- Make sure the DNS server addresses configured in the `/etc/resolv.conf` file are available on all cluster nodes. Otherwise, the cluster may experience domain name resolution issues.
- Make sure the `sudo`, `tar`, `curl`, and `openssl` commands are available on all cluster nodes.
- Make sure the clocks of all cluster nodes are synchronized.

> **Note**: KubeKey depends on the `tar` utility for compression and decompression of software packages. Make sure it is installed.

## Configure Firewall Rules

Kubernetes requires specific ports and protocols for communication between services. If firewall is enabled in your infrastructure environment, you need to allow the required ports and protocols in the firewall settings. If firewall is not enabled in your infrastructure environment, you can skip this step.

The following table lists the ports and protocols that need to be allowed in the firewall.

| Service | Protocol | Action | Start Port | End Port | Remarks |
|---|---|---|---|---|---|
| ssh | TCP | allow | 22 | N/A | — |
| etcd | TCP | allow | 2379 | 2380 | — |
| apiserver | TCP | allow | 6443 | N/A | — |
| calico | TCP | allow | 9099 | 9100 | — |
| bgp | TCP | allow | 179 | N/A | — |
| nodeport | TCP | allow | 30000 | 32767 | — |
| master | TCP | allow | 10250 | 10258 | — |
| worker | TCP | allow | 10250 | N/A | — |
| dns | TCP | allow | 53 | N/A | — |
| dns | UDP | allow | 53 | N/A | — |
| rpcbind | TCP | allow | 111 | N/A | Required when using NFS as persistent storage |
| ipip | IPENCAP / IPIP | allow | N/A | N/A | Required when using Calico |

If you need to temporarily disable the firewall in a test environment, you can run `systemctl stop firewalld`; in production environments, please allow the required ports precisely according to the table above.

## Install Kubernetes

Only the command line installation method is currently supported.

### 1. Download KubeKey

If your access to GitHub / Google APIs is restricted, set the following environment variable:

```shell
export KKZONE=cn
```

Execute the following command to download the latest version of KubeKey:

```shell
curl -sfL https://get-kk.kubesphere.io | SKIP_WEB_INSTALLER=true SKIP_PACKAGE=true sh -
```

> **Note**: `SKIP_WEB_INSTALLER=true` skips the download of Web Installer resources, and `SKIP_PACKAGE=true` skips the download of the offline packaging script. Command line installation only requires the `kk` binary.

After execution, the `kk` binary will be generated in the current directory.

### 2. Create Node Configuration File

Execute the following command to create the node configuration file `inventory.yaml`:

```shell
./kk create inventory -o .
```

`inventory.yaml` mainly sets the connection information of each node in the cluster. After execution, the node configuration file will be generated, for example:

```yaml
apiVersion: kubekey.kubesphere.io/v1
kind: Inventory
metadata:
  name: default
spec:
  hosts: # you can set all nodes here. or set nodes on special groups.
    # localhost:
    #   connector:
    #     password: 123456
    # node1:
    #   connector:
    #     type: ssh
    #     host: node1
    #     port: 22
    #     user: root
    #     password: 123456
    #   internal_ipv4: 1.1.1.1
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

`spec.hosts` is used to configure the connection parameters of each node. Each node uses the node name as the key, for example `localhost` or `node1`.

| Parameter | Description |
|---|---|
| `<key>` | Node name |
| `<key>.connector.type` | Node connection type. Supports `local` (local connection) and `ssh` (remote connection). It is identified automatically based on the node name or IP |
| `<key>.connector.host` | Address when using SSH to connect to the node |
| `<key>.connector.port` | Port when using SSH to connect to the node. Default: `22` |
| `<key>.connector.user` | Username when using SSH to connect to the node. Default: `root` |
| `<key>.connector.password` | Password for connecting to the node. For `local` connections this is the sudo password; for `ssh` connections this is the SSH password |
| `<key>.connector.private_key` | Path to the SSH private key file. Either password or key must be provided |
| `<key>.connector.private_key_content` | Content of the SSH private key. The key content can be used instead of the key file path |
| `<key>.internal_ipv4` | IPv4 address used for cluster-internal communication |
| `<key>.internal_ipv6` | IPv6 address used for cluster-internal communication |

`spec.groups` is used to configure the role information of nodes.

| Parameter | Description |
|---|---|
| `k8s_cluster` | Kubernetes cluster organization. Contains `kube_control_plane` and `kube_worker`, no additional configuration needed |
| `kube_control_plane` | Control plane nodes in the Kubernetes cluster. Configure node names defined in `spec.hosts` under `kube_control_plane.hosts` |
| `kube_worker` | Worker nodes in the Kubernetes cluster. Configure node names defined in `spec.hosts` under `kube_worker.hosts` |
| `etcd` | etcd nodes in the Kubernetes cluster. Configure node names defined in `spec.hosts` under `etcd.hosts` |
| `image_registry` | Nodes used to create a private image registry. Not required for online installation |

### 3. Create Installation Configuration File

Execute the following command to create the installation configuration file `config.yaml`:

```shell
./kk create config --with-kubernetes <Kubernetes version> -o .
```

Replace `<Kubernetes version>` with the actual version you need, for example `v1.36.4`. Kubernetes `v1.23` ~ `v1.36` is supported by default.

After execution, the installation configuration file `config-<Kubernetes version>.yaml` will be generated.

### 4. Configure Cluster Parameters

Configure the Kubernetes cluster information in `config-<Kubernetes version>.yaml`. The commonly used parameters are as follows:

| Parameter | Description |
|---|---|
| `zone` | Download zone for files and images. If your access to GitHub / Google APIs is restricted, set this value to `cn` |
| `kubernetes` | Kubernetes cluster configuration, including version, control plane endpoint, network CIDR, and so on |
| `etcd` | etcd cluster configuration, including installation method (built-in or external), data directory, and so on |
| `image_registry` | Private image registry configuration. Default values are usually used for online installation |
| `cri` | Container runtime configuration. Can be set to `docker` or `containerd` |
| `cni` | Network plugin configuration. Can be set to `calico`, `cilium`, `flannel`, and so on |
| `storage_class` | Storage plugin configuration. Supports `local`, `nfs`, and so on |
| `dns` | Domain name resolution configuration |
| `image_manifests` | List of additional images to be downloaded |

> **Note**: For the complete field descriptions of each parameter, refer to the [Configuration Reference](https://github.com/kubesphere/kubekey/blob/main/docs/zh/reference/config.md).

### 5. Install Kubernetes

Execute the following command to install Kubernetes:

```shell
./kk create cluster -i inventory.yaml -c config-<Kubernetes version>.yaml
```

If you have renamed the installation configuration file to `config.yaml`, you can use the following command:

```shell
./kk create cluster -i inventory.yaml -c config.yaml
```

After the installation is complete, you can check the cluster node status with `kubectl get nodes`:

```shell
kubectl get nodes
```

## FAQ

**Q: Download times out or is slow when accessing GitHub?**
A: First run `export KKZONE=cn` to use domestic sources before downloading. After switching to domestic sources, only a limited number of Kubernetes versions are supported. Set the Kubernetes version to one of the versions listed in the **Kubernetes tab** on the [get-images.kubesphere.io](https://get-images.kubesphere.io) page (specified via `./kk create config --with-kubernetes <version> -o .`).

**Q: How do I re-run the installation?**
A: First clean up the nodes with `kk delete cluster --all --with-data`, then re-run `kk create cluster`.

**Q: How do I keep the CA certificate for adding nodes later?**
A: If you use the KubeKey default certificate, keep the `<working directory>/kubekey/pki/root.crt` file after the installation (the default working directory is `/root/kubekey`). This certificate may be needed when adding nodes later.

**Q: What extra recommendations are there for production environments?**
A: It is recommended to prepare at least 5 nodes, configure high availability in advance, and configure external persistent storage to avoid using node-local disk space as persistent storage.
