package pipelines

import (
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/modules"
	"github.com/kubesphere/kubekey/pkg/core/pipeline"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"github.com/kubesphere/kubekey/pkg/pipelines/initialization"
)

func NewCreateClusterPipeline(runtime connector.Runtime) error {

	m := []modules.Module{
		&initialization.NodeInitializationModule{},
		&initialization.ConfirmModule{},
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

func CreateCluster(clusterCfgFile, k8sVersion, ksVersion string, ksEnabled, verbose, skipCheck, skipPullImages, inCluster, deployLocalStorage bool) error {
	arg := common.Argument{
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
		loaderType = common.File
	} else {
		loaderType = common.AllInOne
	}

	runtime, err := common.NewKubeRuntime(loaderType, arg)
	if err != nil {
		return err
	}

	if err := NewCreateClusterPipeline(runtime); err != nil {
		return err
	}
	return nil
}
