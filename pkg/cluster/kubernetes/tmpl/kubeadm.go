/*
Copyright 2020 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tmpl

import (
	"fmt"
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/api/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/cluster/preinstall"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/lithammer/dedent"
	"text/template"
)

var KubeadmCfgTempl = template.Must(template.New("kubeadmCfg").Parse(
	dedent.Dedent(`---
apiVersion: kubeadm.k8s.io/v1beta2
kind: ClusterConfiguration
etcd:
  external:
    endpoints:
    {{- range .ExternalEtcd.Endpoints }}
    - {{ . }}
    {{- end }}
    caFile: {{ .ExternalEtcd.CaFile }}
    certFile: {{ .ExternalEtcd.CertFile }}
    keyFile: {{ .ExternalEtcd.KeyFile }}
dns:
  type: CoreDNS
  imageRepository: {{ .CorednsRepo }}coredns
  imageTag: {{ .CorednsTag }}
imageRepository: {{ .ImageRepo }}
kubernetesVersion: {{ .Version }}
certificatesDir: /etc/kubernetes/pki
clusterName: {{ .ClusterName }}
controlPlaneEndpoint: {{ .ControlPlaneEndpoint }}
networking:
  dnsDomain: {{ .ClusterName }}
  podSubnet: {{ .PodSubnet }}
  serviceSubnet: {{ .ServiceSubnet }}
apiServer:
  extraArgs:
    anonymous-auth: "True"
    bind-address: 0.0.0.0
    insecure-port: "0"
    profiling: "False"
    apiserver-count: "1"
    endpoint-reconciler-type: lease
    authorization-mode: Node,RBAC
    enable-aggregator-routing: "False"
    allow-privileged: "true"
    storage-backend: etcd3
    feature-gates: CSINodeInfo=true,VolumeSnapshotDataSource=true,ExpandCSIVolumes=true,RotateKubeletClientCertificate=true
  certSANs:
    {{- range .CertSANs }}
    - {{ . }}
    {{- end }}
controllerManager:
  extraArgs:
    node-cidr-mask-size: "{{ .NodeCidrMaskSize }}"
    experimental-cluster-signing-duration: 87600h
    bind-address: 0.0.0.0
    profiling: "False"
    port: "10252"
    feature-gates: CSINodeInfo=true,VolumeSnapshotDataSource=true,ExpandCSIVolumes=true,RotateKubeletClientCertificate=true
  extraVolumes:
  - name: host-time
    hostPath: /etc/localtime
    mountPath: /etc/localtime
    readOnly: true
scheduler:
  extraArgs:
    bind-address: 0.0.0.0
    port: "10251"
    feature-gates: CSINodeInfo=true,VolumeSnapshotDataSource=true,ExpandCSIVolumes=true,RotateKubeletClientCertificate=true

---
apiVersion: kubeproxy.config.k8s.io/v1alpha1
kind: KubeProxyConfiguration
bindAddress: 0.0.0.0
clientConnection:
 acceptContentTypes: 
 burst: 10
 contentType: application/vnd.kubernetes.protobuf
 kubeconfig: 
 qps: 5
clusterCIDR: {{ .PodSubnet }}
configSyncPeriod: 15m0s
conntrack:
 maxPerCore: 32768
 min: 131072
 tcpCloseWaitTimeout: 1h0m0s
 tcpEstablishedTimeout: 24h0m0s
enableProfiling: False
healthzBindAddress: 0.0.0.0:10256
iptables:
 masqueradeAll: {{ .MasqueradeAll }}
 masqueradeBit: 14
 minSyncPeriod: 0s
 syncPeriod: 30s
ipvs:
 excludeCIDRs: []
 minSyncPeriod: 0s
 scheduler: rr
 syncPeriod: 30s
 strictARP: False
mode: {{ .ProxyMode }}

---
apiVersion: kubelet.config.k8s.io/v1beta1
kind: KubeletConfiguration
clusterDomain: {{ .ClusterName }}
clusterDNS:
- {{ .ClusterIP }}
maxPods: {{ .MaxPods }}
rotateCertificates: true
kubeReserved:
  cpu: 200m
  memory: 250Mi
systemReserved:
  cpu: 200m
  memory: 250Mi
evictionHard:
  memory.available: 5%
evictionSoft:
  memory.available: 10%
evictionSoftGracePeriod: 
  memory.available: 2m
evictionMaxPodGracePeriod: 120
evictionPressureTransitionPeriod: 30s
featureGates:
  CSINodeInfo: true
  VolumeSnapshotDataSource: true
  ExpandCSIVolumes: true
  RotateKubeletClientCertificate: true

    `)))

func GenerateKubeadmCfg(mgr *manager.Manager) (string, error) {
	var externalEtcd kubekeyapiv1alpha1.ExternalEtcd
	var endpointsList []string
	var caFile, certFile, keyFile string

	for _, host := range mgr.EtcdNodes {
		endpoint := fmt.Sprintf("https://%s:%s", host.InternalAddress, kubekeyapiv1alpha1.DefaultEtcdPort)
		endpointsList = append(endpointsList, endpoint)
	}
	externalEtcd.Endpoints = endpointsList

	caFile = "/etc/ssl/etcd/ssl/ca.pem"
	certFile = fmt.Sprintf("/etc/ssl/etcd/ssl/node-%s.pem", mgr.EtcdNodes[0].Name)
	keyFile = fmt.Sprintf("/etc/ssl/etcd/ssl/node-%s-key.pem", mgr.EtcdNodes[0].Name)

	externalEtcd.CaFile = caFile
	externalEtcd.CertFile = certFile
	externalEtcd.KeyFile = keyFile

	var imageRepo string
	if mgr.Cluster.Registry.PrivateRegistry != "" {
		imageRepo = fmt.Sprintf("%s/%s", mgr.Cluster.Registry.PrivateRegistry, mgr.Cluster.Kubernetes.ImageRepo)
	} else {
		imageRepo = mgr.Cluster.Kubernetes.ImageRepo
	}

	var corednsRepo string
	if mgr.Cluster.Registry.PrivateRegistry != "" {
		corednsRepo = fmt.Sprintf("%s/", mgr.Cluster.Registry.PrivateRegistry)
	} else {
		corednsRepo = ""
	}
	return util.Render(KubeadmCfgTempl, util.Data{
		"ImageRepo":            imageRepo,
		"CorednsRepo":          corednsRepo,
		"CorednsTag":           preinstall.GetImage(mgr, "coredns").Tag,
		"Version":              mgr.Cluster.Kubernetes.Version,
		"ClusterName":          mgr.Cluster.Kubernetes.ClusterName,
		"ControlPlaneEndpoint": fmt.Sprintf("%s:%s", mgr.Cluster.ControlPlaneEndpoint.Domain, mgr.Cluster.ControlPlaneEndpoint.Port),
		"PodSubnet":            mgr.Cluster.Network.KubePodsCIDR,
		"ServiceSubnet":        mgr.Cluster.Network.KubeServiceCIDR,
		"CertSANs":             mgr.Cluster.GenerateCertSANs(),
		"ExternalEtcd":         externalEtcd,
		"ClusterIP":            "169.254.25.10",
		"MasqueradeAll":        mgr.Cluster.Kubernetes.MasqueradeAll,
		"NodeCidrMaskSize":     mgr.Cluster.Kubernetes.NodeCidrMaskSize,
		"MaxPods":              mgr.Cluster.Kubernetes.MaxPods,
		"ProxyMode":            mgr.Cluster.Kubernetes.ProxyMode,
	})
}
