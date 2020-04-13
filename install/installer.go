package install

import (
	"github.com/sirupsen/logrus"

	kubekeyapi "github.com/pixiake/kubekey/apis/v1alpha1"
	//"github.com/kubermatic/kubeone/pkg/configupload"
	"github.com/pixiake/kubekey/util/dialer/ssh"
	"github.com/pixiake/kubekey/util/state"
)

// Options groups the various possible options for running
// the Kubernetes installation.
type Options struct {
	Verbose         bool
	Manifest        string
	CredentialsFile string
	BackupFile      string
	DestroyWorkers  bool
	RemoveBinaries  bool
}

// Installer is entrypoint for installation process
type Installer struct {
	cluster *kubekeyapi.ClusterCfg
	logger  *logrus.Logger
}

// NewInstaller returns a new installer, responsible for dispatching
// between the different supported Kubernetes versions and running the
func NewInstaller(cluster *kubekeyapi.ClusterCfg, logger *logrus.Logger) *Installer {
	return &Installer{
		cluster: cluster,
		logger:  logger,
	}
}

// Install run the installation process
func (i *Installer) Install() error {
	s, err := i.createState()
	if err != nil {
		return err
	}
	return ExecTasks(s)
}

// Reset resets cluster:
// * destroys all the worker machines
// * kubeadm reset masters
//func (i *Installer) Reset(options *Options) error {
//	s, err := i.createState(options)
//	if err != nil {
//		return err
//	}
//	return installation.Reset(s)
//}

// createState creates a basic, non-host bound state with
// all relevant information, but *no* Runner yet. The various
// task helper functions will take care of setting up Runner
// structs for each task individually.
func (i *Installer) createState() (*state.State, error) {
	s := &state.State{}

	s.Cluster = i.cluster
	s.Connector = ssh.NewConnector()
	//s.Configuration = configupload.NewConfiguration()
	s.WorkDir = "kubekey"
	s.Logger = i.logger
	s.Verbose = true
	//s.ManifestFilePath = options.Manifest
	//s.CredentialsFilePath = options.CredentialsFile
	//s.BackupFile = options.BackupFile
	//s.DestroyWorkers = options.DestroyWorkers
	//s.RemoveBinaries = options.RemoveBinaries
	return s, nil
}
