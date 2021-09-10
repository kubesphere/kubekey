package pipelines

import (
	"github.com/kubesphere/kubekey/pkg/core/modules"
	"github.com/kubesphere/kubekey/pkg/core/pipeline"
	"github.com/kubesphere/kubekey/pkg/pipelines/bootstrap/confirm"
	"github.com/kubesphere/kubekey/pkg/pipelines/bootstrap/os"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"github.com/kubesphere/kubekey/pkg/pipelines/kubernetes"
)

func NewDeleteClusterPipeline(runtime *common.KubeRuntime) error {
	m := []modules.Module{
		&confirm.DeleteClusterConfirmModule{},
		&kubernetes.ResetClusterModule{},
		&os.ClearOSEnvironmentModule{},
	}

	p := pipeline.Pipeline{
		Name:    "DeleteClusterPipeline",
		Modules: m,
		Runtime: runtime,
	}
	if err := p.Start(); err != nil {
		return err
	}
	return nil
}

func DeleteCluster(args common.Argument) error {
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

	if err := NewDeleteClusterPipeline(runtime); err != nil {
		return err
	}
	return nil
}
