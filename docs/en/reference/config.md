# Configuration Reference

This document summarizes all available parameters in KubeKey's built-in default configuration files.
These defaults are located in the `builtin/core/roles/defaults/defaults/main/` directory.
You can refer to this document when writing or modifying your own cluster configuration file.

---

## Main Configuration (01-main.yaml)

### Default Configuration

```yaml
# KubeKey working directory
work_dir: /root/kubekey
# KubeKey binary file storage directory
binary_dir: >-
  {{ .work_dir }}/kubekey
# Scripts storage directory
scripts_dir: >-
  {{ .work_dir }}/scripts
# Artifacts storage directory
artifact_dir: >-
  {{ .work_dir }}/artifact
# Temporary directory
tmp_dir: /tmp/kubekey

# Map common machine architecture names to standard forms
transform_architectures:
  amd64:
    - amd64
    - x86_64
  arm64:
    - arm64
    - aarch64

# If set to "cn", online downloads will use available domestic sources whenever possible
zone: ""

# Enable enhanced security features to meet stricter cluster security requirements
security_enhancement: false

# Enable Kubernetes audit logs
# Audit logs record and track critical operations within the cluster, helping administrators monitor security events, troubleshoot issues, and meet compliance requirements (e.g., SOC2, ISO 27001)
audit: false

delete:
# When removing a node, also uninstall the node's container runtime (CRI), such as Docker or containerd
# deleteCRI: true
  cri: false

# When removing a node, also uninstall etcd from the node
# deleteETCD: true
  etcd: false

# When removing a node, restore the node's DNS configuration
# deleteDNS: true
  dns: false

# When removing a node, also uninstall any private image registry (such as Harbor or registry) installed on the node
# Typically used in conjunction with nodes defined in inventory.groups.image_registry
# deleteImageRegistry: false
  image_registry: false

# When removing a node, also delete data directories (harbor data, registry data, etc.)
# Typically used with the --with-data flag or delete.data: true
# deleteData: false
  data: false

# List of container images to be synchronized to the private registry
image_manifests: []
```

### Parameter Descriptions

| Parameter | Description |
|-----------|-------------|
| `work_dir` | Root working directory used by KubeKey during installation and operation. |
| `binary_dir` | Directory for KubeKey binaries and related tools, auto-generated based on `work_dir`. |
| `scripts_dir` | Directory for scripts required during installation, auto-generated based on `work_dir`. |
| `artifact_dir` | Directory for offline packages (artifact), auto-generated based on `work_dir`. |
| `tmp_dir` | Directory for temporary files during installation. |
| `transform_architectures` | Machine architecture name standardization mapping, used to unify `amd64`/`x86_64`, `arm64`/`aarch64`, etc. |
| `zone` | Region setting. Set to `"cn"` to prioritize domestic download acceleration sources. |
| `security_enhancement` | Whether to enable cluster enhanced security features. |
| `audit` | Whether to enable Kubernetes audit logging. |
| `delete` | Resource cleanup switches when removing nodes. Includes `cri`, `etcd`, `dns`, `image_registry`, `data`. |
| `image_manifests` | Custom container image list for synchronizing to a private image registry. |

---

## Cluster Requirements (01-cluster_require.yaml)

### Default Configuration

```yaml
# Cluster parameter boundaries and compatibility requirements
cluster_require:
  # Maximum etcd disk WAL fsync duration at the 99th percentile (nanoseconds)
  etcd_disk_wal_fysnc_duration_seconds: 10000000
  # Allow installation on unsupported Linux distributions
  allow_unsupported_distribution_setup: false
  # Supported operating system distributions
  supported_os_distributions:
    - ubuntu
    - '"ubuntu"'
    - centos
    - '"centos"'
    - kylin
    - '"kylin"'
    - rocky
    - '"rocky"'
  # Supported network plugins
  require_network_plugin: ['calico', 'flannel', 'cilium', 'hybridnet', 'kube-ovn']
  # Minimum supported Kubernetes version
  kube_version_min_required: v1.23.0
  # Minimum memory (MB) required for each control plane node
  minimal_master_memory_mb: 10
  # Minimum memory (MB) required for each worker node
  minimal_node_memory_mb: 10
  # Supported etcd deployment types
  require_etcd_deployment_type: ['internal', 'external']
  # Supported container runtimes
  require_container_manager: ['docker', 'containerd']
  # Minimum required version of containerd
  containerd_min_version_required: v1.6.0
  # Supported CPU architectures
  supported_architectures:
    - amd64
    - x86_64
    - arm64
    - aarch64
  # Minimum required Linux kernel version
  min_kernel_version: 4.9.17

  # Compatible Kubernetes version matrix by Calico version
  calico_allowed_versions:
    v3.25: ["v1.23", "v1.24", "v1.25", "v1.26", "v1.27", "v1.28"]
    v3.26: ["v1.23", "v1.24", "v1.25", "v1.26", "v1.27", "v1.28"]
    v3.27: ["v1.23", "v1.24", "v1.25", "v1.26", "v1.27", "v1.28", "v1.29"]
    v3.28: ["v1.23", "v1.24", "v1.25", "v1.26", "v1.27", "v1.28", "v1.29", "v1.30"]
    v3.29: ["v1.23", "v1.24", "v1.25", "v1.26", "v1.27", "v1.28", "v1.29", "v1.30", "v1.31", "v1.32"]
    v3.30: ["v1.23", "v1.24", "v1.25", "v1.26", "v1.27", "v1.28", "v1.29", "v1.30", "v1.31", "v1.32", "v1.33"]
    v3.31: ["v1.23", "v1.24", "v1.25", "v1.26", "v1.27", "v1.28", "v1.29", "v1.30", "v1.31", "v1.32", "v1.33", "v1.34"]

  # Compatible Kubernetes version matrix by Cilium version
  cilium_allowed_versions:
    "1.14": ["v1.23", "v1.24", "v1.25", "v1.26", "v1.27"]
    "1.15": ["v1.23", "v1.24", "v1.25", "v1.26", "v1.27", "v1.28", "v1.29"]
    "1.16": ["v1.23", "v1.24", "v1.25", "v1.26", "v1.27", "v1.28", "v1.29", "v1.30"]
    "1.17": ["v1.23", "v1.24", "v1.25", "v1.26", "v1.27", "v1.28", "v1.29", "v1.30", "v1.31", "v1.32"]
    "1.18": ["v1.23", "v1.24", "v1.25", "v1.26", "v1.27", "v1.28", "v1.29", "v1.30", "v1.31", "v1.32", "v1.33"]
    "1.19": ["v1.23", "v1.24", "v1.25", "v1.26", "v1.27", "v1.28", "v1.29", "v1.30", "v1.31", "v1.32", "v1.33", "v1.34"]

  # Compatible Kubernetes version matrix by kube-ovn version
  kubeovn_allowed_versions:
    v1.12: ["v1.23", "v1.24", "v1.25", "v1.26", "v1.27", "v1.28"]
    v1.13: ["v1.23", "v1.24", "v1.25", "v1.26", "v1.27", "v1.28"]
    v1.14: ["v1.23", "v1.24", "v1.25", "v1.26", "v1.27", "v1.28", "v1.29", "v1.30", "v1.31", "v1.32", "v1.33", "v1.34"]
    v1.15: ["v1.23", "v1.24", "v1.25", "v1.26", "v1.27", "v1.28", "v1.29", "v1.30", "v1.31", "v1.32", "v1.33", "v1.34"]

  # Minimum compatible etcd version matrix by Kubernetes version
  etcd_min_versions:
    v1.23: v3.2.18
    v1.24: v3.2.18
    v1.25: v3.2.18
    v1.26: v3.2.18
    v1.27: v3.2.18
    v1.28: v3.4.13-4
    v1.29: v3.4.13-4
    v1.30: v3.4.13-4
    v1.31: v3.5.11-0
    v1.31.14: v3.5.24-0
    v1.32: v3.5.11-0
    v1.32.10: v3.5.24-0
    v1.32.11: v3.5.24-0
    v1.33.0: v3.5.11-0
    v1.33.1: v3.5.11-0
    v1.33.2: v3.5.11-0
    v1.33.3: v3.5.11-0
    v1.33.4: v3.5.11-0
    v1.33.5: v3.5.11-0
    v1.33: v3.5.21-0
    v1.34.0: v3.5.21-0
    v1.34.1: v3.5.21-0
    v1.34: v3.5.24-0
```

### Parameter Descriptions

| Parameter | Description |
|-----------|-------------|
| `cluster_require.etcd_disk_wal_fysnc_duration_seconds` | Maximum allowed etcd disk WAL fsync duration at the 99th percentile (nanoseconds), used for performance boundary checks. |
| `cluster_require.allow_unsupported_distribution_setup` | Whether to allow installation on unsupported operating system distributions. |
| `cluster_require.supported_os_distributions` | List of explicitly supported operating system distributions by KubeKey. |
| `cluster_require.require_network_plugin` | List of supported network plugins. The selected plugin will be validated against this list during deployment. |
| `cluster_require.kube_version_min_required` | Minimum Kubernetes version allowed for installation. |
| `cluster_require.minimal_master_memory_mb` | Minimum memory required for each control plane node (MB). |
| `cluster_require.minimal_node_memory_mb` | Minimum memory required for each worker node (MB). |
| `cluster_require.require_etcd_deployment_type` | Supported etcd deployment methods: `internal` (deployed within cluster) or `external` (existing external cluster). |
| `cluster_require.require_container_manager` | List of supported container runtimes: `docker`, `containerd`. |
| `cluster_require.containerd_min_version_required` | Minimum required version when using containerd. |
| `cluster_require.supported_architectures` | List of supported CPU architectures. |
| `cluster_require.min_kernel_version` | Minimum required Linux kernel version for nodes. |
| `cluster_require.calico_allowed_versions` | Compatible Kubernetes version matrix by Calico version. |
| `cluster_require.cilium_allowed_versions` | Compatible Kubernetes version matrix by Cilium version. |
| `cluster_require.kubeovn_allowed_versions` | Compatible Kubernetes version matrix by kube-ovn version. |
| `cluster_require.etcd_min_versions` | Minimum compatible etcd version matrix by Kubernetes version. |

---

## Certificate Configuration (02-certs.yaml)

### Default Configuration

```yaml
# Certificate generation configuration
# The following certificates will be generated:
# - etcd certificates
# - Kubernetes cluster certificates (replacing the CA certificate generated by kubeadm, which is limited to a 10-year validity)
# - Image registry certificates (for Harbor and similar registries)

# Certificate chain structure:
# CA (self-signed or provided)
# |- etcd.cert
# |- etcd.key
# |- etcd-client.cert
# |- etcd-client.key
# |
# |- image_registry.cert
# |- image_registry.key
# |- image-registry-client.cert
# |- image-registry-client.key
# |
# |- kubernetes.cert
# |- kubernetes.key
# |     |- kubeadm uses this to generate server certificates (kube-apiserver certificate)
# |- front-proxy.cert
# |- front-proxy.key
# |
# |- image-registry.cert
# |- image-registry.key

certs:
  # CA certificate settings
  ca:
    # CA certificate expiration time
    date: 87600h
    # Certificate generation policy:
    # IfNotPresent: Validate the certificate if it exists; generate a self-signed certificate only if it does not exist
    gen_cert_policy: IfNotPresent
  kubernetes_ca:
    date: 87600h
    # How to generate the certificate file. Supported values: IfNotPresent, Always
    gen_cert_policy: IfNotPresent
  front_proxy_ca:
    date: 87600h
    # How to generate the certificate file. Supported values: IfNotPresent, Always
    gen_cert_policy: IfNotPresent
  # etcd certificate
  etcd:
    date: 87600h
    # How to generate the certificate file. Supported values: IfNotPresent, Always
    gen_cert_policy: IfNotPresent
  # image_registry certificate
  image_registry:
    date: 87600h
    # How to generate the certificate file. Supported values: IfNotPresent, Always
    gen_cert_policy: IfNotPresent
```

### Parameter Descriptions

| Parameter | Description |
|-----------|-------------|
| `certs` | Defines all certificates that KubeKey needs to generate or manage. |
| `certs.ca` | Cluster root CA certificate configuration, affecting CA for etcd, kubernetes, and image registry services. |
| `certs.ca.date` | CA certificate validity period, e.g., `87600h` represents 10 years. |
| `certs.ca.gen_cert_policy` | CA certificate generation policy. `IfNotPresent` means generate only if missing; `Always` means regenerate every time. |
| `certs.kubernetes_ca` | Kubernetes cluster CA certificate configuration. |
| `certs.front_proxy_ca` | Kubernetes front-proxy CA certificate configuration, used for the aggregation layer (e.g., metrics-server). |
| `certs.etcd` | CA and node certificate configuration for the etcd cluster. |
| `certs.image_registry` | TLS certificate configuration for the private image registry (e.g., Harbor). |

---

## Image Registry Configuration (02-image_registry.yaml)

### Default Configuration

```yaml
# In an online environment (when image_registry.auth.registry is empty), images are pulled directly from their original registries to the cluster.
# In an offline environment (when image_registry.auth.registry is set), images are first pulled from the source registry, cached locally,
# then pushed to a private registry (such as Harbor), and then used by the cluster.

image_registry:
  # Specify which image registry to install. Supported values: harbor, docker-registry
  # If left empty, no image registry will be installed (assumes an existing registry is already available).
  type: ""
  # Virtual IP (VIP) for high availability of the image registry
  ha_vip: ""
  harbor:
    http_port: 80
    https_port: 443
  # Image registry authentication settings
  auth:
    registry: >-
      {{- if .image_registry.type | empty | not -}}
        {{- if .image_registry.ha_vip | empty | not -}}
      {{ .image_registry.ha_vip }}
        {{- else if .groups.image_registry | default list | empty | not -}}
          {{- $internalIPv4 := index .hostvars (.groups.image_registry | default list | first) "internal_ipv4" | default "" -}}
          {{- $internalIPv6 := index .hostvars (.groups.image_registry | default list | first) "internal_ipv6" | default "" -}}
          {{- if $internalIPv4 | empty | not -}}
      {{ $internalIPv4 }}
          {{- else if $internalIPv6 | empty | not -}}
      {{ $internalIPv6 }}
          {{- end -}}
        {{- end -}}
      {{- else -}}
        {{- if .zone | eq "cn" -}}
      hub.kubesphere.com.cn
        {{- end -}}
      {{- end -}}
    username: >-
      {{- if .image_registry.type | empty | not -}}
      admin
      {{- end }}
    password: >-
      {{- if .image_registry.type | empty | not -}}
      Harbor12345
      {{- end }}
    skip_tls_verify: >-
      {{- .image_registry.type | empty -}}
    ca_file: >-
      {{- if .groups.image_registry | default list | empty | not -}}
      {{ .binary_dir }}/pki/root.crt
      {{- end -}}
    cert_file: >-
      {{- if .groups.image_registry | default list | empty | not -}}
      {{ .binary_dir }}/pki/image-registry-client.crt
      {{- end -}}
    key_file: >-
      {{- if .groups.image_registry | default list | empty | not -}}
      {{ .binary_dir }}/pki/image-registry-client.key
      {{- end -}}
  # Registry endpoint for images from docker.io
  dockerio_registry: >-
    {{- .image_registry.auth.registry | empty | ternary "docker.io" .image_registry.auth.registry -}}

  # Registry endpoint for images from quay.io
  quayio_registry: >-
    {{- .image_registry.auth.registry | empty | ternary "quay.io" .image_registry.auth.registry -}}

  # Registry endpoint for images from ghcr.io
  ghcrio_registry: >-
    {{- .image_registry.auth.registry | empty | ternary "ghcr.io" .image_registry.auth.registry -}}

  # Registry endpoint for images from registry.k8s.io
  k8sio_registry: >-
    {{- .image_registry.auth.registry | empty | ternary "registry.k8s.io" .image_registry.auth.registry -}}
```

### Parameter Descriptions

| Parameter | Description |
|-----------|-------------|
| `image_registry.type` | Type of image registry to deploy: `harbor`, `docker-registry`, or `""` (use existing registry). |
| `image_registry.ha_vip` | Virtual IP used when deploying high-availability registries such as Harbor. |
| `image_registry.harbor.http_port` | Harbor HTTP service port. Defaults to `80`. |
| `image_registry.harbor.https_port` | Harbor HTTPS service port. Defaults to `443`. |
| `image_registry.auth.registry` | Actual image registry address used by the cluster. If a registry is deployed, it is automatically rendered based on `ha_vip` or node IP; empty in online mode; if zone is `cn`, defaults to `hub.kubesphere.com.cn`. |
| `image_registry.auth.username` | Username for logging into the image registry. Defaults to `admin` when deploying Harbor. |
| `image_registry.auth.password` | Password for logging into the image registry. Defaults to `Harbor12345` when deploying Harbor. |
| `image_registry.auth.skip_tls_verify` | Whether to skip TLS certificate verification. Defaults to `false` when deploying Harbor. |
| `image_registry.auth.ca_file` | Image registry CA certificate path. |
| `image_registry.auth.cert_file` | Client certificate path. |
| `image_registry.auth.key_file` | Client private key path. |
| `image_registry.dockerio_registry` | Image registry endpoint to replace `docker.io`. Defaults to `docker.io` if no private registry is configured. |
| `image_registry.quayio_registry` | Image registry endpoint to replace `quay.io`. |
| `image_registry.ghcrio_registry` | Image registry endpoint to replace `ghcr.io`. |
| `image_registry.k8sio_registry` | Image registry endpoint to replace `registry.k8s.io`. |

---

## Native Mode Configuration (02-native.yaml)

### Default Configuration

```yaml
# Basic operating system configuration settings
native:
  ntp:
    # List of NTP server addresses used for time synchronization
    servers:
      - "cn.pool.ntp.org"
    # Whether to enable the NTP service
    enabled: true
  # System timezone configuration
  timezone: Asia/Shanghai

  # NFS service configuration for nodes assigned the 'nfs' role in the inventory
  nfs:
    # Directories to be shared via NFS
    share_dir:
      - /share/
  # Whether to set the node's hostname to the value defined in inventory.hosts
  set_hostname: true
  # List of DNS configuration files to update on each node
  # This ensures that, during cluster installation, critical hostnames can be resolved locally even if no DNS service is available
  # For example:
  #   [control plane endpoint] -> master node IP
  #   [hostname of the node being installed] -> corresponding node IP
  localDNS:
    - /etc/hosts
```

### Parameter Descriptions

| Parameter | Description |
|-----------|-------------|
| `native.ntp.servers` | List of NTP server addresses used for time synchronization. |
| `native.ntp.enabled` | Whether to enable the NTP service to maintain consistent time across nodes. |
| `native.timezone` | System timezone of the node, e.g., `Asia/Shanghai`. |
| `native.nfs.share_dir` | NFS shared directories, used by nodes marked with the `nfs` role. |
| `native.set_hostname` | Whether to automatically set the node hostname according to the inventory definition during installation. |
| `native.localDNS` | List of local DNS resolution files (e.g., `/etc/hosts`), used to provide temporary domain name resolution during installation. |

---

## Kubernetes Configuration (03-kubernetes.yaml)

### Default Configuration

```yaml
kubernetes:
  # Name of the cluster to be installed
  cluster_name: kubekey

  # Image repository for built-in Kubernetes images
  image_repository: >-
    {{ .image_registry.k8sio_registry }}{{ if .image_registry.auth.registry | empty | not }}/kubernetes{{ end }}

  # Pause/Sandbox image configuration
  sandbox_image:
    # Pause image registry address
    registry: >-
      {{ .image_registry.k8sio_registry }}
    # Pause image repository path
    repository: >-
      {{- .image_registry.auth.registry | empty | ternary "pause" "kubernetes/pause" -}}

  # Kubernetes network configuration
  # kube-apiserver parameters
  apiserver:
    # kube-apiserver listening port
    port: 6443
    # List of additional SANs to be added to the apiserver certificate
    certSANs: []
    # Extra startup arguments for kube-apiserver
    extra_args:
      # Example: feature-gates: ExpandCSIVolumes=true,CSIStorageCapacity=true,RotateKubeletServerCertificate=true

  # kube-controller-manager parameters
  controller_manager:
    # Extra startup arguments for kube-controller-manager
    extra_args:
      # Cluster certificate signing validity period
      cluster-signing-duration: 87600h
      # Example: feature-gates: ExpandCSIVolumes=true,CSIStorageCapacity=true,RotateKubeletServerCertificate=true

  # kube-scheduler parameters
  scheduler:
    # Extra startup arguments for kube-scheduler
    extra_args:
      # Example: feature-gates: ExpandCSIVolumes=true,CSIStorageCapacity=true,RotateKubeletServerCertificate=true

  # kube-proxy parameters
  kube_proxy:
    # Whether to take over the deployment of kube-proxy
    manage:
      enabled: false
      # affinity:
      #   nodeAffinity:
      #     requiredDuringSchedulingIgnoredDuringExecution:
      #       nodeSelectorTerms:
      #         - matchExpressions:
      #             - key: kubernetes.io/os
      #               operator: In
      #               values:
      #                 - linux
    # kube-proxy proxy mode: ipvs, iptables
    mode: "ipvs"
    # kube-proxy configuration
    config:
      iptables:
        masqueradeAll: false
        masqueradeBit: 14
        minSyncPeriod: 0s
        syncPeriod: 30s

  # kubelet service parameters
  kubelet:
    # Maximum number of Pods per node
    max_pods: 110
    # PID limit per Pod
    pod_pids_limit: 10000
    # feature_gates:
    # Maximum size of container log files
    container_log_max_size: 5Mi
    # Number of container log files to retain
    container_log_max_files: 3
    # extra_args:

  # Specify a stable IP address or DNS name for the control plane
  # For high availability, it is recommended to set control_plane_endpoint to a DNS name
  # Configuration guidance:
  # 1. If a DNS name is available:
  #    - Set control_plane_endpoint to that DNS name and ensure it resolves to all control plane node IPs
  # 2. If no DNS name is available:
  #    - You can set a DNS name now and add the resolution later
  #    - Add the resolution to each node's local DNS file, for example:
  #      {{ vip }} {{ control_plane_endpoint }}
  #    - If you have a VIP (Virtual IP):
  #        Deploy kube-vip on control plane nodes to map the VIP to the actual node IPs
  #    - If you do not have a VIP:
  #        Deploy HAProxy on worker nodes, use a fixed IP (such as 127.0.0.2) as the VIP, and forward to all control plane node IPs
  #
  # For non-HA scenarios (manual configuration only, not automatically installed):
  # You can set the VIP to the IP of a single control plane node
  control_plane_endpoint:
    # Control plane endpoint hostname or IP
    host: lb.kubesphere.local
    # Control plane endpoint port, defaults to the apiserver port
    port: "{{ .kubernetes.apiserver.port }}"
    # Load balancing type: local, kube-vip, haproxy
    # When type is local, configure as follows:
    #   - On control-plane nodes: 127.0.0.1 {{ .kubernetes.control_plane_endpoint.host }}
    #   - On worker nodes: {{ .init_kubernetes_node }} {{ .kubernetes.control_plane_endpoint.host }}
    type: local
    local:
      # When using local load balancing, you can specify an external load balancer address here
      # Note: You must set up the actual load balancing yourself; this setting is only for DNS resolution
      address: ""
    kube_vip:
      # The IP address of the node's network interface (e.g., "eth0")
      address: ""
      # Supported modes: ARP, BGP
      mode: ARP
      image:
        # kube-vip image registry
        registry: >-
          {{ .image_registry.dockerio_registry }}
        # kube-vip image repository path
        repository: plndr/kube-vip
        # kube-vip image tag
        tag: v0.7.2
    haproxy:
      # The IP address on the node's "lo" (loopback) interface
      address: 127.0.0.1
      # HAProxy health check port
      health_port: 8081
      image:
        # HAProxy image registry
        registry: >-
          {{ .image_registry.dockerio_registry }}
        # HAProxy image repository path
        repository: library/haproxy
        # HAProxy image tag
        tag: 2.9.6-alpine

  # Whether to automatically renew Kubernetes certificates
  certs:
    # There are three ways to provide the Kubernetes CA (Certificate Authority) files:
    # 1. kubeadm: Leave ca_cert and ca_key empty, and kubeadm will generate them automatically. These certificates are valid for 10 years and will not change.
    # 2. kubekey: Set ca_cert to {{ .binary_dir }}/pki/ca.cert and ca_key to {{ .binary_dir }}/pki/ca.key.
    #    These certificates are generated by kubekey, valid for 10 years, and can be updated via `cert.ca_date`.
    # 3. Custom: Manually specify the absolute paths for ca_cert and ca_key to use your own CA files.
    #
    # To use custom CA files, fill in the absolute paths below.
    # If left empty, the default behavior (kubeadm or kubekey) will be used.
    ca_cert: ""
    ca_key: ""
    # The following fields are for the Kubernetes front-proxy CA certificate and key.
    # To use custom front-proxy CA files, fill in the absolute paths below.
    # If left empty, the default behavior will be used.
    front_proxy_cert: ""
    front_proxy_key: ""
    # Automatically renew service certificates (Note: CA certificates cannot be renewed automatically)
    renew: true

  # Customize Kubernetes components via patch files
  patches: []
    # Patches are applied via a directory containing patch files.
    # - name: kube-apiserver0+merge.yaml
    #   path: /etc/kubernetes/kube-apiserver-patch.yaml
    #   content: |
    #     apiVersion: v1
    #     kind: Pod
    #     spec:
    #       containers:
    #         - name: kube-apiserver
    #           command:
    #             - kube-apiserver
    #             - --service-account-issuer=https://kubernetes.default.svc.cluster.local
    #             - --service-account-jwks-uri=https://kubernetes.default.svc.cluster.local/openid/v1/jwks
    #   # The directory contains files named "target[suffix][+patchtype].extension".
    #   # "target" can be one of: kube-apiserver, kube-controller-manager, kube-scheduler,
    #   #                       etcd, kubeletconfiguration, corednsdeployment
    #   # "patchtype" can be: strategic (default), merge, json
    #   # "extension" can be: yaml or json
    #   # "suffix" (optional) determines apply order (alpha-numeric).
    #   # Examples:
    #   #   kube-apiserver+merge.yaml          # merge patch for kube-apiserver
    #   #   kube-apiserver001+strategic.yaml   # strategic patch with ordering suffix
    #   #   kube-controller-manager+merge.yaml
    #   #   kube-scheduler+json.yaml
    #   #   kubeletconfiguration+merge.yaml

  # Phases to skip during kubeadm init
  skip_phases: []
    # - addon/kube-proxy
```

### Parameter Descriptions

| Parameter | Description |
|-----------|-------------|
| `kubernetes.cluster_name` | Name of the Kubernetes cluster. |
| `kubernetes.image_repository` | Repository prefix for pulling Kubernetes core component images, automatically calculated from `k8sio_registry` by default. |
| `kubernetes.sandbox_image` | Complete configuration of the pause (sandbox) container image, including registry and repository. |
| `kubernetes.apiserver.port` | HTTPS listening port for kube-apiserver, default `6443`. |
| `kubernetes.apiserver.certSANs` | List of additional addresses to be added to the kube-apiserver certificate Subject Alternative Names. |
| `kubernetes.apiserver.extra_args` | Extra command-line arguments passed to kube-apiserver. |
| `kubernetes.controller_manager.extra_args` | Extra command-line arguments passed to kube-controller-manager. |
| `kubernetes.scheduler.extra_args` | Extra command-line arguments passed to kube-scheduler. |
| `kubernetes.kube_proxy.manage.enabled` | Whether KubeKey takes over the deployment of kube-proxy (instead of the default kubeadm deployment). |
| `kubernetes.kube_proxy.mode` | Working mode of kube-proxy, `ipvs` or `iptables`. |
| `kubernetes.kube_proxy.config.iptables` | Detailed configuration items in iptables mode. |
| `kubernetes.kubelet.max_pods` | Maximum number of Pods allowed to be scheduled on a single node. |
| `kubernetes.kubelet.pod_pids_limit` | Maximum number of PIDs that each Pod can use. |
| `kubernetes.kubelet.container_log_max_size` | Maximum size of a single container log file before rotation. |
| `kubernetes.kubelet.container_log_max_files` | Number of old container log files to retain. |
| `kubernetes.control_plane_endpoint.host` | Stable access address (IP or DNS) for the control plane. |
| `kubernetes.control_plane_endpoint.port` | Control plane endpoint port. |
| `kubernetes.control_plane_endpoint.type` | Load balancing implementation type: `local` (local resolution), `kube-vip` (VIP-based), `haproxy`. |
| `kubernetes.control_plane_endpoint.local.address` | When using `local` mode, an external load balancer address can be specified for resolution only. |
| `kubernetes.control_plane_endpoint.kube_vip.address` | Network interface name or IP bound by kube-vip. |
| `kubernetes.control_plane_endpoint.kube_vip.mode` | kube-vip working mode: `ARP` or `BGP`. |
| `kubernetes.control_plane_endpoint.kube_vip.image` | kube-vip container image configuration. |
| `kubernetes.control_plane_endpoint.haproxy.address` | Address that HAProxy listens on the local loopback interface. |
| `kubernetes.control_plane_endpoint.haproxy.health_port` | HAProxy health check port. |
| `kubernetes.control_plane_endpoint.haproxy.image` | HAProxy container image configuration. |
| `kubernetes.certs.ca_cert` | Custom Kubernetes CA certificate path (leave empty to use kubeadm/kubekey generated). |
| `kubernetes.certs.ca_key` | Custom Kubernetes CA private key path. |
| `kubernetes.certs.front_proxy_cert` | Custom front-proxy CA certificate path. |
| `kubernetes.certs.front_proxy_key` | Custom front-proxy CA private key path. |
| `kubernetes.certs.renew` | Whether to automatically renew service certificates in the cluster (CA itself will not be automatically renewed). |
| `kubernetes.patches` | Patch Kubernetes static Pods or component configurations via files or inline content. |
| `kubernetes.skip_phases` | List of phases to explicitly skip during `kubeadm init` execution. |

---

## CNI Network Plugin Configuration (04-cni.yaml)

### Default Configuration

```yaml
cni:
  # CNI plugin type to use
  # Specify the network plugin to install for the cluster. Supported: calico, cilium, flannel, hybridnet, kube-ovn, other
  type: calico
  # The complete Pod IP pool for the cluster. Supports IPv4, IPv6, and dual-stack
  pod_cidr: 10.233.64.0/18
  # IPv4 subnet mask length for pod allocation per node. Determines the number of Pod IPs each node can allocate
  ipv4_mask_size: 24
  # IPv6 subnet mask length for pod allocation per node
  ipv6_mask_size: 64
  # The complete Service IP pool for the cluster. Supports IPv4, IPv6, and dual-stack
  service_cidr: 10.233.0.0/18

  # Multi-CNI type configuration. Supported: multus, none
  multi_cni: "none"
  # Network enhancement plugin (Multus) providing multiple network interfaces for Pods
  multus:
    image:
      # Multus image registry
      registry: >-
        {{ .image_registry.ghcrio_registry }}
      # Multus image repository path
      repository: k8snetworkplumbingwg/multus-cni
      # tag: v4.3.0
```

### Parameter Descriptions

| Parameter | Description |
|-----------|-------------|
| `cni.type` | Cluster network plugin type, optional: `calico`, `cilium`, `flannel`, `hybridnet`, `kubeovn`, `other`. |
| `cni.pod_cidr` | CIDR segment for the entire cluster Pod network. |
| `cni.ipv4_mask_size` | IPv4 subnet mask length allocated to each node. For example, using `/24` mask in a `/18` network, each node can get about 256 Pod IPs. |
| `cni.ipv6_mask_size` | IPv6 subnet mask length allocated to each node. |
| `cni.service_cidr` | CIDR segment for the entire cluster Service network. |
| `cni.multi_cni` | Whether to enable multi-CNI support. `multus` means enable Multus, `none` means do not enable. |
| `cni.multus.image` | Multus CNI container image configuration (registry, repository, tag). |

---

## Container Runtime (CRI) Configuration (04-cri.yaml)

### Default Configuration

```yaml
cri:
  # Container runtime type. Supported: containerd, docker
  container_manager: containerd
  # Cgroup driver for the container runtime. Supported: systemd, cgroupfs
  cgroup_driver: systemd
    # tag: "3.9"
  # CRI socket endpoint for the selected container runtime
  cri_socket: >-
    {{- if .cri.container_manager | eq "containerd" -}}
    unix:///var/run/containerd/containerd.sock
    {{- else if and (.cri.container_manager | eq "docker") (.kubernetes.kube_version | semverCompare ">=v1.24.0") -}}
    unix:///var/run/cri-dockerd.sock
    {{- end -}}

  # Registry configuration for CRI, including image mirrors, insecure registries, and authentication
  registry:
    # Image mirror address list
    mirrors: ["https://registry-1.docker.io"]
    # List of insecure registries (allowing HTTP access)
    insecure_registries: []
    # List of private registry authentication information
    auths: []
    # Configuration example:
    # auths:
    #   - registry: docker.io
    #     username: MyDockerAccount
    #     password: my_password
    #     skip_tls_verify: true
    #     ca_cert: /etc/docker/certs.d/docker.io/ca.crt
    #     cert_file: /etc/docker/certs.d/docker.io/cert.crt
    #     key_file: /etc/docker/certs.d/docker.io/key.crt

  # Docker configuration
  docker:
    # Docker daemon configuration
    daemon:
      # Docker data root directory
      data-root: "{{ .cri.docker.data_root | default \"/var/lib/docker\" }}"
      # Container log configuration
      log-opts:
        # Maximum size of a single log file
        max-size: "{{ .kubernetes.kubelet.container_log_max_size | default \"5Mi\" | toLowerByteUnit }}"
        # Number of log files to retain
        max-file: "{{ .kubernetes.kubelet.container_log_max_files | default 3  | toString | toJson }}"
      # Enable live-restore
      live-restore: true
      # Container exec options
      exec-opts:
        - "native.cgroupdriver={{ .cri.cgroup_driver | default \"systemd\" }}"

  # containerd configuration
  containerd:
    config:
      # containerd data root directory
      root: "{{ .cri.containerd.data_root | default \"/var/lib/containerd\" }}"
      # containerd configuration file version
      version: 2
      # containerd runtime state directory
      state: "/run/containerd"
      grpc:
        address: "/run/containerd/containerd.sock"
        uid: 0
        gid: 0
        max_recv_message_size: 16777216
        max_send_message_size: 16777216
      ttrpc:
        address: ""
        uid: 0
        gid: 0
      debug:
        address: ""
        uid: 0
        gid: 0
        level: ""
      metrics:
        address: ""
        grpc_histogram: false
      cgroup:
        path: ""
      timeouts:
        "io.containerd.timeout.shim.cleanup": "5s"
        "io.containerd.timeout.shim.load": "5s"
        "io.containerd.timeout.shim.shutdown": "3s"
        "io.containerd.timeout.task.state": "2s"
      plugins:
        "io.containerd.grpc.v1.cri":
          containerd:
            runtimes:
              runc:
                runtime_type: "io.containerd.runc.v2"
                options:
                  SystemdCgroup: "{{ .cri.cgroup_driver | eq \"systemd\" }}"
          cni:
            bin_dir: "/opt/cni/bin"
            conf_dir: "/etc/cni/net.d"
            max_conf_num: 1
            conf_template: ""

```

### Parameter Descriptions

| Parameter | Description |
|-----------|-------------|
| `cri.container_manager` | Container runtime manager, optional: `containerd` or `docker`. |
| `cri.cgroup_driver` | Cgroup driver used by the container runtime, recommended `systemd` (compatible with most modern OS init systems). |
| `cri.cri_socket` | CRI socket path corresponding to the current container runtime, automatically selected based on `container_manager` and Kubernetes version. |
| `cri.registry.mirrors` | Image mirror addresses, can be configured with domestic mirror sources to improve pull speed. |
| `cri.registry.insecure_registries` | List of image registry addresses allowed to access using HTTP (non-HTTPS). |
| `cri.registry.auths` | List of authentication information for private image registries, including username, password, and optional TLS certificate configuration. |
| `cri.docker.daemon` | Docker daemon configuration items, mapped to `/etc/docker/daemon.json`. |
| `cri.docker.daemon.data-root` | Docker data root directory. |
| `cri.docker.daemon.log-opts.max-size` | Maximum size of a single container log file. |
| `cri.docker.daemon.log-opts.max-file` | Number of old container log files to retain. |
| `cri.docker.daemon.live-restore` | Whether to enable Docker live-restore. |
| `cri.docker.daemon.exec-opts` | Docker exec options list, e.g., cgroup driver. |
| `cri.containerd.config` | containerd configuration, mapped to `/etc/containerd/config.toml`. |
| `cri.containerd.config.root` | containerd data persistence root directory. |
| `cri.containerd.config.version` | containerd configuration file version. |
| `cri.containerd.config.state` | containerd runtime state directory. |
| `cri.containerd.config.grpc.address` | containerd gRPC socket address. |
| `cri.containerd.config.grpc.uid` | Owner UID of the containerd gRPC socket. |
| `cri.containerd.config.grpc.gid` | Owner GID of the containerd gRPC socket. |
| `cri.containerd.config.grpc.max_recv_message_size` | containerd gRPC maximum received message size. |
| `cri.containerd.config.grpc.max_send_message_size` | containerd gRPC maximum sent message size. |
| `cri.containerd.config.ttrpc` | containerd TTRPC configuration. |
| `cri.containerd.config.debug` | containerd debug configuration. |
| `cri.containerd.config.metrics` | containerd metrics configuration. |
| `cri.containerd.config.cgroup` | containerd cgroup configuration. |
| `cri.containerd.config.timeouts` | Timeout settings for containerd operations. |
| `cri.containerd.config.plugins` | containerd CRI plugin configuration, including runtime and CNI settings. |
| `cri.containerd.config.plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc.runtime_type` | runc runtime type. |
| `cri.containerd.config.plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc.options.SystemdCgroup` | Whether runc uses systemd cgroup. |
| `cri.containerd.config.plugins."io.containerd.grpc.v1.cri".cni` | CNI plugin configuration, including binary and configuration directories. |

---

## etcd Configuration (04-etcd.yaml)

### Default Configuration

```yaml
# etcd service configuration
etcd:
  # etcd deployment type:
  # - external: Use an external etcd cluster
  # - internal: Deploy etcd as static Pods within the cluster
  deployment_type: external
  image:
    # etcd image registry
    registry: >-
      {{ .image_registry.dockerio_registry }}
    # etcd image repository path
    repository: kubesphere/etcd
    # etcd image tag
    tag: "{{ .etcd.etcd_version }}"
  # etcd client port
  port: 2379
  # etcd peer-to-peer communication port
  peer_port: 2380
  # etcd service environment variables
  env:
    # Leader election timeout (milliseconds)
    election_timeout: 5000
    # Heartbeat interval (milliseconds)
    heartbeat_interval: 250
    # Data compaction retention duration (hours)
    compaction_retention: 8
    # Number of transactions to trigger a snapshot
    snapshot_count: 10000
    # etcd data directory
    data_dir: /var/lib/etcd
    # etcd cluster token
    token: k8s_etcd
    # metrics: basic
    # quota_backend_bytes: 100
    # max_request_bytes: 100
    # max_snapshots: 100
    # max_wals: 5
    # log_level: info
    # unsupported_arch: arm64
  # etcd backup configuration
  backup:
    # Directory for storing backup files
    backup_dir: /var/lib/etcd-backup
    # Number of backup files to retain
    keep_backup_number: 5
    # Backup script name
    etcd_backup_script: "backup.sh"
    # Scheduled backup cycle (systemd OnCalendar format)
    on_calendar: "*-*-* *:00/30:00"
  # Enable etcd performance tuning
  performance: false
  # Enable etcd traffic priority control
  traffic_priority: false
  # CA certificate path
  ca_file: >-
    {{ .binary_dir }}/pki/root.crt
  # Server certificate path
  server_cert_file: >-
    {{ .binary_dir }}/pki/etcd-{{ "{{ . }}" }}.crt
  # Server private key path
  server_key_file: >-
    {{ .binary_dir }}/pki/etcd-{{ "{{ . }}" }}.key
  # Client certificate path
  client_cert_file: >-
    {{ .binary_dir }}/pki/etcd-client.crt
  # Client private key path
  client_key_file: >-
    {{ .binary_dir }}/pki/etcd-client.key
```

### Parameter Descriptions

| Parameter | Description |
|-----------|-------------|
| `etcd.deployment_type` | etcd deployment mode. `external` uses an existing external etcd; `internal` deploys etcd as static Pods within the Kubernetes cluster. |
| `etcd.image` | etcd container image configuration (registry, repository, tag). |
| `etcd.port` | etcd client request port, default `2379`. |
| `etcd.peer_port` | etcd peer-to-peer communication port, default `2380`. |
| `etcd.env.election_timeout` | Leader election timeout in milliseconds. |
| `etcd.env.heartbeat_interval` | Heartbeat interval between nodes in milliseconds. |
| `etcd.env.compaction_retention` | Duration of data history retained by automatic data compaction in hours. |
| `etcd.env.snapshot_count` | Number of transactions required to trigger a snapshot. |
| `etcd.env.data_dir` | etcd data persistence directory. |
| `etcd.env.token` | Shared token for cluster initialization, used for member discovery. |
| `etcd.backup.backup_dir` | Directory for etcd backup files. |
| `etcd.backup.keep_backup_number` | Number of backup copies retained locally. |
| `etcd.backup.etcd_backup_script` | Name of the backup script executed. |
| `etcd.backup.on_calendar` | Scheduled backup cycle format based on systemd timer, e.g., every 30 minutes. |
| `etcd.performance` | Whether to enable etcd performance tuning parameters. |
| `etcd.traffic_priority` | Whether to enable etcd network traffic priority control. |
| `etcd.ca_file` | etcd CA certificate file path. |
| `etcd.server_cert_file` | etcd server certificate path. |
| `etcd.server_key_file` | etcd server private key path. |
| `etcd.client_cert_file` | etcd client certificate path. |
| `etcd.client_key_file` | etcd client private key path. |

---

## DNS Configuration (05-dns.yaml)

### Default Configuration

```yaml
dns:
  # ====== In-Cluster DNS Service Configuration ======
  # DNS domain suffix used for services and pods within the cluster
  domain: cluster.local

  # NodeLocalDNS pod configuration
  nodelocaldns:
    # Whether to enable NodeLocalDNS
    enabled: true
    # IP address that NodeLocalDNS binds to on each node
    ip: 169.254.25.10
    # NodeLocalDNS image configuration
    image:
      # NodeLocalDNS image registry
      registry: >-
        {{ .image_registry.k8sio_registry }}
      # NodeLocalDNS image repository path
      repository: >-
        dns/k8s-dns-node-cache
      # tag: 1.24.0

  # CoreDNS pod configuration
  coredns:
    # IP address of the cluster DNS service
    ip: >-
      {{ index (.cni.service_cidr | ipInCIDR) 2 }}
    # CoreDNS image configuration
    image:
      # CoreDNS image registry
      registry: >-
        {{ .image_registry.k8sio_registry }}
      # CoreDNS image repository path
      repository: >-
        coredns
      # tag: v1.11.1
    # Custom hosts entries
    dns_etc_hosts: []
    # DNS zone matching configuration
    zone_configs:
      # Each entry defines which DNS zones to match. The default port is 53
      # ".": matches all DNS zones
      # "example.com": matches *.example.com using DNS server on port 53
      # "example.com:54": matches *.example.com using DNS server on port 54
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
          zones:
            - "{{ .dns.domain }}"
        # If internal DNS message rewriting is needed, configure it here
        # rewrite:
        #   - rule: continue
        #     field: name
        #     type: exact
        #     value: "example.com example2.com"
        #     options: ""
        forward:
          # DNS query forwarding rules
          - from: "."
            # Forwarding destination endpoints. The 'to' syntax allows protocol specification
            to: ["/etc/resolv.conf"]
            # Domains to exclude from forwarding
            except: []
            # Use TCP for forwarding, even if the original request was UDP
            force_tcp: false
            # Prefer UDP for forwarding; fallback to TCP if the response is truncated
            prefer_udp: false
            # Maximum consecutive failed health checks before marking upstream as unavailable
            # max_fails: 2
            # Cached connection expiration time
            # expire: 10s
            # TLS properties for secure connections can be set here
            # tls:
            #   cert_file: ""
            #   key_file: ""
            #   ca_file: ""
            # tls_servername: ""
            # Policy for selecting upstream servers: random (default), round_robin, sequential
            # policy: "random"
            # Upstream server health check configuration
            # health_check: ""
            # Maximum number of concurrent DNS queries allowed
            max_concurrent: 1000
```

### Parameter Descriptions

| Parameter | Description |
|-----------|-------------|
| `dns.domain` | Default DNS domain suffix for the cluster (e.g., `cluster.local`). |
| `dns.nodelocaldns.enabled` | Whether to enable NodeLocalDNS to improve cluster DNS resolution performance and reduce CoreDNS load. |
| `dns.nodelocaldns.ip` | Link-local IP that NodeLocalDNS DaemonSet binds on each node, default `169.254.25.10`. |
| `dns.nodelocaldns.image` | NodeLocalDNS container image configuration. |
| `dns.coredns.ip` | CoreDNS cluster service IP, usually the 3rd address in the Service CIDR. |
| `dns.coredns.image` | CoreDNS container image configuration. |
| `dns.coredns.dns_etc_hosts` | Custom `/etc/hosts` format entries injected into CoreDNS. |
| `dns.coredns.zone_configs` | List of CoreDNS Corefile zone configurations, can define matching domains, cache, rewrite, forwarding, and other rules. |
| `dns.coredns.zone_configs[].zones` | List of DNS domains and ports matched by this zone rule. |
| `dns.coredns.zone_configs[].additional_configs` | List of additional CoreDNS plugin directives (e.g., `errors`, `ready`, `prometheus`, `loop`, `reload`, `loadbalance`). |
| `dns.coredns.zone_configs[].cache` | DNS record cache time (seconds). |
| `dns.coredns.zone_configs[].kubernetes.zones` | Cluster DNS domains resolved by the CoreDNS Kubernetes plugin. |
| `dns.coredns.zone_configs[].forward` | List of forwarding rules for queries that cannot be resolved locally. |
| `dns.coredns.zone_configs[].forward[].from` | Source domain that needs forwarding resolution. |
| `dns.coredns.zone_configs[].forward[].to` | List of upstream DNS server or resolution file addresses. |
| `dns.coredns.zone_configs[].forward[].except` | List of exception domains that are not forwarded upstream. |
| `dns.coredns.zone_configs[].forward[].force_tcp` | Whether to force using TCP to forward queries upstream. |
| `dns.coredns.zone_configs[].forward[].prefer_udp` | Whether to prefer using UDP to forward queries upstream. |
| `dns.coredns.zone_configs[].forward[].max_concurrent` | Maximum number of concurrent queries allowed for this forwarding rule. |

---

## Storage Class Configuration (05-storage_class.yaml)

### Default Configuration

```yaml
# Storage class configuration for Kubernetes persistent storage integration
storage_class:
  # Local storage class configuration
  local:
    enabled: true    # Whether to enable local storage class
    default: true    # Whether to set as default storage class
    path: /var/openebs/local  # Host path for local storage volumes

  # NFS storage class configuration
  nfs:
    # Ensure nfs-utils is installed on every node in the k8s_cluster group
    enabled: false   # Whether to enable NFS storage class
    default: false   # Whether to set as default storage class
    # NFS server address
    server: >-
      {{ .groups.nfs | default list | first }}
    path: /share/kubernetes  # NFS export path for persistent volumes
```

### Parameter Descriptions

| Parameter | Description |
|-----------|-------------|
| `storage_class.local.enabled` | Whether to create and enable a `local` StorageClass based on node local disks. |
| `storage_class.local.default` | Whether to mark the `local` StorageClass as the cluster default storage class. |
| `storage_class.local.path` | Actual host path on the node for local storage volumes. |
| `storage_class.nfs.enabled` | Whether to create and enable an NFS-based StorageClass. |
| `storage_class.nfs.default` | Whether to mark the NFS StorageClass as the cluster default storage class. |
| `storage_class.nfs.server` | NFS server address, defaults to the first node in the `nfs` group in the inventory. |
| `storage_class.nfs.path` | Exported shared directory path on the NFS server. |

---

## Download Configuration (10-download.yaml)

### Default Configuration

```yaml
download:
  # Download timeout
  timeout: 300s
  # Default file storage host for China region
  cn_host: kubekey.pek3b.qingstor.com
  # Target operating system
  os: linux
  # Target CPU architecture list
  arch: [ "amd64" ]
  # KubeKey offline artifact package file path
  artifact_file: ""
  # MD5 checksum file of the artifact package
  artifact_md5: ""
  # Whether to download software packages, Helm charts, container images, etc. online
  # Set to false if all required images and packages are already available locally and no remote validation is needed
  fetch: true
  # Download URL templates for each component
  artifact_url:
    # etcd binary package
    etcd: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      github.com/etcd-io/etcd/releases/download/{{ "{{ .version }}" }}/etcd-{{ "{{ .version }}" }}-linux-{{ "{{ .arch }}" }}.tar.gz
    # kubelet binary
    kubelet: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      dl.k8s.io/release/{{ "{{ .version }}" }}/bin/linux/{{ "{{ .arch }}" }}/kubelet
    # kubeadm binary
    kubeadm: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      dl.k8s.io/release/{{ "{{ .version }}" }}/bin/linux/{{ "{{ .arch }}" }}/kubeadm
    # kubectl binary
    kubectl: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      dl.k8s.io/release/{{ "{{ .version }}" }}/bin/linux/{{ "{{ .arch }}" }}/kubectl
    # CNI plugins package
    cni_plugins: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      github.com/containernetworking/plugins/releases/download/{{ "{{ .version }}" }}/cni-plugins-linux-{{ "{{ .arch }}" }}-{{ "{{ .version }}" }}.tgz
    # Helm binary package
    helm: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      get.helm.sh/helm-{{ "{{ .version }}" }}-linux-{{ "{{ .arch }}" }}.tar.gz
    # crictl tool package
    crictl: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      github.com/kubernetes-sigs/cri-tools/releases/download/{{ "{{ .version }}" }}/crictl-{{ "{{ .version }}" }}-linux-{{ "{{ .arch }}" }}.tar.gz
    # Docker binary package
    docker: >-
      https://mirrors.aliyun.com/docker-ce/linux/static/stable/
      {{- "{{ if eq .arch \"amd64\" }}x86_64{{ else if eq .arch \"arm64\" }}aarch64{{ end }}" -}}
      /docker-{{ "{{ .version }}" }}.tgz
    # cri-dockerd package
    cridockerd: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      github.com/Mirantis/cri-dockerd/releases/download/{{ "{{ .version }}" }}/cri-dockerd-{{ "{{ .version | default \"\" | trimPrefix \"v\" }}" }}.{{ "{{ .arch }}" }}.tgz
    # containerd binary package
    containerd: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      github.com/containerd/containerd/releases/download/{{ "{{ .version }}" }}/containerd-{{ "{{ .version | default \"\" | trimPrefix \"v\" }}" }}-linux-{{ "{{ .arch }}" }}.tar.gz
    # runc binary
    runc: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      github.com/opencontainers/runc/releases/download/{{ "{{ .version }}" }}/runc.{{ "{{ .arch }}" }}
    # calicoctl binary
    calicoctl: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      github.com/projectcalico/calico/releases/download/{{ "{{ .version }}" }}/calicoctl-linux-{{ "{{ .arch }}" }}
    # docker-registry package
    docker_registry: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      docker.io/registry/{{ "{{ .version }}" }}/docker-registry-{{ "{{ .version }}" }}-linux-{{ "{{ .arch }}" }}.tgz
    # docker-compose binary
    docker_compose: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      github.com/docker/compose/releases/download/{{ "{{ .version }}" }}/docker-compose-linux-
      {{- "{{ if eq .arch \"amd64\" }}x86_64{{ else if eq .arch \"arm64\" }}aarch64{{ end }}" -}}
    # Harbor offline installer package
    harbor: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      github.com/
      {{- "{{ if eq .arch \"amd64\" }}goharbor/harbor{{ else if eq .arch \"arm64\" }}kubesphere/kubekey{{ end }}" -}}
      /releases/download/
      {{- "{{ if eq .arch \"amd64\" }}{{ .version }}{{ else if eq .arch \"arm64\" }}iso-latest{{ end }}" -}}
      /harbor-offline-installer-{{ "{{ .version }}" }}.tgz
    # keepalived package
    keepalived: >-
      https://{{ .download.cn_host}}/osixia/keepalived/{{ "{{ .version }}" }}/keepalived-{{ "{{ .version }}" }}-linux-{{ "{{ .arch }}" }}.tgz
    # Helm Chart package: Calico
    calico: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      github.com/projectcalico/calico/releases/download/{{ "{{ .version }}" }}/tigera-operator-{{ "{{ .version }}" }}.tgz
    # Helm Chart package: Cilium
    cilium: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      helm.cilium.io/cilium-{{ "{{ .version }}" }}.tgz
    # Helm Chart package: Flannel
    flannel: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      github.com/flannel-io/flannel/releases/download/{{ "{{ .version }}" }}/flannel.tgz
    # Helm Chart package: Kube-OVN
    kubeovn: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      kubeovn.github.io/kube-ovn/kube-ovn-{{ "{{ .version }}" }}.tgz
    # Helm Chart package: Hybridnet
    hybridnet: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      github.com/alibaba/hybridnet/releases/download/helm-chart-{{ "{{ .version }}" }}/hybridnet-{{ "{{ .version }}" }}.tgz
    # Helm Chart package: OpenEBS LocalPV Provisioner
    localpv_provisioner: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      openebs.github.io/dynamic-localpv-provisioner/localpv-provisioner-{{ "{{ .version }}" }}.tgz
    # Helm Chart package: NFS Subdir External Provisioner
    nfs_subdir_external_provisioner: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      github.com/kubernetes-sigs/nfs-subdir-external-provisioner/releases/download/nfs-subdir-external-provisioner-{{ "{{ .version }}" }}/nfs-subdir-external-provisioner-{{ "{{ .version }}" }}.tgz
    # Helm Chart package: Spiderpool
    spiderpool: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      github.com/spidernet-io/spiderpool/releases/download/{{ "{{ .version }}" }}/spiderpool-{{ "{{ .version | default \"\" | trimPrefix \"v\" }}" }}.tgz
  # Additional tools to package
  tools:
    # ORAS tool package
    oras: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      github.com/oras-project/oras/releases/download/v1.3.0/oras_1.3.0_linux_{{ "{{ \"{{ .arch }}\" }}" }}.tar.gz
    # nerdctl tool package
    nerdctl: >-
      {{- .zone | eq "cn" | ternary (tpl "https://{{ .download.cn_host}}/" .) "https://" -}}
      github.com/containerd/nerdctl/releases/download/v2.2.1/nerdctl-2.2.1-linux-{{ "{{ \"{{ .arch }}\" }}" }}.tar.gz
  # Additional Helm Chart list
  charts: []
  # charts:
  #   # Repository Chart
  #   - url: flannel@https://flannel-io.github.io/flannel/
  #     version: 0.28.1
  #   # OCI Chart
  #   - url: oci://ghcr.io/flannel-io/flannel-chart
  #     version: 0.28.1
  # Operating system ISO/package list
  iso:
    - "almalinux-9.0-rpms"
    - "centos-8-rpms"
    - "debian-10-debs"
    - "debian-11-debs"
    - "kylin-v10SP3-2403-rpms"
    - "kylin-v10SP3-rpms"
    - "kylin-v10SP2-rpms"
    - "kylin-v10SP1-rpms"
    - "ubuntu-18.04-debs"
    - "ubuntu-20.04-debs"
    - "ubuntu-22.04-debs"
    - "ubuntu-24.04-debs"
  # CNI type list (for download)
  cni:
    type: []
  # Storage class toggle in download configuration
  storage_class:
    local:
      enabled: true
    nfs:
      enabled: false
  # Container runtime list dynamically determined by K8s version (for download)
  cri:
    container_manager: >
      {{- $container_manager := list }}
      {{- range .download.kubernetes.kube_version }}
        {{- if and (. | semverCompare "<v1.24.0") ($container_manager | has "docker" | not) }}
          {{- $container_manager = append $container_manager "docker" }}
        {{- else if and (. | semverCompare ">=v1.24.0") ($container_manager | has "containerd" | not) }}
          {{- $container_manager = append $container_manager "containerd" }}
        {{- end }}
      {{- end }}
      {{- $container_manager | toJson }}

  # Container image list required by each component
  images:
    manifests: []
    # Default registry address used when downloading images
    registry: >-
      {{- if .zone | eq "cn" }}
      hub.kubesphere.com.cn
      {{- end }}
    # Image pull policy, supports strict, warn
    policy: "strict"
    # Kubernetes related image list (organized by Helm Chart and version)
    openebs-localpv/localpv-provisioner:
      "4.4.0":
        - docker.io/openebs/linux-utils:4.3.0
        - docker.io/openebs/provisioner-localpv:4.4.0
    nfs-subdir-external-provisioner/nfs-subdir-external-provisioner:
      "4.0.18":
        - registry.k8s.io/sig-storage/nfs-subdir-external-provisioner:v4.0.2
    projectcalico/tigera-operator:
      v3.25.2:
        - quay.io/tigera/operator:v1.29.6
        - docker.io/calico/apiserver:v3.25.2
        - docker.io/calico/cni:v3.25.2
        - docker.io/calico/csi:v3.25.2
        - docker.io/calico/ctl:v3.25.2
        - docker.io/calico/kube-controllers:v3.25.2
        - docker.io/calico/node-driver-registrar:v3.25.2
        - docker.io/calico/typha:v3.25.2
        - docker.io/calico/node:v3.25.2
        - docker.io/calico/pod2daemon-flexvol:v3.25.2
        - docker.io/flannel/flannel:v0.24.4
      v3.26.5:
        - quay.io/tigera/operator:v1.30.11
        - docker.io/calico/ctl:v3.26.5
        - docker.io/calico/typha:v3.26.5
        - quay.io/calico/node:v3.26.5
        - docker.io/calico/cni:v3.26.5
        - docker.io/calico/csi:v3.26.5
        - docker.io/calico/apiserver:v3.26.5
        - docker.io/calico/kube-controllers:v3.26.5
        - docker.io/calico/flannel-migration-controller:v3.26.5
        - docker.io/flannel/flannel:v0.24.4
        - docker.io/calico/dikastes:v3.26.5
        - docker.io/calico/node-driver-registrar:v3.26.5
        - quay.io/calico/pod2daemon-flexvol:v3.26.5
      v3.28.5:
        - quay.io/tigera/operator:v1.34.13
        - docker.io/calico/ctl:v3.28.5
        - docker.io/calico/typha:v3.28.5
        - quay.io/calico/node:v3.28.5
        # - docker.io/calico/node-windows:v3.28.5
        - docker.io/calico/cni:v3.28.5
        # - docker.io/calico/cni-windows:v3.28.5
        - docker.io/calico/csi:v3.28.5
        - docker.io/calico/apiserver:v3.28.5
        - docker.io/calico/kube-controllers:v3.28.5
        - docker.io/calico/flannel-migration-controller:v3.28.5
        - docker.io/flannel/flannel:v0.24.4
        - docker.io/calico/dikastes:v3.28.5
        - docker.io/calico/node-driver-registrar:v3.28.5
        - quay.io/calico/pod2daemon-flexvol:v3.28.5
      v3.29.7:
        - quay.io/tigera/operator:v1.36.16
        - docker.io/calico/ctl:v3.29.7
        - docker.io/calico/typha:v3.29.7
        - quay.io/calico/node:v3.29.7
        # - docker.io/calico/node-windows:v3.29.7
        - docker.io/calico/cni:v3.29.7
        # - docker.io/calico/cni-windows:v3.29.7
        - docker.io/calico/csi:v3.29.7
        - docker.io/calico/apiserver:v3.29.7
        - docker.io/calico/kube-controllers:v3.29.7
        - docker.io/calico/flannel-migration-controller:v3.29.7
        # - docker.io/calico/windows:v3.29.7
        - docker.io/flannel/flannel:v0.24.4
        - docker.io/calico/dikastes:v3.29.7
        - docker.io/calico/node-driver-registrar:v3.29.7
        - quay.io/calico/pod2daemon-flexvol:v3.29.7
      v3.30.5:
        - quay.io/tigera/operator:v1.38.9
        - docker.io/calico/ctl:v3.30.5
        - docker.io/calico/typha:v3.30.5
        - quay.io/calico/node:v3.30.5
        # - docker.io/calico/node-windows:v3.30.5
        - docker.io/calico/cni:v3.30.5
        # - docker.io/calico/cni-windows:v3.30.5
        - docker.io/calico/csi:v3.30.5
        - docker.io/calico/apiserver:v3.30.5
        - docker.io/calico/kube-controllers:v3.30.5
        - docker.io/calico/envoy-gateway:v3.30.5
        - docker.io/calico/envoy-proxy:v3.30.5
        - docker.io/calico/envoy-ratelimit:v3.30.5
        - docker.io/calico/flannel-migration-controller:v3.30.5
        - docker.io/flannel/flannel:v0.24.4
        - docker.io/calico/dikastes:v3.30.5
        - docker.io/calico/node-driver-registrar:v3.30.5
        - quay.io/calico/pod2daemon-flexvol:v3.30.5
        - docker.io/calico/csi:v3.30.5
        - docker.io/calico/key-cert-provisioner:v3.30.5
        - docker.io/calico/goldmane:v3.30.5
        - docker.io/calico/whisker:v3.30.5
        - docker.io/calico/whisker-backend:v3.30.5
      v3.31.3:
        - quay.io/tigera/operator:v1.40.3
        - quay.io/calico/ctl:v3.31.3
        - docker.io/calico/typha:v3.31.3
        - quay.io/calico/node:v3.31.3
        # - docker.io/calico/node-windows:v3.31.3
        - docker.io/calico/cni:v3.31.3
        # - docker.io/calico/cni-windows:v3.31.3
        - docker.io/calico/csi:v3.31.3
        - docker.io/calico/apiserver:v3.31.3
        - docker.io/calico/kube-controllers:v3.31.3
        - docker.io/calico/envoy-gateway:v3.31.3
        - docker.io/calico/envoy-proxy:v3.31.3
        - docker.io/calico/envoy-ratelimit:v3.31.3
        - docker.io/calico/flannel-migration-controller:v3.31.3
        - docker.io/flannel/flannel:v0.24.4
        - docker.io/calico/dikastes:v3.31.3
        - docker.io/calico/node-driver-registrar:v3.31.3
        - quay.io/calico/pod2daemon-flexvol:v3.31.3
        - docker.io/calico/csi:v3.31.3
        - docker.io/calico/key-cert-provisioner:v3.31.3
        - docker.io/calico/goldmane:v3.31.3
        - docker.io/calico/whisker:v3.31.3
        - docker.io/calico/whisker-backend:v3.31.3
    cilium/cilium:
      "1.14.19":
        - quay.io/cilium/cilium:v1.14.19
        - quay.io/cilium/certgen:v0.1.16
        - quay.io/cilium/hubble-relay:v1.14.19
        - quay.io/cilium/hubble-ui-backend:v0.13.1
        - quay.io/cilium/hubble-ui:v0.13.1
        - quay.io/cilium/cilium-envoy:v1.30.9-1734953328-6db0e437ba7ed2169f032ceec25922dd06e0b12b
        # - quay.io/cilium/cilium-etcd-operator:v2.0.7
        - quay.io/cilium/operator:v1.14.19
        - quay.io/cilium/startup-script:c54c7edeab7fde4da68e59acd319ab24af242c3f
        - quay.io/cilium/clustermesh-apiserver:v1.14.19
        - quay.io/coreos/etcd:v3.5.4
        - quay.io/cilium/kvstoremesh:v1.14.19
        - ghcr.io/spiffe/spire-agent:1.6.3
        - ghcr.io/spiffe/spire-server:1.6.3
      "1.15.19":
        - quay.io/cilium/cilium:v1.15.19
        - quay.io/cilium/certgen:v0.1.19
        - quay.io/cilium/hubble-relay:v1.15.19
        - quay.io/cilium/hubble-ui-backend:v0.13.2
        - quay.io/cilium/hubble-ui:v0.13.2
        - quay.io/cilium/cilium-envoy:v1.33.4-1752151664-7c2edb0b44cf95f326d628b837fcdd845102ba68
        # - quay.io/cilium/cilium-etcd-operator:v2.0.7
        - quay.io/cilium/operator:v1.15.19
        - quay.io/cilium/startup-script:c54c7edeab7fde4da68e59acd319ab24af242c3f
        - quay.io/cilium/clustermesh-apiserver:v1.15.19
        - docker.io/library/busybox:1.36.1
        - ghcr.io/spiffe/spire-agent:1.8.5
        - ghcr.io/spiffe/spire-server:1.8.5
      "1.16.19":
        - quay.io/cilium/cilium:v1.16.19
        - quay.io/cilium/certgen:v0.3.1
        - quay.io/cilium/hubble-relay:v1.16.19
        - quay.io/cilium/hubble-ui-backend:v0.13.3
        - quay.io/cilium/hubble-ui:v0.13.3
        - quay.io/cilium/cilium-envoy:v1.34.12-1767177245-7935d4d711cb6f8020385a50c996b90896e16a71
        - quay.io/cilium/operator:v1.16.19
        - quay.io/cilium/startup-script:1755531540-60ee83e
        - quay.io/cilium/clustermesh-apiserver:v1.16.19
        - docker.io/library/busybox:1.36.1
        - ghcr.io/spiffe/spire-agent:1.9.6
        - ghcr.io/spiffe/spire-server:1.9.6
      "1.17.15":
        - quay.io/cilium/cilium:v1.17.15
        - quay.io/cilium/certgen:v0.4.1
        - quay.io/cilium/hubble-relay:v1.17.15
        - quay.io/cilium/hubble-ui-backend:v0.13.3
        - quay.io/cilium/hubble-ui:v0.13.3
        - quay.io/cilium/cilium-envoy:v1.36.6-1776000132-2437d2edeaf4d9b56ef279bd0d71127440c067aa
        - quay.io/cilium/operator:v1.17.15
        - quay.io/cilium/startup-script:1755531540-60ee83e
        - quay.io/cilium/clustermesh-apiserver:v1.17.15
        - docker.io/library/busybox:1.37.0
        - ghcr.io/spiffe/spire-agent:1.9.6
        - ghcr.io/spiffe/spire-server:1.9.6
      "1.18.9":
        - quay.io/cilium/cilium:v1.18.9
        - quay.io/cilium/certgen:v0.4.1
        - quay.io/cilium/hubble-relay:v1.18.9
        - quay.io/cilium/hubble-ui-backend:v0.13.3
        - quay.io/cilium/hubble-ui:v0.13.3
        - quay.io/cilium/cilium-envoy:v1.36.6-1776000132-2437d2edeaf4d9b56ef279bd0d71127440c067aa
        - quay.io/cilium/operator:v1.18.9
        - quay.io/cilium/startup-script:1755531540-60ee83e
        - quay.io/cilium/clustermesh-apiserver:v1.18.9
        - docker.io/library/busybox:1.37.0
        - ghcr.io/spiffe/spire-agent:1.12.4
        - ghcr.io/spiffe/spire-server:1.12.4
      "1.19.3":
        - quay.io/cilium/cilium:v1.19.3
        - docker.io/istio/ztunnel:1.28.0-distroless
        - quay.io/cilium/certgen:v0.4.1
        - quay.io/cilium/hubble-relay:v1.19.3
        - quay.io/cilium/hubble-ui-backend:v0.13.3
        - quay.io/cilium/hubble-ui:v0.13.3
        - quay.io/cilium/cilium-envoy:v1.36.6-1776000132-2437d2edeaf4d9b56ef279bd0d71127440c067aa
        - quay.io/cilium/operator:v1.19.3
        - quay.io/cilium/startup-script:1763560095-8f36c34
        - quay.io/cilium/clustermesh-apiserver:v1.19.3
        - docker.io/library/busybox:1.37.0
        - ghcr.io/spiffe/spire-agent:1.9.6
        - ghcr.io/spiffe/spire-server:1.9.6
    flannel/flannel:
      v0.27.4:
        - ghcr.io/flannel-io/flannel-cni-plugin:v1.8.0-flannel1
        - ghcr.io/flannel-io/flannel:v0.27.4
    hybridnet/hybridnet:
      0.6.8:
        - docker.io/hybridnetdev/hybridnet:v0.8.8
    kubeovn/kube-ovn:
      v1.13.15:
        - docker.io/kubeovn/kube-ovn:v1.13.15
        - docker.io/kubeovn/vpc-nat-gateway:v1.13.15
      v1.15.0:
        - docker.io/kubeovn/kube-ovn:v1.15.0
        - docker.io/kubeovn/vpc-nat-gateway:v1.15.0
    spiderpool/spiderpool:
      v1.1.1:
        - ghcr.io/spidernet-io/spiderpool/spiderpool-plugins:27c4f118b1cec3773f2679b772e7583fc77e5686
        - ghcr.io/k8snetworkplumbingwg/multus-cni:v4.1.4
        - ghcr.io/spidernet-io/spiderpool/spiderpool-agent:v1.1.1
        - ghcr.io/spidernet-io/spiderpool/spiderpool-controller:v1.1.1
```

### Parameter Descriptions

| Parameter | Description |
|-----------|-------------|
| `download.timeout` | Timeout for downloading binaries, images, and other resources. |
| `download.cn_host` | Default download acceleration domain when `zone` is set to `cn`. |
| `download.os` | Target operating system for downloaded resources, default `linux`. |
| `download.arch` | List of target CPU architectures for downloaded resources, default `["amd64"]`. |
| `download.artifact_file` | Local path to the offline artifact package, used for offline installation. |
| `download.artifact_md5` | Path to the MD5 checksum file corresponding to the offline artifact package. |
| `download.fetch` | Whether to perform online downloads. If all resources are already prepared locally, can be set to `false`. |
| `download.artifact_url` | Download URL templates for each component binary and Helm Chart, supporting automatic switching to domestic sources based on `zone`. |
| `download.tools` | Additional tools that need to be downloaded and packaged, such as `oras`, `nerdctl`. |
| `download.charts` | List of additional Helm Charts to pull beyond the default components (supports repository or OCI format). |
| `download.iso` | List of operating system RPM/DEB packages to include when creating offline packages. |
| `download.cni.type` | Which CNI plugin types need to be prepared during the download phase. |
| `download.storage_class` | Toggle for pre-packaging storage class related packages/images during the download phase. |
| `download.cri.container_manager` | Dynamically calculated list of container runtimes to download based on the target Kubernetes version. |
| `download.images` | Defines the list of container images required by each component (organized by Helm Chart). |
| `download.images.manifests` | Additional custom image manifest to download and push to the private registry. |
| `download.images.registry` | Default registry address used when downloading images. |
| `download.images.policy` | Image download/verification policy: `strict` (strict verification) or `warn` (warning only). |
| `download.images.<chart_name>` | Image mapping keyed by Helm Chart name; value is a mapping from version number to required image list. |

---
