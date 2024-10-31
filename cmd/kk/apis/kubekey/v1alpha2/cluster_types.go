/*
 Copyright 2021 The KubeSphere Authors.

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

package v1alpha2

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/connector"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/logger"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/util"
)

// ClusterSpec defines the desired state of Cluster
type ClusterSpec struct {
	Hosts                []HostCfg            `yaml:"hosts" json:"hosts,omitempty"`
	RoleGroups           map[string][]string  `yaml:"roleGroups" json:"roleGroups,omitempty"`
	ControlPlaneEndpoint ControlPlaneEndpoint `yaml:"controlPlaneEndpoint" json:"controlPlaneEndpoint,omitempty"`
	System               System               `yaml:"system" json:"system,omitempty"`
	Etcd                 EtcdCluster          `yaml:"etcd" json:"etcd,omitempty"`
	DNS                  DNS                  `yaml:"dns" json:"dns,omitempty"`
	Kubernetes           Kubernetes           `yaml:"kubernetes" json:"kubernetes,omitempty"`
	Network              NetworkConfig        `yaml:"network" json:"network,omitempty"`
	Storage              StorageConfig        `yaml:"storage" json:"storage,omitempty"`
	Registry             RegistryConfig       `yaml:"registry" json:"registry,omitempty"`
	Addons               []Addon              `yaml:"addons" json:"addons,omitempty"`
	KubeSphere           KubeSphere           `json:"kubesphere,omitempty"`
}

type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ClusterSpec `json:"spec,omitempty"`
}

// HostCfg defines host information for cluster.
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
	Timeout         *int64 `yaml:"timeout,omitempty" json:"timeout,omitempty"`

	// Labels defines the kubernetes labels for the node.
	Labels map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
}

// ControlPlaneEndpoint defines the control plane endpoint information for cluster.
type ControlPlaneEndpoint struct {
	InternalLoadbalancer string  `yaml:"internalLoadbalancer" json:"internalLoadbalancer,omitempty"`
	Domain               string  `yaml:"domain" json:"domain,omitempty"`
	ExternalDNS          *bool   `yaml:"externalDNS" json:"externalDNS"`
	Address              string  `yaml:"address" json:"address,omitempty"`
	Port                 int     `yaml:"port" json:"port,omitempty"`
	KubeVip              KubeVip `yaml:"kubevip" json:"kubevip,omitempty"`
}

type KubeVip struct {
	Mode string `yaml:"mode" json:"mode,omitempty"`
}

// CustomScripts defines the custom shell scripts for each node to exec before and finished kubernetes install.
type CustomScripts struct {
	Name      string   `yaml:"name" json:"name,omitempty"`
	Bash      string   `yaml:"bash" json:"bash,omitempty"`
	Role      string   `yaml:"role" json:"role,omitempty"`
	Materials []string `yaml:"materials" json:"materials,omitempty"`
}

// System defines the system config for each node in cluster.
type System struct {
	NtpServers         []string        `yaml:"ntpServers" json:"ntpServers,omitempty"`
	Timezone           string          `yaml:"timezone" json:"timezone,omitempty"`
	Rpms               []string        `yaml:"rpms" json:"rpms,omitempty"`
	Debs               []string        `yaml:"debs" json:"debs,omitempty"`
	PreInstall         []CustomScripts `yaml:"preInstall" json:"preInstall,omitempty"`
	PostClusterInstall []CustomScripts `yaml:"postClusterInstall" json:"postClusterInstall,omitempty"`
	PostInstall        []CustomScripts `yaml:"postInstall" json:"postInstall,omitempty"`
	SkipConfigureOS    bool            `yaml:"skipConfigureOS" json:"skipConfigureOS,omitempty"`
}

// RegistryConfig defines the configuration information of the image's repository.
type RegistryConfig struct {
	Type               string               `yaml:"type" json:"type,omitempty"`
	RegistryMirrors    []string             `yaml:"registryMirrors" json:"registryMirrors,omitempty"`
	InsecureRegistries []string             `yaml:"insecureRegistries" json:"insecureRegistries,omitempty"`
	PrivateRegistry    string               `yaml:"privateRegistry" json:"privateRegistry,omitempty"`
	DataRoot           string               `yaml:"dataRoot" json:"dataRoot,omitempty"`
	NamespaceOverride  string               `yaml:"namespaceOverride" json:"namespaceOverride,omitempty"`
	BridgeIP           string               `yaml:"bridgeIP" json:"bridgeIP,omitempty"`
	Auths              runtime.RawExtension `yaml:"auths" json:"auths,omitempty"`
	Port               int                  `yaml:"port" json:"port,omitempty"`
}

// KubeSphere defines the configuration information of the KubeSphere.
type KubeSphere struct {
	Enabled        bool   `json:"enabled,omitempty"`
	Version        string `json:"version,omitempty"`
	Configurations string `json:"configurations,omitempty"`
}

// GenerateCertSANs is used to generate cert sans for cluster.
func (cfg *ClusterSpec) GenerateCertSANs() []string {
	clusterSvc := fmt.Sprintf("kubernetes.default.svc.%s", cfg.Kubernetes.DNSDomain)
	defaultCertSANs := []string{"kubernetes", "kubernetes.default", "kubernetes.default.svc", clusterSvc, "localhost", "127.0.0.1"}
	extraCertSANs := make([]string, 0)

	extraCertSANs = append(extraCertSANs, cfg.ControlPlaneEndpoint.Domain)

	if cfg.ControlPlaneEndpoint.Address != "" {
		extraCertSANs = append(extraCertSANs, cfg.ControlPlaneEndpoint.Address)
	}

	for _, host := range cfg.Hosts {
		extraCertSANs = append(extraCertSANs, host.Name)
		extraCertSANs = append(extraCertSANs, fmt.Sprintf("%s.%s", host.Name, cfg.Kubernetes.DNSDomain))
		if host.Address != cfg.ControlPlaneEndpoint.Address {
			extraCertSANs = append(extraCertSANs, host.Address)
		}

		nodeAddresses := strings.Split(host.InternalAddress, ",")
		InternalIPv4Address := nodeAddresses[0]
		if InternalIPv4Address != host.Address && InternalIPv4Address != cfg.ControlPlaneEndpoint.Address {
			extraCertSANs = append(extraCertSANs, InternalIPv4Address)
		}
		if len(nodeAddresses) == 2 {
			InternalIPv6Address := nodeAddresses[1]
			extraCertSANs = append(extraCertSANs, InternalIPv6Address)
		}
	}

	extraCertSANs = append(extraCertSANs, util.ParseIp(strings.Split(cfg.Network.KubeServiceCIDR, ",")[0])[0])

	defaultCertSANs = append(defaultCertSANs, extraCertSANs...)

	if cfg.Kubernetes.ApiserverCertExtraSans != nil {
		defaultCertSANs = append(defaultCertSANs, cfg.Kubernetes.ApiserverCertExtraSans...)
	}

	return defaultCertSANs
}

// GroupHosts is used to group hosts according to the configuration file.s
func (cfg *ClusterSpec) GroupHosts() map[string][]*KubeHost {
	hostMap := make(map[string]*KubeHost)
	for _, hostCfg := range cfg.Hosts {
		host := toHosts(hostCfg)
		hostMap[host.Name] = host
	}

	roleGroups := cfg.ParseRolesList(hostMap)

	//Check that the parameters under roleGroups are incorrect
	if len(roleGroups[Master]) == 0 && len(roleGroups[ControlPlane]) == 0 {
		logger.Log.Fatal(errors.New("The number of master/control-plane cannot be 0"))
	}
	if len(roleGroups[Etcd]) == 0 && cfg.Etcd.Type == KubeKey {
		logger.Log.Fatal(errors.New("The number of etcd cannot be 0"))
	}

	for _, host := range roleGroups[ControlPlane] {
		host.SetRole(Master)
		roleGroups[Master] = append(roleGroups[Master], host)
	}

	return roleGroups
}

// +kubebuilder:object:generate=false
type KubeHost struct {
	*connector.BaseHost
	Labels map[string]string
}

func toHosts(cfg HostCfg) *KubeHost {
	host := connector.NewHost()
	host.Name = cfg.Name
	host.Address = cfg.Address
	host.InternalAddress = cfg.InternalAddress
	host.Port = cfg.Port
	host.User = cfg.User
	host.Password = cfg.Password
	host.PrivateKey = cfg.PrivateKey
	host.PrivateKeyPath = cfg.PrivateKeyPath
	host.Arch = cfg.Arch
	host.Timeout = *cfg.Timeout

	kubeHost := &KubeHost{
		BaseHost: host,
		Labels:   cfg.Labels,
	}
	return kubeHost
}

// ClusterIP is used to get the kube-apiserver service address inside the cluster.
func (cfg *ClusterSpec) ClusterIP() string {
	return util.ParseIp(strings.Split(cfg.Network.KubeServiceCIDR, ",")[0])[0]
}

// CorednsClusterIP is used to get the coredns service address inside the cluster.
func (cfg *ClusterSpec) CorednsClusterIP() string {
	return util.ParseIp(strings.Split(cfg.Network.KubeServiceCIDR, ",")[0])[2]
}

// ClusterDNS is used to get the dns server address inside the cluster.
func (cfg *ClusterSpec) ClusterDNS() string {
	if cfg.Kubernetes.EnableNodelocaldns() {
		return "169.254.25.10"
	} else {
		return cfg.CorednsClusterIP()
	}
}

// ParseRolesList is used to parse the host grouping list.
func (cfg *ClusterSpec) ParseRolesList(hostMap map[string]*KubeHost) map[string][]*KubeHost {
	roleGroupLists := make(map[string][]*KubeHost)
	for role, hosts := range cfg.RoleGroups {
		roleGroup := make([]string, 0)
		for _, host := range hosts {
			h := make([]string, 0)
			if strings.Contains(host, "[") && strings.Contains(host, "]") && strings.Contains(host, ":") {
				rangeHosts := getHostsRange(host, hostMap, role)
				h = append(h, rangeHosts...)
			} else {
				if err := hostVerify(hostMap, host, role); err != nil {
					logger.Log.Fatal(err)
				}
				h = append(h, host)
			}

			roleGroup = append(roleGroup, h...)
			for _, hostName := range h {
				if h, ok := hostMap[hostName]; ok {
					roleGroupAppend(roleGroupLists, role, h)
				} else {
					logger.Log.Fatal(fmt.Errorf("incorrect nodeName under roleGroups/%s in the configuration file", role))
				}
			}
		}
	}

	return roleGroupLists
}

func roleGroupAppend(roleGroupLists map[string][]*KubeHost, role string, host *KubeHost) {
	host.SetRole(role)
	r := roleGroupLists[role]
	r = append(r, host)
	roleGroupLists[role] = r
}

func getHostsRange(rangeStr string, hostMap map[string]*KubeHost, group string) []string {
	hostRangeList := make([]string, 0)
	r := regexp.MustCompile(`\[(\d+)\:(\d+)\]`)
	nameSuffix := r.FindStringSubmatch(rangeStr)
	namePrefix := strings.Split(rangeStr, nameSuffix[0])[0]
	nameSuffixStart, _ := strconv.Atoi(nameSuffix[1])
	nameSuffixEnd, _ := strconv.Atoi(nameSuffix[2])
	for i := nameSuffixStart; i <= nameSuffixEnd; i++ {
		if err := hostVerify(hostMap, fmt.Sprintf("%s%d", namePrefix, i), group); err != nil {
			logger.Log.Fatal(err)
		}
		hostRangeList = append(hostRangeList, fmt.Sprintf("%s%d", namePrefix, i))
	}
	return hostRangeList
}

func hostVerify(hostMap map[string]*KubeHost, hostName string, group string) error {
	if _, ok := hostMap[hostName]; !ok {
		return fmt.Errorf("[%s] is in [%s] group, but not in hosts list", hostName, group)
	}
	return nil
}

func (c ControlPlaneEndpoint) IsInternalLBEnabled() bool {
	return c.InternalLoadbalancer == Haproxy
}

func (c ControlPlaneEndpoint) IsInternalLBEnabledVip() bool {
	return c.InternalLoadbalancer == Kubevip
}

// EnableExternalDNS is used to determine whether to use external dns to resolve kube-apiserver domain.
func (c *ControlPlaneEndpoint) EnableExternalDNS() bool {
	if c.ExternalDNS == nil {
		return false
	}
	return *c.ExternalDNS
}

func (r *RegistryConfig) GetHost() string {
	if r.PrivateRegistry == "" {
		return ""
	}
	return strings.Split(r.PrivateRegistry, "/")[0]
}

func (r *RegistryConfig) GetPort() int {
	if r.Port == 0 {
		return 80
	}
	return r.Port
}
