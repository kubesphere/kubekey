package module

import (
	"github.com/kubesphere/kubekey/pkg/core/ending"
	"github.com/kubesphere/kubekey/pkg/core/hook"
)

type PostHookInterface interface {
	hook.Interface
	Init(module Module, result *ending.ModuleResult)
}

type PostHook struct {
	Module Module
	Result *ending.ModuleResult
}

func (p *PostHook) Try() error {
	panic("implement me")
}

func (p *PostHook) Catch(err error) error {
	return err
}

func (p *PostHook) Finally() {
}

func (p *PostHook) Init(module Module, result *ending.ModuleResult) {
	p.Module = module
	p.Result = result
}
