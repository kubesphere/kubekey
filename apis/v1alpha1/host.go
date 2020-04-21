package v1alpha1

type HostCfg struct {
	HostName        string   `yaml:"hostName" json:"hostName,omitempty"`
	SSHAddress      string   `yaml:"sshAddress" json:"sshAddress,omitempty"`
	InternalAddress string   `yaml:"internalAddress" json:"internalAddress,omitempty"`
	Port            string   `yaml:"port" json:"port,omitempty"`
	User            string   `yaml:"user" json:"user,omitempty"`
	Password        string   `yaml:"password, omitempty" json:"password,omitempty"`
	SSHKeyPath      string   `yaml:"sshKeyPath, omitempty" json:"sshKeyPath,omitempty"`
	Role            []string `yaml:"role" json:"role,omitempty" norman:"type=array[enum],options=etcd|master|worker|client"`
	ID              int      `yaml:"omitempty" json:"-"`
	IsEtcd          bool     `yaml:"omitempty"`
	IsMaster        bool     `yaml:"omitempty"`
	IsWorker        bool     `yaml:"omitempty"`
	IsClient        bool     `yaml:"omitempty"`
}

type Hosts struct {
	Hosts []HostCfg
}

func (cfg *ClusterCfg) GroupHosts() (*Hosts, *Hosts, *Hosts, *Hosts, *Hosts) {
	allHosts := Hosts{}
	etcdHosts := Hosts{}
	masterHosts := Hosts{}
	workerHosts := Hosts{}
	k8sHosts := Hosts{}

	for _, host := range cfg.Hosts {
		//clusterNode := HostCfg{}
		for _, role := range host.Role {
			if role == "etcd" {
				host.IsEtcd = true
			}
			if role == "master" {
				host.IsMaster = true
			}
			if role == "worker" {
				host.IsWorker = true
			}
		}
		if host.IsEtcd == true {
			etcdHosts.Hosts = append(etcdHosts.Hosts, host)
		}
		if host.IsMaster == true {
			masterHosts.Hosts = append(masterHosts.Hosts, host)
		}
		if host.IsWorker == true {
			workerHosts.Hosts = append(workerHosts.Hosts, host)
		}
		if host.IsMaster == true || host.IsWorker == true {
			k8sHosts.Hosts = append(k8sHosts.Hosts, host)
		}
		allHosts.Hosts = append(allHosts.Hosts, host)
	}
	return &allHosts, &etcdHosts, &masterHosts, &workerHosts, &k8sHosts
}
