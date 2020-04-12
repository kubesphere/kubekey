package v1alpha1

type HostConfig struct {
	ID                int    `json:"-"`
	PublicAddress     string `json:"publicAddress"`
	PrivateAddress    string `json:"privateAddress"`
	SSHPort           int    `json:"sshPort"`
	SSHUsername       string `json:"sshUsername"`
	SSHPrivateKeyFile string `json:"sshPrivateKeyFile"`
	SSHAgentSocket    string `json:"sshAgentSocket"`
	Bastion           string `json:"bastion"`
	BastionPort       int    `json:"bastionPort"`
	BastionUser       string `json:"bastionUser"`
	Hostname          string `json:"hostname"`
	IsLeader          bool   `json:"isLeader"`
	Untaint           bool   `json:"untaint"`

	// Information populated at the runtime
	OperatingSystem string `json:"-"`
}
