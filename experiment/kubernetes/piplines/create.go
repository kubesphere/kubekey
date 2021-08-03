package piplines

import (
	"github.com/kubesphere/kubekey/experiment/kubernetes/control-plane/tasks"
	"github.com/kubesphere/kubekey/experiment/utils/pipeline"
)

var (
	CreateClusterPipeline = pipeline.Pipeline{Modules: []pipeline.Module{
		{Tasks: []pipeline.Task{
			tasks.InitCluster,
			tasks.GetKubeConfig,
		}}}}
)
