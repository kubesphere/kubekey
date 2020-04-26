package scale

import (
	kubekeyapi "github.com/pixiake/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/pixiake/kubekey/pkg/util/manager"
	ssh "github.com/pixiake/kubekey/pkg/util/ssh"
	log "github.com/sirupsen/logrus"
)

type Executor struct {
	cluster *kubekeyapi.K2ClusterSpec
	logger  *log.Logger
}

func NewExecutor(cluster *kubekeyapi.K2ClusterSpec, logger *log.Logger) *Executor {
	return &Executor{
		cluster: cluster,
		logger:  logger,
	}
}

func (executor *Executor) Execute() error {
	mgr, err := executor.createManager()
	if err != nil {
		return err
	}
	return ExecTasks(mgr)
}

func (executor *Executor) createManager() (*manager.Manager, error) {
	mgr := &manager.Manager{}
	allNodes, etcdNodes, masterNodes, workerNodes, k8sNodes, clientNode := executor.cluster.GroupHosts()
	mgr.AllNodes = allNodes
	mgr.EtcdNodes = etcdNodes
	mgr.MasterNodes = masterNodes
	mgr.WorkerNodes = workerNodes
	mgr.K8sNodes = k8sNodes
	mgr.ClientNode = clientNode
	mgr.Cluster = executor.cluster
	mgr.Connector = ssh.NewConnector()
	mgr.Logger = executor.logger
	mgr.Verbose = true

	return mgr, nil
}
