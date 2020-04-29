```yaml
apiVersion: kubekey.io/v1alpha1
kind: K2Cluster
metadata:
  name: demo
spec:
  hosts:
  - hostName: node1
    sshAddress: 172.16.0.2
    internalAddress: 172.16.0.2
    port: "22"
    user: ubuntu
    password: Qcloud@123
    sshKeyPath: ""
    role:
    - etcd
    - master
    - worker
  lbKubeapiserver:
    domain: lb.kubesphere.local
    address: ""
    port: "6443"
  kubeCluster:
    version: v1.17.4
    imageRepo: kubekey
    clusterName: cluster.local
  network:
    plugin: calico
    kube_pods_cidr: 10.233.64.0/18
    kube_service_cidr: 10.233.0.0/18
  registry:
    registryMirrors: []
    insecureRegistries: []
  plugins:
    localVolume:
      enabled: true
      isDefaultClass: true
    nfsClient:
      enabled: true
      isDefaultClass: false
      nfsServer: 172.16.0.2
      nfsPath: /mnt/nfs
      nfsVrs3Enabled: false
      nfsArchiveOnDelete: false
```