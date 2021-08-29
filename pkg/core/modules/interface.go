package modules

import (
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/connector"
)

type Module interface {
	Default(runtime connector.Runtime, rootCache *cache.Cache, moduleCache *cache.Cache)
	Init()
	Is() string
	Run() error
	AutoAssert()
}
