package pipelines

import (
	"github.com/kubesphere/kubekey/experiment/core/config"
	"github.com/kubesphere/kubekey/experiment/core/modules"
	"github.com/kubesphere/kubekey/experiment/core/pipeline"
	"github.com/kubesphere/kubekey/experiment/pipelines/initialization"
)

func NewCreateClusterPipeline(runtime *config.Runtime) error {

	modules := []modules.Module{
		&initialization.InitializationModule{},
		&initialization.ConfirmModule{},
	}

	p := pipeline.Pipeline{
		Name:    "CreateClusterPipeline",
		Modules: modules,
		Runtime: runtime,
	}
	if err := p.Start(); err != nil {
		return err
	}
	return nil
}

func CreateCluster(clusterCfgFile, k8sVersion, ksVersion string, ksEnabled, verbose, skipCheck, skipPullImages, inCluster, deployLocalStorage bool) error {
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

	var loaderType string
	if clusterCfgFile != "" {
		loaderType = config.File
	} else {
		loaderType = config.AllInOne
	}

	runtime, err := config.NewRuntime(loaderType, arg)
	if err != nil {
		return err
	}

	if err := NewCreateClusterPipeline(runtime); err != nil {
		return err
	}
	return nil
}
