# KubeKey

Since v3.0, KubeSphere changes the ansible-based installer to the new installer called KubeKey that is developed in Go language. With KubeKey, you can install Kubernetes and KubeSphere separately or as a whole easily, efficiently and flexibly.

## Motivation

* Ansible-based installer has a bunch of software dependency such as Python. KubeKey is developed in Go language to get rid of the problem in a variety of environment so that increasing the success rate of installation.
* KubeKey uses Kubeadm to install K8s cluster on nodes in parallel as much as possible in order to reduce installation complexity and improve efficiency. It will greatly save installation time compared to the older installer.
* KubeKey supports for scaling cluster from allinone to multi-node cluster.
* KubeKey aims to install cluster as an object, i.e., CaaO.

## Quick Start

### Prepare

Please follow the list to prepare environment.

#### Supported Linux Distributions

* **Ubuntu**  *16.04, 18.04*
* **Debian**  *Buster, Stretch*
* **CentOS/RHEL**  *7*

#### Requirements and Recommendations

* Require SSH can access to all nodes. `sudo` and `curl` can be used in all nodes.
* It's recommended that Your OS is clean (without any other software installed), otherwise there may be conflicts.
* OS requirements (For Minimal Installation of KubeSphere only)：at least 2 vCPUs and 4 GB RAM.
* Make sure the storage service is available if you want to deploy a cluster with KubeSphere.
  The relevant client should be installed on all nodes in cluster, if you storage server is [nfs / ceph / glusterfs](./docs/storage-client.md).
* Make sure the DNS address in /etc/resolv.conf is available. Otherwise, it may cause some issues of DNS in cluster.
* If your network configuration uses Firewall or Security Group，you must ensure infrastructure components can communicate with each other through specific ports. It's recommended that you turn off the firewall or follow the link configuriation: [NetworkAccess](./docs/network-access.md).
* A container image mirror (accelerator) is recommended to be prepared, if you have trouble downloading images from dockerhub.io.

### Usage

#### Download binary

```shell script
curl -O -k https://kubernetes.pek3b.qingstor.com/tools/kubekey/kk
chmod +x kk
```

#### Deploy a cluster

* allinone

```shell script
./kk create cluster
```

* multi-node

```shell script
# 1. Create an example configuration file by the following command or the docs/config-example.md
$ ./kk create config      # Only for kubernetes
$ ./kk create config --add localVolume      # Add plugins (eg: localVolume / nfsClient / localVolume,nfsClient)

# 2. Please fill in the configuration file under the current path (k2cluster-example.yaml) according to the environmental information.

# 3. Deploy cluster
$ ./kk create cluster -f ./k2cluster-example.yaml
```

#### Add Nodes

> Add new node's information to the cluster config file, then apply the changes.

```shell script
./kk scale -f ./k2cluster-example.yaml
```

#### Reset Cluster

* allinone

```shell script
./kk reset
```

* multi-node

```shell script
./kk reset -f ./k2cluster-example.yaml
```

## Build

```shell script
git clone https://github.com/kubesphere/kubekey.git
cd kubekey
./build.sh
```

**Note:**

* Docker needs to be installed before building.
* If you have problem to access `https://proxy.golang.org/` in China mainland, please open the build.sh to use the Go module proxy in China.

## Road Map

* CaaO (Cluster as an Object)
* Support more container runtimes: cri-o containerd
