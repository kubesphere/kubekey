package action

import (
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/config"
	"github.com/kubesphere/kubekey/pkg/core/ending"
	"github.com/kubesphere/kubekey/pkg/core/vars"
	"github.com/pkg/errors"
)

type BaseAction struct {
	Cache     *cache.Cache
	RootCache *cache.Cache
	Result    *ending.Result
}

func (b *BaseAction) Init(cache *cache.Cache, rootCache *cache.Cache) {
	b.Cache = cache
	b.RootCache = rootCache
	b.Result = ending.NewResult()
}

func (b *BaseAction) Execute(runtime *config.Runtime, vars vars.Vars) error {
	return nil
}

// todo: if the action result need to store more info

func (b *BaseAction) WrapResult(err error) *ending.Result {
	defer b.Result.SetEndTime()
	if err != nil {
		b.Result.ErrResult(errors.WithStack(err))
		return b.Result
	}
	b.Result.NormalResult(int(ending.SUCCESS), "", "")
	return b.Result
}
