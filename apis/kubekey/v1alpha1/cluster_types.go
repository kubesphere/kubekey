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

package v1alpha1

import (
	"errors"
	"fmt"
	"github.com/kubesphere/kubekey/pkg/util"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"regexp"
	"strconv"
	"strings"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ClusterSpec defines the desired state of Cluster
type ClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Cluster. Edit Cluster_types.go to remove/update
	Hosts                []HostCfg            `yaml:"hosts" json:"hosts,omitempty"`
	RoleGroups           RoleGroups           `yaml:"roleGroups" json:"roleGroups,omitempty"`
	ControlPlaneEndpoint ControlPlaneEndpoint `yaml:"controlPlaneEndpoint" json:"controlPlaneEndpoint,omitempty"`
	Kubernetes           Kubernetes           `yaml:"kubernetes" json:"kubernetes,omitempty"`
	Network              NetworkConfig        `yaml:"network" json:"network,omitempty"`
	Registry             RegistryConfig       `yaml:"registry" json:"registry,omitempty"`
	Addons               []Addon              `yaml:"addons" json:"addons,omitempty"`
	KubeSphere           KubeSphere           `json:"kubesphere,omitempty"`
}

// ClusterStatus defines the observed state of Cluster
type ClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	JobInfo       JobInfo      `json:"jobInfo,omitempty"`
	Version       string       `json:"version,omitempty"`
	NetworkPlugin string       `json:"networkPlugin,omitempty"`
	NodesCount    int          `json:"nodesCount,omitempty"`
	EtcdCount     int          `json:"etcdCount,omitempty"`
	MasterCount   int          `json:"masterCount,omitempty"`
	WorkerCount   int          `json:"workerCount,omitempty"`
	Nodes         []NodeStatus `json:"nodes,omitempty"`
	Conditions    []Condition  `json:"Conditions,omitempty"`
}

type JobInfo struct {
	Namespace string    `json:"namespace,omitempty"`
	Name      string    `json:"name,omitempty"`
	Pods      []PodInfo `json:"pods,omitempty"`
}

type PodInfo struct {
	Name       string          `json:"name,omitempty"`
	Containers []ContainerInfo `json:"containers,omitempty"`
}
type ContainerInfo struct {
	Name string `json:"name,omitempty"`
}
type NodeStatus struct {
	InternalIP string          `json:"internalIP,omitempty"`
	Hostname   string          `json:"hostname,omitempty"`
	Roles      map[string]bool `json:"roles,omitempty"`
}

type Condition struct {
	Step      string      `json:"step,omitempty"`
	StartTime metav1.Time `json:"startTime,omitempty"`
	EndTime   metav1.Time `json:"endTime,omitempty"`
	Status    bool        `json:"status,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Cluster is the Schema for the clusters API
// +kubebuilder:resource:path=clusters,scope=Cluster
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterList contains a list of Cluster
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}

type HostCfg struct {
	Name            string `yaml:"name,omitempty" json:"name,omitempty"`
	Address         string `yaml:"address,omitempty" json:"address,omitempty"`
	InternalAddress string `yaml:"internalAddress,omitempty" json:"internalAddress,omitempty"`
	Port            int    `yaml:"port,omitempty" json:"port,omitempty"`
	User            string `yaml:"user,omitempty" json:"user,omitempty"`
	Password        string `yaml:"password,omitempty" json:"password,omitempty"`
	PrivateKey      string `yaml:"privateKey,omitempty" json:"privateKey,omitempty"`
	PrivateKeyPath  string `yaml:"privateKeyPath,omitempty" json:"privateKeyPath,omitempty"`
	Arch            string `yaml:"arch,omitempty" json:"arch,omitempty"`
	ID              int    `json:"-"`
	IsEtcd          bool   `json:"-"`
	IsMaster        bool   `json:"-"`
	IsWorker        bool   `json:"-"`
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
}

type ControlPlaneEndpoint struct {
	Domain  string `yaml:"domain" json:"domain,omitempty"`
	Address string `yaml:"address" json:"address,omitempty"`
	Port    int    `yaml:"port" json:"port,omitempty"`
}

type RegistryConfig struct {
	RegistryMirrors    []string `yaml:"registryMirrors" json:"registryMirrors,omitempty"`
	InsecureRegistries []string `yaml:"insecureRegistries" json:"insecureRegistries,omitempty"`
	PrivateRegistry    string   `yaml:"privateRegistry" json:"privateRegistry,omitempty"`
}

type KubeSphere struct {
	Enabled        bool   `json:"enabled,omitempty"`
	Version        string `json:"version,omitempty"`
	Configurations string `json:"configurations,omitempty"`
}

type ExternalEtcd struct {
	Endpoints []string
	CaFile    string
	CertFile  string
	KeyFile   string
}

func (cfg *ClusterSpec) GenerateCertSANs() []string {
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

	if cfg.Kubernetes.ApiserverCertExtraSans != nil {
		defaultCertSANs = append(defaultCertSANs, cfg.Kubernetes.ApiserverCertExtraSans...)
	}

	return defaultCertSANs
}

func (cfg *ClusterSpec) GroupHosts(logger *log.Logger) (*HostGroups, error) {
	clusterHostsGroups := HostGroups{}

	hostList := map[string]string{}
	for _, host := range cfg.Hosts {
		hostList[host.Name] = host.Name
	}

	etcdGroup, masterGroup, workerGroup, err := cfg.ParseRolesList(hostList, logger)
	if err != nil {
		return nil, err
	}
	for index, host := range cfg.Hosts {
		host.ID = index
		if len(etcdGroup) > 0 {
			for _, hostName := range etcdGroup {
				if host.Name == hostName {
					host.IsEtcd = true
					break
				}
			}
		}

		if len(masterGroup) > 0 {
			for _, hostName := range masterGroup {
				if host.Name == hostName {
					host.IsMaster = true
					break
				}
			}
		}

		if len(workerGroup) > 0 {
			for _, hostName := range workerGroup {
				if hostName != "" && host.Name == hostName {
					host.IsWorker = true
					break
				}
			}
		}

		if host.IsEtcd {
			clusterHostsGroups.Etcd = append(clusterHostsGroups.Etcd, host)
		}
		if host.IsMaster {
			clusterHostsGroups.Master = append(clusterHostsGroups.Master, host)
		}
		if host.IsWorker {
			clusterHostsGroups.Worker = append(clusterHostsGroups.Worker, host)
		}
		if host.IsMaster || host.IsWorker {
			clusterHostsGroups.K8s = append(clusterHostsGroups.K8s, host)
		}
		clusterHostsGroups.All = append(clusterHostsGroups.All, host)
	}

	//Check that the parameters under roleGroups are incorrect
	if len(masterGroup) == 0 {
		logger.Fatal(errors.New("The number of master cannot be 0."))
	}
	if len(etcdGroup) == 0 {
		logger.Fatal(errors.New("The number of etcd cannot be 0."))
	}

	if len(masterGroup) != len(clusterHostsGroups.Master) {
		return nil, errors.New("Incorrect nodeName under roleGroups/master in the configuration file, Please check before installing.")
	}
	if len(etcdGroup) != len(clusterHostsGroups.Etcd) {
		return nil, errors.New("Incorrect nodeName under roleGroups/etcd in the configuration file, Please check before installing.")
	}
	if len(workerGroup) != len(clusterHostsGroups.Worker) {
		return nil, errors.New("Incorrect nodeName under roleGroups/work in the configuration file, Please check before installing.")
	}

	return &clusterHostsGroups, nil
}

func (cfg *ClusterSpec) ClusterIP() string {
	return util.ParseIp(cfg.Network.KubeServiceCIDR)[2]
}

func (cfg *ClusterSpec) ParseRolesList(hostList map[string]string, logger *log.Logger) ([]string, []string, []string, error) {
	etcdGroupList := []string{}
	masterGroupList := []string{}
	workerGroupList := []string{}

	for _, host := range cfg.RoleGroups.Etcd {
		if strings.Contains(host, "[") && strings.Contains(host, "]") && strings.Contains(host, ":") {
			etcdGroupList = append(etcdGroupList, getHostsRange(host, hostList, "etcd", logger)...)
		} else {
			if err := hostVerify(hostList, host, "etcd"); err != nil {
				logger.Fatal(err)
			}
			etcdGroupList = append(etcdGroupList, host)
		}
	}

	for _, host := range cfg.RoleGroups.Master {
		if strings.Contains(host, "[") && strings.Contains(host, "]") && strings.Contains(host, ":") {
			masterGroupList = append(masterGroupList, getHostsRange(host, hostList, "master", logger)...)
		} else {
			if err := hostVerify(hostList, host, "master"); err != nil {
				logger.Fatal(err)
			}
			masterGroupList = append(masterGroupList, host)
		}
	}

	for _, host := range cfg.RoleGroups.Worker {
		if strings.Contains(host, "[") && strings.Contains(host, "]") && strings.Contains(host, ":") {
			workerGroupList = append(workerGroupList, getHostsRange(host, hostList, "worker", logger)...)
		} else {
			if err := hostVerify(hostList, host, "worker"); err != nil {
				logger.Fatal(err)
			}
			workerGroupList = append(workerGroupList, host)
		}
	}
	return etcdGroupList, masterGroupList, workerGroupList, nil
}

func getHostsRange(rangeStr string, hostList map[string]string, group string, logger *log.Logger) []string {
	hostRangeList := []string{}
	r := regexp.MustCompile(`\[(\d+)\:(\d+)\]`)
	nameSuffix := r.FindStringSubmatch(rangeStr)
	namePrefix := strings.Split(rangeStr, nameSuffix[0])[0]
	nameSuffixStart, _ := strconv.Atoi(nameSuffix[1])
	nameSuffixEnd, _ := strconv.Atoi(nameSuffix[2])
	for i := nameSuffixStart; i <= nameSuffixEnd; i++ {
		if err := hostVerify(hostList, fmt.Sprintf("%s%d", namePrefix, i), group); err != nil {
			logger.Fatal(err)
		}
		hostRangeList = append(hostRangeList, fmt.Sprintf("%s%d", namePrefix, i))
	}
	return hostRangeList
}

func hostVerify(hostList map[string]string, hostName string, group string) error {
	if _, ok := hostList[hostName]; !ok {
		return errors.New(fmt.Sprintf("[%s] is in [%s] group, but not in hosts list.", hostName, group))
	}
	return nil
}
