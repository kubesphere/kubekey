# KubeKey
Start Kubernetes and KubeSphere flexibly and easily.
## Motivation
* Kubekey is developed with golang to reduce software dependency issues.
* Kubekey uses kubeadm to concurrently install k8s cluster to reduce installation complexity and improve installation efficiency.
* Support for scaling cluster from allinone to multi-node.
## Quick Start
### Prepare
Please follow the list to prepare environment.
#### Supported Linux Distributions
* **Ubuntu**  *16.04, 18.04*
* **Debian**  *Buster, Stretch*
* **CentOS/RHEL**  *7*
#### Requirements and Recommendations
* Require SSH can access to all nodes, `sudo` and `curl` can be used in all nodes.
* It's recommended that Your OS is clean (without any other software installed), otherwise there may be conflicts.
* OS requirements (For Minimal Installation of KubeSphere only)：at least 2 vCPUs and 4GB RAM.
* Make sure the storage service is available if you want to deploy a cluster with KubeSphere.<br>
  The relevant client should be installed on all nodes in cluster, if you storage server is [nfs / ceph / glusterfs](./docs/storage-client.md).
* Make sure the DNS address in /etc/resolv.conf is available. Otherwise, it may cause some issues of DNS in cluster.
* If your network configuration uses Firewall or Security Group，you must ensure infrastructure components can communicate with each other through specific ports.<br>
  It's recommended that you turn off the firewall or follow the link configuriation: [NetworkAccess](./docs/network-access.md)
* A container image mirror (accelerator) is recommended to be prepared, if you have trouble downloading images from dockerhub.io.          

### Usage
* Download binary
```shell script
curl -O -k https://kubernetes.pek3b.qingstor.com/tools/kubekey/kk
chmod +x kk
```
* Deploy a cluster
```shell script
# allinone
./kk create cluster

# multiNodes
# 1. Create a example configuration file by following command or reference docs/config-example.md
./kk create config      # Only kubernetes
./kk create config --add localVolume      # Add plugins (eg: localVolume / nfsClient / localVolume,nfsClient)

# 2. Please fill in the configuration file under the current path (k2cluster-example.yaml) according to the environmental information.

# 3. Deploy cluster
./kk create cluster -f ./k2cluster-example.yaml
```
* Add Nodes
> Add new node's information to the cluster config file
```shell script
./kk scale -f ./k2cluster-example.yaml
```
* Reset Cluster
```shell script
# allinone
./kk reset

# multinodes
./kk reset -f ./k2cluster-example.yaml
```

## Build
```shell script
git clone https://github.com/kubesphere/kubekey.git
cd kubekey
./build.sh
```
> Note: Docker needs to be installed before building.
## Road Map
* CaaO (Cluster as a Object)
* Support more container runtimes: cri-o containerd

