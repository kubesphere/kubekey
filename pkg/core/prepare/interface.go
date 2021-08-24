package prepare

import (
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/config"
)

type Prepare interface {
	PreCheck(runtime *config.Runtime) (bool, error)
	Init(cache *cache.Cache, rootCache *cache.Cache)
}
