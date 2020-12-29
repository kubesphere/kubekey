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
	"strings"
	"text/template"

	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/cluster/preinstall"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/lithammer/dedent"
	"github.com/pkg/errors"
)

// KubeadmCfgTempl defines the template of kubeadm configuration file.
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
  imageRepository: {{ .CorednsRepo }}
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
    anonymous-auth: "true"
    bind-address: 0.0.0.0
    insecure-port: "0"
    profiling: "false"
    apiserver-count: "1"
    endpoint-reconciler-type: lease
    authorization-mode: Node,RBAC
    enable-aggregator-routing: "false"
    allow-privileged: "true"
    storage-backend: etcd3
    audit-log-maxage: "30"
    audit-log-maxbackup: "10"
    audit-log-maxsize: "100"
    audit-log-path: /var/log/apiserver/audit.log
    feature-gates: CSINodeInfo=true,VolumeSnapshotDataSource=true,ExpandCSIVolumes=true,RotateKubeletClientCertificate=true
  certSANs:
    {{- range .CertSANs }}
    - {{ . }}
    {{- end }}
controllerManager:
  extraArgs:
    node-cidr-mask-size: "{{ .NodeCidrMaskSize }}"
    experimental-cluster-signing-duration: 87600h
    bind-address: 127.0.0.1
    profiling: "false"
    port: "10252"
    terminated-pod-gc-threshold: "10"
    feature-gates: RotateKubeletServerCertificate=true,CSINodeInfo=true,VolumeSnapshotDataSource=true,ExpandCSIVolumes=true,RotateKubeletClientCertificate=true
  extraVolumes:
  - name: host-time
    hostPath: /etc/localtime
    mountPath: /etc/localtime
    readOnly: true
scheduler:
  extraArgs:
    profiling: "false"
    bind-address: 127.0.0.1
    port: "10251"
    feature-gates: CSINodeInfo=true,VolumeSnapshotDataSource=true,ExpandCSIVolumes=true,RotateKubeletClientCertificate=true

{{- if .CriSock }}
---
apiVersion: kubeadm.k8s.io/v1beta2
kind: InitConfiguration
nodeRegistration:
  criSocket: {{ .CriSock }}
{{- end }}
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
{{- if .CriSock }}
containerLogMaxSize: 5Mi
containerLogMaxFiles: 3
{{- if .CgroupDriver }}
cgroupDriver: systemd
{{- end }}
{{- end }}
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

// GenerateKubeadmCfg create kubeadm configuration file to initialize the cluster.
func GenerateKubeadmCfg(mgr *manager.Manager) (string, error) {
	// generate etcd configuration
	var externalEtcd kubekeyapiv1alpha1.ExternalEtcd
	var endpointsList []string
	var caFile, certFile, keyFile, containerRuntimeEndpoint string

	for _, host := range mgr.EtcdNodes {
		endpoint := fmt.Sprintf("https://%s:%s", host.InternalAddress, kubekeyapiv1alpha1.DefaultEtcdPort)
		endpointsList = append(endpointsList, endpoint)
	}
	externalEtcd.Endpoints = endpointsList

	caFile = "/etc/ssl/etcd/ssl/ca.pem"
	certFile = fmt.Sprintf("/etc/ssl/etcd/ssl/node-%s.pem", mgr.MasterNodes[0].Name)
	keyFile = fmt.Sprintf("/etc/ssl/etcd/ssl/node-%s-key.pem", mgr.MasterNodes[0].Name)

	externalEtcd.CaFile = caFile
	externalEtcd.CertFile = certFile
	externalEtcd.KeyFile = keyFile

	// generate cri configuration
	switch mgr.Cluster.Kubernetes.ContainerManager {
	case "docker":
		containerRuntimeEndpoint = ""
	case "crio":
		containerRuntimeEndpoint = kubekeyapiv1alpha1.DefaultCrioEndpoint
	case "containerd":
		containerRuntimeEndpoint = kubekeyapiv1alpha1.DefaultContainerdEndpoint
	case "isula":
		containerRuntimeEndpoint = kubekeyapiv1alpha1.DefaultIsulaEndpoint
	default:
		containerRuntimeEndpoint = ""
	}

	if mgr.Cluster.Kubernetes.ContainerRuntimeEndpoint != "" {
		containerRuntimeEndpoint = mgr.Cluster.Kubernetes.ContainerRuntimeEndpoint
	}

	cgroupDriver, err := getKubeletCgroupDriver(mgr)
	if err != nil {
		return "", err
	}

	return util.Render(KubeadmCfgTempl, util.Data{
		"ImageRepo":            strings.TrimSuffix(preinstall.GetImage(mgr, "kube-apiserver").ImageRepo(), "/kube-apiserver"),
		"CorednsRepo":          strings.TrimSuffix(preinstall.GetImage(mgr, "coredns").ImageRepo(), "/coredns"),
		"CorednsTag":           preinstall.GetImage(mgr, "coredns").Tag,
		"Version":              mgr.Cluster.Kubernetes.Version,
		"ClusterName":          mgr.Cluster.Kubernetes.ClusterName,
		"ControlPlaneEndpoint": fmt.Sprintf("%s:%d", mgr.Cluster.ControlPlaneEndpoint.Domain, mgr.Cluster.ControlPlaneEndpoint.Port),
		"PodSubnet":            mgr.Cluster.Network.KubePodsCIDR,
		"ServiceSubnet":        mgr.Cluster.Network.KubeServiceCIDR,
		"CertSANs":             mgr.Cluster.GenerateCertSANs(),
		"ExternalEtcd":         externalEtcd,
		"ClusterIP":            "169.254.25.10",
		"MasqueradeAll":        mgr.Cluster.Kubernetes.MasqueradeAll,
		"NodeCidrMaskSize":     mgr.Cluster.Kubernetes.NodeCidrMaskSize,
		"MaxPods":              mgr.Cluster.Kubernetes.MaxPods,
		"ProxyMode":            mgr.Cluster.Kubernetes.ProxyMode,
		"CriSock":              containerRuntimeEndpoint,
		"CgroupDriver":         cgroupDriver,
	})
}

func getKubeletCgroupDriver(mgr *manager.Manager) (string, error) {
	var cmd, kubeletCgroupDriver string
	switch mgr.Cluster.Kubernetes.ContainerManager {
	case "docker":
		cmd = "docker info | grep 'Cgroup Driver' | awk -F': ' '{ print $2; }'"
	case "crio":
		cmd = "crio config | grep cgroup_manager | awk -F'= ' '{ print $2; }'"
	case "containerd":
		cmd = "containerd config dump | grep systemd_cgroup | awk -F'= ' '{ print $2; }'"
	case "isula":
		cmd = "isula info | grep 'Cgroup Driver' | awk -F': ' '{ print $2; }'"
	default:
		kubeletCgroupDriver = ""
	}

	checkResult, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo env PATH=$PATH /bin/sh -c \"%s\"", cmd), 3, false)
	if err != nil {
		return "", errors.Wrap(errors.WithStack(err), "Failed to get container runtime cgroup driver.")
	}
	if strings.Contains(checkResult, "systemd") && !strings.Contains(checkResult, "false") {
		kubeletCgroupDriver = "systemd"
	} else {
		kubeletCgroupDriver = ""
	}

	return kubeletCgroupDriver, nil
}
