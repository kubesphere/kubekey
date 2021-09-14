package modules

import (
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/connector"
)

type Module interface {
	IsSkip() bool
	Default(runtime connector.Runtime, pipelineCache *cache.Cache, moduleCache *cache.Cache)
	Init()
	Is() string
	Run() error
	Slogan()
	AutoAssert()
}
