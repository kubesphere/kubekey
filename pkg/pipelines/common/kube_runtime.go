package common

import (
	"fmt"
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	kubekeyclientset "github.com/kubesphere/kubekey/clients/clientset/versioned"
	kubekeycontroller "github.com/kubesphere/kubekey/controllers/kubekey"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
)

type KubeRuntime struct {
	connector.BaseRuntime
	Cluster    *kubekeyapiv1alpha1.ClusterSpec
	Kubeconfig string
	Conditions []kubekeyapiv1alpha1.Condition
	ClientSet  *kubekeyclientset.Clientset
	Arg        Argument
}

type Argument struct {
	FilePath           string
	KubernetesVersion  string
	KsEnable           bool
	KsVersion          string
	Debug              bool
	SkipCheck          bool
	SkipPullImages     bool
	AddImagesRepo      bool
	DeployLocalStorage bool
	SourcesDir         string
	InCluster          bool
}

func NewKubeRuntime(flag string, arg Argument) (connector.Runtime, error) {
	loader := NewLoader(flag, arg)
	cluster, err := loader.Load()
	if err != nil {
		return nil, err
	}
	clusterSpec := &cluster.Spec

	defaultCluster, hostGroups, err := clusterSpec.SetDefaultClusterSpec(arg.InCluster)
	if err != nil {
		return nil, err
	}

	var clientset *kubekeyclientset.Clientset
	if arg.InCluster {
		c, err := kubekeycontroller.NewKubekeyClient()
		if err != nil {
			return nil, err
		}
		clientset = c
	}

	base := connector.BaseRuntime{
		ObjName:      cluster.Name,
		ClusterHosts: generateHosts(hostGroups, defaultCluster),
		WorkDir:      generateWorkDir(),
		AllHosts:     make([]connector.Host, 0, 0),
		RoleHosts:    make(map[string][]connector.Host),
	}
	base.SetConnector(connector.NewDialer())
	for _, v := range hostGroups.All {
		host := ToHosts(v)
		if v.IsMaster {
			host.SetRole(Master)
		}
		if v.IsWorker {
			host.SetRole(Worker)
		}
		if v.IsEtcd {
			host.SetRole(Etcd)
		}
		host.SetRole(K8s)
		base.AppendHost(host)
		base.AppendRoleMap(host)
	}

	r := &KubeRuntime{
		Cluster:   defaultCluster,
		ClientSet: clientset,
		Arg:       arg,
	}
	r.BaseRuntime = base

	var runtime connector.Runtime
	runtime = r
	return runtime, nil
}

// Copy is used to create a copy for Runtime.
func (k *KubeRuntime) Copy() connector.Runtime {
	runtime := *k
	return &runtime
}

func ToHosts(cfg kubekeyapiv1alpha1.HostCfg) *connector.BaseHost {
	host := &connector.BaseHost{
		Name:           cfg.Name,
		Address:        cfg.InternalAddress,
		Port:           cfg.Port,
		User:           cfg.User,
		Password:       cfg.Password,
		PrivateKey:     cfg.PrivateKey,
		PrivateKeyPath: cfg.PrivateKeyPath,
		Arch:           cfg.Arch,
		Roles:          make([]string, 0, 0),
		RoleTable:      make(map[string]bool),
	}
	return host
}

func generateHosts(hostGroups *kubekeyapiv1alpha1.HostGroups, cfg *kubekeyapiv1alpha1.ClusterSpec) []string {
	var lbHost string
	var hostsList []string

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

func generateWorkDir() string {
	currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		logger.Log.Fatal(errors.Wrap(err, "Failed to get current dir"))
	}
	return fmt.Sprintf("%s/%s", currentDir, kubekeyapiv1alpha1.DefaultPreDir)
}
