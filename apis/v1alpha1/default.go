package v1alpha1

//const (
//	DefaultPreDir         = "/tmp/kubekey"
//	DefaultSSHPort        = "22"
//	DefaultDockerSockPath = "/var/run/docker.sock"
//	DefaultLBPort         = "6443"
//	DefaultLBDomain       = "lb.kubesphere.local"
//	DefaultNetworkPlugin  = "calico"
//	DefaultPodsCIDR       = "10.233.64.0/18"
//	DefaultServiceCIDR    = "10.233.0.0/18"
//	DefaultKubeImageRepo  = "kubekey"
//	DefaultClusterName    = "cluster.local"
//	DefaultArch           = "amd64"
//	DefaultHostName       = "allinone"
//	DefaultEtcdRepo       = "kubekey/etcd"
//	DefaultEtcdVersion    = "v3.3.12"
//	DefaultEtcdPort       = "2379"
//	DefaultKubeVersion    = "v1.17.4"
//	DefaultCniVersion     = "v0.8.2"
//	DefaultHelmVersion    = "v3.1.2"
//	ETCDRole              = "etcd"
//	MasterRole            = "master"
//	WorkerRole            = "worker"
//)
//
//type HostConfig struct {
//	ID                int    `json:"-"`
//	PublicAddress     string `json:"publicAddress"`
//	PrivateAddress    string `json:"privateAddress"`
//	SSHPort           int    `json:"sshPort"`
//	SSHUsername       string `json:"sshUsername"`
//	SSHPrivateKeyFile string `json:"sshPrivateKeyFile"`
//	SSHAgentSocket    string `json:"sshAgentSocket"`
//	Bastion           string `json:"bastion"`
//	BastionPort       int    `json:"bastionPort"`
//	BastionUser       string `json:"bastionUser"`
//	Hostname          string `json:"hostname"`
//	IsLeader          bool   `json:"isLeader"`
//	Untaint           bool   `json:"untaint"`
//
//	// Information populated at the runtime
//	OperatingSystem string `json:"-"`
//}
