package v1alpha1

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// K2ClusterSpec defines the desired state of K2Cluster
type K2ClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Hosts           []HostCfg          `yaml:"hosts" json:"hosts,omitempty"`
	LBKubeApiserver LBKubeApiserverCfg `yaml:"lbKubeapiserver" json:"lbKubeapiserver,omitempty"`
	KubeCluster     KubeCluster        `yaml:"kubeCluster" json:"kubeCluster,omitempty"`
	Network         NetworkConfig      `yaml:"network" json:"network,omitempty"`
	Registry        RegistryConfig     `yaml:"registry" json:"registry,omitempty"`
	Plugins         PluginsCfg         `yaml:"plugins" json:"plugins,omitempty"`
}

type KubeCluster struct {
	Version     string `yaml:"version" json:"version,omitempty"`
	ImageRepo   string `yaml:"imageRepo" json:"imageRepo,omitempty"`
	ClusterName string `yaml:"clusterName" json:"clusterName,omitempty"`
}

// K2ClusterStatus defines the observed state of K2Cluster
type K2ClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// K2Cluster is the Schema for the k2clusters API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=k2clusters,scope=Namespaced
type K2Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   K2ClusterSpec   `json:"spec,omitempty"`
	Status K2ClusterStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// K2ClusterList contains a list of K2Cluster
type K2ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []K2Cluster `json:"items"`
}

//func init() {
//	SchemeBuilder.Register(&K2Cluster{}, &K2ClusterList{})
//}

type HostCfg struct {
	HostName        string   `yaml:"hostName" json:"hostName,omitempty"`
	SSHAddress      string   `yaml:"sshAddress" json:"sshAddress,omitempty"`
	InternalAddress string   `yaml:"internalAddress" json:"internalAddress,omitempty"`
	Port            string   `yaml:"port" json:"port,omitempty"`
	User            string   `yaml:"user" json:"user,omitempty"`
	Password        string   `yaml:"password" json:"password,omitempty"`
	SSHKeyPath      string   `yaml:"sshKeyPath" json:"sshKeyPath,omitempty"`
	Role            []string `yaml:"role" json:"role,omitempty" norman:"type=array[enum],options=etcd|master|worker|client"`
	ID              int      `json:"-"`
	IsEtcd          bool     `json:"-"`
	IsMaster        bool     `json:"-"`
	IsWorker        bool     `json:"-"`
	IsClient        bool     `json:"-"`
}

type Hosts struct {
	Hosts []HostCfg
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

type RegistryConfig struct {
	RegistryMirrors    []string `yaml:"registryMirrors" json:"registryMirrors,omitempty"`
	InsecureRegistries []string `yaml:"insecureRegistries" json:"insecureRegistries,omitempty"`
	PrivateRegistry    string   `yaml:"privateRegistry" json:"privateRegistry,omitempty"`
}

type ExternalEtcd struct {
	Endpoints []string
	CaFile    string
	CertFile  string
	KeyFile   string
}

func (cfg *K2ClusterSpec) GenerateHosts() []string {
	var lbHost string
	hostsList := []string{}

	_, _, masters, _, _, _ := cfg.GroupHosts()
	if cfg.LBKubeApiserver.Address != "" {
		lbHost = fmt.Sprintf("%s  %s", cfg.LBKubeApiserver.Address, cfg.LBKubeApiserver.Domain)
	} else {
		lbHost = fmt.Sprintf("%s  %s", masters.Hosts[0].InternalAddress, DefaultLBDomain)
	}

	for _, host := range cfg.Hosts {
		if host.HostName != "" {
			hostsList = append(hostsList, fmt.Sprintf("%s  %s.%s %s", host.InternalAddress, host.HostName, cfg.KubeCluster.ClusterName, host.HostName))
		}
	}

	hostsList = append(hostsList, lbHost)
	return hostsList
}

func (cfg *K2ClusterSpec) GenerateCertSANs() []string {
	clusterSvc := fmt.Sprintf("kubernetes.default.svc.%s", cfg.KubeCluster.ClusterName)
	defaultCertSANs := []string{"kubernetes", "kubernetes.default", "kubernetes.default.svc", clusterSvc, "localhost", "127.0.0.1"}
	extraCertSANs := []string{}

	extraCertSANs = append(extraCertSANs, cfg.LBKubeApiserver.Domain)
	extraCertSANs = append(extraCertSANs, cfg.LBKubeApiserver.Address)

	for _, host := range cfg.Hosts {
		extraCertSANs = append(extraCertSANs, host.HostName)
		extraCertSANs = append(extraCertSANs, fmt.Sprintf("%s.%s", host.HostName, cfg.KubeCluster.ClusterName))
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

func (cfg *K2ClusterSpec) GroupHosts() (*Hosts, *Hosts, *Hosts, *Hosts, *Hosts, *Hosts) {
	allHosts := Hosts{}
	etcdHosts := Hosts{}
	masterHosts := Hosts{}
	workerHosts := Hosts{}
	k8sHosts := Hosts{}
	clientHost := Hosts{}

	for _, host := range cfg.Hosts {
		//clusterNode := HostCfg{}
		if host.IsEtcd {
			etcdHosts.Hosts = append(etcdHosts.Hosts, host)
		}
		if host.IsMaster {
			masterHosts.Hosts = append(masterHosts.Hosts, host)
		}
		if host.IsWorker {
			workerHosts.Hosts = append(workerHosts.Hosts, host)
		}
		if host.IsMaster || host.IsWorker {
			k8sHosts.Hosts = append(k8sHosts.Hosts, host)
		}
		if host.IsClient {
			clientHost.Hosts = append(clientHost.Hosts, host)
		}
		allHosts.Hosts = append(allHosts.Hosts, host)
	}
	return &allHosts, &etcdHosts, &masterHosts, &workerHosts, &k8sHosts, &clientHost
}

func (cfg *K2ClusterSpec) ClusterIP() string {
	return util.ParseIp(cfg.Network.KubeServiceCIDR)[2]
}
