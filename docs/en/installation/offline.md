# Offline Installation of Kubernetes and KubeSphere

This section describes how to deploy Kubernetes and KubeSphere using offline packages in an environment without Internet access.

> **Prerequisite**: The installation process depends on the `tar` utility for compression and decompression. Please ensure it is pre-installed in your system environment. If `charts` is configured in `config.yaml`, make sure Helm is pre-installed on the packaging node.

---

## Preparation

Prepare Linux hosts according to the following minimum configuration requirements.

| Role | Host Count | Minimum Requirements (per node) | Network |
|------|-----------|--------------------------------|---------|
| Packaging node | 1 | CPU: 1 core, Memory: 1 GB, Disk: 150 GB | |
| Deployment node (runs Web Installer) | 1 | CPU: 1 core, Memory: 1 GB, Disk: 150 GB | Network connected to Kubernetes nodes |
| Private image registry node | 1 | CPU: 8 cores, Memory: 16 GB, Disk: 100 GB | Network connected to Kubernetes nodes |
| Kubernetes node | ≥ 1 | CPU: 2 cores, Memory: 4 GB, Disk: 40 GB | Inter-node network connected |

> **Notes**
>
> - A single host can simultaneously assume multiple roles, e.g. both a deployment node and a private image registry node, or both a deployment node and a Kubernetes node.
> - **The private image registry node and Kubernetes nodes cannot be the same host.**

**Role descriptions:**

- **Packaging node**: Prepare at least 1 Linux server as the packaging node. This node will download required software packages and images from the Internet, so it must be able to access: `github.com`, `docker.io`, `quay.io`.
- **Deployment node** (runs Web Installer services): During installation, `kk` commands need to be executed on this node to run the installation service. This node must maintain network connectivity with the private image registry nodes and Kubernetes nodes.
- **Private image registry node**: If no private image registry has been deployed, prepare at least 1 Linux server. This node must maintain network connectivity with all Kubernetes nodes.
- **Kubernetes nodes**: Prepare at least 1 Linux server as a cluster node (no need to pre-install Kubernetes).

---

## Build Offline Package

### Create Configuration File

> You can generate it via the https://get-images.kubesphere.io page.

Log in to the packaging node and create a `config.yaml` file on the packaging node:

```yaml
apiVersion: kubekey.kubesphere.io/v1
kind: Config
spec:
  zone: "cn"
  download:
    arch:
      - amd64
      - arm64
    images:
      policy: warn
      registry: hub.kubesphere.com.cn
    kubernetes:
      kube_version:
        - v1.23.17
        - v1.24.17
        - v1.25.16
        - v1.26.15
        - v1.27.16
        - v1.28.15
        - v1.29.15
        - v1.30.14
        - v1.31.12
        - v1.32.11
        - v1.33.7
        - v1.34.3
    cni:
      type:
        - calico
        - cilium
        - flannel
        - kubeovn
        - hybridnet
      multi_cni:
        - multus
        - spiderpool
    cri:
      container_manager:
        - containerd
        - docker
    storage_class:
      local:
        enabled: true
      nfs:
        enabled: true
    image_registry:
      type:
        - harbor
        - docker-registry
    iso:
      - "almalinux-9.0-rpms"
      - "kylin-v10SP3-rpms"
      - "ubuntu-22.04-debs"
      - "centos-8-rpms"
      - "kylin-v10SP2-rpms"
      - "ubuntu-24.04-debs"
      - "debian-10-debs"
      - "kylin-v10SP1-rpms"
      - "debian-11-debs"
      - "ubuntu-18.04-debs"
      - "kylin-v10SP3-2403-rpms"
      - "ubuntu-20.04-debs"
```

**Field descriptions:**

| Field | Description |
|-------|-------------|
| `apiVersion` | API version of the configuration file. Fixed value: `kubekey.kubesphere.io/v1` |
| `kind` | Resource type. Fixed value: `Config` |
| `spec.zone` | Download zone for packages. `cn` means using domestic mirrors |
| `spec.download.arch` | CPU architectures to download. Supports `amd64` and `arm64` |
| `spec.download.images.policy` | Image download policy. `warn` means only warning if the image does not exist |
| `spec.download.images.registry` | Image registry address |
| `spec.download.kubernetes.kube_version` | List of Kubernetes versions to include |
| `spec.download.cni.type` | CNI plugin types to include |
| `spec.download.cni.multi_cni` | Multi-CNI management components to include |
| `spec.download.cri.container_manager` | Container runtime types. Supports `containerd` and `docker` |
| `spec.download.storage_class` | Storage classes to include. Supports `local` and `nfs` |
| `spec.download.image_registry.type` | Image registry types. Supports `harbor` and `docker-registry` |
| `spec.download.iso` | List of operating systems for building ISO dependency packages |

### Get kk and Web Installer

If your access to GitHub/GoogleAPIs is restricted, set the following environment variable:

```shell
export KKZONE=cn
```

Execute the following command to download KubeKey and Web Installer:

```shell
curl -sfL https://get-kk.kubesphere.io | sh -
```

After execution, the following files will be generated in the current directory:

| Original File | Extracted File |
|--------|-----------|
| `kubekey-v4.x.x-linux-amd64.tar.gz` | kk: KubeKey binary |
| `web-installer.tgz` | dist: Web UI resources<br>host-check.yaml, kubernetes, kubesphere: Task template files<br>schema: Configuration files<br>README.md: Installation documentation |
| `package.sh` | Offline package build script |

### Build Offline Package

Execute the build script:

```shell
./package.sh config.yaml
```

When `Offline package artifact.tgz has been created successfully.` is printed, the build is successful. The offline package is `artifact.tgz`.

Offline package contents:

```text
artifact/
├── artifact/kubekey-artifact.tgz    # Complete offline resource package
└── artifact/tools/                  # Tool packages for different architectures
    ├── amd64/
    │   ├── kubekey-v4.x.x-linux-amd64.tar.gz
    │   ├── nerdctl-2.2.1-linux-amd64.tar.gz
    │   └── oras_1.3.0_linux_amd64.tar.gz
    └── arm64/
        ├── kubekey-v4.x.x-linux-arm64.tar.gz
        ├── nerdctl-2.2.1-linux-arm64.tar.gz
        └── oras_1.3.0_linux_arm64.tar.gz
```

---

## Install Cluster Using Offline Package

Before installing the cluster, you need to specify a private image registry address. There are two options:

- **Option 1**: Install a private image registry separately. Please refer to [Image Registry Installation](../image-registry/README.md).
- **Option 2**: Install the image registry together with the cluster. You need to add the corresponding configuration in `inventory.yaml` and `config.yaml` (see the command line installation steps below).

### Extract the Offline Package

```shell
tar -zxvf artifact.tgz
```

### Method 1: Web Installer

> **Tip**: The Web Installer does not currently support installing a private image registry. Please install it separately beforehand by referring to [Image Registry Installation](../image-registry/README.md).

#### 1. Enter the Offline Package Directory and Extract Tools

KubeKey tools are located in the `tools/{arch}/` directory. Extract the corresponding tool based on your machine's architecture:

```shell
# Check machine architecture
uname -m
```

Extract KubeKey to the offline package directory

```shell
cd artifact/
tar -zxvf tools/{arch}/kubekey-v4.x.x-linux-{arch}.tar.gz .
```

#### 2. Push Images to the Private Image Registry

Execute the following command to push images from the offline package to the deployed private image registry:

```shell
kk artifact images --push -c config.yaml -a kubekey-artifact.tgz
```

> **Note**: Before executing, make sure the private image registry address is correctly configured in `config.yaml`.

#### 3. Start the Web Installer

```shell
kk web --port 8080 --schema-path web-installer/schema --ui-path web-installer/dist
```

If the following information is displayed, the Web Installer has started successfully:

```
Web server started successfully on port 8080
```

Do not close the command terminal. Open KubeKey's UI page in the browser via `http://<bootstrap-node-ip>:8080`.

### Method 2: Command Line Installation

> **Tip**: The command line supports installing the [Image Registry](../image-registry/README.md) separately. It also supports synchronous installation during cluster creation (just modify `inventory.yaml` and `config.yaml`).

#### 1. Enter the Offline Package Directory

KubeKey tools are located in the `tools/{arch}/` directory. Extract the corresponding tool based on your machine's architecture:

```shell
# Check machine architecture
uname -m
```

Extract KubeKey to the offline package directory

```shell
tar -zxvf tools/{arch}/kubekey-v4.x.x-linux-{arch}.tar.gz .
```

#### 2. Push Images to the Private Image Registry

Execute the following command to push images from the offline package to the deployed private image registry:

```shell
kk artifact images --push -c config.yaml -a kubekey-artifact.tgz
```

> **Note**: Before executing, make sure the private image registry address is correctly configured in `config.yaml`.

#### 3. Create Node Configuration File

Execute the following command to create the node configuration file `inventory.yaml`:

```shell
./kk create inventory -o .
```

`inventory.yaml` mainly defines the connection information of each node in the cluster. After execution, the node configuration file will be generated. Example:

```yaml
apiVersion: kubekey.kubesphere.io/v1
kind: Inventory
metadata:
  name: default
spec:
  hosts:
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
    # etcd nodes (when etcd_deployment_type is external)
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

**Configure node connection parameters in `spec:hosts`:**

| Parameter | Description |
|-----------|-------------|
| `<key>` | Node name |
| `<key>:connector` | Node connection info |
| `<key>:connector:type` | Connection type. Supports `local` (local connection) and `ssh` (remote connection). Automatically identifies local or ssh based on the node name or IP |
| `<key>:connector:host` | Address when using ssh to connect to the node |
| `<key>:connector:port` | Port when using ssh to connect to the node. Default: `22` |
| `<key>:connector:user` | Username when using ssh to connect to the node. Default: `root` |
| `<key>:connector:password` | Password for the connection. For local connections this is the sudo password; for ssh connections this is the ssh password |
| `<key>:connector:private_key` | Path to the ssh private key file. Either password or key must be provided |
| `<key>:connector:private_key_content` | Content of the ssh private key. Can be used instead of a key file path |
| `<key>:internal_ipv4` | IPv4 address used for cluster-internal communication |
| `<key>:internal_ipv6` | IPv6 address used for cluster-internal communication |

**Configure node role information in `spec:groups`:**

| Parameter | Description |
|-----------|-------------|
| `k8s_cluster` | Kubernetes cluster organization. Contains `kube_control_plane` and `kube_worker`, no additional configuration needed |
| `kube_control_plane` | Control plane nodes in the Kubernetes cluster. Configure node names defined in `spec:hosts` under `kube_control_plane:hosts` |
| `kube_worker` | Worker nodes in the Kubernetes cluster. Configure node names defined in `spec:hosts` under `kube_worker:hosts` |
| `etcd` | etcd nodes in the Kubernetes cluster. Configure node names defined in `spec:hosts` under `etcd:hosts` |
| `image_registry` | Nodes used to create a private image registry. Usually required for offline installation |

>
> If you choose to install the image registry together with the cluster, you need to add the `image_registry` node and group in `inventory.yaml`. Example:
>
> ```yaml
> spec:
>   hosts:
>     harbor1:
>       connector:
>         type: ssh
>         host: 172.16.0.1
>         port: 22
>         user: root
>         password: 123456
>       internal_ipv4: 172.16.0.1
>   groups:
>     image_registry:
>       hosts:
>         - harbor1
> ```

#### 3. Create Installation Configuration File

Execute the following command to create the installation configuration file `config.yaml`:

```shell
./kk create config --with-kubernetes v1.32.13 -o .
```

Replace `v1.32.13` with the actual version you need. KubeKey supports Kubernetes `v1.23~v1.34` by default.

After execution, the installation configuration file `config-v1.32.13.yaml` will be generated.

>
> If you choose to install the image registry together with the cluster, you need to add the image registry configuration in `config.yaml`:
>
> ```yaml
> spec:
>   image_registry:
>     # Image registry type. Supports harbor, docker-registry. Leave empty to skip installation.
>     type: "harbor"
>     auth:
>       # Address of the private image registry
>       registry: "dockerhub.kubekey.local"
> ```

#### 4. Configure Cluster Parameters

Configure Kubernetes cluster information in `config-v1.32.13.yaml`:

| Parameter | Description |
|-----------|-------------|
| `zone` | Download zone for files and images. Not needed for network access during offline installation, but effective when building the offline package |
| `kubernetes` | Kubernetes-related configuration |
| `etcd` | etcd-related configuration |
| `image_registry` | Private image registry-related configuration |
| `cri` | Container runtime-related configuration |
| `cni` | Network plugin-related configuration |
| `storage_class` | Storage plugin-related configuration |
| `dns` | DNS-related configuration |
| `image_manifests` | Additional images to be downloaded |

> **Note**: For complete configuration reference, please refer to [Configuration Reference](../reference/config.md).

#### 5. Install the Cluster

```shell
kk create cluster -a kubekey-artifact.tgz -i inventory.yaml -c config.yaml
```
