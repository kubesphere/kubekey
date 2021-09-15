package pipelines

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/core/modules"
	"github.com/kubesphere/kubekey/pkg/core/pipeline"
	"github.com/kubesphere/kubekey/pkg/pipelines/addons"
	"github.com/kubesphere/kubekey/pkg/pipelines/binaries"
	"github.com/kubesphere/kubekey/pkg/pipelines/bootstrap/confirm"
	"github.com/kubesphere/kubekey/pkg/pipelines/bootstrap/os"
	"github.com/kubesphere/kubekey/pkg/pipelines/bootstrap/precheck"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"github.com/kubesphere/kubekey/pkg/pipelines/container"
	"github.com/kubesphere/kubekey/pkg/pipelines/etcd"
	"github.com/kubesphere/kubekey/pkg/pipelines/images"
	"github.com/kubesphere/kubekey/pkg/pipelines/kubernetes"
	"github.com/kubesphere/kubekey/pkg/pipelines/loadbalancer"
	"github.com/kubesphere/kubekey/pkg/pipelines/plugins/dns"
	"github.com/kubesphere/kubekey/pkg/pipelines/plugins/network"
)

func NewCreateClusterPipeline(runtime *common.KubeRuntime) error {

	isK3s := runtime.Cluster.Kubernetes.Type == "k3s"

	m := []modules.Module{
		&precheck.NodePreCheckModule{Skip: isK3s},
		&confirm.InstallConfirmModule{Skip: isK3s},
		&binaries.NodeBinariesModule{},
		&os.ConfigureOSModule{},
		&kubernetes.KubernetesStatusModule{},
		&container.InstallContainerModule{Skip: isK3s},
		&images.ImageModule{Skip: isK3s || runtime.Arg.SkipPullImages},
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
		&addons.AddonsModule{},
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

	if err := NewCreateClusterPipeline(runtime); err != nil {
		return err
	}
	return nil
}
