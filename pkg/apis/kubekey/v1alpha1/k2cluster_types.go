package v1alpha1

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"regexp"
	"strconv"
	"strings"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// K2ClusterSpec defines the desired state of K2Cluster
type K2ClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Hosts                []HostCfg            `yaml:"hosts" json:"hosts,omitempty"`
	RoleGroups           RoleGroups           `yaml:"roleGroups" json:"roleGroups,omitempty"`
	ControlPlaneEndpoint ControlPlaneEndpoint `yaml:"controlPlaneEndpoint" json:"controlPlaneEndpoint,omitempty"`
	Kubernetes           Kubernetes           `yaml:"kubernetes" json:"kubernetes,omitempty"`
	Network              NetworkConfig        `yaml:"network" json:"network,omitempty"`
	Registry             RegistryConfig       `yaml:"registry" json:"registry,omitempty"`
	Storage              Storage              `yaml:"stroage" json:"stroage,omitempty"`
	KubeSphere           KubeSphere           `yaml:"kubesphere" json:"kubephere,omitempty"`
}

type Kubernetes struct {
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
	Name            string `yaml:"name,omitempty" json:"name,omitempty"`
	Address         string `yaml:"address,omitempty" json:"address,omitempty"`
	InternalAddress string `yaml:"internalAddress,omitempty" json:"internalAddress,omitempty"`
	Port            string `yaml:"port,omitempty" json:"port,omitempty"`
	User            string `yaml:"user,omitempty" json:"user,omitempty"`
	Password        string `yaml:"password,omitempty" json:"password,omitempty"`
	PrivateKeyPath  string `yaml:"privateKeyPath,omitempty" json:"privateKeyPath,omitempty"`
	ID              int    `json:"-"`
	IsEtcd          bool   `json:"-"`
	IsMaster        bool   `json:"-"`
	IsWorker        bool   `json:"-"`
	IsClient        bool   `json:"-"`
}

type RoleGroups struct {
	Etcd   []string `yaml:"etcd" json:"etcd,omitempty"`
	Master []string `yaml:"master" json:"master,omitempty"`
	Worker []string `yaml:"worker" json:"worker,omitempty"`
}

type HostGroups struct {
	All    []HostCfg
	Etcd   []HostCfg
	Master []HostCfg
	Worker []HostCfg
	K8s    []HostCfg
	Client []HostCfg
}

type NetworkConfig struct {
	Plugin          string `yaml:"plugin" json:"plugin,omitempty"`
	KubePodsCIDR    string `yaml:"kube_pods_cidr" json:"kube_pods_cidr,omitempty"`
	KubeServiceCIDR string `yaml:"kube_service_cidr" json:"kube_service_cidr,omitempty"`
}

type ControlPlaneEndpoint struct {
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

func (cfg *K2ClusterSpec) GenerateCertSANs() []string {
	clusterSvc := fmt.Sprintf("kubernetes.default.svc.%s", cfg.Kubernetes.ClusterName)
	defaultCertSANs := []string{"kubernetes", "kubernetes.default", "kubernetes.default.svc", clusterSvc, "localhost", "127.0.0.1"}
	extraCertSANs := []string{}

	extraCertSANs = append(extraCertSANs, cfg.ControlPlaneEndpoint.Domain)
	extraCertSANs = append(extraCertSANs, cfg.ControlPlaneEndpoint.Address)

	for _, host := range cfg.Hosts {
		extraCertSANs = append(extraCertSANs, host.Name)
		extraCertSANs = append(extraCertSANs, fmt.Sprintf("%s.%s", host.Name, cfg.Kubernetes.ClusterName))
		if host.Address != cfg.ControlPlaneEndpoint.Address {
			extraCertSANs = append(extraCertSANs, host.Address)
		}
		if host.InternalAddress != host.Address && host.InternalAddress != cfg.ControlPlaneEndpoint.Address {
			extraCertSANs = append(extraCertSANs, host.InternalAddress)
		}
	}

	extraCertSANs = append(extraCertSANs, util.ParseIp(cfg.Network.KubeServiceCIDR)[0])

	defaultCertSANs = append(defaultCertSANs, extraCertSANs...)
	return defaultCertSANs
}

func (cfg *K2ClusterSpec) GroupHosts() *HostGroups {
	clusterHostsGroups := HostGroups{}
	etcdGroup, masterGroup, workerGroup := cfg.ParseRolesList()
	for index, host := range cfg.Hosts {
		host.ID = index
		for _, hostName := range etcdGroup {
			if host.Name == hostName {
				host.IsEtcd = true
				clusterHostsGroups.Etcd = append(clusterHostsGroups.Etcd, host)
				break
			}
		}

		for _, hostName := range masterGroup {
			if host.Name == hostName {
				host.IsMaster = true
				clusterHostsGroups.Master = append(clusterHostsGroups.Master, host)
				break
			}
		}

		for _, hostName := range workerGroup {
			if host.Name == hostName {
				host.IsWorker = true
				clusterHostsGroups.Worker = append(clusterHostsGroups.Worker, host)
				break
			}
		}

		clusterHostsGroups.All = append(clusterHostsGroups.All, host)
	}

	for _, host := range clusterHostsGroups.All {
		if host.IsMaster || host.IsWorker {
			clusterHostsGroups.K8s = append(clusterHostsGroups.K8s, host)
		}
	}

	clusterHostsGroups.Client = append(clusterHostsGroups.Client, clusterHostsGroups.Master[0])
	return &clusterHostsGroups
}

func (cfg *K2ClusterSpec) ClusterIP() string {
	return util.ParseIp(cfg.Network.KubeServiceCIDR)[2]
}

func (cfg *K2ClusterSpec) ParseRolesList() ([]string, []string, []string) {
	etcdGroupList := []string{}
	masterGroupList := []string{}
	workerGroupList := []string{}

	for _, host := range cfg.RoleGroups.Etcd {
		if strings.Contains(host, "[") && strings.Contains(host, "]") && strings.Contains(host, ":") {
			etcdGroupList = append(etcdGroupList, getHostsRange(host)...)
		} else {
			etcdGroupList = append(etcdGroupList, host)
		}

	}

	for _, host := range cfg.RoleGroups.Master {
		if strings.Contains(host, "[") && strings.Contains(host, "]") && strings.Contains(host, ":") {
			masterGroupList = append(masterGroupList, getHostsRange(host)...)
		} else {
			masterGroupList = append(masterGroupList, host)
		}
	}

	for _, host := range cfg.RoleGroups.Worker {
		if strings.Contains(host, "[") && strings.Contains(host, "]") && strings.Contains(host, ":") {
			workerGroupList = append(workerGroupList, getHostsRange(host)...)
		} else {
			workerGroupList = append(workerGroupList, host)
		}
	}
	return etcdGroupList, masterGroupList, workerGroupList
}

func getHostsRange(rangeStr string) []string {
	hostRangeList := []string{}
	r := regexp.MustCompile(`\[(\d+)\:(\d+)\]`)
	nameSuffix := r.FindStringSubmatch(rangeStr)
	namePrefix := strings.Split(rangeStr, nameSuffix[0])[0]
	nameSuffixStart, _ := strconv.Atoi(nameSuffix[1])
	nameSuffixEnd, _ := strconv.Atoi(nameSuffix[2])
	for i := nameSuffixStart; i <= nameSuffixEnd; i++ {
		hostRangeList = append(hostRangeList, fmt.Sprintf("%s%d", namePrefix, i))
	}
	return hostRangeList
}
