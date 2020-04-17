package v1alpha1

import (
	"fmt"
	"github.com/pixiake/kubekey/util"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	DefaultPreDir         = "kubekey"
	DefaultSSHPort        = "22"
	DefaultDockerSockPath = "/var/run/docker.sock"
	DefaultLBPort         = "6443"
	DefaultLBDomain       = "lb.kubesphere.local"
	DefaultNetworkPlugin  = "calico"
	DefaultPodsCIDR       = "10.233.64.0/18"
	DefaultServiceCIDR    = "10.233.0.0/18"
	DefaultKubeImageRepo  = "kubekey"
	DefaultClusterName    = "cluster.local"
	DefaultArch           = "amd64"
	DefaultHostName       = "allinone"
	DefaultEtcdRepo       = "kubekey/etcd"
	DefaultEtcdVersion    = "v3.3.12"
	DefaultEtcdPort       = "2379"
	DefaultKubeVersion    = "v1.17.4"
	DefaultCniVersion     = "v0.8.2"
	DefaultHelmVersion    = "v3.1.2"
	ETCDRole              = "etcd"
	MasterRole            = "master"
	WorkerRole            = "worker"
)

type ClusterCfg struct {
	Hosts           []HostCfg          `yaml:"hosts" json:"hosts,omitempty"`
	LBKubeApiserver LBKubeApiserverCfg `yaml:"lbKubeapiserver" json:"lbKubeapiserver,omitempty"`
	KubeVersion     string             `yaml:"kubeVersion" json:"kubeVersion,omitempty"`
	KubeImageRepo   string             `yaml:"kubeImageRepo" json:"kubeImageRepo,omitempty"`
	KubeClusterName string             `yaml:"kubeClusterName" json:"kubeClusterName,omitempty"`
	Network         NetworkConfig      `yaml:"network" json:"network,omitempty"`
}

func (c ClusterCfg) GetObjectKind() schema.ObjectKind {
	panic("implement me")
}

func (c ClusterCfg) DeepCopyObject() runtime.Object {
	panic("implement me")
}

type Taint struct {
	Key    string      `json:"key,omitempty" yaml:"key"`
	Value  string      `json:"value,omitempty" yaml:"value"`
	Effect TaintEffect `json:"effect,omitempty" yaml:"effect"`
}

type TaintEffect string

const (
	TaintEffectNoSchedule       TaintEffect = "NoSchedule"
	TaintEffectPreferNoSchedule TaintEffect = "PreferNoSchedule"
	TaintEffectNoExecute        TaintEffect = "NoExecute"
)

type NodeInfo struct {
	HostName string
}

type NetworkConfig struct {
	Plugin          string `yaml:"plugin" json:"plugin,omitempty"`
	KubePodsCIDR    string `yaml:"kube_pods_cidr" json:"kube_pods_cidr,omitempty"`
	KubeServiceCIDR string `yaml:"kube_service_cidr" json:"kube_service_cidr,omitempty"`
}

type LBKubeApiserverCfg struct {
	Domain  string `yaml:"domain" json:"domain,omitempty"`
	Address string `yaml:"address" json:"address,omitempty"`
	Port    string `yaml:"port" json:"port,omitempty"`
}

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

type ExternalEtcd struct {
	Endpoints []string
	CaFile    string
	CertFile  string
	KeyFile   string
}

func RegisterDefaults(scheme *runtime.Scheme) error {
	scheme.AddTypeDefaultingFunc(&ClusterCfg{}, func(obj interface{}) { SetDefaultClusterCfg(obj.(*ClusterCfg)) })
	return nil
}

func (cfg *ClusterCfg) GenerateHosts() []string {
	var lbHost string
	hostsList := []string{}

	_, _, masters, _, _ := cfg.GroupHosts()
	if cfg.LBKubeApiserver.Address != "" {
		lbHost = fmt.Sprintf("%s  %s", cfg.LBKubeApiserver.Address, cfg.LBKubeApiserver.Domain)
	} else {
		lbHost = fmt.Sprintf("%s  %s", masters.Hosts[0].InternalAddress, DefaultLBDomain)
	}

	for _, host := range cfg.Hosts {
		if host.HostName != "" {
			hostsList = append(hostsList, fmt.Sprintf("%s  %s.%s %s", host.InternalAddress, host.HostName, cfg.KubeClusterName, host.HostName))
		}
	}

	hostsList = append(hostsList, lbHost)
	return hostsList
}

func (cfg *ClusterCfg) GenerateCertSANs() []string {
	clusterSvc := fmt.Sprintf("kubernetes.default.svc.%s", cfg.KubeClusterName)
	defaultCertSANs := []string{"kubernetes", "kubernetes.default", "kubernetes.default.svc", clusterSvc, "localhost", "127.0.0.1"}
	extraCertSANs := []string{}

	extraCertSANs = append(extraCertSANs, cfg.LBKubeApiserver.Domain)
	extraCertSANs = append(extraCertSANs, cfg.LBKubeApiserver.Address)

	for _, host := range cfg.Hosts {
		extraCertSANs = append(extraCertSANs, host.HostName)
		extraCertSANs = append(extraCertSANs, fmt.Sprintf("%s.%s", host.HostName, cfg.KubeClusterName))
		if host.SSHAddress != cfg.LBKubeApiserver.Address {
			extraCertSANs = append(extraCertSANs, host.SSHAddress)
		}
		if host.InternalAddress != host.SSHAddress && host.InternalAddress != cfg.LBKubeApiserver.Address {
			extraCertSANs = append(extraCertSANs, host.InternalAddress)
		}
	}

	extraCertSANs = append(extraCertSANs, util.ParseIp(cfg.Network.KubeServiceCIDR)[0])

	defaultCertSANs = append(defaultCertSANs, extraCertSANs...)
	return defaultCertSANs
}
