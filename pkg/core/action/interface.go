package action

import (
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/config"
	"github.com/kubesphere/kubekey/pkg/core/vars"
)

type Action interface {
	Execute(runtime *config.Runtime, vars vars.Vars) (err error)
	Init(cache *cache.Cache, rootCache *cache.Cache)
}
