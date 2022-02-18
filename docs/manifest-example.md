# QuickStart
The following is an example of a manifest file for a kubernetes v1.21.5 cluster. It contains the repositories for `ubuntu 20.04` and `centos 7`, some necessary components, private registry，and necessary images.
```yaml
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Manifest
metadata:
  name: sample
spec:
  arches: 
  - amd64
  operatingSystems: 
  - arch: amd64
    type: linux
    id: ubuntu
    version: "20.04"
    osImage: Ubuntu 20.04.3 LTS
    repository: 
      iso:
        localPath: 
        url: https://github.com/pixiake/k8s-dependencies/releases/download/v1.0.0/ubuntu-20.04-amd64-debs.iso
  - arch: amd64
    type: linux
    id: centos
    version: "7"
    osImage: CentOS Linux 7 (Core)
    repository:
      iso:
        localPath:
        url: https://github.com/pixiake/k8s-dependencies/releases/download/v1.0.0/centos-7-amd64-rpms.iso
  kubernetesDistributions: 
  - type: kubernetes
    version: v1.21.5
  components: 
    helm:
      version: v3.6.3
    cni:
      version: v0.9.1
    etcd:
      version: v3.4.13
    containerRuntimes:
    - type: docker
      version: 20.10.8
    crictl:
      version: v1.22.0
    docker-registry:
      version: "2"
    harbor:
      version: v2.4.1
    docker-compose:
      version: v2.2.2
  images:
  - docker.io/calico/cni:v3.20.0
  - docker.io/calico/kube-controllers:v3.20.0
  - docker.io/calico/node:v3.20.0
  - docker.io/calico/pod2daemon-flexvol:v3.20.0
  - docker.io/coredns/coredns:1.8.0
  - docker.io/kubesphere/k8s-dns-node-cache:1.15.12
  - docker.io/kubesphere/kube-apiserver:v1.21.5
  - docker.io/kubesphere/kube-controller-manager:v1.21.5
  - docker.io/kubesphere/kube-proxy:v1.21.5
  - docker.io/kubesphere/kube-scheduler:v1.21.5
  - docker.io/kubesphere/pause:3.4.1
```

# The Manifest Definition
The following is a full fields definition of the manifest file.
```yaml
apiVersion: kubekey.kubesphere.io/v1alpha2
kind: Manifest
metadata:
  name: sample
spec:
  arches: # Define the architecture that will be included in the artifact.
  - amd64
  - arm64
  operatingSystems: # Define the operating system that will be included in the artifact.
  - arch: amd64
    type: linux
    id: ubuntu
    version: "20.04"
    osImage: Ubuntu 20.04.3 LTS
    repository: # Define the operating system repository iso file that will be included in the artifact.
      iso:
        localPath: ./ubuntu.iso # Define getting the iso file from the local path.
        url: # Define getting the iso file from the URL.
  - arch: amd64
    type: linux
    id: centos
    version: "7"
    osImage: CentOS Linux 7 (Core)
    repository:
      iso:
        localPath:
        url: https://github.com/pixiake/k8s-dependencies/releases/download/v1.0.0/centos-7-amd64-rpms.iso
  kubernetesDistributions: # Define the kubernetes distribution that will be included in the artifact.
  - type: kubernetes
    version: v1.21.5
  - type: kubernetes
    version: v1.22.1
  ## The following components' versions are automatically generated based on the default configuration of KubeKey.
  components: 
    helm:
      version: v3.6.3
    cni:
      version: v0.9.1
    etcd:
      version: v3.4.13
    ## For now, if your cluster container runtime is containerd, KubeKey will add a docker 20.10.8 container runtime in the below list.
    ## The reason is KubeKey creates a cluster with containerd by installing a docker first and making kubelet connect the socket file of containerd which docker contained.
    containerRuntimes:
    - type: docker
      version: 20.10.8
    crictl:
      version: v1.22.0
    ## The following components define the private registry files that will be included in the artifact.
    docker-registry:
      version: "2"
    harbor:
      version: v2.4.1
    docker-compose:
      version: v2.2.2
  ## Define the images that will be included in the artifact.
  ## When you generate this file using KubeKey, all the images contained on the cluster hosts will be automatically added. 
  ## Of course, you can also modify this list of images manually.
  images:
  - docker.io/calico/cni:v3.20.0
  - docker.io/calico/kube-controllers:v3.20.0
  - docker.io/calico/node:v3.20.0
  - docker.io/calico/pod2daemon-flexvol:v3.20.0
  - docker.io/coredns/coredns:1.8.0
  - docker.io/kubesphere/k8s-dns-node-cache:1.15.12
  - docker.io/kubesphere/kube-apiserver:v1.21.5
  - docker.io/kubesphere/kube-controller-manager:v1.21.5
  - docker.io/kubesphere/kube-proxy:v1.21.5
  - docker.io/kubesphere/kube-scheduler:v1.21.5
  - docker.io/kubesphere/pause:3.4.1
  - dockerhub.kubekey.local/kubesphere/kube-apiserver:v1.22.1
  - dockerhub.kubekey.local/kubesphere/kube-controller-manager:v1.22.1
  - dockerhub.kubekey.local/kubesphere/kube-proxy:v1.22.1
  - dockerhub.kubekey.local/kubesphere/kube-scheduler:v1.22.1
  - dockerhub.kubekey.local/kubesphere/pause:3.5
  ## Define the authentication information if you need to pull images from a registry that requires authorization.
  registry:
    auths:
    "dockerhub.kubekey.local":
      username: "xxx"
      password: "***"
      plainHTTP: false # If the registry is serving for http request, set this to true.
```