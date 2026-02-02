# kubernetes

内建的playbook实现了kubernetes的一整套生命周期管理，包含创建集群，删除集群，添加节点，删除节点，升级集群等。

> **说明**：本文档详细列出了 KubeKey 支持的 Kubernetes 集群各组件的版本兼容性信息。这些版本信息基于官方文档和源码分析，确保各组件之间的兼容性和稳定性。在使用 KubeKey 部署或升级集群时，建议参考本文档选择合适的组件版本。

## 组件

> **注意**：以下列出的所有组件版本均经过测试验证，确保与对应 Kubernetes 版本的兼容性。版本选择遵循以下原则：
> - 优先使用官方推荐的稳定版本
> - 默认版本通常选择该 Kubernetes 版本范围内最新的稳定版本
> - 可通过配置参数自定义组件版本，但需确保版本兼容性

### [私有镜像仓库](./image_registry.md)

> **说明**：KubeKey 支持部署私有镜像仓库，用于在离线环境或内网环境中存储和管理容器镜像。选择合适的镜像仓库方案取决于您的具体需求（如安全性、功能丰富度、部署复杂度等）。

**Harbor**：一个用于存储和管理容器镜像的企业级开源镜像仓库，支持镜像访问控制、镜像签名、漏洞扫描及多种身份认证方式，适合大规模生产环境使用。

- **默认版本**：v2.10.2
- **适用场景**：企业级生产环境、需要高级安全功能的场景
- **主要特性**：RBAC权限控制、镜像扫描、Helm Chart仓库、多租户支持

**Docker Registry**：Docker官方的开源镜像仓库服务，用于保存和分发Docker镜像，功能简洁、部署方便，适合小型或自定义需求的场景。

- **默认版本**：2.8.3
- **适用场景**：小型团队、简单部署、快速搭建
- **主要特性**：轻量级、易于部署和维护

### etcd

> **说明**：etcd 是 Kubernetes 集群的核心数据存储组件，负责存储集群的所有配置数据、状态信息和元数据。etcd 的版本选择至关重要，必须与 Kubernetes 版本保持兼容。

> **重要提示**：
> - etcd 版本不兼容可能导致集群无法启动或数据丢失
> - 升级 etcd 前请务必备份数据
> - 建议使用官方推荐的默认版本，除非有特殊需求

各个kubernetes版本中etcd推荐列表：
| kubernetes 版本 | etcd 默认版本 | etcd 最小要求版本 | 来源 |
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
| 1.31.0\~1.31.13 | 3.5.15 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.31.0/cmd/kubeadm/app/constants/constants.go |
| 1.31.14 | 3.5.24 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.31.14/cmd/kubeadm/app/constants/constants.go |
| 1.32.0\~1.32.9 | 3.5.16 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.32.0/cmd/kubeadm/app/constants/constants.go |
| 1.32.10\~1.32.11 | 3.5.24 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.32.10/cmd/kubeadm/app/constants/constants.go |
| 1.33.0\~1.33.5 | 3.5.21 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.33.0/cmd/kubeadm/app/constants/constants.go |
| 1.33.6\~1.33.7 | 3.5.24 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.33.6/cmd/kubeadm/app/constants/constants.go |
| 1.34.0\~1.34.1 | 3.6.4 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.34.0/cmd/kubeadm/app/constants/constants.go |
| 1.34.2\~1.34.3 | 3.6.5 | 3.2.18 | https://github.com/kubernetes/kubernetes/blob/v1.34.2/cmd/kubeadm/app/constants/constants.go |

**kubekey config 中 etcd 默认值**：

> **说明**：KubeKey 为每个 Kubernetes 中版本（如 1.23、1.24）选择该版本范围内最大的 etcd 版本作为默认值，以确保最佳兼容性和稳定性。

> **自定义版本**：可通过 `--set etcd.etcd_version="v3.6.5"` 指定安装的 etcd 版本，但需确保该版本与您的 Kubernetes 版本兼容（参考上表）。
| kubernetes 版本 | etcd 默认版本 |
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

### 容器运行时

> **说明**：容器运行时是 Kubernetes 节点上负责运行容器的底层软件。KubeKey 支持多种容器运行时，包括 containerd、CRI-O 和 Docker（通过 cri-dockerd）。

> **版本选择**：默认版本根据 [cri-dockerd](https://github.com/Mirantis/cri-dockerd) 项目中的代码来设置，确保与 Kubernetes 版本的兼容性。

> **注意**：Docker 作为容器运行时在 Kubernetes 1.24+ 版本中已被弃用，建议使用 containerd 或 CRI-O。


### 容器网络插件

> **说明**：容器网络插件（CNI）是 Kubernetes 集群中负责 Pod 之间网络通信的核心组件。不同的 CNI 插件提供不同的网络功能特性，如网络策略、多租户隔离、服务网格集成等。

> **重要提示**：
> - KubeKey 默认只会安装一个网络插件，避免多个 CNI 插件冲突
> - 默认安装 Calico，如需使用其他插件请通过配置指定
> - 可通过 `--set cni.type="none"` 不安装任何插件（适用于已有网络插件或自定义网络方案）
> - 选择网络插件后，更换插件需要重新部署集群，请谨慎选择

#### [calico](https://github.com/projectcalico/calico)
> **说明**：Calico 是一个功能强大的网络和网络安全解决方案，支持 BGP 路由、网络策略、IP 地址管理等功能。适合需要细粒度网络控制和策略管理的生产环境。

> **安装方式**：
> - 通过 `--set cni.type="calico"` 指定安装 Calico 作为容器网络插件
> 通过 `--set cni.calico_version=“v3.31.3` 指定安装的 calico 版本

| kubernetes 版本 | 推荐 calico 版本 | kubekey 默认版本 | 来源 |
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

> **说明**：Cilium 是基于 eBPF 技术的新一代网络和安全解决方案，提供高性能的网络转发、网络策略、服务网格和可观测性功能。适合需要高性能和丰富功能的企业级环境。

> **安装方式**：
> - 通过 `--set cni.type="cilium"` 指定安装 Cilium 作为容器网络插件
> - 通过 `--set cni.cilium_version="1.18.5"` 指定安装的 Cilium 版本（不指定则使用默认版本）

> **主要特性**：基于 eBPF 的高性能数据平面、网络策略、服务网格集成、可观测性、多集群支持

| kubernetes 版本 | 推荐 cilium 版本 | kubekey 默认版本 | 来源 |
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

> **说明**：Flannel 是一个简单可靠的网络插件，为 Kubernetes 集群提供覆盖网络（overlay network）。Flannel 设计简洁，易于部署和维护，适合中小型集群或对网络功能要求不高的场景。

> **安装方式**：
> - 通过 `--set cni.type="flannel"` 指定安装 Flannel 作为容器网络插件
> - 通过 `--set cni.flannel_version="0.27.4"` 指定安装的 Flannel 版本（不指定则使用默认版本）

> **主要特性**：简单易用、轻量级、支持多种后端（VXLAN、host-gw、UDP）、跨节点通信

| kubernetes 版本 | 推荐 flannel 版本 | kubekey 默认版本 | 来源 |
|---|---|---|---|
| 1.23\~1.34 | 0.19.0+ | 0.27.4 | https://github.com/flannel-io/flannel/blob/master/Documentation/kubernetes.md | 


#### [hybridnet](https://github.com/alibaba/hybridnet)

> **说明**：HybridNet 是阿里巴巴开源的多网络平面 CNI 插件，支持 Underlay 和 Overlay 网络混合部署，提供灵活的 IP 地址管理和网络隔离能力。适合需要多网络平面和精细 IP 管理的场景。

> **安装方式**：
> - 通过 `--set cni.type="hybridnet"` 指定安装 HybridNet 作为容器网络插件
> - 通过 `--set cni.hybridnet_version="0.6.8"` 指定安装的 HybridNet 版本（不指定则使用默认版本）

> **注意**：HybridNet 官方未明确说明支持的 Kubernetes 版本范围。KubeKey 项目中默认版本为 0.6.8，建议在生产环境使用前进行充分测试。

> **主要特性**：多网络平面支持、Underlay/Overlay 混合部署、灵活的 IP 地址管理、网络隔离

#### [kubeovn](https://github.com/kubeovn/kube-ovn)

> **说明**：Kube-OVN 是基于 OVN（Open Virtual Network）的 Kubernetes 网络插件，提供企业级的网络功能，包括子网管理、QoS、网络策略、静态 IP 分配等。适合需要丰富网络功能和精细控制的场景。

> **安装方式**：
> - 通过 `--set cni.type="kubeovn"` 指定安装 Kube-OVN 作为容器网络插件
> - 通过 `--set cni.kubeovn_version="v1.15.0"` 指定安装的 Kube-OVN 版本（不指定则使用默认版本）

> **主要特性**：子网管理、QoS 流量控制、网络策略、静态 IP 分配、多租户支持、VPC 网络

| kubernetes 版本 | 推荐 kubeovn 版本 | kubekey 默认版本 | 来源 |
|---|---|---|---|
| 1.23\~1.28 | 1.12, 1.13 | v1.13.15 | https://kubeovn.github.io/docs/v1.12.x/en/start/prepare/<br>https://kubeovn.github.io/docs/v1.13.x/en/start/prepare/ |
| 1.29\~1.34 | 1.14, 1.15 | v1.15.0 | https://kubeovn.github.io/docs/v1.14.x/en/start/prepare/<br>https://kubeovn.github.io/docs/v1.15.x/en/start/prepare/ |


### 存储

> **说明**：存储插件为 Kubernetes 集群提供持久化存储能力，允许 Pod 挂载持久卷（PV）来保存数据。KubeKey 支持安装多个存储插件，以满足不同的存储需求。

> **重要提示**：
> - KubeKey 默认会安装 LocalPV 存储插件，提供本地存储能力
> - 可以根据实际需求启用其他存储插件（如 NFS）
> - 多个存储插件可以同时存在，通过 StorageClass 区分不同的存储类型

#### [localpv](https://github.com/openebs/dynamic-localpv-provisioner)

> **说明**：LocalPV 提供本地持久化存储能力，使用节点上的本地磁盘或目录作为存储后端。适合需要高性能本地存储的场景，如数据库、缓存等应用。

> **安装方式**：
> - 通过 `--set storage_class.local.enabled=true` 开启 LocalPV（默认已开启）
> - 通过 `--set storage_class.localpv_provisioner_version="4.4.0"` 指定版本（不指定则使用默认版本）

> **主要特性**：本地高性能存储、自动 PV 创建、支持多种存储类型（hostpath、device、lvm）

| kubernetes 版本 | 推荐 localpv 版本 | kubekey 默认版本 | 来源 |
|---|---|---|---|
| 1.23\~1.33 | v4.0.x, v4.1.x, v4.2.x, HEAD | 4.4.0 | https://github.com/openebs/dynamic-localpv-provisioner?tab=readme-ov-file#kubernetes-compatibility-matrix |

#### [nfs](https://github.com/kubernetes-sigs/nfs-subdir-external-provisioner)

> **说明**：NFS 存储插件提供基于 NFS（Network File System）的共享存储能力，允许多个 Pod 共享同一个存储卷。适合需要共享存储的场景，如文件服务、内容管理等应用。

> **安装方式**：
> - 通过 `--set storage_class.nfs.enabled=true` 开启 NFS 存储插件
> - 通过 `--set storage_class.nfs_provisioner_version="4.0.18"` 指定版本（不指定则使用默认版本）

> **前置条件**：需要预先配置 NFS 服务器，并在 KubeKey 配置中指定 NFS 服务器地址和路径

> **主要特性**：共享存储、多 Pod 访问、易于扩展、成本较低

| kubernetes 版本 | 推荐 nfs 版本 | kubekey 默认版本 | 来源 |
|---|---|---|---|
| 1.23\~1.33 | v4.0.0+ | 4.0.18 | https://github.com/kubernetes-sigs/nfs-subdir-external-provisioner/blob/nfs-subdir-external-provisioner-4.0.0/charts/nfs-subdir-external-provisioner/README.md#prerequisites |

### 域名服务

> **说明**：域名服务（DNS）是 Kubernetes 集群中负责服务发现和域名解析的核心组件。KubeKey 会自动部署 DNS 服务，确保集群内的 Pod 和服务可以通过域名相互访问。

#### coredns

> **说明**：CoreDNS 是 Kubernetes 集群的默认 DNS 服务器，负责解析集群内的 Service 和 Pod 域名。CoreDNS 是集群正常运行所必需的组件，KubeKey 会自动安装。

> **版本指定**：可通过 `--set dns.dns_image.tag="v1.12.1"` 指定 CoreDNS 版本（不指定则使用默认版本）

> **主要功能**：Service 域名解析、Pod 域名解析、自定义 DNS 规则、上游 DNS 转发

| kubernetes 版本 | 推荐 coredns 版本 | kubekey 默认 coredns 版本 | 来源 |
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

> **说明**：NodeLocalDNS 是一个 DNS 缓存守护进程，运行在每个节点上，用于缓存 DNS 查询结果，减少对 CoreDNS 的请求压力，提高 DNS 解析性能。NodeLocalDNS 是可选的性能优化组件。

> **版本指定**：可通过 `--set dns.dns_cache_image.tag="v1.26.4"` 指定 NodeLocalDNS 版本（不指定则使用默认版本）

> **主要功能**：DNS 查询缓存、减少 DNS 查询延迟、降低 CoreDNS 负载、提高集群 DNS 性能

> **注意**：NodeLocalDNS 不是必需组件，但建议在生产环境中启用以提升 DNS 性能

| kubernetes 版本 | 推荐 nodelocaldns 版本 | kubekey 默认 nodelocaldns 版本 | 来源 |
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

### pause 镜像

> **说明**：pause 镜像是 Kubernetes 中一个特殊的容器镜像，每个 Pod 都会包含一个 pause 容器。pause 容器作为 Pod 的网络命名空间和 PID 命名空间的占位符，用于保持 Pod 的网络配置和进程隔离。pause 镜像版本必须与 Kubernetes 版本匹配。

> **版本指定**：可通过 `--set kubernetes.sandbox_image="3.10.1"` 指定 pause 镜像版本（不指定则使用默认版本）

> **重要提示**：
> - pause 镜像版本由 Kubernetes 版本决定，不建议随意修改
> - pause 容器对用户不可见，但它是 Pod 正常运行的基础
> - 版本不匹配可能导致 Pod 无法正常启动

| kubernetes 版本 | pause 版本 | kubekey 默认 pause 版本 | 来源 |
|---|---|---|---|
| 1.23 | 3.6 | 3.6 | https://github.com/kubernetes/kubernetes/blob/v1.23.0/cmd/kubeadm/app/constants/constants.go#L412 |
| 1.24 | 3.7 | 3.7 | https://github.com/kubernetes/kubernetes/blob/v1.24.0/cmd/kubeadm/app/constants/constants.go#L428 |
| 1.25 | 3.8 | 3.8 | https://github.com/kubernetes/kubernetes/blob/v1.25.0/cmd/kubeadm/app/constants/constants.go#L424 |
| 1.26\~1.30 | 3.9 | 3.9 | https://github.com/kubernetes/kubernetes/blob/v1.26.0/cmd/kubeadm/app/constants/constants.go#L420<br>https://github.com/kubernetes/kubernetes/blob/v1.27.0/cmd/kubeadm/app/constants/constants.go#L420<br>https://github.com/kubernetes/kubernetes/blob/v1.28.0/cmd/kubeadm/app/constants/constants.go#L419<br>https://github.com/kubernetes/kubernetes/blob/v1.29.0/cmd/kubeadm/app/constants/constants.go#L423<br>https://github.com/kubernetes/kubernetes/blob/v1.30.0/cmd/kubeadm/app/constants/constants.go#L436 |
| 1.31\~1.33 | 3.10 | 3.10 | https://github.com/kubernetes/kubernetes/blob/v1.31.0/cmd/kubeadm/app/constants/constants.go#L438<br>https://github.com/kubernetes/kubernetes/blob/v1.32.0/cmd/kubeadm/app/constants/constants.go#L445<br>https://github.com/kubernetes/kubernetes/blob/v1.33.0/cmd/kubeadm/app/constants/constants.go#L445 |
| 1.34 | 3.10.1 | 3.10.1 | https://github.com/kubernetes/kubernetes/blob/v1.34.0/cmd/kubeadm/app/constants/constants.go#L445 |


