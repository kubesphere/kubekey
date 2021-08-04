package action

import (
	"github.com/kubesphere/kubekey/experiment/utils/cache"
	"github.com/kubesphere/kubekey/experiment/utils/config"
	"github.com/kubesphere/kubekey/experiment/utils/ending"
	"github.com/kubesphere/kubekey/experiment/utils/vars"
)

type BaseAction struct {
	Manager *config.Manager
	Pool    *cache.Cache
}

func (b *BaseAction) Init(mgr *config.Manager, pool *cache.Cache) {
	b.Manager = mgr
	b.Pool = pool
}

func (b *BaseAction) Execute(vars vars.Vars) (result *ending.Result) {
	return nil
}
