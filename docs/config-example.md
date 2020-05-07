```yaml
apiVersion: kubekey.kubesphere.io/v1alpha1
kind: Cluster
metadata:
  name: example
spec:
  hosts:
  - {name: node1, address: 172.16.0.2, internalAddress: 172.16.0.2, user: ubuntu, password: Qcloud@123}
  - {name: node2, address: 172.16.0.3, internalAddress: 172.16.0.3, password: Qcloud@123}
  - {name: node3, address: 172.16.0.4, internalAddress: 172.16.0.4, privateKeyPath: "~/.ssh/id_rsa"}
  roleGroups:
    etcd:
     - node1
    master: 
     - node1
     - node[2:10]
    worker:
     - node1
     - node[10:100]
  controlPlaneEndpoint:
    domain: lb.kubesphere.local
    address: ""
    port: "6443"
  kubernetes:
    version: v1.17.4
    imageRepo: kubekey
    clusterName: cluster.local
  network:
    plugin: calico
    podNetworkCidr: 10.233.64.0/18
    serviceNetworkCidr: 10.233.0.0/18
  registry:
    registryMirrors: []
    insecureRegistries: []
  storage:
    defaultStorageClass: localVolume
    nfsClient:
      nfsServer: 172.16.0.2
      nfsPath: /mnt/nfs
      nfsVrs3Enabled: false
      nfsArchiveOnDelete: false
```