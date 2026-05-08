# Online Installation of Kubernetes and KubeSphere

This section describes how to install Kubernetes and KubeSphere in an Internet-accessible environment.

The installation process uses KubeKey. For more information about KubeKey, visit the [GitHub KubeKey repository](https://github.com/kubesphere/kubekey).

## Prerequisites

The installation process depends on the `tar` utility for compression and decompression. Please ensure it is pre-installed in your system environment.

## Install Kubernetes and KubeSphere

Two installation methods are supported: **command line** and **Web UI**.

### Command Line Installation

#### 1. Download KubeKey

If your access to GitHub/GoogleAPIs is restricted, set the following environment variable:

```shell
export KKZONE=cn
```

Execute the following command to download the latest version of KubeKey:

```shell
curl -sfL https://get-kk.kubesphere.io | SKIP_WEB_INSTALLER=true SKIP_PACKAGE=true sh -
```

After execution, the following files will be generated in the current directory:

| Original File | Extracted File |
|---------------|----------------|
| kubekey-v4.x.x-linux-amd64.tar.gz | kk: KubeKey binary |

#### 2. Create Node Configuration File

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
| `image_registry` | Nodes used to create a private image registry. Not required for online installation |

#### 3. Create Installation Configuration File

Execute the following command to create the installation configuration file `config.yaml`:

```shell
./kk create config --with-kubernetes <Kubernetes version> -o .
```

Replace `<Kubernetes version>` with the actual version you need, e.g. `v1.27.4`. KubeKey supports Kubernetes `v1.23~v1.34` by default.

After execution, the installation configuration file `config-<Kubernetes version>.yaml` will be generated.

> **Note**: Do not delete the installation configuration file after installation, as it is still needed for subsequent operations such as adding nodes. If the file is lost, you will need to recreate it.

#### 4. Configure Cluster Parameters

Configure Kubernetes cluster information in `config-<Kubernetes version>.yaml`:

| Parameter | Description |
|-----------|-------------|
| `zone` | Download zone for files and images. If your access to GitHub/GoogleAPIs is restricted, set this to `cn` |
| `kubernetes` | Kubernetes-related configuration |
| `etcd` | etcd-related configuration |
| `image_registry` | Private image registry-related configuration |
| `cri` | Container runtime-related configuration |
| `cni` | Network plugin-related configuration |
| `storage_class` | Storage plugin-related configuration |
| `dns` | DNS-related configuration |
| `image_manifests` | Additional images to be downloaded |

> **Note**: For complete configuration reference, please refer to [Configuration Reference](../reference/config-reference.md).

#### 5. Install Kubernetes

Execute the following command to install Kubernetes:

```shell
./kk create cluster -i inventory.yaml -c config.yaml
```

#### 6. Install KubeSphere

Execute the following command to install KubeSphere:

```shell
chart=oci://hub.kubesphere.com.cn/kse/ks-core
version=1.2.4
helm upgrade --install -n kubesphere-system --create-namespace ks-core $chart \
  --debug --wait --version $version --reset-values --take-ownership \
  --set global.imageRegistry=hub.kubesphere.com.cn,extension.imageRegistry=hub.kubesphere.com.cn
```

> **Note**: Helm version must be >= 3.17.0

If the following information is displayed, KubeSphere has been installed successfully:

```
NOTES:
Thank you for choosing KubeSphere Helm Chart.

Please be patient and wait for several seconds for the KubeSphere deployment to complete.

1. Wait for Deployment Completion

    Confirm that all KubeSphere components are running by executing the following command:

    kubectl get pods -n kubesphere-system

2. Access the KubeSphere Console

    Once the deployment is complete, you can access the KubeSphere console using the following URL:

    http://192.168.6.10:30880

3. Login to KubeSphere Console

    Use the following credentials to log in:

    Account: admin
    Password: P@88w0rd

NOTE: It is highly recommended to change the default password immediately after the first login.
```

### Web UI Installation

#### 1. Download KubeKey (with Web Installer)

If your access to GitHub/GoogleAPIs is restricted, set the following environment variable:

```shell
export KKZONE=cn
```

Execute the following command to download the latest version of KubeKey (including Web Installer):

```shell
curl -sfL https://get-kk.kubesphere.io | SKIP_PACKAGE=true sh -
```

After execution, the following files will be generated in the current directory:

| Original File | Extracted File |
|---------------|----------------|
| kubekey-v4.x.x-linux-amd64.tar.gz | kk: KubeKey binary |
| web-installer.tgz | dist: Web UI resources<br>host-check.yaml, kubernetes, kubesphere: Task template files<br>schema: Configuration files<br>README.md: Installation documentation |

#### 2. Start Web Installer

Execute the following command to start the Web Installer:

```shell
./kk web --port 8080 --schema-path web-installer/schema --ui-path web-installer/dist
```

If the following message is displayed, the Web Installer has started successfully:

```
Web server started successfully on port 8080
```

Do not close the terminal. Open KubeKey's UI page in the browser via `http://<bootstrap-node-ip>:8080`.

> **Tip**: The Web Installer includes Kubernetes and KubeSphere installation modules by default. If you don't need to install KubeSphere, delete `web-installer/schema/kubesphere.json`. For more Web UI operation steps, please refer to the [KubeSphere documentation](https://docs.kubesphere.com.cn/v4.2.1/03-installation-and-upgrade/02-install-kubesphere/02-offline-install-kubernetes-and-kubesphere/#_步骤_5部署_kubernetes_和_kubesphere).
