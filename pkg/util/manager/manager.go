package manager

import (
	ssh2 "github.com/kubesphere/kubekey/pkg/util/ssh"
	"github.com/sirupsen/logrus"

	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"

	"github.com/kubesphere/kubekey/pkg/util/runner"
	//"k8s.io/client-go/rest"
	//bootstraputil "k8s.io/cluster-bootstrap/token/util"
	//dynclient "sigs.k8s.io/manager-runtime/pkg/client"
)

type Manager struct {
	Cluster      *kubekeyapi.ClusterSpec
	Logger       logrus.FieldLogger
	Connector    *ssh2.Dialer
	Runner       *runner.Runner
	AllNodes     []kubekeyapi.HostCfg
	EtcdNodes    []kubekeyapi.HostCfg
	MasterNodes  []kubekeyapi.HostCfg
	WorkerNodes  []kubekeyapi.HostCfg
	K8sNodes     []kubekeyapi.HostCfg
	ClientNode   []kubekeyapi.HostCfg
	ClusterHosts []string
	WorkDir      string
	JoinCommand  string
	JoinToken    string
	Verbose      bool
}

func (mgr *Manager) KubeadmVerboseFlag() string {
	if mgr.Verbose {
		return "--v=6"
	}
	return ""
}

func (mgr *Manager) Clone() *Manager {
	newManager := *mgr
	return &newManager
}
