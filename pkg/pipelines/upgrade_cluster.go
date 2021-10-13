package pipelines

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/core/modules"
	"github.com/kubesphere/kubekey/pkg/core/pipeline"
	"github.com/kubesphere/kubekey/pkg/pipelines/bootstrap/confirm"
	"github.com/kubesphere/kubekey/pkg/pipelines/bootstrap/os"
	"github.com/kubesphere/kubekey/pkg/pipelines/bootstrap/precheck"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"github.com/kubesphere/kubekey/pkg/pipelines/kubernetes"
	"github.com/kubesphere/kubekey/pkg/pipelines/loadbalancer"
)

func NewUpgradeClusterPipeline(runtime *common.KubeRuntime) error {
	isK3s := runtime.Cluster.Kubernetes.Type == "k3s"

	m := []modules.Module{
		&precheck.NodePreCheckModule{Skip: isK3s},
		&precheck.ClusterPreCheckModule{},
		&confirm.UpgradeConfirmModule{},
		&os.ConfigureOSModule{},
		&kubernetes.ProgressiveUpgradeModule{Step: kubernetes.ToV121},
		&kubernetes.ProgressiveUpgradeModule{Step: kubernetes.ToV122},
		&loadbalancer.HaproxyModule{Skip: !runtime.Cluster.ControlPlaneEndpoint.IsInternalLBEnabled()},
	}

	p := pipeline.Pipeline{
		Name:    "UpgradeClusterPipeline",
		Modules: m,
		Runtime: runtime,
	}
	if err := p.Start(); err != nil {
		return err
	}
	return nil
}

func UpgradeCluster(args common.Argument, downloadCmd string) error {
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

	if err := NewUpgradeClusterPipeline(runtime); err != nil {
		return err
	}
	return nil
}
