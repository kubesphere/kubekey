# Cluster Configuration Sample
```yaml
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Cluster
metadata:
  name: sample
spec:
  hosts:
  # Assume that the default port for SSH is 22. Otherwise, add the port number after the IP address. 
  # If you install Kubernetes on ARM, add "arch: arm64". For example, {...user: ubuntu, password: Qcloud@123, arch: arm64}.
  - {name: node1, address: 172.16.0.2, internalAddress: "172.16.0.2,2022::2", port: 8022, user: ubuntu, password: "Qcloud@123"}
  # For default root user.
  # Kubekey will parse `labels` field and automatically label the node.
  - {name: node2, address: 172.16.0.3, internalAddress: "172.16.0.3,2022::3", password: "Qcloud@123", labels: {disk: SSD, role: backend}}
  # For password-less login with SSH keys.
  - {name: node3, address: 172.16.0.4, internalAddress: "172.16.0.4,2022::4", privateKeyPath: "~/.ssh/id_rsa"}
  roleGroups:
    etcd:
    - node1 # All the nodes in your cluster that serve as the etcd nodes.
    master:
    - node1
    - node[2:10] # From node2 to node10. All the nodes in your cluster that serve as the master nodes.
    worker:
    - node1
    - node[10:100] # All the nodes in your cluster that serve as the worker nodes.
    ## Specify the node role as registry. Only one node can be set as registry. For more information check docs/registry.md
    registry:
    - node1
  controlPlaneEndpoint:
    # Internal loadbalancer for apiservers. Support: haproxy, kube-vip [Default: ""]
    internalLoadbalancer: haproxy
    # Determines whether to use external dns to resolve the control-plane domain. 
    # If 'externalDNS' is set to 'true', the 'address' needs to be set to "".
    externalDNS: false  
    domain: lb.kubesphere.local
    # The IP address of your load balancer. If you use internalLoadblancer in "kube-vip" mode, a VIP is required here.
    address: ""      
    port: 6443
  system:
    # The ntp servers of chrony.
    ntpServers:
      - time1.cloud.tencent.com
      - ntp.aliyun.com
      - node1 # Set the node name in `hosts` as ntp server if no public ntp servers access.
    timezone: "Asia/Shanghai"
    # Specify additional packages to be installed. The ISO file which is contained in the artifact is required.
    rpms:
      - nfs-utils
    # Specify additional packages to be installed. The ISO file which is contained in the artifact is required.
    debs: 
      - nfs-common
    #preInstall:  # Specify custom init shell scripts for each nodes, and execute according to the list order at the first stage.
    #  - name: format and mount disk  
    #    bash: /bin/bash -x setup-disk.sh
    #    materials: # scripts can has some dependency materials. those will copy to the node        
    #      - ./setup-disk.sh # the script which shell execute need
    #      -  xxx            # other tools materials need by this script
    #postInstall: # Specify custom finish clean up shell scripts for each nodes after the Kubernetes install.
    #  - name: clean tmps files
    #    bash: |
    #       rm -fr /tmp/kubekey/*
    #skipConfigureOS: true # Do not pre-configure the host OS (e.g. kernel modules, /etc/hosts, sysctl.conf, NTP servers, etc). You will have to set these things up via other methods before using KubeKey.

  kubernetes:
    #kubelet start arguments
    #kubeletArgs:
      # Directory path for managing kubelet files (volume mounts, etc).
    #  - --root-dir=/var/lib/kubelet
    version: v1.21.5
    # Optional extra Subject Alternative Names (SANs) to use for the API Server serving certificate. Can be both IP addresses and DNS names.
    apiserverCertExtraSans:  
      - 192.168.8.8
      - lb.kubespheredev.local
    # Container Runtime, support: containerd, cri-o, isula. [Default: docker]
    containerManager: docker
    clusterName: cluster.local
    # Whether to install a script which can automatically renew the Kubernetes control plane certificates. [Default: false]
    autoRenewCerts: true
    # masqueradeAll tells kube-proxy to SNAT everything if using the pure iptables proxy mode. [Default: false].
    masqueradeAll: false
    # maxPods is the number of Pods that can run on this Kubelet. [Default: 110]
    maxPods: 110
    # podPidsLimit is the maximum number of PIDs in any pod. [Default: 10000]
    podPidsLimit: 10000
    # The internal network node size allocation. This is the size allocated to each node on your network. [Default: 24]
    nodeCidrMaskSize: 24
    # Specify which proxy mode to use. [Default: ipvs]
    proxyMode: ipvs
    # enable featureGates, [Default: {"ExpandCSIVolumes":true,"RotateKubeletServerCertificate": true,"CSIStorageCapacity":true, "TTLAfterFinished":true}]
    featureGates: 
      CSIStorageCapacity: true
      ExpandCSIVolumes: true
      RotateKubeletServerCertificate: true
      TTLAfterFinished: true
    ## support kata and NFD
    # kata:
    #   enabled: true
    # nodeFeatureDiscovery
    #   enabled: true
    # additional kube-proxy configurations
    kubeProxyConfiguration:
      ipvs:
        # CIDR's to exclude when cleaning up IPVS rules.
        # necessary to put node cidr here when internalLoadbalancer=kube-vip and proxyMode=ipvs
        # refer to: https://github.com/kubesphere/kubekey/issues/1702
        excludeCIDRs:
          - 172.16.0.2/24
  etcd:
    # Specify the type of etcd used by the cluster. When the cluster type is k3s, setting this parameter to kubeadm is invalid. [kubekey | kubeadm | external] [Default: kubekey]
    type: kubekey  
    ## The following parameters need to be added only when the type is set to external.
    ## caFile, certFile and keyFile need not be set, if TLS authentication is not enabled for the existing etcd.
    # external:
    #   endpoints:
    #     - https://192.168.6.6:2379
    #   caFile: /pki/etcd/ca.crt
    #   certFile: /pki/etcd/etcd.crt
    #   keyFile: /pki/etcd/etcd.key
    dataDir: "/var/lib/etcd"
    # Time (in milliseconds) of a heartbeat interval.
    heartbeatInterval: 250
    # Time (in milliseconds) for an election to timeout. 
    electionTimeout: 5000
    # Number of committed transactions to trigger a snapshot to disk.
    snapshotCount: 10000
    # Auto compaction retention for mvcc key value store in hour. 0 means disable auto compaction.
    autoCompactionRetention: 8
    # Set level of detail for etcd exported metrics, specify 'extensive' to include histogram metrics.
    metrics: basic
    ## Etcd has a default of 2G for its space quota. If you put a value in etcd_memory_limit which is less than
    ## etcd_quota_backend_bytes, you may encounter out of memory terminations of the etcd cluster. Please check
    ## etcd documentation for more information.
    # 8G is a suggested maximum size for normal environments and etcd warns at startup if the configured value exceeds it.
    quotaBackendBytes: 2147483648 
    # Maximum client request size in bytes the server will accept.
    # etcd is designed to handle small key value pairs typical for metadata.
    # Larger requests will work, but may increase the latency of other requests
    maxRequestBytes: 1572864
    # Maximum number of snapshot files to retain (0 is unlimited)
    maxSnapshots: 5
    # Maximum number of wal files to retain (0 is unlimited)
    maxWals: 5
    # Configures log level. Only supports debug, info, warn, error, panic, or fatal.
    logLevel: info
  network:
    plugin: calico
    calico:
      ipipMode: Always  # IPIP Mode to use for the IPv4 POOL created at start up. If set to a value other than Never, vxlanMode should be set to "Never". [Always | CrossSubnet | Never] [Default: Always]
      vxlanMode: Never  # VXLAN Mode to use for the IPv4 POOL created at start up. If set to a value other than Never, ipipMode should be set to "Never". [Always | CrossSubnet | Never] [Default: Never]
      vethMTU: 0  # The maximum transmission unit (MTU) setting determines the largest packet size that can be transmitted through your network. By default, MTU is auto-detected. [Default: 0]
    kubePodsCIDR: 10.233.64.0/18,fc00::/48
    kubeServiceCIDR: 10.233.0.0/18,fd00::/108
  storage:
    openebs:
      basePath: /var/openebs/local # base path of the local PV provisioner
  registry:
    registryMirrors: []
    insecureRegistries: []
    privateRegistry: "dockerhub.kubekey.local"
    namespaceOverride: ""
    auths: # if docker add by `docker login`, if containerd append to `/etc/containerd/config.toml`
      "dockerhub.kubekey.local":
        username: "xxx"
        password: "***"
        skipTLSVerify: false # Allow contacting registries over HTTPS with failed TLS verification.
        plainHTTP: false # Allow contacting registries over HTTP.
        certsPath: "/etc/docker/certs.d/dockerhub.kubekey.local" # Use certificates at path (*.crt, *.cert, *.key) to connect to the registry.
  addons: [] # You can install cloud-native addons (Chart or YAML) by using this field.
  #dns:
  #  ## Optional hosts file content to coredns use as /etc/hosts file.
  #  dnsEtcHosts: |
  #    192.168.0.100 api.example.com
  #    192.168.0.200 ingress.example.com
  #  coredns:
  #    ## additionalConfigs adds any extra configuration to coredns
  #    additionalConfigs: |
  #      whoami
  #      log
  #    ## Array of optional external zones to coredns forward queries to. It's injected into coredns' config file before
  #    ## default kubernetes zone. Use it as an optimization for well-known zones and/or internal-only domains, i.e. VPN for internal networks (default is unset)
  #    externalZones:
  #    - zones:
  #      - example.com
  #      - example.io:1053
  #      nameservers:
  #      - 1.1.1.1
  #      - 2.2.2.2
  #      cache: 5
  #    - zones:
  #      - mycompany.local:4453
  #      nameservers:
  #      - 192.168.0.53
  #      cache: 10
  #    - zones:
  #      - mydomain.tld
  #      nameservers:
  #      - 10.233.0.3
  #      cache: 5
  #      rewrite:
  #      - name substring website.tld website.namespace.svc.cluster.local
  #    ## Rewrite plugin block to perform internal message rewriting.
  #    rewriteBlock: |
  #      rewrite stop {
  #        name regex (.*)\.my\.domain {1}.svc.cluster.local
  #        answer name (.*)\.svc\.cluster\.local {1}.my.domain
  #      }
  #    ## DNS servers to be added *after* the cluster DNS. These serve as backup
  #    ## DNS servers in early cluster deployment when no cluster DNS is available yet.
  #    upstreamDNSServers:
  #    - 8.8.8.8
  #    - 1.2.4.8
  #    - 114.114.114.114
  #  nodelocaldns:
  #    ## It's possible to extent the nodelocaldns' configuration by adding an array of external zones.
  #    externalZones:
  #    - zones:
  #      - example.com
  #      - example.io:1053
  #      nameservers:
  #      - 1.1.1.1
  #      - 2.2.2.2
  #      cache: 5
  #    - zones:
  #      - mycompany.local:4453
  #      nameservers:
  #      - 192.168.0.53
  #      cache: 10
  #    - zones:
  #      - mydomain.tld
  #      nameservers:
  #      - 10.233.0.3
  #      cache: 5
  #      rewrite:
  #      - name substring website.tld website.namespace.svc.cluster.local

```

# Network Configuration sample
## Hybridnet
To learn more about hybridnet, check out https://github.com/alibaba/hybridnet
```yaml
  network:
    plugin: hybridnet
    hybridnet:
      defaultNetworkType: Overlay
      enableNetworkPolicy: false
      init: false
      preferVxlanInterfaces: eth0
      preferVlanInterfaces: eth0
      preferBGPInterfaces: eth0
      networks:
      - name: "net1"
        type: Underlay
        nodeSelector:
          network: "net1"
        subnets:
          - name: "subnet-10"
            netID: 10
            cidr: "192.168.10.0/24"
            gateway: "192.168.10.1"
          - name: "subnet-11"
            netID: 11
            cidr: "192.168.11.0/24"
            gateway: "192.168.11.1"
      - name: "net2"
        type: Underlay
        nodeSelector:
          network: "net2"
        subnets:
          - name: "subnet-30"
            netID: 30
            cidr: "192.168.30.0/24"
            gateway: "192.168.30.1"
          - name: "subnet-31"
            netID: 31
            cidr: "192.168.31.0/24"
            gateway: "192.168.31.1"
      - name: "net3"
        type: Underlay
        netID: 0
        nodeSelector:
          network: "net3"
        subnets:
          - name: "subnet-50"
            cidr: "192.168.50.0/24"
            gateway: "192.168.50.1"
            start: "192.168.50.100"
            end: "192.168.50.200"
            reservedIPs: ["192.168.50.101","192.168.50.102"]
            excludeIPs: ["192.168.50.111","192.168.50.112"]
```
