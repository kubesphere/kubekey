package pipelines

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/bootstrap/confirm"
	"github.com/kubesphere/kubekey/pkg/bootstrap/os"
	"github.com/kubesphere/kubekey/pkg/bootstrap/precheck"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/modules"
	"github.com/kubesphere/kubekey/pkg/core/pipeline"
	"github.com/kubesphere/kubekey/pkg/kubernetes"
	"github.com/kubesphere/kubekey/pkg/kubesphere"
	"github.com/kubesphere/kubekey/pkg/loadbalancer"
)

func NewUpgradeClusterPipeline(runtime *common.KubeRuntime) error {
	m := []modules.Module{
		&precheck.NodePreCheckModule{},
		&precheck.ClusterPreCheckModule{},
		&confirm.UpgradeConfirmModule{},
		&os.ConfigureOSModule{},
		&kubernetes.SetUpgradePlanModule{Step: kubernetes.ToV121},
		&kubernetes.ProgressiveUpgradeModule{Step: kubernetes.ToV121},
		&loadbalancer.HaproxyModule{Skip: !runtime.Cluster.ControlPlaneEndpoint.IsInternalLBEnabled()},
		&kubesphere.ConvertModule{},
		&kubesphere.DeployModule{},
		&kubesphere.CheckResultModule{},
		&kubernetes.SetUpgradePlanModule{Step: kubernetes.ToV122},
		&kubernetes.ProgressiveUpgradeModule{Step: kubernetes.ToV122},
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
