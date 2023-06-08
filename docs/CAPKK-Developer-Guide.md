# CAPKK Developer Guide

CAPKK (Cluster-API-Provider-Kubekey) is an SSH-based infrastructure provider that follows the development documentation of the cluster-api Provider contracts. (https://cluster-api.sigs.k8s.io/developer/providers/contracts.html#provider-contract)

In the CAPKK project, there are three main resources and corresponding controllers: KKCluster, KKMachine, and KKInstance.

## Project structure

```bash
├── api
│   └── v1beta1              ## CR defined in CAPKK
├── bootstrap                ## bootstrap provider for k3s version
├── config                   ## Kubernetes yaml files for CAPKK project
├── controllers              ## controllers included in CAPKK
│   ├── kkcluster
│   ├── kkinstance
│   └── kkmachine
├── controlplane             ## control-plane provider for k3s version
├── docs                     ## documentation
├── exp                      ## experimental features
├── hack
├── pkg
│   ├── clients              ## clients used in CAPKK
│   │   └── ssh
│   ├── rootfs               ## data storage directory for CAPKK
│   ├── scope                ## interfaces defined by functionality in CAPKK, mainly used to split KKCluster, KKMachine, and KKInstance attribute fields.
│   ├── service              ## concrete implementation of various operations in CAPKK
│   │   ├── binary           ## operations on binary files in CAPKK
│   │   ├── bootstrap        ## operations on machine environment initialization in CAPKK
│   │   ├── containermanager ## operations on container runtime in CAPKK
│   │   ├── operation        ## basic operations on Linux machines in CAPKK
│   │   ├── provisioning     ## parsing cloud-init or ignition files in CAPKK and mapping them to SSH commands
│   │   ├── repository       ## operations on rpm packages in CAPKK
│   │   └── util
│   └── util
│       ├── filesystem
│       └── hash
├── scripts
├── templates                ## used to generate automated scripts in sample yaml for GitHub releases
├── test                     ## e2e testing in CAPKK
│   └── e2e
├── util                     ## general utility class for KubeKey
└── version                  ## contains KubeKey's version and SHA256 of binary
````

## KKCluster
KKCluster is used to define the global basic information of a cluster.
```go
type KKClusterSpec struct {
     // Define the distribution of Kubernetes, with currently supported parameters including "kubernetes" and "k3s".
     Distribution string `json:"distribution,omitempty"`
     
     // Define the SSH authentication information for the nodes.
     Nodes Nodes `json:"nodes"`
     
     // Define the endpoints for the control plane.
     // Optional
     ControlPlaneEndpoint clusterv1.APIEndpoint `json:"controlPlaneEndpoint"`
     
     // Define the address for the load balancer for the control plane.
     ControlPlaneLoadBalancer *KKLoadBalancerSpec `json:"controlPlaneLoadBalancer,omitempty"`
     
     // Define the download address for the cluster component binaries, which is usually used to overwrite default addresses in offline deployment scenarios.
     // Optional
     Component *Component `json:"component,omitempty"`
     
     // Define information for the private image repository.
     // Optional
     Registry Registry `json:"registry,omitempty"`
}
```
In the above content, the following points should be noted:
* The nodes field specifies the SSH information of all nodes included in the cluster.
* CAPKK uses kube-vip by default to achieve high availability of the cluster control plane, so the controlPlaneLoadBalancer can be set to any unused IP within the subnet.
* CAPKK uses the official download address of binary components by default. If you want to use the domestic resource address maintained by the KubeSphere team, you can configure it as follows:
```yaml
spec:
  component:
    zone: cn
```
* The component field can also be used to specify a custom file server address and configure the overrides array to modify the path and SHA256 checksum of different binaries in the file server used by CAPKK.
  
The following is a common KKCluster sample YAML file:
```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: KKCluster
metadata:
  name: quick-start
spec:
  component:
    zone: ""
  nodes:
    auth:
      user: ubuntu
      password: P@88w0rd
    instances:
      - name: test1
        address: 192.168.0.3
      - name: test2
        address: 192.168.0.4
      - name: test3
        address: 192.168.0.5
  controlPlaneLoadBalancer:
    host: 192.168.0.100
```
After applying this YAML file, the webhook of KKCluster will add some default values, such as `roles: [control-plane, worker]`.
## KKMachine
KKMachine is mainly used to represent the mapping between Kubernetes node and infrastructure host.
```go
type KKMachineSpec struct {
     // Unique ID of the infrastructure provider
     ProviderID *string `json:"providerID,omitempty"`
     
     // Name of the corresponding KKInstance resource
     InstanceID *string `json:"instanceID,omitempty"`
     
     // Role of the node
     // Optional
     Roles []Role `json:"roles"`
     
     // Container runtime of the corresponding node for this resource
     // Optional
     ContainerManager ContainerManager `json:"containerManager"`
     
     // software packages of the corresponding node for this resource
     // Optional
     Repository *Repository `json:"repository,omitempty"`
}
```
Usually, we do not directly use KKMachine resources, but declare KKMachineTemplate resources for reference by cluster-api resources.


### KKMachineTemplate
The xxxTemplate resource is the template type of this type of resource, which contains all the contents in the spec of the corresponding resource. In the usage of cluster-api, we declare resources with the replica field, such as KubeadmControlPlane (KCP) and MachineDeployment (MD), indicating that this type of resource can be dynamically scaled. These resources also contain an infrastructureRef field, which refers to a specific infrastructure template resource (xxxMachineTemplate). When scaling up, KCP and MD resources will create corresponding, real single instances of the resource based on the template resource.
The following is an example of a common control plane KKMachineTemplate yaml:
```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: KKMachineTemplate
metadata:
  name: quick-start
  namespace: default
spec:
  template:
    spec:
      roles:
        - control-plane
      containerManager:
        type: containerd     
      repository:
        iso: "none"
        update: false
```
Notes:
- The roles specify control-plane, which means that the resources that reference this KKMachineTemplate will only select machines with corresponding roles in KKCluster for bootstrapping.
- containerManager specifies that the container runtime configured in the KKMachine created by this template is containerd.
- repository specifies the policy for the KKMachine to operate rpm software sources created by this template. In this case, the policy indicates that CAPKK will use the default software source on the corresponding Linux machine to install default software dependencies (conntrack, socat, etc.).
## KKInstance
In the developer documentation of cluster-api, the definition of infra only includes xxxCluster and xxxMachine resources, while KKInstance is a resource exclusive to CAPKK. The additional definition of KKInstance in CAPKK aims to decouple the logic of the operator and controller, i.e., KKMachine is focused on maintaining interactions with cluster-api CR, while KKInstance focuses on maintaining Linux machines (by performing command-based operations on the machines via SSH).
The property fields of KKInstance mainly consist of the collection of nodes and KKMachine fields in KKCluster, which will not be elaborated here.

The controller of KKInstance mainly performs the following operations on the machine:

* Bootstrop: Some machine environment initialization or cleanup operations defined in CAPKK, such as creating corresponding directories, adding Linux users, running initialization shell scripts, resetting the network, deleting files, and resetting Kubeadm.
* Repository: RPM software source handling operations defined in CAPKK, such as mounting ISO software packages, updating sources, and installing dependency software packages.
* Binary: Cluster component binary file handling operations defined in CAPKK, such as downloading binary files.
* ContainerManager: Machine container runtime operations defined in CAPKK, such as checking container runtime and installing container runtime.
* Provisioning: Parsing cloud-init or ignition files provided by cluster-api for the corresponding machine in CAPKK and mapping them to SSH commands. This cloud-init file will include operations such as "kubeadm init" and "kubeadm join".

For the interface definitions of these operations, see [interface](https://github.com/kubesphere/kubekey/blob/master/pkg/service/interface.go).
