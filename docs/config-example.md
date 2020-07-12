```yaml
apiVersion: kubekey.kubesphere.io/v1alpha1
kind: Cluster
metadata:
  name: example
spec:
  hosts:
  - {name: node1, address: 172.16.0.2, internalAddress: 172.16.0.2, port: 8022, user: ubuntu, password: Qcloud@123} # Assume that the default port for SSH is 22, otherwise add the port number after the IP address as above
  - {name: node2, address: 172.16.0.3, internalAddress: 172.16.0.3, password: Qcloud@123}  # the default root user
  - {name: node3, address: 172.16.0.4, internalAddress: 172.16.0.4, privateKeyPath: "~/.ssh/id_rsa"} # password-less login with SSH keys
  roleGroups:
    etcd:
     - node1
    master:
     - node1
     - node[2:10] # the nodes from node2, node3,..., to node10
    worker:
     - node1
     - node[10:100]
  controlPlaneEndpoint:
    domain: lb.kubesphere.local
    address: ""
    port: "6443"
  kubernetes:
    version: v1.17.8
    imageRepo: kubesphere
    clusterName: cluster.local
    masqueradeAll: false  # masqueradeAll tells kube-proxy to SNAT everything if using the pure iptables proxy mode. [Default: false]
    maxPods: 110  # maxPods is the number of pods that can run on this Kubelet. [Default: 110]
    nodeCidrMaskSize: 24  # internal network node size allocation. This is the size allocated to each node on your network. [Default: 24]
  network:
    plugin: calico
    calico:
      ipipMode: Always  # IPIP Mode to use for the IPv4 POOL created at start up. If set to a value other than Never, vxlanMode should be set to "Never". [Always | CrossSubnet | Never] [Default: Always]
      vxlanMode: Never  # VXLAN Mode to use for the IPv4 POOL created at start up. If set to a value other than Never, ipipMode should be set to "Never". [Always | CrossSubnet | Never] [Default: Never]
      vethMTU: 1440  # The maximum transmission unit (MTU) setting determines the largest packet size that can be transmitted through your network. [Default: 1440]
    podNetworkCidr: 10.233.64.0/18
    serviceNetworkCidr: 10.233.0.0/18
  registry:
    registryMirrors: []
    insecureRegistries: []
    privateRegistry: ""
  storage:
    defaultStorageClass: localVolume
    localVolume:
      storageClassName: local
    nfsClient:
      storageClassName: nfs-client
      # Hostname of the NFS server(ip or hostname)
      nfsServer: SHOULD_BE_REPLACED
      # Basepath of the mount point
      nfsPath: SHOULD_BE_REPLACED
      nfsVrs3Enabled: false
      nfsArchiveOnDelete: false
    rbd:
      storageClassName: rbd
      # Ceph rbd monitor endpoints, for example
      # monitors:
      #   - 172.25.0.1:6789
      #   - 172.25.0.2:6789
      #   - 172.25.0.3:6789
      monitors:
      - SHOULD_BE_REPLACED
      adminID: admin
      # ceph admin secret, for example,
      # adminSecret: AQAnwihbXo+uDxAAD0HmWziVgTaAdai90IzZ6Q==
      adminSecret: TYPE_ADMIN_ACCOUNT_HERE
      userID: admin
      # ceph user secret, for example,
      # userSecret: AQAnwihbXo+uDxAAD0HmWziVgTaAdai90IzZ6Q==
      userSecret: TYPE_USER_SECRET_HERE
      pool: rbd
      fsType: ext4
      imageFormat: 2
      imageFeatures: layering
    glusterfs:
      storageClassName: glusterfs
      restAuthEnabled: true
      # e.g. glusterfs_provisioner_resturl: http://192.168.0.4:8080
      restUrl: SHOULD_BE_REPLACED
      # e.g. glusterfs_provisioner_clusterid: 6a6792ed25405eaa6302da99f2f5e24b
      clusterID: SHOULD_BE_REPLACED
      restUser: admin
      secretName: heketi-secret
      gidMin: 40000
      gidMax: 50000
      volumeType: replicate:2
      # e.g. jwt_admin_key: 123456
      jwtAdminKey: SHOULD_BE_REPLACED

```
