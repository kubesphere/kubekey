package install

import (
	kubekeyapi "github.com/pixiake/kubekey/apis/v1alpha1"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"kubeone/pkg/configupload"
	"kubeone/pkg/credentials"
	"kubeone/pkg/installer"
	"kubeone/pkg/installer/installation"
	"kubeone/pkg/state"
	//"kubeone/pkg/credentials"
	//"kubeone/pkg/installer"
	//"kubeone/pkg/installer/installation"
)

type Installer struct {
	cluster *kubekeyapi.ClusterCfg
	logger  *log.Logger
}

type Options struct {
	Verbose         bool
	Manifest        string
	CredentialsFile string
	BackupFile      string
	DestroyWorkers  bool
	RemoveBinaries  bool
}

func NewInstaller(cluster *kubekeyapi.ClusterCfg, logger *log.Logger) *Installer {
	return &Installer{
		cluster: cluster,
		logger:  logger,
	}
}

func (i *Installer) Install(options *Options) error {
	s, err := i.createState(options)
	if err != nil {
		return err
	}
	return installation.Install(s)
}

func (i *Installer) createState(options *Options) (*state.State, error) {
	s, err := state.New()
	if err != nil {
		return nil, err
	}

	s.Cluster = i.cluster
	s.Connector = ssh.NewConnector()
	s.Configuration = configupload.NewConfiguration()
	s.WorkDir = "kubeone"
	s.Logger = i.logger
	s.Verbose = options.Verbose
	s.ManifestFilePath = options.Manifest
	s.CredentialsFilePath = options.CredentialsFile
	s.BackupFile = options.BackupFile
	s.DestroyWorkers = options.DestroyWorkers
	s.RemoveBinaries = options.RemoveBinaries
	return s, nil
}

func CreateCluster(clusterCfgFile string, addons string, pkg string) {
	cluster := kubekeyapi.GetClusterCfg(clusterCfgFile)
}

func runInstall(logger *log.Logger) error {
	cluster, err := loadClusterConfig(installOptions.Manifest, installOptions.TerraformState, installOptions.CredentialsFilePath, logger)
	if err != nil {
		return errors.Wrap(err, "failed to load cluster")
	}

	options, err := createInstallerOptions(installOptions.Manifest, cluster, installOptions)
	if err != nil {
		return errors.Wrap(err, "failed to create installer options")
	}

	// Validate credentials
	_, err = credentials.ProviderCredentials(cluster.CloudProvider.Name, installOptions.CredentialsFilePath)
	if err != nil {
		return errors.Wrap(err, "failed to validate credentials")
	}

	return installer.NewInstaller(cluster, logger).Install(options)
}
