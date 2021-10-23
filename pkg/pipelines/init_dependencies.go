package pipelines

import (
	"github.com/kubesphere/kubekey/pkg/bootstrap/os"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/module"
	"github.com/kubesphere/kubekey/pkg/core/pipeline"
)

func NewInitDependenciesPipeline(runtime *common.KubeRuntime) error {
	m := []module.Module{
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
