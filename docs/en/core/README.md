# kubernetes

The built-in playbook implements a complete lifecycle management of Kubernetes, including cluster creation, cluster deletion, node addition, node removal, cluster upgrade, etc.

> **Note**: This document details the version compatibility information for all components of Kubernetes clusters supported by KubeKey. These version details are based on official documentation and source code analysis to ensure compatibility and stability between components. When deploying or upgrading clusters with KubeKey, it is recommended to refer to this document to select appropriate component versions.

## Components

> **Note**: All component versions listed below have been tested and verified to ensure compatibility with the corresponding Kubernetes versions. Version selection follows these principles:
> - Prioritize officially recommended stable versions
> - Default versions are typically the latest stable versions within the Kubernetes version range
> - Component versions can be customized through configuration parameters, but version compatibility must be ensured

### [Private Image Registry](./image_registry.md)

> **Note**: KubeKey supports deploying private image registries for storing and managing container images in offline or internal network environments. Choosing the appropriate image registry solution depends on your specific requirements (such as security, feature richness, deployment complexity, etc.).

**Harbor**: An enterprise-grade open-source image registry for storing and managing container images, supporting image access control, image signing, vulnerability scanning, and multiple authentication methods. Suitable for large-scale production environments.

- **Default Version**: v2.10.2
- **Use Cases**: Enterprise production environments, scenarios requiring advanced security features
- **Key Features**: RBAC access control, image scanning, Helm Chart repository, multi-tenant support

**Docker Registry**: Docker's official open-source image registry service for storing and distributing Docker images. Simple functionality and easy deployment, suitable for small teams or custom requirement scenarios.

- **Default Version**: 2.8.3
- **Use Cases**: Small teams, simple deployments, quick setup
- **Key Features**: Lightweight, easy to deploy and maintain

### etcd

> **Note**: etcd is the core data storage component of Kubernetes clusters, responsible for storing all configuration data, state information, and metadata of the cluster. etcd version selection is crucial and must be compatible with the Kubernetes version.

> **Important Notes**:
> - Incompatible etcd versions may cause cluster startup failures or data loss
> - Always backup data before upgrading etcd
> - It is recommended to use the officially recommended default version unless there are special requirements

Recommended etcd versions for each Kubernetes version:

| kubernetes version | etcd default version | etcd minimum required version | source |
|---|---|---|---|
| 1.23.0\~1.23.13 | 3.5.1 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.23.0/cmd/kubeadm/app/constants/constants.go |
| 1.23.14 | 3.5.5 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.23.14/cmd/kubeadm/app/constants/constants.go |
| 1.23.15\~1.23.17 | 3.5.6 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.23.15/cmd/kubeadm/app/constants/constants.go |
| 1.24.0\~1.24.7 | 3.5.3 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.24.0/cmd/kubeadm/app/constants/constants.go |
| 1.24.8 | 3.5.5 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.24.8/cmd/kubeadm/app/constants/constants.go |
| 1.24.9\~1.24.17 | 3.5.6 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.24.9/cmd/kubeadm/app/constants/constants.go |
| 1.25.0\~1.25.3 | 3.5.4 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.25.0/cmd/kubeadm/app/constants/constants.go |
| 1.25.4 | 3.5.5 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.25.4/cmd/kubeadm/app/constants/constants.go |
| 1.25.5\~1.25.14 | 3.5.6 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.25.5/cmd/kubeadm/app/constants/constants.go |
| 1.25.15\~1.25.16 | 3.5.9 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.25.15/cmd/kubeadm/app/constants/constants.go |
| 1.26.0\~1.26.9 | 3.5.6 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.26.0/cmd/kubeadm/app/constants/constants.go |
| 1.26.10\~1.26.12 | 3.5.9 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.26.10/cmd/kubeadm/app/constants/constants.go |
| 1.26.13\~1.26.15 | 3.5.10 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.26.13/cmd/kubeadm/app/constants/constants.go |
| 1.27.0\~1.27.6 | 3.5.7 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.27.0/cmd/kubeadm/app/constants/constants.go |
| 1.27.7\~1.27.9 | 3.5.9 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.27.7/cmd/kubeadm/app/constants/constants.go |
| 1.27.10\~1.27.11 | 3.5.10 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.27.10/cmd/kubeadm/app/constants/constants.go |
| 1.27.12\~1.27.16 | 3.5.12 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.27.12/cmd/kubeadm/app/constants/constants.go |
| 1.28.0\~1.28.5 | 3.5.9 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.28.0/cmd/kubeadm/app/constants/constants.go |
| 1.28.6\~1.28.7 | 3.5.10 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.28.6/cmd/kubeadm/app/constants/constants.go |
| 1.28.8\~1.28.13 | 3.5.12 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.28.8/cmd/kubeadm/app/constants/constants.go |
| 1.28.14\~1.28.15 | 3.5.15 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.28.14/cmd/kubeadm/app/constants/constants.go |
| 1.29.0\~1.29.2 | 3.5.10 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.29.0/cmd/kubeadm/app/constants/constants.go |
| 1.29.3\~1.29.8 | 3.5.12 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.29.3/cmd/kubeadm/app/constants/constants.go |
| 1.29.9\~1.29.10 | 3.5.15 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.29.9/cmd/kubeadm/app/constants/constants.go |
| 1.29.11\~1.29.15 | 3.5.16 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.29.11/cmd/kubeadm/app/constants/constants.go |
| 1.30.0\~1.30.4 | 3.5.12 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.30.0/cmd/kubeadm/app/constants/constants.go |
| 1.30.5\~1.30.14 | 3.5.15 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.30.5/cmd/kubeadm/app/constants/constants.go |
| 1.31.0\~1.30.13 | 3.5.15 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.31.0/cmd/kubeadm/app/constants/constants.go |
| 1.31.14 | 3.5.24 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.31.14/cmd/kubeadm/app/constants/constants.go |
| 1.32.0\~1.32.9 | 3.5.16 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.32.0/cmd/kubeadm/app/constants/constants.go |
| 1.32.10\~1.32.11 | 3.5.24 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.32.10/cmd/kubeadm/app/constants/constants.go |
| 1.33.0\~1.33.5 | 3.5.21 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.33.0/cmd/kubeadm/app/constants/constants.go |
| 1.33.6\~1.33.7 | 3.5.24 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.33.6/cmd/kubeadm/app/constants/constants.go |
| 1.34.0\~1.34.1 | 3.6.4 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.34.0/cmd/kubeadm/app/constants/constants.go |
| 1.34.2\~1.34.3 | 3.6.5 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.34.2/cmd/kubeadm/app/constants/constants.go |

**etcd default values in kubekey config**:

> **Note**: KubeKey selects the maximum etcd version within each Kubernetes minor version (e.g., 1.23, 1.24) as the default value to ensure optimal compatibility and stability.

> **Custom Version**: You can specify the etcd version to install via `--set etcd.etcd_version="v3.6.5"`, but ensure that the version is compatible with your Kubernetes version (refer to the table above).

| kubernetes version | etcd default version |
|---|---|
| 1.23 | v3.5.6 |
| 1.24 | v3.5.6 |
| 1.25 | v3.5.9 |
| 1.26 | v3.5.10 |
| 1.27 | v3.5.12 |
| 1.28 | v3.5.15 |
| 1.29 | v3.5.15 |
| 1.30 | v3.5.15 |
| 1.31 | v3.5.24 |
| 1.32 | v3.5.24 |
| 1.33 | v3.5.24 |
| 1.34 | v3.6.5 |

### Container Runtime

> **Note**: Container runtime is the underlying software on Kubernetes nodes responsible for running containers. KubeKey supports multiple container runtimes, including containerd, CRI-O, and Docker (via cri-dockerd).

> **Version Selection**: Default versions are set based on the code in the [cri-dockerd](https://github.com/Mirantis/cri-dockerd) project to ensure compatibility with Kubernetes versions.

> **Note**: Docker as a container runtime has been deprecated in Kubernetes 1.24+, and it is recommended to use containerd or CRI-O.

### Container Network Plugin

> **Note**: Container Network Plugin (CNI) is the core component in Kubernetes clusters responsible for network communication between Pods. Different CNI plugins provide different network features, such as network policies, multi-tenant isolation, service mesh integration, etc.

> **Important Notes**:
> - KubeKey will only install one network plugin by default to avoid conflicts between multiple CNI plugins
> - Calico is installed by default. If you need to use other plugins, please specify them through configuration
> - You can use `--set cni.type="none"` to not install any plugin (suitable for existing network plugins or custom network solutions)
> - After selecting a network plugin, changing plugins requires redeploying the cluster, so choose carefully

#### [calico](https://github.com/projectcalico/calico)

> **Note**: Calico is a powerful networking and network security solution that supports BGP routing, network policies, IP address management, and more. Suitable for production environments requiring fine-grained network control and policy management.

> **Installation**:
> - Use `--set cni.type="calico"` to specify Calico as the container network plugin
> - Use `--set cni.calico_version="v3.31.3"` to specify the Calico version to install (if not specified, the default version will be used)

> **Key Features**: Network policies, BGP routing, IP pool management, multi-tenant support, eBPF data plane

| kubernetes version | recommended calico version | kubekey default version | source |
|---|---|---|---|
| 1.23 | 3.25 | v3.25.2 | https://archive-os-3-25.netlify.app/calico/3.25/getting-started/kubernetes/requirements/#kubernetes-requirements |
| 1.24 | 3.25, 3.26 | v3.26.5 | https://archive-os-3-25.netlify.app/calico/3.25/getting-started/kubernetes/requirements/#kubernetes-requirements<br>https://archive-os-3-26.netlify.app/calico/3.26/getting-started/kubernetes/requirements/#kubernetes-requirements |
| 1.25 | 3.25, 3.26 | v3.26.5 | https://archive-os-3-25.netlify.app/calico/3.25/getting-started/kubernetes/requirements/#kubernetes-requirements<br>https://archive-os-3-26.netlify.app/calico/3.26/getting-started/kubernetes/requirements/#kubernetes-requirements |
| 1.26 | 3.25, 3.26 | v3.26.5 | https://archive-os-3-25.netlify.app/calico/3.25/getting-started/kubernetes/requirements/#kubernetes-requirements<br>https://archive-os-3-26.netlify.app/calico/3.26/getting-started/kubernetes/requirements/#kubernetes-requirements |
| 1.27 | 3.25, 3.26, 3.27, 3.28 | v3.28.5 | https://archive-os-3-25.netlify.app/calico/3.25/getting-started/kubernetes/requirements/#kubernetes-requirements<br>https://archive-os-3-26.netlify.app/calico/3.26/getting-started/kubernetes/requirements/#kubernetes-requirements<br>https://archive-os-3-27.netlify.app/calico/3.27/getting-started/kubernetes/requirements/#kubernetes-requirements<br>https://archive-os-3-28.netlify.app/calico/3.28/getting-started/kubernetes/requirements/#kubernetes-requirements |
| 1.28 | 3.25, 3.26, 3.27, 3.28 | v3.28.5 | https://archive-os-3-25.netlify.app/calico/3.25/getting-started/kubernetes/requirements/#kubernetes-requirements<br>https://archive-os-3-26.netlify.app/calico/3.26/getting-started/kubernetes/requirements/#kubernetes-requirements<br>https://archive-os-3-27.netlify.app/calico/3.27/getting-started/kubernetes/requirements/#kubernetes-requirements<br>https://archive-os-3-28.netlify.app/calico/3.28/getting-started/kubernetes/requirements/#kubernetes-requirements |
| 1.29 | 3.27, 3.28, 3.29 | v3.29.7 | https://archive-os-3-27.netlify.app/calico/3.27/getting-started/kubernetes/requirements/#kubernetes-requirements<br>https://archive-os-3-28.netlify.app/calico/3.28/getting-started/kubernetes/requirements/#kubernetes-requirements<br>https://docs.tigera.io/calico/3.29/getting-started/kubernetes/requirements#kubernetes-requirements |
| 1.30 | 3.28, 3.29 | v3.29.7 | https://archive-os-3-28.netlify.app/calico/3.28/getting-started/kubernetes/requirements/#kubernetes-requirements<br>https://docs.tigera.io/calico/3.29/getting-started/kubernetes/requirements#kubernetes-requirements |
| 1.31 | 3.29, 3.30 | v3.30.5 | https://docs.tigera.io/calico/3.29/getting-started/kubernetes/requirements#kubernetes-requirements<br>https://docs.tigera.io/calico/3.30/getting-started/kubernetes/requirements#kubernetes-requirements |
| 1.32 | 3.29, 3.30, 3.31 | v3.31.3 | https://docs.tigera.io/calico/3.29/getting-started/kubernetes/requirements#kubernetes-requirements<br>https://docs.tigera.io/calico/3.30/getting-started/kubernetes/requirements#kubernetes-requirements<br>https://docs.tigera.io/calico/latest/getting-started/kubernetes/requirements#kubernetes-requirements |
| 1.33 | 3.30, 3.31 | v3.31.3 | https://docs.tigera.io/calico/3.30/getting-started/kubernetes/requirements#kubernetes-requirements<br>https://docs.tigera.io/calico/latest/getting-started/kubernetes/requirements#kubernetes-requirements |
| 1.34 | 3.31 | v3.31.3 | https://docs.tigera.io/calico/latest/getting-started/kubernetes/requirements#kubernetes-requirements |


#### [cilium](https://github.com/cilium/cilium)

> **Note**: Cilium is a next-generation networking and security solution based on eBPF technology, providing high-performance network forwarding, network policies, service mesh, and observability features. Suitable for enterprise environments requiring high performance and rich features.

> **Installation**:
> - Use `--set cni.type="cilium"` to specify Cilium as the container network plugin
> - Use `--set cni.cilium_version="1.18.5"` to specify the Cilium version to install (if not specified, the default version will be used)

> **Key Features**: High-performance eBPF-based data plane, network policies, service mesh integration, observability, multi-cluster support

| kubernetes version | recommended cilium version | kubekey default version | source |
|---|---|---|---|
| 1.23 | 1.14 | 1.14.19 | https://docs.cilium.io/en/v1.14/network/kubernetes/compatibility/ |
| 1.24 | 1.14 | 1.14.19  | https://docs.cilium.io/en/v1.14/network/kubernetes/compatibility/ |
| 1.25 | 1.14 | 1.14.19 | https://docs.cilium.io/en/v1.14/network/kubernetes/compatibility/ |
| 1.26 | 1.14, 1.15 | 1.15.19 | https://docs.cilium.io/en/v1.14/network/kubernetes/compatibility/<br>https://docs.cilium.io/en/v1.15/network/kubernetes/compatibility/ |
| 1.27 | 1.14, 1.15, 1.16 | 1.16.18 | https://docs.cilium.io/en/v1.14/network/kubernetes/compatibility/<br>https://docs.cilium.io/en/v1.15/network/kubernetes/compatibility/<br>https://docs.cilium.io/en/v1.16/network/kubernetes/compatibility/ |
| 1.28 | 1.15, 1.16 | 1.16.18 | https://docs.cilium.io/en/v1.15/network/kubernetes/compatibility/<br>https://docs.cilium.io/en/v1.16/network/kubernetes/compatibility/ |
| 1.29 | 1.15, 1.16, 1.17 | 1.17.11 | https://docs.cilium.io/en/v1.15/network/kubernetes/compatibility/<br>https://docs.cilium.io/en/v1.16/network/kubernetes/compatibility/<br>https://docs.cilium.io/en/v1.17/network/kubernetes/compatibility/ |
| 1.30 | 1.16, 1.17, 1.18 | 1.18.5 | https://docs.cilium.io/en/v1.16/network/kubernetes/compatibility/<br>https://docs.cilium.io/en/v1.17/network/kubernetes/compatibility/<br>https://docs.cilium.io/en/v1.18/network/kubernetes/compatibility/ |
| 1.31 | 1.17, 1.18 | 1.18.5 | https://docs.cilium.io/en/v1.17/network/kubernetes/compatibility/<br>https://docs.cilium.io/en/v1.18/network/kubernetes/compatibility/ |
| 1.32 | 1.17, 1.18 | 1.18.5 | https://docs.cilium.io/en/v1.17/network/kubernetes/compatibility/<br>https://docs.cilium.io/en/v1.18/network/kubernetes/compatibility/ |
| 1.33 | 1.18 | 1.18.5 | https://docs.cilium.io/en/v1.18/network/kubernetes/compatibility/ |
| 1.34 | - | 1.18.5 | - |

#### [flannel](https://github.com/flannel-io/flannel)

> **Note**: Flannel is a simple and reliable network plugin that provides overlay network for Kubernetes clusters. Flannel is designed to be simple, easy to deploy and maintain, suitable for small to medium-sized clusters or scenarios with low network feature requirements.

> **Installation**:
> - Use `--set cni.type="flannel"` to specify Flannel as the container network plugin
> - Use `--set cni.flannel_version="0.27.4"` to specify the Flannel version to install (if not specified, the default version will be used)

> **Key Features**: Simple and easy to use, lightweight, supports multiple backends (VXLAN, host-gw, UDP), cross-node communication

| kubernetes version | recommended flannel version | kubekey default version | source |
|---|---|---|---|
| 1.23\~1.34 | 0.19.0+ | 0.27.4 | https://github.com/flannel-io/flannel/blob/master/Documentation/kubernetes.md | 


#### [hybridnet](https://github.com/alibaba/hybridnet)

> **Note**: HybridNet is an open-source multi-network plane CNI plugin from Alibaba, supporting hybrid deployment of Underlay and Overlay networks, providing flexible IP address management and network isolation capabilities. Suitable for scenarios requiring multiple network planes and fine-grained IP management.

> **Installation**:
> - Use `--set cni.type="hybridnet"` to specify HybridNet as the container network plugin
> - Use `--set cni.hybridnet_version="0.6.8"` to specify the HybridNet version to install (if not specified, the default version will be used)

> **Note**: HybridNet official documentation does not clearly specify the supported Kubernetes version range. The default version in the KubeKey project is 0.6.8. It is recommended to fully test before using in production environments.

> **Key Features**: Multi-network plane support, Underlay/Overlay hybrid deployment, flexible IP address management, network isolation

#### [kubeovn](https://github.com/kubeovn/kube-ovn)

> **Note**: Kube-OVN is a Kubernetes network plugin based on OVN (Open Virtual Network), providing enterprise-grade network features including subnet management, QoS, network policies, static IP allocation, etc. Suitable for scenarios requiring rich network features and fine-grained control.

> **Installation**:
> - Use `--set cni.type="kubeovn"` to specify Kube-OVN as the container network plugin
> - Use `--set cni.kubeovn_version="v1.15.0"` to specify the Kube-OVN version to install (if not specified, the default version will be used)

> **Key Features**: Subnet management, QoS traffic control, network policies, static IP allocation, multi-tenant support, VPC network

| kubernetes version | recommended kubeovn version | kubekey default version | source |
|---|---|---|---|
| 1.23\~1.28 | 1.12, 1.13 | v1.13.15 | https://kubeovn.github.io/docs/v1.12.x/en/start/prepare/<br>https://kubeovn.github.io/docs/v1.13.x/en/start/prepare/ |
| 1.29\~1.34 | 1.14, 1.15 | v1.15.0 | https://kubeovn.github.io/docs/v1.14.x/en/start/prepare/<br>https://kubeovn.github.io/docs/v1.15.x/en/start/prepare/ |


### Storage

> **Note**: Storage plugins provide persistent storage capabilities for Kubernetes clusters, allowing Pods to mount persistent volumes (PV) to save data. KubeKey supports installing multiple storage plugins to meet different storage requirements.

> **Important Notes**:
> - KubeKey will install the LocalPV storage plugin by default, providing local storage capabilities
> - Other storage plugins (such as NFS) can be enabled according to actual requirements
> - Multiple storage plugins can coexist, distinguished by StorageClass for different storage types

#### [localpv](https://github.com/openebs/dynamic-localpv-provisioner)

> **Note**: LocalPV provides local persistent storage capabilities, using local disks or directories on nodes as storage backends. Suitable for scenarios requiring high-performance local storage, such as databases, caches, and other applications.

> **Installation**:
> - Use `--set storage_class.local.enabled=true` to enable LocalPV (enabled by default)
> - Use `--set storage_class.localpv_provisioner_version="4.4.0"` to specify the version (if not specified, the default version will be used)

> **Key Features**: Local high-performance storage, automatic PV creation, supports multiple storage types (hostpath, device, lvm)

| kubernetes version | recommended localpv version | kubekey default version | source |
|---|---|---|---|
| 1.23\~1.33 | v4.0.x, v4.1.x, v4.2.x, HEAD | 4.4.0 | https://github.com/openebs/dynamic-localpv-provisioner?tab=readme-ov-file#kubernetes-compatibility-matrix |

#### [nfs](https://github.com/kubernetes-sigs/nfs-subdir-external-provisioner)

> **Note**: The NFS storage plugin provides shared storage capabilities based on NFS (Network File System), allowing multiple Pods to share the same storage volume. Suitable for scenarios requiring shared storage, such as file services, content management, and other applications.

> **Installation**:
> - Use `--set storage_class.nfs.enabled=true` to enable the NFS storage plugin
> - Use `--set storage_class.nfs_provisioner_version="4.0.18"` to specify the version (if not specified, the default version will be used)

> **Prerequisites**: An NFS server must be pre-configured, and the NFS server address and path must be specified in the KubeKey configuration

> **Key Features**: Shared storage, multi-Pod access, easy to scale, low cost

| kubernetes version | recommended nfs version | kubekey default version | source |
|---|---|---|---|
| 1.23\~1.33 | v4.0.0+ | 4.0.18 | https://github.com/kubernetes-sigs/nfs-subdir-external-provisioner/blob/nfs-subdir-external-provisioner-4.0.0/charts/nfs-subdir-external-provisioner/README.md#prerequisites |

### DNS Service

> **Note**: DNS service is the core component in Kubernetes clusters responsible for service discovery and domain name resolution. KubeKey will automatically deploy DNS services to ensure Pods and services within the cluster can access each other through domain names.

#### coredns

> **Note**: CoreDNS is the default DNS server for Kubernetes clusters, responsible for resolving Service and Pod domain names within the cluster. CoreDNS is a required component for normal cluster operation, and KubeKey will install it automatically.

> **Version Specification**: You can specify the CoreDNS version via `--set dns.dns_image.tag="v1.12.1"` (if not specified, the default version will be used)

> **Key Functions**: Service domain name resolution, Pod domain name resolution, custom DNS rules, upstream DNS forwarding

| kubernetes version | recommended coredns version | kubekey default coredns version | source |
|---|---|---|---|
| 1.23\~1.24 | v1.8.6 | v1.8.6 ｜ https://github.com/kubernetes/kubernetes/blob/v1.23.0/cluster/addons/dns/coredns/coredns.yaml.base#L142<br>https://github.com/kubernetes/kubernetes/blob/v1.24.0/cluster/addons/dns/coredns/coredns.yaml.base#L142 |
| 1.25\~1.26 | v1.9.3 | v1.9.3 ｜ https://github.com/kubernetes/kubernetes/blob/v1.25.0/cluster/addons/dns/coredns/coredns.yaml.base#L142<br>https://github.com/kubernetes/kubernetes/blob/v1.26.0/cluster/addons/dns/coredns/coredns.yaml.base#L142 |
| 1.27\~1.28 | v1.10.1 | v1.10.1 ｜ https://github.com/kubernetes/kubernetes/blob/v1.27.0/cluster/addons/dns/coredns/coredns.yaml.base#L142<br>https://github.com/kubernetes/kubernetes/blob/v1.28.0/cluster/addons/dns/coredns/coredns.yaml.base#L136 |
| 1.29 | v1.11.1 | v1.11.1 ｜ https://github.com/kubernetes/kubernetes/blob/v1.29.0/cluster/addons/dns/coredns/coredns.yaml.base#L136 |
| 1.30.0\~1.30.4 | v1.11.1 | v1.11.3 | https://github.com/kubernetes/kubernetes/blob/v1.30.0/cluster/addons/dns/coredns/coredns.yaml.base#L136 |
| 1.30.5\~1.30.14 | v1.11.3 | v1.11.3 | https://github.com/kubernetes/kubernetes/blob/v1.30.5/cluster/addons/dns/coredns/coredns.yaml.base#L136 |
| 1.31.0 | v1.11.1 | v1.11.3 | https://github.com/kubernetes/kubernetes/blob/v1.31.0/cluster/addons/dns/coredns/coredns.yaml.base#L136 |
| 1.31.1\~1.31.14 | v1.11.3 | v1.11.3 | https://github.com/kubernetes/kubernetes/blob/v1.31.1/cluster/addons/dns/coredns/coredns.yaml.base#L136 |
| 1.32.0\~1.32.11 | v1.11.3 | v1.11.3 | https://github.com/kubernetes/kubernetes/blob/v1.32.0/cluster/addons/dns/coredns/coredns.yaml.base#L136 |
| 1.33.0\~1.33.7 | v1.12.0 | v1.12.0 | https://github.com/kubernetes/kubernetes/blob/v1.33.0/cluster/addons/dns/coredns/coredns.yaml.base#L136 |
| 1.34.0\~1.34.3 | v1.12.1 | v1.12.1 | https://github.com/kubernetes/kubernetes/blob/v1.34.0/cluster/addons/dns/coredns/coredns.yaml.base#L136 |

#### nodelocaldns

> **Note**: NodeLocalDNS is a DNS cache daemon that runs on each node to cache DNS query results, reducing the request pressure on CoreDNS and improving DNS resolution performance. NodeLocalDNS is an optional performance optimization component.

> **Version Specification**: You can specify the NodeLocalDNS version via `--set dns.dns_cache_image.tag="v1.26.4"` (if not specified, the default version will be used)

> **Key Functions**: DNS query caching, reduced DNS query latency, reduced CoreDNS load, improved cluster DNS performance

> **Note**: NodeLocalDNS is not a required component, but it is recommended to enable it in production environments to improve DNS performance

| kubernetes version | recommended nodelocaldns version | kubekey default nodelocaldns version | source |
|---|---|---|---|
| 1.23\~1.24 | 1.21.1 | v1.21.1 | https://github.com/kubernetes/kubernetes/blob/v1.23.0/cluster/addons/dns/nodelocaldns/nodelocaldns.yaml#L141<br>https://github.com/kubernetes/kubernetes/blob/v1.24.0/cluster/addons/dns/coredns/coredns.yaml.base#L142 |
| 1.25 | 1.22.8 | v1.22.8 | https://github.com/kubernetes/kubernetes/blob/v1.25.0/cluster/addons/dns/coredns/coredns.yaml.base#L142 |
| 1.26 | 1.22.13 | v1.22.13 | https://github.com/kubernetes/kubernetes/blob/v1.26.0/cluster/addons/dns/coredns/coredns.yaml.base#L142 |
| 1.27 | 1.22.20 | v1.22.20 | https://github.com/kubernetes/kubernetes/blob/v1.27.0/cluster/addons/dns/coredns/coredns.yaml.base#L142 |
| 1.28\~1.29 | 1.22.23 | v1.22.23 | https://github.com/kubernetes/kubernetes/blob/v1.28.0/cluster/addons/dns/coredns/coredns.yaml.base#L142<br>https://github.com/kubernetes/kubernetes/blob/v1.29.0/cluster/addons/dns/coredns/coredns.yaml.base#L142 |
| 1.30 | 1.22.28 | v1.22.28 | https://github.com/kubernetes/kubernetes/blob/v1.30.0/cluster/addons/dns/coredns/coredns.yaml.base#L142 |
| 1.31\~1.32 | 1.23.1 | v1.23.1 | https://github.com/kubernetes/kubernetes/blob/v1.31.0/cluster/addons/dns/coredns/coredns.yaml.base#L142<br>https://github.com/kubernetes/kubernetes/blob/v1.32.0/cluster/addons/dns/coredns/coredns.yaml.base#L142 |
| 1.33 | 1.25.0 | v1.25.0 | https://github.com/kubernetes/kubernetes/blob/v1.33.0/cluster/addons/dns/coredns/coredns.yaml.base#L142 |
| 1.34 | 1.26.4 | v1.26.4 | https://github.com/kubernetes/kubernetes/blob/v1.34.0/cluster/addons/dns/coredns/coredns.yaml.base#L142 |

### pause Image

> **Note**: The pause image is a special container image in Kubernetes. Each Pod contains a pause container. The pause container serves as a placeholder for the Pod's network namespace and PID namespace, maintaining the Pod's network configuration and process isolation. The pause image version must match the Kubernetes version.

> **Version Specification**: You can specify the pause image version via `--set kubernetes.sandbox_image="3.10.1"` (if not specified, the default version will be used)

> **Important Notes**:
> - The pause image version is determined by the Kubernetes version, and it is not recommended to modify it arbitrarily
> - The pause container is invisible to users, but it is the foundation for normal Pod operation
> - Version mismatches may cause Pods to fail to start normally

| kubernetes version | pause version | kubekey default pause version | source |
|---|---|---|---|
| 1.23 | 3.6 | 3.6 | https://github.com/kubernetes/kubernetes/blob/v1.23.0/cmd/kubeadm/app/constants/constants.go#L412 |
| 1.24 | 3.7 | 3.7 | https://github.com/kubernetes/kubernetes/blob/v1.24.0/cmd/kubeadm/app/constants/constants.go#L428 |
| 1.25 | 3.8 | 3.8 | https://github.com/kubernetes/kubernetes/blob/v1.25.0/cmd/kubeadm/app/constants/constants.go#L424 |
| 1.26\~1.30 | 3.9 | 3.9 | https://github.com/kubernetes/kubernetes/blob/v1.26.0/cmd/kubeadm/app/constants/constants.go#L420<br>https://github.com/kubernetes/kubernetes/blob/v1.27.0/cmd/kubeadm/app/constants/constants.go#L420<br>https://github.com/kubernetes/kubernetes/blob/v1.28.0/cmd/kubeadm/app/constants/constants.go#L419<br>https://github.com/kubernetes/kubernetes/blob/v1.29.0/cmd/kubeadm/app/constants/constants.go#L423<br>https://github.com/kubernetes/kubernetes/blob/v1.30.0/cmd/kubeadm/app/constants/constants.go#L436 |
| 1.31\~1.33 | 3.10 | 3.10 | https://github.com/kubernetes/kubernetes/blob/v1.31.0/cmd/kubeadm/app/constants/constants.go#L438<br>https://github.com/kubernetes/kubernetes/blob/v1.32.0/cmd/kubeadm/app/constants/constants.go#L445<br>https://github.com/kubernetes/kubernetes/blob/v1.33.0/cmd/kubeadm/app/constants/constants.go#L445 |
| 1.34 | 3.10.1 | 3.10.1 | https://github.com/kubernetes/kubernetes/blob/v1.34.0/cmd/kubeadm/app/constants/constants.go#L445 |

