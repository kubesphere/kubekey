package kubernetes

import (
	"github.com/kubesphere/kubekey/experiment/core/config"
	"github.com/kubesphere/kubekey/experiment/core/pipeline"
	"github.com/kubesphere/kubekey/experiment/pipeline/kubernetes/module/control_plane"
)

// todo: 有个缺点，会在第一时间初始化所有的module对象，然后才进行执行。考虑是否修改为责任链模式
func NewGetClusterInfoPipeline(runtime *config.Runtime) error {

	modules := []pipeline.Module{
		control_plane.NewGetClusterStatusModule(runtime),
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
