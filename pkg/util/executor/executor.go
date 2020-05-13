package executor

import (
	"fmt"
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/kubesphere/kubekey/pkg/util/ssh"
	log "github.com/sirupsen/logrus"
)

type Executor struct {
	cluster *kubekeyapi.K2ClusterSpec
	logger  *log.Logger
	Verbose bool
}

func NewExecutor(cluster *kubekeyapi.K2ClusterSpec, logger *log.Logger, verbose bool) *Executor {
	return &Executor{
		cluster: cluster,
		logger:  logger,
		Verbose: verbose,
	}
}

func (executor *Executor) CreateManager() (*manager.Manager, error) {
	mgr := &manager.Manager{}
	defaultCluster, hostGroups := executor.cluster.SetDefaultK2ClusterSpec()
	mgr.AllNodes = hostGroups.All
	mgr.EtcdNodes = hostGroups.Etcd
	mgr.MasterNodes = hostGroups.Master
	mgr.WorkerNodes = hostGroups.Worker
	mgr.K8sNodes = hostGroups.K8s
	mgr.ClientNode = hostGroups.Client
	mgr.Cluster = defaultCluster
	mgr.ClusterHosts = GenerateHosts(hostGroups, defaultCluster)
	mgr.Connector = ssh.NewConnector()
	mgr.Logger = executor.logger
	mgr.Verbose = executor.Verbose

	return mgr, nil
}

func GenerateHosts(hostGroups *kubekeyapi.HostGroups, cfg *kubekeyapi.K2ClusterSpec) []string {
	var lbHost string
	hostsList := []string{}

	if cfg.ControlPlaneEndpoint.Address != "" {
		lbHost = fmt.Sprintf("%s  %s", cfg.ControlPlaneEndpoint.Address, cfg.ControlPlaneEndpoint.Domain)
	} else {
		lbHost = fmt.Sprintf("%s  %s", hostGroups.Master[0].InternalAddress, cfg.ControlPlaneEndpoint.Domain)
	}

	for _, host := range cfg.Hosts {
		if host.Name != "" {
			hostsList = append(hostsList, fmt.Sprintf("%s  %s.%s %s", host.InternalAddress, host.Name, cfg.Kubernetes.ClusterName, host.Name))
		}
	}

	hostsList = append(hostsList, lbHost)
	return hostsList
}
