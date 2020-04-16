package manager

import (
	ssh2 "github.com/pixiake/kubekey/util/ssh"
	"github.com/sirupsen/logrus"

	kubekeyapi "github.com/pixiake/kubekey/apis/v1alpha1"

	"github.com/pixiake/kubekey/util/runner"
	//"k8s.io/client-go/rest"
	//bootstraputil "k8s.io/cluster-bootstrap/token/util"
	//dynclient "sigs.k8s.io/manager-runtime/pkg/client"
)

//func New() (*State, error) {
//	joinToken, err := bootstraputil.GenerateBootstrapToken()
//	return &State{
//		JoinToken: joinToken,
//	}, err
//}

// State holds together currently test flags and parsed info, along with
// utilities like logger
type Manager struct {
	Cluster   *kubekeyapi.ClusterCfg
	Logger    logrus.FieldLogger
	Connector *ssh2.Dialer
	//Configuration             *configupload.Configuration
	Runner      *runner.Runner
	WorkDir     string
	JoinCommand string
	JoinToken   string
	//RESTConfig                *rest.Config
	//DynamicClient             dynclient.Client
	Verbose bool
	//BackupFile                string
	//DestroyWorkers            bool
	//RemoveBinaries            bool
	//ForceUpgrade              bool
	//UpgradeMachineDeployments bool
	//PatchCNI                  bool
	//CredentialsFilePath       string
	//ManifestFilePath          string
}

func (mgr *Manager) KubeadmVerboseFlag() string {
	if mgr.Verbose {
		return "--v=6"
	}
	return ""
}

// Clone returns a shallow copy of the State.
func (mgr *Manager) Clone() *Manager {
	newState := *mgr
	return &newState
}
