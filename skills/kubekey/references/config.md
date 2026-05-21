# Config Complete Configuration Reference

Config file defines all cluster configuration options.

## Basic Structure

```yaml
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Config
spec:
  # Cluster parameter constraints
  cluster_require:
    # etcd WAL fsync max latency (nanoseconds)
    etcd_disk_wal_fysnc_duration_seconds: 10000000
    # Allow installation on unsupported distributions
    allow_unsupported_distribution_setup: false
    # Supported operating systems
    supported_os_distributions:
      - ubuntu
      - centos
      - kylin
      - rocky
    # Required network plugins
    require_network_plugin: ['calico', 'flannel', 'cilium', 'hybridnet', 'kube-ovn']
    # Minimum Kubernetes version
    kube_version_min_required: v1.23.0
    # Minimum memory requirements (MB)
    minimal_master_memory_mb: 1024
    minimal_node_memory_mb: 512
    # Supported deployment types
    require_etcd_deployment_type: ['internal', 'external']
    require_container_manager: ['docker', 'containerd']
    # Supported CPU architectures
    supported_architectures:
      - amd64
      - x86_64
      - arm64
      - aarch64
    # Minimum kernel version
    min_kernel_version: 4.9.17

  # Working directory
  work_dir: /root/kubekey
  tmp_dir: /tmp/kubekey

  # Zone (set to "cn" for domestic sources in China)
  zone: ""

  # Security enhancement
  security_enhancement: false

  # Audit logging
  audit: false

  # Cleanup options when deleting nodes
  delete:
    cri: false
    etcd: false
    dns: false
    image_registry: false
    data: false

  # Certificate configuration
  certs:
    ca:
      date: 87600h
      gen_cert_policy: IfNotPresent
    kubernetes_ca:
      date: 87600h
      gen_cert_policy: IfNotPresent
    front_proxy_ca:
      date: 87600h
      gen_cert_policy: IfNotPresent
    etcd:
      date: 87600h
      gen_cert_policy: IfNotPresent
    image_registry:
      date: 87600h
      gen_cert_policy: IfNotPresent

  # Image registry configuration
  image_registry:
    type: ""  # harbor, docker-registry
    ha_vip: ""
    auth:
      registry: ""
      username: ""
      password: ""

  # OS configuration
  native:
    ntp:
      servers:
        - "cn.pool.ntp.org"
      enabled: true
    timezone: Asia/Shanghai
    nfs:
      share_dir:
        - /share/
    set_hostname: true
    localDNS:
      - /etc/hosts
```

## Kubernetes Configuration

```yaml
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Config
spec:
  kubernetes:
    # Cluster name
    cluster_name: kubekey

    # Image repository
    image_repository: "{{ .image_registry.k8sio_registry }}"

    # Pause image
    sandbox_image:
      registry: "{{ .image_registry.k8sio_registry }}"
      repository: "pause"

    # kube-apiserver configuration
    apiserver:
      port: 6443
      certSANs: []
      extra_args:
        # feature-gates: ExpandCSIVolumes=true,CSIStorageCapacity=true

    # kube-controller-manager configuration
    controller_manager:
      extra_args:
        cluster-signing-duration: 87600h

    # kube-scheduler configuration
    scheduler:
      extra_args: {}

    # kube-proxy configuration
    kube_proxy:
      manage:
        enabled: false
      mode: "ipvs"  # ipvs, iptables
      config:
        iptables:
          masqueradeAll: false
          masqueradeBit: 14
          minSyncPeriod: 0s
          syncPeriod: 30s

    # kubelet configuration
    kubelet:
      max_pods: 110
      pod_pids_limit: 10000
      container_log_max_size: 5Mi
      container_log_max_files: 3

    # Control plane endpoint
    control_plane_endpoint:
      host: lb.kubesphere.local
      port: 6443
      type: local  # local, kube_vip, haproxy
      local:
        address: ""
      kube_vip:
        address: ""
        mode: ARP
      haproxy:
        address: 127.0.0.1
        health_port: 8081

    # Certificate configuration
    certs:
      ca_cert: ""
      ca_key: ""
      front_proxy_cert: ""
      front_proxy_key: ""
      renew: true

    # Custom patches
    patches: []

    # Skip kubeadm phases
    skip_phases: []
```

### kube-apiserver extra_args Example

```yaml
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Config
spec:
  kubernetes:
    apiserver:
      extra_args:
        feature-gates: ExpandCSIVolumes=true,CSIStorageCapacity=true
        anonymous-auth: "false"
        enable-admission-plugins: NodeRestriction,ServiceAccount
        default-not-ready-toleration-seconds: "300"
        default-unreachable-toleration-seconds: "300"
```

### kube-controller_manager extra_args Example

```yaml
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Config
spec:
  kubernetes:
    controller_manager:
      extra_args:
        cluster-signing-duration: 87600h
        feature-gates: ExpandCSIVolumes=true,CSIStorageCapacity=true
        terminated-pod-gc-period: "30s"
```

### kubelet extra_args Example

```yaml
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Config
spec:
  kubernetes:
    kubelet:
      extra_args:
        container-log-max-files: "3"
        container-log-max-size: "5Mi"
        max-pods: "110"
        pod-pids_limit: "10000"
```

## CNI Network Configuration

```yaml
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Config
spec:
  cni:
    # Network plugin: calico, cilium, flannel, hybridnet, kubeovn
    type: calico

    # Pod CIDR
    pod_cidr: 10.233.64.0/18
    ipv4_mask_size: 24
    ipv6_mask_size: 64

    # Service CIDR
    service_cidr: 10.233.0.0/18

    # Multi-CNI: multus, spiderpool
    multi_cni: "none"
```

### Different CNI Plugin Configurations

#### Calico

```yaml
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Config
spec:
  cni:
    type: calico
    pod_cidr: 10.233.64.0/18
    service_cidr: 10.233.0.0/18
    ipv4_mask_size: 24
```

#### Flannel

```yaml
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Config
spec:
  cni:
    type: flannel
    pod_cidr: 10.233.64.0/18
    service_cidr: 10.233.0.0/18
```

#### Cilium

```yaml
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Config
spec:
  cni:
    type: cilium
    pod_cidr: 10.233.64.0/18
    service_cidr: 10.233.0.0/18
```

#### Kube-OVN

```yaml
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Config
spec:
  cni:
    type: kubeovn
    pod_cidr: 10.233.64.0/18
    service_cidr: 10.233.0.0/18
    ipv4_mask_size: 24
```

## CRI Container Runtime Configuration

```yaml
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Config
spec:
  cri:
    # Container runtime: containerd, docker
    container_manager: containerd

    # Cgroup driver: systemd, cgroupfs
    cgroup_driver: systemd

    # Registry configuration
    registry:
      mirrors:
        - https://registry-1.docker.io
        - https://mirror.gcr.io
      insecure_registries: []
      auths: []
      # Example:
      # auths:
      #   - registry: docker.io
      #     username: MyDockerAccount
      #     password: my_password
      #     skip_tls_verify: true
    skip_tls_verify: false
```

### containerd Additional Configuration

```yaml
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Config
spec:
  cri:
    container_manager: containerd
    containerd:
      version: "3.9"
      # Additional registry mirrors
      registry:
        mirrors:
          - "https://docker.m.daocloud.io"
```

## etcd Configuration

```yaml
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Config
spec:
  etcd:
    # Deployment type: internal, external
    deployment_type: external
    port: 2379
    peer_port: 2380
    env:
      election_timeout: 5000
      heartbeat_interval: 250
      compaction_retention: 8
      snapshot_count: 10000
      data_dir: /var/lib/etcd
      token: k8s_etcd

    # Backup configuration (internal type only)
    backup:
      backup_dir: /var/lib/etcd-backup
      keep_backup_number: 5
      etcd_backup_script: "backup.sh"
      on_calendar: "*-*-* *:00/30:00"

    # Performance tuning
    performance: false
    traffic_priority: false
```

### external etcd Configuration

```yaml
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Config
spec:
  etcd:
    deployment_type: external
    # Need to configure etcd nodes in inventory
    # External etcd cluster needs to be prepared in advance
```

### internal etcd Configuration (managed by kubekey)

```yaml
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Config
spec:
  etcd:
    deployment_type: internal
    # Data directory
    env:
      data_dir: /var/lib/etcd
    # Auto backup
    backup:
      backup_dir: /var/lib/etcd-backup
      keep_backup_number: 5
```

## DNS Configuration

```yaml
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Config
spec:
  dns:
    # Cluster domain
    domain: cluster.local

    # NodeLocalDNS
    nodelocaldns:
      enabled: true
      ip: 169.254.25.10

    # CoreDNS
    coredns:
      dns_etc_hosts: []
      zone_configs:
        - zones: [".:53"]
          additional_configs:
            - errors
            - ready
            - prometheus :9153
            - loop
            - reload
            - loadbalance
          cache: 30
          kubernetes:
            zone: ""
          forward:
            - from: "."
              to: ["/etc/resolv.conf"]
              except: []
              force_tcp: false
              prefer_udp: false
              max_concurrent: 1000
```

### Custom DNS Forwarding

```yaml
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Config
spec:
  dns:
    domain: cluster.local
    coredns:
      zone_configs:
        - zones: [".:53"]
          forward:
            - from: "."
              to: ["/etc/resolv.conf"]
            - from: "example.com"
              to: ["10.0.0.1:53"]
```

## Storage Class Configuration

```yaml
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Config
spec:
  storage_class:
    # Local storage
    local:
      enabled: true
      default: true
      path: /var/openebs/local

    # NFS storage
    nfs:
      enabled: false
      default: false
      server: ""
      path: /share/kubernetes
```

### Enable NFS Storage

```yaml
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Config
spec:
  storage_class:
    nfs:
      enabled: true
      default: false
      server: 192.168.1.100
      path: /share/kubernetes
```

### Disable Local Storage

```yaml
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Config
spec:
  storage_class:
    local:
      enabled: false
      default: false
```

## Image Registry Configuration

```yaml
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Config
spec:
  image_registry:
    # Install type: harbor, docker-registry (empty means no installation)
    type: harbor
    
    # HA VIP (high availability mode)
    ha_vip: ""
    
    auth:
      # Registry address
      registry: my-registry.example.com
      # Authentication
      username: admin
      password: Harbor12345
```

### Harbor Configuration Example

```yaml
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Config
spec:
  image_registry:
    type: harbor
    auth:
      registry: harbor.kubesphere.local
      username: admin
      password: Harbor12345
```

### Docker Registry Configuration Example

```yaml
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Config
spec:
  image_registry:
    type: docker-registry
    auth:
      registry: registry.kubesphere.local
      username: admin
      password: Reg12345
```

## Security Enhancement Configuration

```yaml
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Config
spec:
  # Enable security enhancement
  security_enhancement: true
  
  # Enable audit logging
  audit: true
```

## Complete Minimal Configuration Example

```yaml
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Config
metadata:
  name: config
spec:
  kubernetes:
    version: v1.33.7
  cni:
    type: calico
    pod_cidr: 10.233.64.0/18
    service_cidr: 10.233.0.0/18
  cri:
    container_manager: containerd
  etcd:
    deployment_type: external
  storage_class:
    local:
      enabled: true
      default: true
      path: /var/openebs/local
```

## High Availability Configuration Example

```yaml
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Config
metadata:
  name: ha-config
spec:
  kubernetes:
    version: v1.33.7
    control_plane_endpoint:
      host: lb.kubesphere.local
      port: 6443
      type: kube_vip
      kube_vip:
        address: 192.168.1.100
        mode: ARP
  cni:
    type: calico
    pod_cidr: 10.233.64.0/18
    service_cidr: 10.233.0.0/18
  cri:
    container_manager: containerd
  etcd:
    deployment_type: internal
    backup:
      backup_dir: /var/lib/etcd-backup
      keep_backup_number: 5
  image_registry:
    type: harbor
    ha_vip: 192.168.1.101
    auth:
      registry: harbor.kubesphere.local
      username: admin
      password: Harbor12345
```
