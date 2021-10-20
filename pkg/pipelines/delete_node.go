package pipelines

import (
	"github.com/kubesphere/kubekey/pkg/bootstrap/config"
	"github.com/kubesphere/kubekey/pkg/bootstrap/confirm"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/modules"
	"github.com/kubesphere/kubekey/pkg/core/pipeline"
	"github.com/kubesphere/kubekey/pkg/kubernetes"
)

func DeleteNodePipeline(runtime *common.KubeRuntime) error {
	m := []modules.Module{
		&confirm.DeleteNodeConfirmModule{},
		&config.ModifyConfigModule{},
		&kubernetes.CompareConfigAndClusterInfoModule{},
		&kubernetes.DeleteKubeNodeModule{},
	}

	p := pipeline.Pipeline{
		Name:    "AddNodesPipeline",
		Modules: m,
		Runtime: runtime,
	}
	if err := p.Start(); err != nil {
		return err
	}
	return nil
}

func DeleteNode(args common.Argument) error {
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

	if err := DeleteNodePipeline(runtime); err != nil {
		return err
	}
	return nil
}
