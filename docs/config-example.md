```yaml
apiVersion: kubekey.kubesphere.io/v1alpha1
kind: Cluster
metadata:
  name: example
spec:
  hosts:
  - {name: node1, address: 172.16.0.2, internalAddress: 172.16.0.2, user: ubuntu, password: Qcloud@123}
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
    version: v1.17.6
    imageRepo: kubesphere
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
  kubesphere:
    console:
      enableMultiLogin: false  # enable/disable multi login
      port: 30880
    common:
      mysqlVolumeSize: 20Gi
      minioVolumeSize: 20Gi
      etcdVolumeSize: 20Gi
      openldapVolumeSize: 2Gi
      redisVolumSize: 2Gi
    monitoring:
      prometheusReplicas: 1
      prometheusMemoryRequest: 400Mi
      prometheusVolumeSize: 20Gi
      grafana:
        enabled: false
    logging:
      enabled: false
      elasticsearchMasterReplicas: 1
      elasticsearchDataReplicas: 1
      logsidecarReplicas: 2
      elasticsearchMasterVolumeSize: 4Gi
      elasticsearchDataVolumeSize: 20Gi
      logMaxAge: 7
      elkPrefix: logstash
      containersLogMountedPath: ""
      kibana:
        enabled: false
    openpitrix:
      enabled: false
    devops:
      enabled: false
      jenkinsMemoryLim: 2Gi
      jenkinsMemoryReq: 1500Mi
      jenkinsVolumeSize: 8Gi
      jenkinsJavaOpts_Xms: 512m
      jenkinsJavaOpts_Xmx: 512m
      jenkinsJavaOpts_MaxRAM: 2g
      sonarqube:
        enabled: false
        postgresqlVolumeSize: 8Gi
    notification:
      enabled: false
    alerting:
      enabled: false
    serviceMesh:
      enabled: false
    metricsServer:
      enabled: false
```
