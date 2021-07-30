package config

import (
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	kubekeyclientset "github.com/kubesphere/kubekey/clients/clientset/versioned"
	"github.com/kubesphere/kubekey/pkg/connector"
	"github.com/kubesphere/kubekey/pkg/util/runner"
	log "github.com/sirupsen/logrus"
	"os"
	"sync"
)

type GlobalConfig struct {
	ObjName            string
	Cluster            *kubekeyapiv1alpha1.ClusterSpec
	Logger             log.FieldLogger
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
	globalConfig          *GlobalConfig
	globalConfigSingleton sync.Once
)

func GetConfig() *GlobalConfig {
	globalConfigSingleton.Do(func() {
		loader := NewLoader(os.Args[0])
		if err := loader.Load(globalConfig); err != nil {
			os.Exit(1)
		}
	})
	return globalConfig
}
