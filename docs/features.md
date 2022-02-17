# Features

### Cluster Lifecycle Management 

* Create clusters
* Delete clusters
* Add nodes
* Update control plane and worker nodes
* Renew cluster certificates

### Supported Components

- Core components
  - Kubernetes/K3s
  - etcd
- Container runtimes
  - Docker
  - containerd
  - CRI-O (not integrated)
  - iSula (not integrated)
  - Kata
- Network plugins
  - Calico
  - Flannel
  - Kube-OVN
  - Cilium
  - Multus CNI
  - No plugin
- Storage
  - OpenEBS Local PV
  - Custom storage (allows users to customize storage service by using [addons](addons.md))
- Container images registries
  - [Docker registry](registry.md)
  - [Harbor](registry.md)
- Applications
  - Node Feature Discovery

### Air-Gapped Installation

- Create a private image registry
- Customize images available in the private image registry 

### Advanced Features

- Custom system component configurations (kube-apiserver/kube-controller-manager/kube-scheduler/kubelet/kube-proxy)
- Command plugins
- [Universal task scheduling framework](developer-guide.md)