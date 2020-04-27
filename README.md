# KubeKey
Deploy a Kubernetes Cluster flexibly and easily
## Quick Start
### Check List
Please follow the list to prepare environment.

|  ID   | Check Item  |
|  :----:  | :----  |
|  1  | Require SSH can access to all nodes.  |
|  2  | It's recommended that Your OS is clean (without any other software installed), otherwise there may be conflicts.  |
|  3  | OS requirements (For Minimal Installation of KubeSphere only)：at least 2 vCPUs and 4GB RAM. |
|  4  | Make sure the storage service is available if you want to deploy a cluster with KubeSphere.<br>The relevant client should be installed on all nodes in cluster, if you storage server is [nfs / ceph / glusterfs](./docs/storage-client.md).   |
|  5  | Make sure the DNS address in /etc/resolv.conf is available. Otherwise, it may cause some issues of DNS in cluster. |
|  6  | If your network configuration uses Firewall or Security Group，you must ensure infrastructure components can communicate with each other through specific ports.<br>It's recommended that you turn off the firewall or follow the link configuriation: [NetworkAccess](./docs/NetworkAccess.md)|
|  7  | A container image mirror is recommended to be prepared, if you have trouble downloading images from dockerhub.io.  |            

### Usage
* Download binary
```shell script
curl -O -k https://kubesphere-installer.pek3b.qingstor.com/kubeocean/ko
chmod +x ko
```
* Deploy a Allinone cluster
```shell script
./ko create cluster

# Kubesphere can be installed by "--add kubesphere"
./ko create cluster --add kubesphere
```
* Deploy a MultiNodes cluster
  
> Create a cluster config file by following command or [example config file](docs/cluster-info.yaml)
```shell script
./ko create config
```
> Deploy cluster
```shell script
./ko create cluster --cluster-info ./cluster-info.yaml

# Kubesphere can be installed by "--add kubesphere"
./ko create cluster --cluster-info ./cluster-info.yaml --add kubesphere
```
* Add Nodes
> Add new node's information to the cluster config file
```shell script
./ko scale --cluster-info ./cluster-info.yaml
```
* Reset Cluster
```shell script
# allinone
./ko reset

# multinodes
./ko reset --cluster-info /home/ubuntu/cluster-info.yaml
```
### Supported
* Deploy allinone cluster
* Deploy multinodes cluster
* Add nodes (masters and nodes)

### Build
```shell script
git clone https://github.com/pixiake/kubeocean.git
cd kubeocean
./build.sh
```
> Note: Docker needs to be installed before building.

