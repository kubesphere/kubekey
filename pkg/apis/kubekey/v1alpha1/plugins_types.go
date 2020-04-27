package v1alpha1

type LocalVolume struct {
	IsDefaultClass bool
}

type NfsClient struct {
	IsDefaultClass     bool
	NfsServer          string
	NfsPath            string
	NfsVrs3Enabled     bool
	NfsArchiveOnDelete bool
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
