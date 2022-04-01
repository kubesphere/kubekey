# Container Image Registry

KubeKey now allows users to create a container image registry ([Docker registry](https://docs.docker.com/registry/) or [Harbor](https://goharbor.io/)).

### How to Use

1. Initialize a local images registry:

   ```
   ./kk init registry [(-f | --filename) path]
   ```

2. Add information about the server used as the image registry to **spec.hosts** and **spec.roleGroups** in the KubeKey configuration file.

   Example:

   ```
   apiVersion: kubekey.kubesphere.io/v1alpha2
   kind: Cluster
   metadata:
     name: sample
   spec:
     hosts:
     - {name: node1, address: 192.168.6.6, internalAddress: 192.168.6.6, password: Qcloud@123}
     roleGroups:
       etcd:
       - node1
       control-plane:
       - node1
       worker:
       - node1
       ## Specify the node role as registry. Only one node can be set as registry.
       registry:
       - node1
     controlPlaneEndpoint:
       ##Internal loadbalancer for apiservers
       #internalLoadbalancer: haproxy
       domain: lb.kubesphere.local
       address: ""
       port: 6443
     kubernetes: 
       version: v1.21.5
       clusterName: cluster.local
     network:
       plugin: calico
       kubePodsCIDR: 10.233.64.0/18
       kubeServiceCIDR: 10.233.0.0/18
       # multus support. https://github.com/k8snetworkplumbingwg/multus-cni
       enableMultusCNI: false
     registry:
       ## `docker registry` is used to create local registry by default.  
       ## `harbor` can be also set for type.
       # type: "harbor"  
       privateRegistry: dockerhub.kubekey.local
       privateRegistryIp: ""
       auths:
         "dockerhub.kubekey.local":
           username: admin
           password: Harbor12345
       registryMirrors: []
       insecureRegistries: []
     addons: []
   ```

