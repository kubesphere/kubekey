package config

import (
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	kubekeyclientset "github.com/kubesphere/kubekey/clients/clientset/versioned"
	"github.com/kubesphere/kubekey/experiment/utils/connector"
	"github.com/kubesphere/kubekey/experiment/utils/logger"
	"github.com/kubesphere/kubekey/experiment/utils/runner"
	"os"
	"sync"
)

type Manager struct {
	ObjName            string
	Cluster            *kubekeyapiv1alpha1.ClusterSpec
	Logger             logger.KubeKeyLog
	Connector          connector.Connector
	Runner             *runner.Runner
	AllNodes           []kubekeyapiv1alpha1.HostCfg
	EtcdNodes          []kubekeyapiv1alpha1.HostCfg
	MasterNodes        []kubekeyapiv1alpha1.HostCfg
	WorkerNodes        []kubekeyapiv1alpha1.HostCfg
	K8sNodes           []kubekeyapiv1alpha1.HostCfg
	EtcdContainer      bool
	ClusterHosts       []string
	WorkDir            string
	KsEnable           bool
	KsVersion          string
	Debug              bool
	SkipCheck          bool
	SkipPullImages     bool
	SourcesDir         string
	AddImagesRepo      bool
	InCluster          bool
	DeployLocalStorage bool
	Kubeconfig         string
	Conditions         []kubekeyapiv1alpha1.Condition
	ClientSet          *kubekeyclientset.Clientset
	DownloadCommand    func(path, url string) string
}

var (
	manager          *Manager
	managerSingleton sync.Once
)

func GetManager() *Manager {
	managerSingleton.Do(func() {
		loader := NewLoader(os.Args[0])
		if err := loader.Load(manager); err != nil {
			os.Exit(1)
		}
	})
	return manager
}

// Copy is used to create a copy for Manager.
func (mgr *Manager) Copy() *Manager {
	newManager := *mgr
	return &newManager
}
