package action

import (
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/connector"
)

type Action interface {
	Execute(runtime connector.Runtime) (err error)
	Init(cache *cache.Cache, rootCache *cache.Cache)
	AutoAssert(runtime connector.Runtime)
}
