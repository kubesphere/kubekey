/*
Copyright 2026 The KubeSphere Authors.

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

package convert

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"sigs.k8s.io/yaml"
)

// V1Alpha2APIVersion is the apiVersion of KubeKey v3 cluster configuration files.
const V1Alpha2APIVersion = "kubekey.kubesphere.io/v1alpha2"

// v3 role group keys, normalized to lowercase.
const (
	v3RoleMaster        = "master"
	v3RoleControlPlane  = "control-plane"
	v3RoleControlPlane2 = "controlplane"
	v3RoleWorker        = "worker"
	v3RoleEtcd          = "etcd"
	v3RoleRegistry      = "registry"
)

// Cluster is a minimal representation of the KubeKey v3 (v1alpha2) Cluster
// configuration. Only fields needed for conversion are declared; unknown
// fields are ignored by the YAML decoder.
type Cluster struct {
	APIVersion string      `json:"apiVersion"`
	Kind       string      `json:"kind"`
	Metadata   ClusterMeta `json:"metadata"`
	Spec       ClusterSpec `json:"spec"`
}

// ClusterMeta holds the object name of the v3 cluster configuration.
type ClusterMeta struct {
	Name string `json:"name,omitempty"`
}

// ClusterSpec mirrors cmd/kk/apis/kubekey/v1alpha2.ClusterSpec of the 3.x branch.
type ClusterSpec struct {
	Hosts                []HostCfg            `json:"hosts,omitempty"`
	RoleGroups           map[string][]string  `json:"roleGroups,omitempty"`
	ControlPlaneEndpoint ControlPlaneEndpoint `json:"controlPlaneEndpoint,omitempty"`
	System               System               `json:"system,omitempty"`
	Etcd                 EtcdCluster          `json:"etcd,omitempty"`
	DNS                  DNS                  `json:"dns,omitempty"`
	Kubernetes           Kubernetes           `json:"kubernetes,omitempty"`
	Network              NetworkConfig        `json:"network,omitempty"`
	Storage              StorageConfig        `json:"storage,omitempty"`
	Registry             RegistryConfig       `json:"registry,omitempty"`
	Addons               []Addon              `json:"addons,omitempty"`
	KubeSphere           KubeSphere           `json:"kubesphere,omitempty"`
}

// HostCfg mirrors v1alpha2.HostCfg. Note that 3.x hosts have no roles field:
// roles are expressed solely through roleGroups.
type HostCfg struct {
	Name            string            `json:"name,omitempty"`
	Address         string            `json:"address,omitempty"`
	InternalAddress string            `json:"internalAddress,omitempty"`
	Port            int               `json:"port,omitempty"`
	User            string            `json:"user,omitempty"`
	Password        string            `json:"password,omitempty"`
	PrivateKey      string            `json:"privateKey,omitempty"`
	PrivateKeyPath  string            `json:"privateKeyPath,omitempty"`
	Arch            string            `json:"arch,omitempty"`
	Timeout         *int64            `json:"timeout,omitempty"`
	Labels          map[string]string `json:"labels,omitempty"`
}

// ControlPlaneEndpoint mirrors v1alpha2.ControlPlaneEndpoint.
type ControlPlaneEndpoint struct {
	InternalLoadbalancer string  `json:"internalLoadbalancer,omitempty"`
	Domain               string  `json:"domain,omitempty"`
	ExternalDNS          *bool   `json:"externalDNS,omitempty"`
	Address              string  `json:"address,omitempty"`
	Port                 int     `json:"port,omitempty"`
	KubeVip              KubeVip `json:"kubevip,omitempty"`
}

// KubeVip mirrors v1alpha2.KubeVip.
type KubeVip struct {
	Mode string `json:"mode,omitempty"`
}

// System mirrors v1alpha2.System.
type System struct {
	NtpServers         []string `json:"ntpServers,omitempty"`
	Timezone           string   `json:"timezone,omitempty"`
	Rpms               []string `json:"rpms,omitempty"`
	Debs               []string `json:"debs,omitempty"`
	PreInstall         []any    `json:"preInstall,omitempty"`
	PostClusterInstall []any    `json:"postClusterInstall,omitempty"`
	PostInstall        []any    `json:"postInstall,omitempty"`
	SkipConfigureOS    bool     `json:"skipConfigureOS,omitempty"`
}

// RegistryConfig mirrors v1alpha2.RegistryConfig.
type RegistryConfig struct {
	Type               string         `json:"type,omitempty"`
	RegistryMirrors    []string       `json:"registryMirrors,omitempty"`
	InsecureRegistries []string       `json:"insecureRegistries,omitempty"`
	PrivateRegistry    string         `json:"privateRegistry,omitempty"`
	ContainerdDataDir  string         `json:"containerdDataDir,omitempty"`
	DockerDataDir      string         `json:"dockerDataDir,omitempty"`
	RegistryDataDir    string         `json:"registryDataDir,omitempty"`
	NamespaceOverride  string         `json:"namespaceOverride,omitempty"`
	BridgeIP           string         `json:"bridgeIP,omitempty"`
	Auths              map[string]any `json:"auths,omitempty"`
	NamespaceRewrite   map[string]any `json:"namespaceRewrite,omitempty"`
	RemoteMirrors      map[string]any `json:"remoteMirrors,omitempty"`
}

// KubeSphere mirrors v1alpha2.KubeSphere.
type KubeSphere struct {
	Enabled        bool   `json:"enabled,omitempty"`
	Version        string `json:"version,omitempty"`
	Configurations string `json:"configurations,omitempty"`
}

// Kubernetes mirrors v1alpha2.Kubernetes.
type Kubernetes struct {
	Type                     string          `json:"type,omitempty"`
	Version                  string          `json:"version,omitempty"`
	ClusterName              string          `json:"clusterName,omitempty"`
	DNSDomain                string          `json:"dnsDomain,omitempty"`
	DisableKubeProxy         bool            `json:"disableKubeProxy,omitempty"`
	MasqueradeAll            bool            `json:"masqueradeAll,omitempty"`
	MaxPods                  int             `json:"maxPods,omitempty"`
	PodPidsLimit             int             `json:"podPidsLimit,omitempty"`
	NodeCidrMaskSize         int             `json:"nodeCidrMaskSize,omitempty"`
	NodeCidrMaskSizeIPv6     int             `json:"nodeCidrMaskSizeIPv6,omitempty"`
	ApiserverCertExtraSans   []string        `json:"apiserverCertExtraSans,omitempty"`
	ProxyMode                string          `json:"proxyMode,omitempty"`
	AutoRenewCerts           *bool           `json:"autoRenewCerts,omitempty"`
	Nodelocaldns             *bool           `json:"nodelocaldns,omitempty"`
	ContainerManager         string          `json:"containerManager,omitempty"`
	ContainerRuntimeEndpoint string          `json:"containerRuntimeEndpoint,omitempty"`
	NodeFeatureDiscovery     map[string]any  `json:"nodeFeatureDiscovery,omitempty"`
	Kata                     map[string]any  `json:"kata,omitempty"`
	ApiServerArgs            []string        `json:"apiserverArgs,omitempty"`
	ControllerManagerArgs    []string        `json:"controllerManagerArgs,omitempty"`
	SchedulerArgs            []string        `json:"schedulerArgs,omitempty"`
	KubeletArgs              []string        `json:"kubeletArgs,omitempty"`
	KubeProxyArgs            []string        `json:"kubeProxyArgs,omitempty"`
	FeatureGates             map[string]bool `json:"featureGates,omitempty"`
	KubeletConfiguration     map[string]any  `json:"kubeletConfiguration,omitempty"`
	KubeProxyConfiguration   map[string]any  `json:"kubeProxyConfiguration,omitempty"`
	Audit                    map[string]any  `json:"audit,omitempty"`
	NvidiaRuntime            *bool           `json:"nvidiaRuntime,omitempty"`
}

// NetworkConfig mirrors v1alpha2.NetworkConfig.
type NetworkConfig struct {
	Plugin          string         `json:"plugin,omitempty"`
	KubePodsCIDR    string         `json:"kubePodsCIDR,omitempty"`
	KubeServiceCIDR string         `json:"kubeServiceCIDR,omitempty"`
	Calico          CalicoCfg      `json:"calico,omitempty"`
	Flannel         map[string]any `json:"flannel,omitempty"`
	Kubeovn         map[string]any `json:"kubeovn,omitempty"`
	MultusCNI       map[string]any `json:"multusCNI,omitempty"`
	Hybridnet       map[string]any `json:"hybridnet,omitempty"`
}

// CalicoCfg mirrors v1alpha2.CalicoCfg.
type CalicoCfg struct {
	IPIPMode              string         `json:"ipipMode,omitempty"`
	IPAutoDetectionMethod string         `json:"ipAutoDetectionMethod,omitempty"`
	VXLANMode             string         `json:"vxlanMode,omitempty"`
	VethMTU               int            `json:"vethMTU,omitempty"`
	Ipv4NatOutgoing       *bool          `json:"ipv4NatOutgoing,omitempty"`
	Ipv6NatOutgoing       *bool          `json:"ipv6NatOutgoing,omitempty"`
	DefaultIPPOOL         *bool          `json:"defaultIPPOOL,omitempty"`
	Typha                 map[string]any `json:"typha,omitempty"`
	Controller            map[string]any `json:"controller,omitempty"`
}

// EtcdCluster mirrors v1alpha2.EtcdCluster.
type EtcdCluster struct {
	Type                    string         `json:"type,omitempty"`
	External                map[string]any `json:"external,omitempty"`
	Port                    *int           `json:"port,omitempty"`
	PeerPort                *int           `json:"peerPort,omitempty"`
	ExtraArgs               []string       `json:"extraArgs,omitempty"`
	BackupDir               string         `json:"backupDir,omitempty"`
	BackupPeriod            int            `json:"backupPeriod,omitempty"`
	KeepBackupNumber        int            `json:"keepBackupNumber,omitempty"`
	BackupScriptDir         string         `json:"backupScript,omitempty"`
	DataDir                 *string        `json:"dataDir,omitempty"`
	HeartbeatInterval       *int           `json:"heartbeatInterval,omitempty"`
	ElectionTimeout         *int           `json:"electionTimeout,omitempty"`
	SnapshotCount           *int           `json:"snapshotCount,omitempty"`
	AutoCompactionRetention *int           `json:"autoCompactionRetention,omitempty"`
	Metrics                 *string        `json:"metrics,omitempty"`
	QuotaBackendBytes       *int64         `json:"quotaBackendBytes,omitempty"`
	MaxRequestBytes         *int64         `json:"maxRequestBytes,omitempty"`
	MaxSnapshots            *int           `json:"maxSnapshots,omitempty"`
	MaxWals                 *int           `json:"maxWals,omitempty"`
	LogLevel                *string        `json:"logLevel,omitempty"`
}

// StorageConfig mirrors v1alpha2.StorageConfig.
type StorageConfig struct {
	OpenEBS OpenEBSCfg `json:"openebs,omitempty"`
}

// OpenEBSCfg mirrors v1alpha2.OpenEBSCfg.
type OpenEBSCfg struct {
	BasePath string `json:"basePath,omitempty"`
}

// DNS mirrors v1alpha2.DNS.
type DNS struct {
	DNSEtcHosts  string         `json:"dnsEtcHosts,omitempty"`
	NodeEtcHosts string         `json:"nodeEtcHosts,omitempty"`
	CoreDNS      map[string]any `json:"coredns,omitempty"`
	NodeLocalDNS map[string]any `json:"nodelocaldns,omitempty"`
}

// Addon mirrors v1alpha2.Addon.
type Addon struct {
	Name      string         `json:"name,omitempty"`
	Namespace string         `json:"namespace,omitempty"`
	Repo      map[string]any `json:"repo,omitempty"`
	Values    string         `json:"values,omitempty"`
}

// ParseCluster parses a KubeKey v3 (v1alpha2) cluster configuration from YAML bytes.
func ParseCluster(data []byte) (*Cluster, error) {
	c := &Cluster{}
	if err := yaml.Unmarshal(data, c); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal v3 cluster configuration")
	}
	if c.Kind != "" && c.Kind != "Cluster" {
		return nil, errors.Errorf("unsupported kind %q, expected \"Cluster\"", c.Kind)
	}
	if c.APIVersion != "" && c.APIVersion != V1Alpha2APIVersion {
		return nil, errors.Errorf("unsupported apiVersion %q, expected %q", c.APIVersion, V1Alpha2APIVersion)
	}
	if len(c.Spec.Hosts) == 0 {
		return nil, errors.New("no hosts found in spec.hosts")
	}
	return c, nil
}

// expandRoleGroups resolves v3 roleGroups into per-role host name lists.
// It expands the node[1:5] range shorthand and verifies that every referenced
// host exists in the hosts list. Returned keys are v3 role names.
func expandRoleGroups(roleGroups map[string][]string, knownHosts []string) (map[string][]string, error) {
	hostSet := make(map[string]struct{}, len(knownHosts))
	for _, h := range knownHosts {
		hostSet[h] = struct{}{}
	}

	result := make(map[string][]string)
	for role, hosts := range roleGroups {
		expanded := make([]string, 0, len(hosts))
		for _, entry := range hosts {
			if strings.Contains(entry, "[") && strings.Contains(entry, "]") && strings.Contains(entry, ":") {
				rangeHosts, err := expandHostsRange(entry, hostSet, role)
				if err != nil {
					return nil, err
				}
				expanded = append(expanded, rangeHosts...)
				continue
			}
			if _, ok := hostSet[entry]; !ok {
				return nil, errors.Errorf("[%s] is in [%s] group, but not in hosts list", entry, role)
			}
			expanded = append(expanded, entry)
		}
		result[strings.ToLower(role)] = expanded
	}
	return result, nil
}

var hostsRangeRegexp = regexp.MustCompile(`\[(\d+)\:(\d+)\]`)

// expandHostsRange expands "node[1:3]" into ["node1", "node2", "node3"],
// mirroring v1alpha2.getHostsRange.
func expandHostsRange(rangeStr string, hostSet map[string]struct{}, group string) ([]string, error) {
	suffix := hostsRangeRegexp.FindStringSubmatch(rangeStr)
	if suffix == nil {
		return nil, errors.Errorf("invalid host range %q in roleGroups/%s, expected format name[start:end]", rangeStr, group)
	}
	prefix := strings.Split(rangeStr, suffix[0])[0]
	start, err := strconv.Atoi(suffix[1])
	if err != nil {
		return nil, errors.Wrapf(err, "invalid range start in %q", rangeStr)
	}
	end, err := strconv.Atoi(suffix[2])
	if err != nil {
		return nil, errors.Wrapf(err, "invalid range end in %q", rangeStr)
	}
	hosts := make([]string, 0, end-start+1)
	for i := start; i <= end; i++ {
		name := fmt.Sprintf("%s%d", prefix, i)
		if _, ok := hostSet[name]; !ok {
			return nil, errors.Errorf("[%s] is in [%s] group, but not in hosts list", name, group)
		}
		hosts = append(hosts, name)
	}
	return hosts, nil
}

// splitInternalAddress splits a v3 internalAddress ("10.0.0.1,fd00::1" or a
// single address) into ipv4 and ipv6 parts.
func splitInternalAddress(internalAddress string) (ipv4, ipv6 string) {
	parts := strings.Split(internalAddress, ",")
	ipv4 = strings.TrimSpace(parts[0])
	if len(parts) > 1 {
		ipv6 = strings.TrimSpace(parts[1])
	}
	return ipv4, ipv6
}
