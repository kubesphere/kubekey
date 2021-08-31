package prepare

import (
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/connector"
)

type Prepare interface {
	PreCheck(runtime connector.Runtime) (bool, error)
	Init(cache *cache.Cache, rootCache *cache.Cache, runtime connector.Runtime)
	AutoAssert()
}
