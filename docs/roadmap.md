# KubeKey's RoadMap

KubeKey Roadmap demonstrates a list of open source product development plans, as well as KubeSphere community's anticipation. Obviously, it details the future's direction of KubeKey, but may change over time. We hope that can help you to get familiar with the project plans and vision through the Roadmap. Of course, if you have any better ideas, welcome to filing Issues.
## v1.1
### Feature:
- [x] Support for deploying k8s in ARM architecture.
- [x] Support more container runtimes: cri-o containerd isula
- [x] Support more cni plugins: cilium kube-ovn
- [x] Support for deploying clusters without cni plugin.
- [x] Support custom components parameters.  
- [x] Support certificate expiration check and certificate update.
- [x] Support backup of etcd data.
- [x] Support for adding labels to nodes when deploying cluster.
- [ ] Support for adding taints to nodes when deploying cluster.
- [x] Support command auto-completion.
- [x] Support for deploying k3s (experimental).

## v1.2.0
### Feature:
- [x] Support for deploying Highly Available clusters by using internal load balancer.
- [x] Support Kubernetes certificate automatic renew.

## v2.0.0
### Feature:
- [x] More flexible task scheduling framework.
- [x] Support easier and more flexible air-gapped installation.
- [x] Support kubekey to independently generate certificate.
- [x] Support custom private registry authorization.
- [x] Enable featureGates in Kubernetes of cluster-config.
- [x] Support Kata and Node Feature Discovery.
- [x] Support customizing dnsDomain for the cluster.
- [x] Add pod PID limit and PID available.
- [x] Support setting NTP server and timezone.

## v2.1.0
### Feature:
- [] Support the use of kubeadm to manage etcd and use of existing etcd. 
- [] Optimize the Container Manager installation process.
- [] Reduce the size of KubeKey artifact.
- [] Support more version of Kubernetes.