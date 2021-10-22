package pipelines

import (
	"fmt"
	kubekeycontroller "github.com/kubesphere/kubekey/controllers/kubekey"
	"github.com/kubesphere/kubekey/pkg/addons"
	"github.com/kubesphere/kubekey/pkg/binaries"
	"github.com/kubesphere/kubekey/pkg/bootstrap/confirm"
	"github.com/kubesphere/kubekey/pkg/bootstrap/os"
	"github.com/kubesphere/kubekey/pkg/bootstrap/precheck"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/container"
	"github.com/kubesphere/kubekey/pkg/core/module"
	"github.com/kubesphere/kubekey/pkg/core/pipeline"
	"github.com/kubesphere/kubekey/pkg/etcd"
	"github.com/kubesphere/kubekey/pkg/images"
	"github.com/kubesphere/kubekey/pkg/k3s"
	"github.com/kubesphere/kubekey/pkg/kubernetes"
	"github.com/kubesphere/kubekey/pkg/kubesphere"
	"github.com/kubesphere/kubekey/pkg/loadbalancer"
	"github.com/kubesphere/kubekey/pkg/plugins/dns"
	"github.com/kubesphere/kubekey/pkg/plugins/network"
	"github.com/kubesphere/kubekey/pkg/plugins/storage"
)

func NewCreateClusterPipeline(runtime *common.KubeRuntime) error {
	noNetworkPlugin := runtime.Cluster.Network.Plugin == "" || runtime.Cluster.Network.Plugin == "none"

	m := []module.Module{
		&precheck.NodePreCheckModule{},
		&confirm.InstallConfirmModule{},
		&binaries.NodeBinariesModule{},
		&os.ConfigureOSModule{},
		&kubernetes.KubernetesStatusModule{},
		&container.InstallContainerModule{},
		&images.ImageModule{Skip: runtime.Arg.SkipPullImages},
		&etcd.ETCDPreCheckModule{},
		&etcd.ETCDCertsModule{},
		&etcd.InstallETCDBinaryModule{},
		&etcd.ETCDModule{},
		&etcd.BackupETCDModule{},
		&kubernetes.InstallKubeBinariesModule{},
		&kubernetes.InitKubernetesModule{},
		&dns.ClusterDNSModule{},
		&kubernetes.KubernetesStatusModule{},
		&kubernetes.JoinNodesModule{},
		&loadbalancer.HaproxyModule{Skip: !runtime.Cluster.ControlPlaneEndpoint.IsInternalLBEnabled()},
		&network.DeployNetworkPluginModule{},
		&addons.AddonsModule{Skip: noNetworkPlugin},
		&storage.DeployLocalVolumeModule{Skip: noNetworkPlugin || (!runtime.Arg.DeployLocalStorage && !runtime.Cluster.KubeSphere.Enabled)},
		&kubesphere.DeployModule{Skip: noNetworkPlugin},
		&kubesphere.CheckResultModule{Skip: noNetworkPlugin},
	}

	p := pipeline.Pipeline{
		Name:    "CreateClusterPipeline",
		Modules: m,
		Runtime: runtime,
	}
	if err := p.Start(); err != nil {
		return err
	}
	return nil
}

func NewK3sCreateClusterPipeline(runtime *common.KubeRuntime) error {
	noNetworkPlugin := runtime.Cluster.Network.Plugin == "" || runtime.Cluster.Network.Plugin == "none"

	m := []module.Module{
		&binaries.K3sNodeBinariesModule{},
		&os.ConfigureOSModule{},
		&k3s.StatusModule{},
		&etcd.ETCDPreCheckModule{},
		&etcd.ETCDCertsModule{},
		&etcd.InstallETCDBinaryModule{},
		&etcd.ETCDModule{},
		&etcd.BackupETCDModule{},
		&k3s.InstallKubeBinariesModule{},
		&k3s.InitClusterModule{},
		&k3s.StatusModule{},
		&k3s.JoinNodesModule{},
		&loadbalancer.K3sHaproxyModule{Skip: !runtime.Cluster.ControlPlaneEndpoint.IsInternalLBEnabled()},
		&network.DeployNetworkPluginModule{},
		&addons.AddonsModule{Skip: noNetworkPlugin},
		&storage.DeployLocalVolumeModule{Skip: noNetworkPlugin || (!runtime.Arg.DeployLocalStorage && !runtime.Cluster.KubeSphere.Enabled)},
		&kubesphere.DeployModule{Skip: noNetworkPlugin},
		&kubesphere.CheckResultModule{Skip: noNetworkPlugin},
	}

	p := pipeline.Pipeline{
		Name:    "K3sCreateClusterPipeline",
		Modules: m,
		Runtime: runtime,
	}
	if err := p.Start(); err != nil {
		return err
	}
	return nil
}

func CreateCluster(args common.Argument, downloadCmd string) error {
	args.DownloadCommand = func(path, url string) string {
		// this is an extension point for downloading tools, for example users can set the timeout, proxy or retry under
		// some poor network environment. Or users even can choose another cli, it might be wget.
		// perhaps we should have a build-in download function instead of totally rely on the external one
		return fmt.Sprintf(downloadCmd, path, url)
	}

	var loaderType string
	if args.FilePath != "" {
		loaderType = common.File
	} else {
		loaderType = common.AllInOne
	}

	runtime, err := common.NewKubeRuntime(loaderType, args)
	if err != nil {
		return err
	}
	if args.InCluster {
		c, err := kubekeycontroller.NewKubekeyClient()
		if err != nil {
			return err
		}
		runtime.ClientSet = c
	}

	switch runtime.Cluster.Kubernetes.Type {
	case common.K3s:
		if err := NewK3sCreateClusterPipeline(runtime); err != nil {
			return err
		}
	case common.Kubernetes:
		if err := NewCreateClusterPipeline(runtime); err != nil {
			return err
		}
	default:
		if err := NewCreateClusterPipeline(runtime); err != nil {
			return err
		}
	}
	return nil
}
