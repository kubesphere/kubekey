package install

import (
	"github.com/pixiake/kubekey/util/manager"
	ssh2 "github.com/pixiake/kubekey/util/ssh"
	"github.com/sirupsen/logrus"

	kubekeyapi "github.com/pixiake/kubekey/apis/v1alpha1"
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
type Executor struct {
	cluster *kubekeyapi.ClusterCfg
	logger  *logrus.Logger
}

// NewInstaller returns a new installer, responsible for dispatching
// between the different supported Kubernetes versions and running the
func NewExecutor(cluster *kubekeyapi.ClusterCfg, logger *logrus.Logger) *Executor {
	return &Executor{
		cluster: cluster,
		logger:  logger,
	}
}

// Install run the installation process
func (executor *Executor) Execute() error {
	mgr, err := executor.createManager()
	if err != nil {
		return err
	}
	return ExecTasks(mgr)
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

func (executor *Executor) createManager() (*manager.Manager, error) {
	mgr := &manager.Manager{}

	mgr.Cluster = executor.cluster
	mgr.Connector = ssh2.NewConnector()
	//s.Configuration = configupload.NewConfiguration()
	//mgr.WorkDir = "kubekey"
	mgr.Logger = executor.logger
	mgr.Verbose = true
	//s.ManifestFilePath = options.Manifest
	//s.CredentialsFilePath = options.CredentialsFile
	//s.BackupFile = options.BackupFile
	//s.DestroyWorkers = options.DestroyWorkers
	//s.RemoveBinaries = options.RemoveBinaries
	return mgr, nil
}
