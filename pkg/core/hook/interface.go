package hook

import (
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/ending"
)

type Interface interface {
	Try() error
	Catch(err error) error
	Finally()
}

type PostHook interface {
	Interface
	Init(runtime connector.Runtime, desc string, result *ending.ModuleResult)
}
