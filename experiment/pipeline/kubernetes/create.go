package kubernetes

import (
	"github.com/kubesphere/kubekey/experiment/core/config"
	"github.com/kubesphere/kubekey/experiment/core/pipeline"
	"github.com/kubesphere/kubekey/experiment/pipeline/kubernetes/module/control_plane"
)

func NewGetClusterInfoPipeline(runtime *config.Runtime) error {

	modules := []pipeline.Module{
		&control_plane.GetClusterStatusModule{},
	}

	p := pipeline.Pipeline{
		Modules: modules,
		Runtime: runtime,
	}
	if err := p.Start(); err != nil {
		return err
	}
	return nil
}

func GetClusterInfo(clusterCfgFile, k8sVersion, ksVersion string, ksEnabled, verbose, skipCheck, skipPullImages, inCluster, deployLocalStorage bool) error {
	arg := config.Argument{
		FilePath:           clusterCfgFile,
		KubernetesVersion:  k8sVersion,
		KsEnable:           ksEnabled,
		KsVersion:          ksVersion,
		SkipCheck:          skipCheck,
		SkipPullImages:     skipPullImages,
		InCluster:          inCluster,
		DeployLocalStorage: deployLocalStorage,
		Debug:              verbose,
	}

	runtime, err := config.NewRuntime(config.File, arg)
	if err != nil {
		return err
	}

	if err := NewGetClusterInfoPipeline(runtime); err != nil {
		return err
	}
	return nil
}
