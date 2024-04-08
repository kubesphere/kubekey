#!/usr/bin/env bash

# Copyright 2022 The KubeSphere Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#####################################################################
#
#  Usage:
#    Specify the component version through environment variables.
#
#    For example:
#
#      KUBERNETES_VERSION=v1.25.3 bash hack/sync-components.sh
#
####################################################################

set -e

KUBERNETES_VERSION=${KUBERNETES_VERSION}
NODE_LOCAL_DNS_VERSION=${NODE_LOCAL_DNS_VERSION}
COREDNS_VERSION=${COREDNS_VERSION}
CALICO_VERSION=${CALICO_VERSION}
KUBE_OVN_VERSION=${KUBE_OVN_VERSION}
CILIUM_VERSION=${CILIUM_VERSION}
OPENEBS_VERSION=${OPENEBS_VERSION}
KUBEVIP_VERSION=${KUBEVIP_VERSION}
HAPROXY_VERSION=${HAPROXY_VERSION}
HELM_VERSION=${HELM_VERSION}
CNI_VERSION=${CNI_VERSION}
ETCD_VERSION=${ETCD_VERSION}
CRICTL_VERSION=${CRICTL_VERSION}
K3S_VERSION=${K3S_VERSION}
CONTAINERD_VERSION=${CONTAINERD_VERSION}
RUNC_VERSION=${RUNC_VERSION}
COMPOSE_VERSION=${COMPOSE_VERSION}
CALICO_VERSION=${CALICO_VERSION}
CRI_DOCKER_VERSION=${CRI_DOCKER_VERSION}

# qsctl
QSCTL_ACCESS_KEY_ID=${QSCTL_ACCESS_KEY_ID}
QSCTL_SECRET_ACCESS_KEY=${QSCTL_SECRET_ACCESS_KEY}

# docker.io
DOCKERHUB_USERNAME=${DOCKERHUB_USERNAME}
DOCKERHUB_PASSWORD=${DOCKERHUB_PASSWORD}

# registry.cn-beijing.aliyuncs.com
ALIYUNCS_USERNAME=${ALIYUNCS_USERNAME}
ALIYUNCS_PASSWORD=${ALIYUNCS_PASSWORD}

DOCKERHUB_NAMESPACE="kubesphere"
ALIYUNCS_NAMESPACE="kubesphereio"

BINARIES=("kubeadm" "kubelet" "kubectl")
ARCHS=("amd64" "arm64")

# Generate qsctl config
if [ $QSCTL_ACCESS_KEY_ID ] && [ $QSCTL_SECRET_ACCESS_KEY ];then
   echo "access_key_id: $QSCTL_ACCESS_KEY_ID" > qsctl-config.yaml
   echo "secret_access_key: $QSCTL_SECRET_ACCESS_KEY" >> qsctl-config.yaml
fi

# Login docker.io
if [ $DOCKERHUB_USERNAME ] && [ $DOCKERHUB_PASSWORD ];then
   skopeo login docker.io -u $DOCKERHUB_USERNAME -p $DOCKERHUB_PASSWORD
fi

# Login registry.cn-beijing.aliyuncs.com
if [ $ALIYUNCS_USERNAME ] && [ $ALIYUNCS_PASSWORD ];then
   skopeo login registry.cn-beijing.aliyuncs.com -u $ALIYUNCS_USERNAME -p $ALIYUNCS_PASSWORD
fi

# Sync Kubernetes Binaries and Images
if [ $KUBERNETES_VERSION ]; then
   for arch in ${ARCHS[@]}
   do
     mkdir -p binaries/kube/$KUBERNETES_VERSION/$arch
     for binary in ${BINARIES[@]}
     do
       echo "Synchronizing $binary-$arch"

       curl -L -o binaries/kube/$KUBERNETES_VERSION/$arch/$binary \
                  https://storage.googleapis.com/kubernetes-release/release/$KUBERNETES_VERSION/bin/linux/$arch/$binary

       qsctl cp binaries/kube/$KUBERNETES_VERSION/$arch/$binary \
             qs://kubernetes-release/release/$KUBERNETES_VERSION/bin/linux/$arch/$binary \
             -c qsctl-config.yaml
     done
   done

   chmod +x binaries/kube/$KUBERNETES_VERSION/amd64/kubeadm
   binaries/kube/$KUBERNETES_VERSION/amd64/kubeadm config images list --kubernetes-version $KUBERNETES_VERSION | xargs -I {} skopeo sync --src docker --dest docker {} docker.io/$DOCKERHUB_NAMESPACE/${image##} --all
   binaries/kube/$KUBERNETES_VERSION/amd64/kubeadm config images list --kubernetes-version $KUBERNETES_VERSION | xargs -I {} skopeo sync --src docker --dest docker {} registry.cn-beijing.aliyuncs.com/$ALIYUNCS_NAMESPACE/${image##} --all

   rm -rf binaries
fi

# Sync Helm Binary
if [ $HELM_VERSION ]; then
   for arch in ${ARCHS[@]}
   do
     mkdir -p binaries/helm/$HELM_VERSION/$arch
     echo "Synchronizing helm-$arch"

     curl -L -o binaries/helm/$HELM_VERSION/$arch/helm-$HELM_VERSION-linux-$arch.tar.gz \
                https://get.helm.sh/helm-$HELM_VERSION-linux-$arch.tar.gz

     tar -zxf binaries/helm/$HELM_VERSION/$arch/helm-$HELM_VERSION-linux-$arch.tar.gz -C binaries/helm/$HELM_VERSION/$arch

     sha256sum binaries/helm/$HELM_VERSION/$arch/linux-$arch/helm

     qsctl cp $KUBERNETES_VERSION/$arch/linux-$arch/helm \
           qs://kubernetes-helm/linux-$arch/$HELM_VERSION/helm \
           -c qsctl-config.yaml

     qsctl cp binaries/helm/$HELM_VERSION/$arch/helm-$HELM_VERSION-linux-$arch.tar.gz \
           qs://kubernetes-helm/linux-$arch/$HELM_VERSION/helm-$HELM_VERSION-linux-$arch.tar.gz \
           -c qsctl-config.yaml
   done

   rm -rf binaries
fi

# Sync ETCD Binary
if [ $ETCD_VERSION ]; then
   for arch in ${ARCHS[@]}
   do
     mkdir -p binaries/etcd/$ETCD_VERSION/$arch
     echo "Synchronizing etcd-$arch"

     curl -L -o binaries/etcd/$ETCD_VERSION/$arch/etcd-$ETCD_VERSION-linux-$arch.tar.gz \
                https://github.com/coreos/etcd/releases/download/$ETCD_VERSION/etcd-$ETCD_VERSION-linux-$arch.tar.gz

     sha256sum binaries/etcd/$ETCD_VERSION/$arch/etcd-$ETCD_VERSION-linux-$arch.tar.gz

     qsctl cp binaries/etcd/$ETCD_VERSION/$arch/etcd-$ETCD_VERSION-linux-$arch.tar.gz \
           qs://kubernetes-release/etcd/release/download/$ETCD_VERSION/etcd-$ETCD_VERSION-linux-$arch.tar.gz \
           -c qsctl-config.yaml
   done

   rm -rf binaries
fi

# Sync CNI Binary
if [ $CNI_VERSION ]; then
   for arch in ${ARCHS[@]}
   do
     mkdir -p binaries/cni/$CNI_VERSION/$arch
     echo "Synchronizing cni-$arch"

     curl -L -o binaries/cni/$CNI_VERSION/$arch/cni-plugins-linux-$arch-$CNI_VERSION.tgz \
                https://github.com/containernetworking/plugins/releases/download/$CNI_VERSION/cni-plugins-linux-$arch-$CNI_VERSION.tgz

     qsctl cp binaries/cni/$CNI_VERSION/$arch/cni-plugins-linux-$arch-$CNI_VERSION.tgz \
           qs://containernetworking/plugins/releases/download/$CNI_VERSION/cni-plugins-linux-$arch-$CNI_VERSION.tgz \
           -c qsctl-config.yaml
   done

   rm -rf binaries
fi

# Sync CALICOCTL Binary
if [ $CALICO_VERSION ]; then
   for arch in ${ARCHS[@]}
   do
     mkdir -p binaries/calicoctl/$CALICO_VERSION/$arch
     echo "Synchronizing calicoctl-$arch"

     curl -L -o binaries/calicoctl/$CALICO_VERSION/$arch/calicoctl-linux-$arch \
                https://github.com/projectcalico/calico/releases/download/$CALICO_VERSION/calicoctl-linux-$arch

     sha256sum binaries/calicoctl/$CALICO_VERSION/$arch/calicoctl-linux-$arch

     qsctl cp binaries/calicoctl/$CALICO_VERSION/$arch/calicoctl-linux-$arch \
           qs://kubernetes-release/projectcalico/calico/releases/download/$CALICO_VERSION/calicoctl-linux-$arch \
           -c qsctl-config.yaml
   done

   rm -rf binaries
fi

# Sync crictl Binary
if [ $CRICTL_VERSION ]; then
   echo "access_key_id: $ACCESS_KEY_ID" > qsctl-config.yaml
   echo "secret_access_key: $SECRET_ACCESS_KEY" >> qsctl-config.yaml

   for arch in ${ARCHS[@]}
   do
     mkdir -p binaries/crictl/$CRICTL_VERSION/$arch
     echo "Synchronizing crictl-$arch"

     sha256sum binaries/crictl/$CRICTL_VERSION/$arch/crictl-$CRICTL_VERSION-linux-$arch.tar.gz

     curl -L -o binaries/crictl/$CRICTL_VERSION/$arch/crictl-$CRICTL_VERSION-linux-$arch.tar.gz \
                https://github.com/kubernetes-sigs/cri-tools/releases/download/$CRICTL_VERSION/crictl-$CRICTL_VERSION-linux-$arch.tar.gz

     qsctl cp binaries/crictl/$CRICTL_VERSION/$arch/crictl-$CRICTL_VERSION-linux-$arch.tar.gz \
           qs://kubernetes-release/cri-tools/releases/download/$CRICTL_VERSION/crictl-$CRICTL_VERSION-linux-$arch.tar.gz \
           -c qsctl-config.yaml
   done

   rm -rf binaries
fi

# Sync k3s Binary
if [ $K3S_VERSION ]; then
   for arch in ${ARCHS[@]}
   do
     mkdir -p binaries/k3s/$K3S_VERSION/$arch
     echo "Synchronizing k3s-$arch"
     if [ $arch != "amd64" ]; then
        curl -L -o binaries/k3s/$K3S_VERSION/$arch/k3s \
                   https://github.com/rancher/k3s/releases/download/$K3S_VERSION+k3s1/k3s-$arch
     else
        curl -L -o binaries/k3s/$K3S_VERSION/$arch/k3s \
                   https://github.com/rancher/k3s/releases/download/$K3S_VERSION+k3s1/k3s
     fi
     qsctl cp binaries/k3s/$K3S_VERSION/$arch/k3s \
           qs://kubernetes-release/k3s/releases/download/$K3S_VERSION+k3s1/linux/$arch/k3s \
           -c qsctl-config.yaml
   done

   rm -rf binaries
fi

# Sync containerd Binary
if [ $CONTAINERD_VERSION ]; then
   for arch in ${ARCHS[@]}
   do
     mkdir -p binaries/containerd/$CONTAINERD_VERSION/$arch
     echo "Synchronizing containerd-$arch"

     curl -L -o binaries/containerd/$CONTAINERD_VERSION/$arch/containerd-$CONTAINERD_VERSION-linux-$arch.tar.gz \
                https://github.com/containerd/containerd/releases/download/v$CONTAINERD_VERSION/containerd-$CONTAINERD_VERSION-linux-$arch.tar.gz

     sha256sum binaries/containerd/$CONTAINERD_VERSION/$arch/containerd-$CONTAINERD_VERSION-linux-$arch.tar.gz

     qsctl cp binaries/containerd/$CONTAINERD_VERSION/$arch/containerd-$CONTAINERD_VERSION-linux-$arch.tar.gz \
           qs://kubernetes-release/containerd/containerd/releases/download/v$CONTAINERD_VERSION/containerd-$CONTAINERD_VERSION-linux-$arch.tar.gz \
           -c qsctl-config.yaml
   done

   rm -rf binaries
fi

# Sync runc Binary
if [ $RUNC_VERSION ]; then
   for arch in ${ARCHS[@]}
   do
     mkdir -p binaries/runc/$RUNC_VERSION/$arch
     echo "Synchronizing runc-$arch"

     curl -L -o binaries/runc/$RUNC_VERSION/$arch/runc.$arch \
                https://github.com/opencontainers/runc/releases/download/$RUNC_VERSION/runc.$arch

     sha256sum binaries/runc/$RUNC_VERSION/$arch/runc.$arch

     qsctl cp binaries/runc/$RUNC_VERSION/$arch/runc.$arch \
           qs://kubernetes-release/opencontainers/runc/releases/download/$RUNC_VERSION/runc.$arch \
           -c qsctl-config.yaml
   done

   rm -rf binaries
fi

# Sync docker-compose Binary
if [ $COMPOSE_VERSION ]; then
   for arch in ${ARCHS[@]}
   do
     mkdir -p binaries/compose/$COMPOSE_VERSION/$arch
     echo "Synchronizing runc-$arch"
     if [ $arch == "amd64" ]; then
        curl -L -o binaries/compose/$COMPOSE_VERSION/$arch/docker-compose-linux-x86_64 \
                   https://github.com/docker/compose/releases/download/$COMPOSE_VERSION/docker-compose-linux-x86_64

        qsctl cp binaries/compose/$COMPOSE_VERSION/$arch/docker-compose-linux-x86_64 \
              qs://kubernetes-release/docker/compose/releases/download/$COMPOSE_VERSION/docker-compose-linux-x86_64 \
              -c qsctl-config.yaml

     elif [ $arch == "arm64" ]; then
        curl -L -o binaries/compose/$COMPOSE_VERSION/$arch/docker-compose-linux-aarch64 \
                   https://github.com/docker/compose/releases/download/$COMPOSE_VERSION/docker-compose-linux-aarch64

        qsctl cp binaries/compose/$COMPOSE_VERSION/$arch/docker-compose-linux-aarch64 \
              qs://kubernetes-release/docker/compose/releases/download/$COMPOSE_VERSION/docker-compose-linux-aarch64 \
              -c qsctl-config.yaml

     fi
   done

   rm -rf binaries
fi

# Sync CRI_DDOCKER Binary
if [ $CRI_DOCKER_VERSION ]; then
   for arch in ${ARCHS[@]}
   do
     mkdir -p binaries/cri-dockerd/$CRI_DOCKER_VERSION/$arch
     echo "Synchronizing cri-dockerd-$arch"

     curl -L -o binaries/cri-dockerd/$CRI_DOCKER_VERSION/$arch/cri-dockerd-$CRI_DOCKER_VERSION.$arch.tgz \
                https://github.com/Mirantis/cri-dockerd/releases/download/v$CRI_DOCKER_VERSION/cri-dockerd-$CRI_DOCKER_VERSION.$arch.tgz

     sha256sum binaries/cri-dockerd/$CRI_DOCKER_VERSION/$arch/cri-dockerd-$CRI_DOCKER_VERSION.$arch.tgz

     qsctl cp binaries/cri-dockerd/$CRI_DOCKER_VERSION/$arch/cri-dockerd-$CRI_DOCKER_VERSION.$arch.tgz \
           qs://kubernetes-release/cri-dockerd/releases/download/v$CRI_DOCKER_VERSION/cri-dockerd-$CRI_DOCKER_VERSION.$arch.tgz \
           -c qsctl-config.yaml
   done

   rm -rf binaries
fi

rm -rf qsctl-config.yaml

# Sync NodeLocalDns Images
if [ $NODE_LOCAL_DNS_VERSION ]; then
   skopeo sync --src docker --dest docker registry.k8s.io/dns/k8s-dns-node-cache:$NODE_LOCAL_DNS_VERSION docker.io/$DOCKERHUB_NAMESPACE/k8s-dns-node-cache:$NODE_LOCAL_DNS_VERSION --all
   skopeo sync --src docker --dest docker registry.k8s.io/dns/k8s-dns-node-cache:$NODE_LOCAL_DNS_VERSION registry.cn-beijing.aliyuncs.com/$ALIYUNCS_NAMESPACE/k8s-dns-node-cache:$NODE_LOCAL_DNS_VERSION --all
fi

# Sync Coredns Images
if [ $COREDNS_VERSION ]; then
   skopeo sync --src docker --dest docker docker.io/coredns/coredns:$COREDNS_VERSION registry.cn-beijing.aliyuncs.com/$ALIYUNCS_NAMESPACE/coredns:$COREDNS_VERSION --all
fi

# Sync Calico Images
if [ $CALICO_VERSION ]; then
   skopeo sync --src docker --dest docker docker.io/calico/kube-controllers:$CALICO_VERSION registry.cn-beijing.aliyuncs.com/$ALIYUNCS_NAMESPACE/kube-controllers:$CALICO_VERSION --all
   skopeo sync --src docker --dest docker docker.io/calico/cni:$CALICO_VERSION registry.cn-beijing.aliyuncs.com/$ALIYUNCS_NAMESPACE/cni:$CALICO_VERSION --all
   skopeo sync --src docker --dest docker docker.io/calico/node:$CALICO_VERSION registry.cn-beijing.aliyuncs.com/$ALIYUNCS_NAMESPACE/node:$CALICO_VERSION --all
   skopeo sync --src docker --dest docker docker.io/calico/pod2daemon-flexvol:$CALICO_VERSION registry.cn-beijing.aliyuncs.com/$ALIYUNCS_NAMESPACE/pod2daemon-flexvol:$CALICO_VERSION --all
   skopeo sync --src docker --dest docker docker.io/calico/typha:$CALICO_VERSION registry.cn-beijing.aliyuncs.com/$ALIYUNCS_NAMESPACE/typha:$CALICO_VERSION --all
fi

# Sync Kube-OVN Images
if [ $KUBE_OVN_VERSION ]; then
   skopeo sync --src docker --dest docker docker.io/kubeovn/kube-ovn:$KUBE_OVN_VERSION registry.cn-beijing.aliyuncs.com/$ALIYUNCS_NAMESPACE/kube-ovn:$KUBE_OVN_VERSION --all
   skopeo sync --src docker --dest docker docker.io/kubeovn/vpc-nat-gateway:$KUBE_OVN_VERSION registry.cn-beijing.aliyuncs.com/$ALIYUNCS_NAMESPACE/vpc-nat-gateway:$KUBE_OVN_VERSION --all
fi

# Sync Cilium Images
if [ $CILIUM_VERSION ]; then
   skopeo sync --src docker --dest docker docker.io/cilium/cilium:$CILIUM_VERSION registry.cn-beijing.aliyuncs.com/$ALIYUNCS_NAMESPACE/cilium:$CILIUM_VERSION --all
   skopeo sync --src docker --dest docker docker.io/cilium/cilium-operator-generic:$CILIUM_VERSION registry.cn-beijing.aliyuncs.com/$ALIYUNCS_NAMESPACE/cilium-operator-generic:$CILIUM_VERSION --all
fi

# Sync OpenEBS Images
if [ $OPENEBS_VERSION ]; then
   skopeo sync --src docker --dest docker docker.io/openebs/provisioner-localpv:$OPENEBS_VERSION registry.cn-beijing.aliyuncs.com/$ALIYUNCS_NAMESPACE/provisioner-localpv:$OPENEBS_VERSION --all
   skopeo sync --src docker --dest docker docker.io/openebs/linux-utils:$OPENEBS_VERSION registry.cn-beijing.aliyuncs.com/$ALIYUNCS_NAMESPACE/linux-utils:$OPENEBS_VERSION --all
fi

# Sync Haproxy Images
if [ $HAPROXY_VERSION ]; then
   skopeo sync --src docker --dest docker docker.io/library/haproxy:$HAPROXY_VERSION registry.cn-beijing.aliyuncs.com/$ALIYUNCS_NAMESPACE/haproxy:$HAPROXY_VERSION --all
fi

# Sync Kube-vip Images
if [ $KUBEVIP_VERSION ]; then
   skopeo sync --src docker --dest docker docker.io/plndr/kubevip:$KUBEVIP_VERSION registry.cn-beijing.aliyuncs.com/$ALIYUNCS_NAMESPACE/kubevip:$KUBEVIP_VERSION --all
fi
