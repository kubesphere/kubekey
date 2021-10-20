package pipelines

import (
	"github.com/kubesphere/kubekey/pkg/bootstrap/confirm"
	"github.com/kubesphere/kubekey/pkg/bootstrap/os"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/modules"
	"github.com/kubesphere/kubekey/pkg/core/pipeline"
	"github.com/kubesphere/kubekey/pkg/k3s"
	"github.com/kubesphere/kubekey/pkg/kubernetes"
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

func NewK3sDeleteClusterPipeline(runtime *common.KubeRuntime) error {
	m := []modules.Module{
		&confirm.DeleteClusterConfirmModule{},
		&k3s.DeleteClusterModule{},
		&os.ClearOSEnvironmentModule{},
	}

	p := pipeline.Pipeline{
		Name:    "K3sDeleteClusterPipeline",
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

	switch runtime.Cluster.Kubernetes.Type {
	case common.K3s:
		if err := NewK3sDeleteClusterPipeline(runtime); err != nil {
			return err
		}
	case common.Kubernetes:
		if err := NewDeleteClusterPipeline(runtime); err != nil {
			return err
		}
	default:
		if err := NewDeleteClusterPipeline(runtime); err != nil {
			return err
		}
	}
	return nil
}
