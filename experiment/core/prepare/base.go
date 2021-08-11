package prepare

import (
	"github.com/kubesphere/kubekey/experiment/core/cache"
	"github.com/kubesphere/kubekey/experiment/core/config"
)

type BasePrepare struct {
	mgr  *config.Runtime
	Pool *cache.Cache
}

func (b *BasePrepare) Init(mgr *config.Runtime, pool *cache.Cache) {
	b.mgr = mgr
	b.Pool = pool
}

func (b *BasePrepare) PreCheck() (bool, error) {
	return true, nil
}
