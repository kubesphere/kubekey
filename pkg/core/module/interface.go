package module

import (
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/ending"
	"github.com/kubesphere/kubekey/pkg/core/hook"
)

type Module interface {
	IsSkip() bool
	Default(runtime connector.Runtime, pipelineCache *cache.Cache, moduleCache *cache.Cache)
	Init()
	Is() string
	Run(result *ending.ModuleResult)
	Until() (*bool, error)
	Slogan()
	AutoAssert()
	AppendPostHook(h hook.PostHook)
	CallPostHook(result *ending.ModuleResult) error
}

type Task interface {
	GetDesc() string
	Init(runtime connector.Runtime, moduleCache *cache.Cache, pipelineCache *cache.Cache)
	Execute() *ending.TaskResult
}
