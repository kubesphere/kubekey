package modules

import (
	"github.com/kubesphere/kubekey/experiment/core/cache"
	"github.com/kubesphere/kubekey/experiment/core/config"
)

type Module interface {
	Default(runtime *config.Runtime, rootCache *cache.Cache)
	Init()
	Is() string
	Run() error
}
