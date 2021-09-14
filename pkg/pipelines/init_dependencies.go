package pipelines

import (
	"github.com/kubesphere/kubekey/pkg/core/modules"
	"github.com/kubesphere/kubekey/pkg/core/pipeline"
	"github.com/kubesphere/kubekey/pkg/pipelines/bootstrap/os"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
)

func NewInitDependenciesPipeline(runtime *common.KubeRuntime) error {
	m := []modules.Module{
		&os.InitDependenciesModule{},
	}

	p := pipeline.Pipeline{
		Name:    "InitDependenciesPipeline",
		Modules: m,
		Runtime: runtime,
	}
	if err := p.Start(); err != nil {
		return err
	}
	return nil
}

func InitDependencies(args common.Argument) error {
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

	if err := NewInitDependenciesPipeline(runtime); err != nil {
		return err
	}
	return nil
}
