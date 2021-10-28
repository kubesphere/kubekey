package task

import (
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/ending"
)

type Interface interface {
	GetDesc() string
	Init(runtime connector.Runtime, moduleCache *cache.Cache, pipelineCache *cache.Cache)
	Execute() *ending.TaskResult
}
