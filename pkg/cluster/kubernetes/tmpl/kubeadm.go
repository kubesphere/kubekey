package tmpl

import (
	"fmt"
	"github.com/lithammer/dedent"
	kubekeyapi "github.com/pixiake/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/pixiake/kubekey/pkg/util"
	"github.com/pixiake/kubekey/pkg/util/manager"
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
  imageRepository: coredns
  imageTag: 1.6.0
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
    authorization-mode: Node,RBAC
  timeoutForControlPlane: 4m0s
  certSANs:
    {{- range .CertSANs }}
    - {{ . }}
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
 masqueradeAll: False
 masqueradeBit: 14
 minSyncPeriod: 0s
 syncPeriod: 30s
ipvs:
 excludeCIDRs: []
 minSyncPeriod: 0s
 scheduler: rr
 syncPeriod: 30s
 strictARP: False
metricsBindAddress: 127.0.0.1:10249
mode: ipvs
nodePortAddresses: []
oomScoreAdj: -999
portRange: 
udpIdleTimeout: 250ms
    `)))

func GenerateKubeadmCfg(mgr *manager.Manager) (string, error) {
	var externalEtcd kubekeyapi.ExternalEtcd
	var endpointsList []string
	var caFile, certFile, keyFile string

	for _, host := range mgr.EtcdNodes.Hosts {
		endpoint := fmt.Sprintf("https://%s:%s", host.InternalAddress, kubekeyapi.DefaultEtcdPort)
		endpointsList = append(endpointsList, endpoint)
	}
	externalEtcd.Endpoints = endpointsList

	caFile = "/etc/ssl/etcd/ssl/ca.pem"
	certFile = fmt.Sprintf("/etc/ssl/etcd/ssl/admin-%s.pem", mgr.EtcdNodes.Hosts[0].HostName)
	keyFile = fmt.Sprintf("/etc/ssl/etcd/ssl/admin-%s-key.pem", mgr.EtcdNodes.Hosts[0].HostName)

	externalEtcd.CaFile = caFile
	externalEtcd.CertFile = certFile
	externalEtcd.KeyFile = keyFile

	return util.Render(KubeadmCfgTempl, util.Data{
		"ImageRepo":            mgr.Cluster.KubeImageRepo,
		"Version":              mgr.Cluster.KubeVersion,
		"ClusterName":          mgr.Cluster.KubeClusterName,
		"ControlPlaneEndpoint": fmt.Sprintf("%s:%s", mgr.Cluster.LBKubeApiserver.Address, mgr.Cluster.LBKubeApiserver.Port),
		"PodSubnet":            mgr.Cluster.Network.KubePodsCIDR,
		"ServiceSubnet":        mgr.Cluster.Network.KubeServiceCIDR,
		"CertSANs":             mgr.Cluster.GenerateCertSANs(),
		"ExternalEtcd":         externalEtcd,
	})
}
