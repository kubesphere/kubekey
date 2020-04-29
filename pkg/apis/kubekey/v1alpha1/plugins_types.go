package v1alpha1

type LocalVolume struct {
	IsDefaultClass bool `yaml:"isDefaultClass" json:"isDefaultClass,omitempty"`
}

type NfsClient struct {
	IsDefaultClass     bool   `yaml:"isDefaultClass" json:"isDefaultClass,omitempty"`
	NfsServer          string `yaml:"nfsServer" json:"nfsServer,omitempty"`
	NfsPath            string `yaml:"nfsPath" json:"nfsPath,omitempty"`
	NfsVrs3Enabled     bool   `yaml:"nfsVrs3Enabled" json:"nfsVrs3Enabled,omitempty"`
	NfsArchiveOnDelete bool   `yaml:"nfsArchiveOnDelete" json:"nfsArchiveOnDelete,omitempty"`
}

type GlusterFS struct {
	IsDefaultClass  bool
	RestAuthEnabled bool
	RestUrl         string
	ClusterID       string
	RestUser        string
	SecretName      string
	GidMin          int
	GidMax          int
	VolumeType      string
	JwtAdminKey     string
}

type CephRBD struct {
	IsDefaultClass bool
	Monitors       []string
	AdminID        string
	AdminSecret    string
	UserID         string
	UserSecret     string
	Pool           string
	FsType         string
	ImageFormat    string
	ImageFeatures  string
}
