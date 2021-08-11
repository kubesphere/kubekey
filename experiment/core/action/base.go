package action

import (
	"github.com/kubesphere/kubekey/experiment/core/cache"
	"github.com/kubesphere/kubekey/experiment/core/config"
	"github.com/kubesphere/kubekey/experiment/core/ending"
	"github.com/kubesphere/kubekey/experiment/core/vars"
)

type BaseAction struct {
	Manager *config.Runtime
	Cache   *cache.Cache
	Result  *ending.Result
}

func (b *BaseAction) Init(mgr *config.Runtime, cache *cache.Cache) {
	b.Manager = mgr
	b.Cache = cache
	b.Result = ending.NewResult()
}

func (b *BaseAction) Execute(vars vars.Vars) error {
	return nil
}

// todo: if the action result need to store more info

func (b *BaseAction) WrapResult(err error) *ending.Result {
	defer b.Result.SetEndTime()
	if err != nil {
		b.Result.ErrResult(err)
		return b.Result
	}
	b.Result.NormalResult(int(ending.SUCCESS), "", "")
	return b.Result
}
