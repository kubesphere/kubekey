package module

import (
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/ending"
	"github.com/kubesphere/kubekey/pkg/core/hook"
	"github.com/pkg/errors"
)

type BaseModule struct {
	Name          string
	Desc          string
	Skip          bool
	ModuleCache   *cache.Cache
	PipelineCache *cache.Cache
	Runtime       connector.ModuleRuntime
	PostHook      []PostHookInterface
}

func (b *BaseModule) IsSkip() bool {
	return b.Skip
}

func (b *BaseModule) Default(runtime connector.Runtime, pipelineCache *cache.Cache, moduleCache *cache.Cache) {
	b.Runtime = runtime
	b.PipelineCache = pipelineCache
	b.ModuleCache = moduleCache
}

func (b *BaseModule) Init() {
	if b.Name == "" {
		b.Name = DefaultModuleName
	}
}

func (b *BaseModule) Is() string {
	return BaseModuleType
}

func (b *BaseModule) Run(result *ending.ModuleResult) {
	panic("implement me")
}

func (b *BaseModule) Until() (*bool, error) {
	return nil, nil
}

func (b *BaseModule) Slogan() {
}

func (b *BaseModule) AutoAssert() {
}

func (b *BaseModule) AppendPostHook(h PostHookInterface) {
	b.PostHook = append(b.PostHook, h)
}

func (b *BaseModule) CallPostHook(result *ending.ModuleResult) error {
	for i := range b.PostHook {
		h := b.PostHook[i]
		h.Init(b, result)
		if err := hook.Call(h); err != nil {
			return errors.Wrapf(err, "Module[%s] call post hook failed", b.Name)
		}
	}
	return nil
}
