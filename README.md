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
  * `ebtables`/`socat`/`ipset`/`conntrack` should be installed in all nodes.
  * `docker` can be installed by yourself or by KubeKey.
  * `Red Hat` includes `SELinux` in its `Linux release`. It is recommended to close SELinux or [switch the mode of SELinux](./docs/turn-off-SELinux.md) to `Permissive`
> * It's recommended that Your OS is clean (without any other software installed), otherwise there may be conflicts.  
> * A container image mirror (accelerator) is recommended to be prepared if you have trouble downloading images from dockerhub.io. [Configure registry-mirrors for the Docker daemon](https://docs.docker.com/registry/recipes/mirror/#configure-the-docker-daemon).
> * KubeKey will install [OpenEBS](https://openebs.io/) to provision LocalPV for development and testing environment by default, this is convenient for new users. For production, please use NFS / Ceph / GlusterFS  or commercial products as persistent storage, and install the [relevant client](docs/storage-client.md) in all nodes.
> * If you encounter `Permission denied` when copying, it is recommended to check [SELinux and turn off it](./docs/turn-off-SELinux.md) first 

* Networking and DNS requirements:
  * Make sure the DNS address in `/etc/resolv.conf` is available. Otherwise, it may cause some issues of DNS in cluster.
  * If your network configuration uses Firewall or Security Group，you must ensure infrastructure components can communicate with each other through specific ports. It's recommended that you turn off the firewall or follow the link configuriation: [NetworkAccess](docs/network-access.md).

## Usage

### Get the Installer Excutable File

* Download Binary

    ```shell script
    curl -O -k https://kubernetes.pek3b.qingstor.com/tools/kubekey/kk
    chmod +x kk
    ```

or

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

### Create a Cluster

#### Quick Start

Quick Start is for `all-in-one` installation which is a good start to get familiar with KubeSphere.

##### Command

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

### Enable kubectl autocompletion

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

## Documents

* [Configuration example](docs/config-example.md)
* [Network access](docs/network-access.md)
* [Storage clients](docs/storage-client.md)
* [Roadmap](docs/roadmap.md)
