package prepare

import (
	"github.com/kubesphere/kubekey/experiment/utils/cache"
	"github.com/kubesphere/kubekey/experiment/utils/config"
)

type BasePrepare struct {
	mgr  *config.Manager
	Pool *cache.Cache
}

func (b *BasePrepare) Init(mgr *config.Manager, pool *cache.Cache) {
	b.mgr = mgr
	b.Pool = pool
}

func (b *BasePrepare) PreCheck() (bool, error) {
	return true, nil
}
