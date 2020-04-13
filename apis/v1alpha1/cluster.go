package v1alpha1

type ClusterCfg struct {
	Hosts           []HostCfg          `yaml:"hosts" json:"hosts,omitempty"`
	LBKubeApiserver LBKubeApiserverCfg `yaml:"lb_kubeapiserver" json:"lb_kubeapiserver,omitempty"`
	KubeVersion     string             `yaml:"kube_version" json:"kube_version,omitempty"`
	KubeImageRepo   string             `yaml:"kube_image_repo" json:"kube_image_repo,omitempty"`
	KubeClusterName string             `yaml:"kube_cluster_name" json:"kube_cluster_name,omitempty"`
	Network         NetworkConfig      `yaml:"network" json:"network,omitempty"`
}

type HostCfg struct {
	HostName        string   `yaml:"hostName,omitempty" json:"hostName,omitempty"`
	Address         string   `yaml:"address" json:"address,omitempty"`
	Port            string   `yaml:"port" json:"port,omitempty"`
	InternalAddress string   `yaml:"internal_address" json:"internalAddress,omitempty"`
	Role            []string `yaml:"role" json:"role,omitempty" norman:"type=array[enum],options=etcd|worker|worker"`
	//HostnameOverride string   `yaml:"hostname_override" json:"hostnameOverride,omitempty"`
	User     string `yaml:"user" json:"user,omitempty"`
	Password string `yaml:"password" json:"password,omitempty"`
	//SSHAgentAuth     bool              `yaml:"ssh_agent_auth,omitempty" json:"sshAgentAuth,omitempty"`
	//SSHKey           string            `yaml:"ssh_key" json:"sshKey,omitempty" norman:"type=password"`
	SSHKeyPath string `yaml:"ssh_key_path" json:"sshKeyPath,omitempty"`
	//SSHCert          string            `yaml:"ssh_cert" json:"sshCert,omitempty"`
	//SSHCertPath      string            `yaml:"ssh_cert_path" json:"sshCertPath,omitempty"`
	//Labels map[string]string `yaml:"labels" json:"labels,omitempty"`
	//Taints []Taint           `yaml:"taints" json:"taints,omitempty"`
	ID int `json:"-"`
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
