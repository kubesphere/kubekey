# Offline Installation of Kubernetes and KubeSphere

This section describes how to deploy Kubernetes and KubeSphere using offline packages in an environment that cannot access the Internet.

The installation process uses the open-source tool KubeKey v4.x. For more information about KubeKey, visit the [GitHub KubeKey repository](https://github.com/kubesphere/kubekey).

> **Note**: KubeSphere Community Edition users need to build the offline package themselves while connected to the Internet; users of other KubeSphere editions can contact KubeSphere delivery service experts to obtain the latest offline package, and can skip the "Build Offline Package" section.

> **Note**: The installation process depends on the `tar` utility for compression and decompression of software packages. Make sure it is pre-installed in your system environment. If `charts` is configured in `config.yaml`, make sure Helm is pre-installed on the packaging node.

## Overview

Compared with online installation, offline installation requires you to first package the required components and images as an offline package on a machine that can access the Internet, and then transfer it to the target environment for installation. The overall flow is as follows:

1. **Build the offline package** (on a networked machine): Download components and images and package them as `artifact.tgz`.
2. **Transfer the offline package**: Copy `artifact.tgz` to the target environment (for example, via a storage medium or an intranet transfer).
3. **Install the cluster** (in the target environment): Extract the offline package, push images to the private image registry (if installing the private image registry separately, i.e., Option 1, image pushing is completed automatically by the installation flow when KubeKey installs the image registry together with the cluster), and then install Kubernetes and KubeSphere.

## Role Descriptions

Offline installation involves the following four roles:

| Role | Responsibility | Minimum Configuration (per node) | Network Requirements |
|---|---|---|---|
| Packaging node | Downloads required software packages and images from the Internet and builds the offline package | CPU: 1 core, Memory: 1 GB, Disk: 150 GB | Must have Internet access |
| Deployment node (runs the Web Installer service) | Executes `kk` commands on this node during installation to run the installation service | CPU: 1 core, Memory: 1 GB, Disk: 150 GB | Network connected to Kubernetes nodes |
| Private image registry node | Stores the container images required by the cluster | CPU: 8 cores, Memory: 16 GB, Disk: 100 GB | Network connected to Kubernetes nodes |
| Kubernetes node | Runs cluster workloads (no need to pre-install Kubernetes) | CPU: 2 cores, Memory: 4 GB, Disk: 40 GB | Inter-node network connected |

> **Note**:
> - A single host can simultaneously assume multiple roles, for example both a deployment node and a private image registry node, or both a deployment node and a Kubernetes node.
> - The private image registry node and the Kubernetes node cannot be the same host, because the Kubernetes installation process restarts the container runtime on that node, which may interrupt the image registry service.

## Prerequisites

> **Note**: The following are the prerequisites that Kubernetes nodes must satisfy.

- You need to prepare at least 1 Linux server as a cluster node. In production environments, to ensure high availability, it is recommended to prepare at least 5 Linux servers, 3 of which act as control plane nodes and another 2 as worker nodes. If you install KubeSphere on multiple Linux servers, make sure all servers belong to the same subnet.
- The operating system and version of the cluster nodes must be Ubuntu 18.04, Ubuntu 20.04, Ubuntu 22.04, Ubuntu 24.04, Debian 10, Debian 11, CentOS 8, AlmaLinux 9.0, or Kylin v10. The operating systems of multiple servers can be different. For support of other operating systems and versions, consult the official solution experts or delivery service experts of QingCloud.
- In production environments, to ensure the cluster has sufficient compute and storage resources, it is recommended that each cluster node be configured with at least 8 CPU cores, 16 GB of memory, and 200 GB of disk space. In addition, it is recommended to mount at least another 200 GB of disk space under `/var/lib/docker` (for Docker) or `/var/lib/containerd` (for containerd) on each cluster node to store container runtime data.
- In production environments, it is recommended to configure high availability for the KubeSphere cluster in advance to avoid service interruption when a single control plane node fails. For more information, see [Configure High Availability](https://docs.kubesphere.com.cn/v4.2.1/03-installation-and-upgrade/01-preparations/02-configure-high-availability/02-configure-k8s-high-availability/).

  > **Note**: If you plan multiple control plane nodes, be sure to configure high availability for the cluster in advance.

- By default, KubeSphere uses the local disk space of cluster nodes as persistent storage. In production environments, it is recommended to configure an external storage system as persistent storage in advance. For more information, see [Configure External Persistent Storage](https://docs.kubesphere.com.cn/v4.2.1/03-installation-and-upgrade/01-preparations/04-configure-external-persistent-storage/).
- Make sure the DNS server addresses configured in the `/etc/resolv.conf` file are available on all cluster nodes. Otherwise, the KubeSphere cluster may experience domain name resolution issues.
- Make sure the `sudo`, `tar`, `curl`, and `openssl` commands are available on all cluster nodes.
- Make sure the clocks of all cluster nodes are synchronized.

## Configure Firewall Rules

KubeSphere requires specific ports and protocols for communication between services. If firewall is enabled in your infrastructure environment, you need to allow the required ports and protocols in the firewall settings. If firewall is not enabled in your infrastructure environment, you can skip this step.

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
| local-registry | TCP | allow | 5000 | N/A | Required in offline environments |
| local-apt | TCP | allow | 5080 | N/A | Required in offline environments |
| rpcbind | TCP | allow | 111 | N/A | Required when using NFS as persistent storage |
| ipip | IPENCAP / IPIP | allow | N/A | N/A | Required when using Calico |

## Build Offline Package

### Create Configuration File

**Tip**: You can visually select the required components and automatically generate `config.yaml` on the [https://get-images.kubesphere.io](https://get-images.kubesphere.io) page, or create it manually by referring to the following example.

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
        - v1.34.3
        # Other versions from v1.23~v1.34 can also be listed
    cni:
      type:
        - calico
        - cilium
        - flannel
        - kubeovn
        - hybridnet
      # multi_cni:          # Optional, multi-CNI management component, e.g. multus
      #   - multus
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
      - "ubuntu-22.04-debs"
      - "centos-8-rpms"
      # Add or remove as needed
```

**Field descriptions:**

| Field | Description |
|---|---|
| `apiVersion` | API version of the configuration file. Fixed value: `kubekey.kubesphere.io/v1` |
| `kind` | Resource type. Fixed value: `Config` |
| `spec.zone` | Region for downloading software packages. `cn` means using domestic sources |
| `spec.download.arch` | CPU architectures to download. Supports `amd64` and `arm64` |
| `spec.download.images.policy` | Image download policy. `warn` means only recording a warning if the image is missing some CPU architectures or operating systems; `strict` means the pulled image must contain all selected CPU architectures and operating systems, otherwise an error is raised |
| `spec.download.images.registry` | Image registry address |
| `spec.download.kubernetes.kube_version` | List of Kubernetes versions to include |
| `spec.download.cni.type` | CNI plugin types to include |
| `spec.download.cni.multi_cni` | Multi-CNI management components to include |
| `spec.download.cri.container_manager` | Container runtime types. Supports `containerd` and `docker` |
| `spec.download.storage_class` | Storage classes to include. Supports `local`, `nfs` |
| `spec.download.image_registry.type` | Image registry types. Supports `harbor` and `docker-registry` |
| `spec.download.iso` | List of operating systems for building ISO dependency packages, used to install system dependencies |

### Get KubeKey and Web Installer

If your access to GitHub or Google APIs is restricted, set the following environment variable:

```bash
export KKZONE=cn
```

Execute the following command to download KubeKey and Web Installer:

```bash
curl -sfL https://get-kk.kubesphere.io | sh -
```

After execution, the following files will be generated in the current directory:

| Original File | Extracted File |
|---|---|
| `kubekey-v4.x.x-linux-amd64.tar.gz` | `kk`: KubeKey binary |
| `web-installer.tgz` | `dist`: Web page resources; `host-check.yaml`, `kubernetes`, `kubesphere`: Task template files; `schema`: Configuration form definitions; `README.md`: Installation documentation |
| `package.sh` | Build script for the offline package (automatically generated by the download command; internally calls `kk artifact export` to complete download and packaging) |

### Build the Offline Package

Execute the build script:

```bash
./package.sh config.yaml
```

When the following information is printed, the build has succeeded:

```text
Offline package artifact.tgz has been created successfully.
```

The offline package is `artifact.tgz` and contains the following:

```text
artifact/
├── kubekey-artifact.tgz    # Complete offline resource package
└── tools/                  # Tool packages for different architectures
    ├── amd64/
    │   ├── kubekey-v4.x.x-linux-amd64.tar.gz
    │   ├── nerdctl-2.2.1-linux-amd64.tar.gz
    │   └── oras_1.3.0_linux_amd64.tar.gz
    └── arm64/
        ├── kubekey-v4.x.x-linux-arm64.tar.gz
        ├── nerdctl-2.2.1-linux-arm64.tar.gz
        └── oras_1.3.0_linux_arm64.tar.gz
```

After transferring `artifact.tgz` to the target environment, the following steps are all executed in the target environment.

## Install the Cluster Using the Offline Package

Before installing the cluster, you need to specify a private image registry address. There are two ways:

- **Option 1**: Install the private image registry separately. Refer to [Image Registry Installation](https://github.com/kubesphere/kubekey/blob/main/docs/zh/image-registry/README.md).
- **Option 2**: Install the image registry together with the cluster. See the installation steps below for details.

### Choose an Installation Method

Both of the following methods can complete the installation of Kubernetes and KubeSphere. **The installation results are identical. You only need to choose one entry point; there is no need to run both**:

- **Method 1: Command Line Installation**: Suitable for scenarios where you are familiar with command line operations and need fine-grained cluster parameter configuration.
- **Method 2: Web Installer Installation**: Suitable for scenarios where you want to complete node addition, parameter configuration, and installation validation through a graphical interface.

> **Note**: Choose either one of the two methods; do not run them repeatedly. Both methods support the two private image registry deployment ways described above (install separately, or install together with the cluster).

### Extract the Offline Package

```bash
tar -zxvf artifact.tgz
```

### Method 1: Command Line Installation

#### 1. Enter the Offline Package Directory and Extract Tools

KubeKey tools are located in the `tools/{arch}/` directory. Extract the corresponding tool based on the architecture of the installation machine.

Check the machine architecture:

```bash
uname -m
```

Enter the offline package directory:

```bash
cd artifact/
```

Extract KubeKey to the offline package directory:

```bash
tar -zxvf tools/$(uname -m)/kubekey-v4.x.x-linux-$(uname -m).tar.gz -C .
```

#### 2. Push Images to the Private Image Registry (Option 1 only: install the private image registry separately)

> **Note**: Only when using **Option 1 (install the private image registry separately)** do you need to manually execute this step to push the images in the offline package to the deployed private image registry. If you use **Option 2 (install the image registry together with the cluster)**, KubeKey automatically deploys the image registry and pushes the images when executing `kk create cluster`. **Skip this step in that case.**

Execute the following command to push the images in the offline package to the deployed private image registry:

```bash
./kk artifact images --push -c config.yaml -a kubekey-artifact.tgz
```

> **Note**: Before execution, make sure the private image registry address is correctly configured in the `config.yaml` used for pushing (that is, the `spec.image_registry.auth` field). This `config.yaml` is the same file used for packaging in the "Build Offline Package" section above.

#### 3. Create Node Configuration File

Execute the following command to create the node configuration file `inventory.yaml`:

```bash
./kk create inventory -o .
```

After execution, the node configuration file `inventory.yaml` will be generated. `inventory.yaml` is mainly used to set the connection information of each node in the cluster, for example:

```yaml
apiVersion: kubekey.kubesphere.io/v1
kind: Inventory
metadata:
  name: default
spec:
  hosts: {}
  groups:
    k8s_cluster:
      groups:
        - kube_control_plane
        - kube_worker
    kube_control_plane:
      hosts:
        - localhost
    kube_worker:
      hosts:
        - localhost
    etcd:
      hosts:
        - localhost
```

Node connection parameters in `spec.hosts`:

| Parameter | Description |
|---|---|
| `<key>` | Node name |
| `<key>.connector.type` | Node connection type. Supports `local` (local connection) and `ssh` (remote connection). KubeKey automatically identifies the connection type based on the node name or IP |
| `<key>.connector.host` | Address when using SSH to connect to the node |
| `<key>.connector.port` | Port when using SSH to connect to the node. Default: `22` |
| `<key>.connector.user` | Username when using SSH to connect to the node. Default: `root` |
| `<key>.connector.password` | Password for connecting to the node. For `local` connections this is the sudo password; for `ssh` connections this is the SSH password |
| `<key>.connector.private_key` | Path to the SSH private key file. Either password or key must be provided |
| `<key>.connector.private_key_content` | Content of the SSH private key. The key content can be used instead of the key file path |
| `<key>.internal_ipv4` | IPv4 address used for cluster-internal communication |
| `<key>.internal_ipv6` | IPv6 address used for cluster-internal communication |

Node role parameters in `spec.groups`:

| Parameter | Description |
|---|---|
| `k8s_cluster` | Kubernetes cluster node group. Contains `kube_control_plane` and `kube_worker`, no additional configuration needed |
| `kube_control_plane` | Control plane nodes in the Kubernetes cluster. Configure node names defined in `spec.hosts` under `kube_control_plane.hosts` |
| `kube_worker` | Worker nodes in the Kubernetes cluster. Configure node names defined in `spec.hosts` under `kube_worker.hosts` |
| `etcd` | etcd nodes in the Kubernetes cluster. Configure node names defined in `spec.hosts` under `etcd.hosts` |
| `image_registry` | Nodes used to create a private image registry. Usually required for offline installation |

If you choose to install the image registry together with the cluster, you need to add the `image_registry` node and group in `inventory.yaml`. Example:

```yaml
spec:
  hosts:
    harbor1:
      connector:
        type: ssh
        host: 172.16.0.1
        port: 22
        user: root
        password: 123456
      internal_ipv4: 172.16.0.1
  groups:
    image_registry:
      hosts:
        - harbor1
```

#### 4. Create Installation Configuration File

> **Note**: The configuration file generated in this step is used for **installing the cluster**, and is not the same file as the `config.yaml` used for **packaging** in the "Build Offline Package" section above.

Execute the following command to create the installation configuration file. The following example uses `v1.34.3`, which is already included in the offline resource list of the `config.yaml` example above:

```bash
./kk create config --with-kubernetes v1.34.3 -o .
```

Replace `v1.34.3` with the actual Kubernetes version you need. Make sure the replaced version is included in the `spec.download.kubernetes.kube_version` list used when building the offline package.

After execution, the installation configuration file `config-v1.34.3.yaml` will be generated.

If you choose to install the image registry together with the cluster, you need to add the image registry configuration in `config-v1.34.3.yaml`:

```yaml
apiVersion: kubekey.kubesphere.io/v1
kind: Config
spec:
  download:
    arch:
      - amd64
      - arm64
    images:
      policy: warn
  image_registry:
    auth:
      registry: dockerhub.kubekey.local  # Replace with your actual private image registry address
      username: admin
      password: Harbor12345
      skip_tls_verify: false
      plain_http: false
```

#### 5. Install the Cluster

```bash
./kk create cluster -a kubekey-artifact.tgz -i inventory.yaml -c config-v1.34.3.yaml
```

#### 6. Install KubeSphere

KubeKey v4.x decouples the installation of Kubernetes and KubeSphere. After Kubernetes is installed, you need to manually execute the following Helm commands to install KubeSphere.

> **Note**: The `ks-core` Helm Chart package is included in `kubekey-artifact.tgz`. After extracting `kubekey-artifact.tgz`, it is located in the `charts/` directory.

```bash
# Extract the offline resource package to obtain the ks-core chart
tar -zxvf kubekey-artifact.tgz
```

```bash
helm upgrade --install \
  -n kubesphere-system \
  --create-namespace \
  ks-core \
  ./charts/ks-core-1.2.5.tgz \
  --debug \
  --wait \
  --reset-values \
  --set auditing.enable=true \
  --set ha.enabled=true \
  --set redisHA.enabled=true \
  --set global.imageRegistry=dockerhub.kubekey.local \
  --set extension.imageRegistry=dockerhub.kubekey.local
```

> **Note**:
> - Replace the registry address above with your actual private image registry address.
> - `auditing.enable`, `ha.enabled`, and `redisHA.enabled` are recommended configuration items for production environments and can be adjusted based on actual needs.
> - Helm version must be >= 3.17.0.

### Method 2: Web Installer Installation

#### 1. Enter the Offline Package Directory and Extract Tools

KubeKey tools are located in the `tools/{arch}/` directory. Extract the corresponding tool based on the architecture of the installation machine.

Check the machine architecture:

```bash
uname -m
```

Enter the offline package directory:

```bash
cd artifact/
```

Extract KubeKey to the offline package directory:

```bash
tar -zxvf tools/$(uname -m)/kubekey-v4.x.x-linux-$(uname -m).tar.gz -C .
```

Extract `kubekey-artifact.tgz` to the working directory. This file contains all the resources required for installation: binaries, Helm chart packages, image files, and so on.

```bash
mkdir kubekey
tar -zxvf kubekey-artifact.tgz -C kubekey
```

#### 2. Start Web Installer

Extract the Web Installer package

```shell
tar -zxvf web-installer.tgz
```

Execute the following command to start the Web Installer page:

```shell
./kk web --port 8080 --schema-path web-installer/schema --ui-path web-installer/dist
```

If the following information is displayed, the Web Installer has started successfully:

```text
Web server started successfully on port 8080
```

Do not close the command terminal.

#### 3. Deploy Kubernetes and KubeSphere via Web Installer

In the browser, visit `http://<Deployment Node IP Address>:8080` and click **Start Installation** to enter the deployment flow.

##### 3.1 Add Cluster Nodes

On the **Basic Information** page, add cluster nodes. Each node needs to be assigned a role, and the following three roles are supported:

- **Master**: Control plane node, responsible for cluster scheduling and management. etcd is automatically installed on Master nodes.
- **Worker**: Worker node, running actual business containers.
- **Image**: Image registry node, used to automatically deploy a private image registry. This role is usually required for offline installation.

The Web Installer supports the following three ways to add nodes:

- **Manual addition**: Suitable for adding a single node. You need to fill in host name, IP address, SSH address, SSH authentication, and other information.
- **File upload**: Suitable for batch adding nodes. Fill in the node information according to the template and upload the file.
- **Node scan**: Suitable for automatically discovering nodes. You can scan nodes by IP CIDR and select the nodes to add based on the scan results.

> **Note**: If you only add one node, the node role must be `Master & Worker`.

##### 3.2 Modify Configuration Parameters

Configure the parameters required for deploying Kubernetes and KubeSphere.

Both the Kubernetes and KubeSphere Core tabs support **Form mode** and **YAML mode**. You can fill in the configuration information in the form (all parameters need to be configured), or directly edit the YAML file. For more configuration information, refer to the [Configuration Example](https://github.com/kubesphere/kubekey/blob/main/docs/zh/reference/config.md).

> **Note**: If you choose to install the image registry, you need to add the following parameters in **YAML mode** to skip the online download:
>
> ```yaml
> download:
>   fetch: false
> ```

**Install Image Registry**

If you need to deploy a private image registry together with the cluster, add an Image role node when adding nodes, and set the image registry type to `harbor` or `docker-registry` in the Kubernetes configuration parameters:

- **Single-node image registry**: Add 1 Image role node and set the image registry type in the Kubernetes configuration parameters.
- **Highly available image registry**: Add multiple Image role nodes (recommended ≥ 3), set the image registry type in the Kubernetes configuration parameters, and configure the highly available virtual IP of the image registry as the access entry.

After configuration, click **Next**.

##### 3.3 Installation Preview

On the **Installation Preview** page, confirm that the version and other information are correct, then click **Next: Execute Installation** to start the installation. You can also go back to the previous step to modify the configuration parameters.

##### 3.4 Installation

Wait patiently for the installation to complete. After the installation is complete, the system will automatically enter the installation validation step.

If an exception occurs during the installation, click **View Logs** to view the log details, and quit or initialize and reinstall.

> **Note**: If you need to reinstall, modify configuration parameters, or clear the configuration on the Web Installer page, click the **Initialize** button on the left. It resets all tasks on the Kubernetes nodes and returns to the Basic Information page. The initialization operation is irreversible, so proceed with caution.

##### 3.5 Installation Validation

1. On the **Installation Validation** page, click **Start Detection**, and the system will automatically run the corresponding detection scripts to verify system availability.
2. If the system detection passes, click **Complete** to view the KubeSphere access address, administrator username, and default password.
3. Enter the access address in the web browser, log in to the KubeSphere Web console, and start using KubeSphere.

> **Note**: Depending on your network environment, you may need to configure traffic forwarding rules and allow port `30880` in the firewall.

## Access the KubeSphere Web Console

After the installation is complete, visit the KubeSphere console address displayed in the installation result in your browser.

Use the administrator username and default password displayed in the installation result to log in to the KubeSphere Web console. It is recommended that you change the default password immediately after the first login.

## FAQ

**Q: The version selected when packaging differs from the version used during installation?**
A: The Kubernetes version used for installation must be included in the `spec.download.kubernetes.kube_version` list when building the offline package; otherwise, the images of that version are not included in the offline package.

**Q: Can the Web Installer install a private image registry?**
A: Yes. Add an Image role node when adding nodes, and set the image registry type (`harbor` or `docker-registry`) in the Kubernetes configuration parameters. To deploy a highly available image registry, add multiple Image role nodes and configure the highly available virtual IP of the image registry as the access entry.

**Q: Image pushing fails?**
A: Make sure the registry address, username, and password in `config.yaml` are correct, and that port `5000` is reachable over the network.

**Q: Why does pushing images pull the hub.kubesphere.com.cn images online?**
A: Before pushing images to the private image registry, KubeKey first checks the integrity of the local image files online (verifying whether the image manifests match the images in the registry). If no online check is needed, you can set `spec.download.fetch` to `false` in `config.yaml` to skip the online check and completely complete the image push offline.

**Q: How do I add a single machine?**
A: The node role must be `Master & Worker`.

**Q: How do I re-run the installation?**
A: On the Web Installer, click the **Initialize** button (this operation is irreversible); for the command line method, clean up the nodes and then re-run `kk create cluster`.

**Q: How do I keep the CA certificate for adding nodes later?**
A: If you use the KubeKey default certificate, keep the `<working directory>/kubekey/pki/root.crt` file after the installation (the default working directory is `/root/kubekey`). This certificate may be needed when adding nodes later.

**Q: How do I install in an environment with Internet access?**
A: Refer to [Online Installation of Kubernetes and KubeSphere](https://docs.kubesphere.com.cn/v4.2.1/03-installation-and-upgrade/02-install-kubesphere/01-online-install-kubernetes-and-kubesphere/).
