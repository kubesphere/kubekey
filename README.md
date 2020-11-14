# KubeKey

[![CI](https://github.com/kubesphere/kubekey/workflows/CI/badge.svg?branch=master&event=push)](https://github.com/kubesphere/kubekey/actions?query=event%3Apush+branch%3Amaster+workflow%3ACI+)

> English | [中文](README_zh-CN.md)

Since v3.0.0, [KubeSphere](https://kubesphere.io) changes the ansible-based installer to the new installer called KubeKey that is developed in Go language. With KubeKey, you can install Kubernetes and KubeSphere separately or as a whole easily, efficiently and flexibly.

There are three scenarios to use KubeKey.

* Install Kubernetes only
* Install Kubernetes and KubeSphere together in one command
* Install Kubernetes first, then deploy KubeSphere on it using [ks-installer](https://github.com/kubesphere/ks-installer)

> **Important:** If you have existing clusters, please refer to [ks-installer (Install KubeSphere on existing Kubernetes cluster)](https://github.com/kubesphere/ks-installer).

## Motivation

* Ansible-based installer has a bunch of software dependency such as Python. KubeKey is developed in Go language to get rid of the problem in a variety of environment so that increasing the success rate of installation.
* KubeKey uses Kubeadm to install K8s cluster on nodes in parallel as much as possible in order to reduce installation complexity and improve efficiency. It will greatly save installation time compared to the older installer.
* KubeKey supports for scaling cluster from allinone to multi-node cluster, even an HA cluster.
* KubeKey aims to install cluster as an object, i.e., CaaO.

## Supported Environment

### Linux Distributions

* **Ubuntu**  *16.04, 18.04*
* **Debian**  *Buster, Stretch*
* **CentOS/RHEL**  *7*
* **SUSE Linux Enterprise Server** *15*


### <span id = "KubernetesVersions">Kubernetes Versions</span> 

* **v1.15**: &ensp; *v1.15.12*
* **v1.16**: &ensp; *v1.16.13*
* **v1.17**: &ensp; *v1.17.9* (default)
* **v1.18**: &ensp; *v1.18.6*
> Looking for more supported versions [Click here](./docs/kubernetes-versions.md)

> Notice: KubeSphere is currently not supported running on k8s v1.19.x
## Requirements and Recommendations

* Minimum resource requirements (For Minimal Installation of KubeSphere only)：
  * 2 vCPUs
  * 4 GB RAM
  * 20 GB Storage

> /var/lib/docker is mainly used to store the container data, and will gradually increase in size during use and operation. In the case of a production environment, it is recommended that /var/lib/docker mounts a drive separately.

* OS requirements:
  * `SSH` can access to all nodes.
  * Time synchronization for all nodes.
  * `sudo`/`curl`/`openssl` should be used in all nodes.
  * `docker` can be installed by yourself or by KubeKey.
  * `Red Hat` includes `SELinux` in its `Linux release`. It is recommended to close SELinux or [switch the mode of SELinux](./docs/turn-off-SELinux.md) to `Permissive`
> * It's recommended that Your OS is clean (without any other software installed), otherwise there may be conflicts.  
> * A container image mirror (accelerator) is recommended to be prepared if you have trouble downloading images from dockerhub.io. [Configure registry-mirrors for the Docker daemon](https://docs.docker.com/registry/recipes/mirror/#configure-the-docker-daemon).
> * KubeKey will install [OpenEBS](https://openebs.io/) to provision LocalPV for development and testing environment by default, this is convenient for new users. For production, please use NFS / Ceph / GlusterFS  or commercial products as persistent storage, and install the [relevant client](docs/storage-client.md) in all nodes.
> * If you encounter `Permission denied` when copying, it is recommended to check [SELinux and turn off it](./docs/turn-off-SELinux.md) first 

* Dependency requirements:

KubeKey can install Kubernetes and KubeSphere together. The dependency that needs to be installed may be different based on the Kubernetes version to be installed. You can refer to the list below to see if you need to install relevant dependencies on your node in advance.

|             | Kubernetes Version ≥ 1.18 | Kubernetes Version < 1.18 |
| ----------- | ------------------------- | ------------------------- |
| `socat`     | Required                  | Optional but recommended  |
| `conntrack` | Required                  | Optional but recommended  |
| `ebtables`  | Optional but recommended  | Optional but recommended  |
| `ipset`     | Optional but recommended  | Optional but recommended  |

* Networking and DNS requirements:
  * Make sure the DNS address in `/etc/resolv.conf` is available. Otherwise, it may cause some issues of DNS in cluster.
  * If your network configuration uses Firewall or Security Group，you must ensure infrastructure components can communicate with each other through specific ports. It's recommended that you turn off the firewall or follow the link configuriation: [NetworkAccess](docs/network-access.md).

## Usage

### Get the Installer Executable File

* Binary downloads of the KubeKey can be found on the [Releases page](https://github.com/kubesphere/kubekey/releases).
  Unpack the binary and you are good to go!

* Build Binary from Source Code

    ```shell script
    git clone https://github.com/kubesphere/kubekey.git
    cd kubekey
    ./build.sh
    ```

> Note:
>
> * Docker needs to be installed before building.
> * If you have problem to access `https://proxy.golang.org/`, excute `build.sh -p` instead.

### YUM

Add YUM source repo by the following command:

```
cat > bintray-kubesphere-rpm.repo <<EOF
#bintraybintray-kubesphere-rpm - packages by kubesphere from Bintray
[bintraybintray-kubesphere-rpm]
name=bintray-kubesphere-rpm
baseurl=https://dl.bintray.com/kubesphere/rpm
gpgcheck=0
repo_gpgcheck=0
enabled=1
EOF
sudo mv bintray-kubesphere-rpm.repo /etc/yum.repos.d/ && sudo yum update
```

then you can install it by: `yum install kk`

### Debian

Add deb source repo by the following command:

```
echo "deb [trusted=yes] https://dl.bintray.com/kubesphere/deb wheezy main" | sudo tee -a /etc/apt/sources.list
sudo apt update
```

then you can install it by: `sudo apt-get install kk`

### Create a Cluster

#### Quick Start

Quick Start is for `all-in-one` installation which is a good start to get familiar with KubeSphere.

> Note: Since Kubernetes temporarily does not support uppercase NodeName, contains uppercase letters in the hostname will lead to subsequent installation error

##### Command

> If you have problem to access `https://storage.googleapis.com`, execute first `export KKZONE=cn`.

```shell script
./kk create cluster [--with-kubernetes version] [--with-kubesphere version]
```

##### Examples

* Create a pure Kubernetes cluster with default version.

    ```shell script
    ./kk create cluster
    ```

* Create a Kubernetes cluster with a specified version ([supported versions](#KubernetesVersions)).

    ```shell script
    ./kk create cluster --with-kubernetes v1.17.9
    ```

* Create a Kubernetes cluster with KubeSphere installed (e.g. `--with-kubesphere v3.0.0`)

    ```shell script
    ./kk create cluster --with-kubesphere [version]
    ```

#### Advanced

You have more control to customize parameters or create a multi-node cluster using the advanced installation. Specifically, create a cluster by specifying a configuration file.

> If you have problem to access `https://storage.googleapis.com`, execute first `export KKZONE=cn`.

1. First, create an example configuration file

    ```shell script
    ./kk create config [--with-kubernetes version] [--with-kubesphere version] [(-f | --file) path]
    ```

   **examples:**

   * create an example config file with default configurations. You also can specify the file that could be a different filename, or in different folder.

    ```shell script
    ./kk create config [-f ~/myfolder/abc.yaml]
    ```

   * with KubeSphere

    ```shell script
    ./kk create config --with-kubesphere
    ```

2. Modify the file config-sample.yaml according to your environment
> Note:  Since Kubernetes temporarily does not support uppercase NodeName, contains uppercase letters in workerNode`s name will lead to subsequent installation error
> 
> A persistent storage is required in the cluster, when kubesphere will be installed. The local volume is used default. If you want to use other persistent storage, please refer to [addons](./docs/addons.md).
3. Create a cluster using the configuration file

    ```shell script
    ./kk create cluster -f config-sample.yaml
    ```

### Enable Multi-cluster Management

By default, KubeKey will only install a **solo** cluster without Kubernetes federation. If you want to set up a multi-cluster control plane to centrally manage multiple clusters using KubeSphere, you need to set the `ClusterRole` in [config-example.yaml](docs/config-example.md). For multi-cluster user guide, please refer to [How to Enable the Multi-cluster Feature](https://github.com/kubesphere/community/tree/master/sig-multicluster/how-to-setup-multicluster-on-kubesphere).

### Enable Pluggable Components

KubeSphere has decoupled some core feature components since v2.1.0. These components are designed to be pluggable which means you can enable them either before or after installation. By default, KubeSphere will be started with a minimal installation if you do not enable them.

You can enable any of them according to your demands. It is highly recommended that you install these pluggable components to discover the full-stack features and capabilities provided by KubeSphere. Please ensure your machines have sufficient CPU and memory before enabling them. See [Enable Pluggable Components](https://github.com/kubesphere/ks-installer#enable-pluggable-components) for the details.


### Add Nodes

Add new node's information to the cluster config file, then apply the changes.

```shell script
./kk add nodes -f config-sample.yaml
```

### Delete Nodes

You can delete the node by the following command，the nodeName that needs to be removed.

```shell script
./kk delete node <nodeName> -f config-sample.yaml
```

### Delete Cluster

You can delete the cluster by the following command:

* If you started with the quick start (all-in-one):

```shell script
./kk delete cluster
```

* If you started with the advanced (created with a configuration file):

```shell script
./kk delete cluster [-f config-sample.yaml]
```
### Upgrade Cluster
#### Allinone
Upgrading cluster with a specified version.
```shell script
./kk upgrade [--with-kubernetes version] [--with-kubesphere version] 
```
* Support upgrading Kubernetes only.
* Support upgrading KubeSphere only.
* Support upgrading Kubernetes and KubeSphere.

#### Multi-nodes
Upgrading cluster with a specified configuration file.
```shell script
./kk upgrade [--with-kubernetes version] [--with-kubesphere version] [(-f | --file) path]
```
* If `--with-kubernetes` or `--with-kubesphere` is specified, the configuration file will be also updated.
* Use `-f` to specify the configuration file which was generated for cluster creation.

> Note: Upgrading multi-nodes cluster need a specified configuration file. If the cluster was installed without kubekey or the configuration file for installation was not found, the configuration file needs to be created by yourself or following command.

Getting cluster info and generating kubekey's configuration file (optional).
```shell script
./kk create config [--from-cluster] [(-f | --file) path] [--kubeconfig path]
```
* `--from-cluster` means fetching cluster's information from an existing cluster. 
* `-f` refers to the path where the configuration file is generated.
* `--kubeconfig` refers to the path where the kubeconfig. 
* After generating the configuration file, some parameters need to be filled in, such as the ssh information of the nodes.

## Documents

* [Configuration example](docs/config-example.md)
* [Addons](docs/addons.md)
* [Network access](docs/network-access.md)
* [Storage clients](docs/storage-client.md)
* [Kubectl autocompletion](docs/kubectl-autocompletion.md)
* [Roadmap](docs/roadmap.md)
