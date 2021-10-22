package hook

import (
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/ending"
)

type Base struct {
	Runtime connector.Runtime
	Desc    string
	Result  *ending.ModuleResult
}

func (b *Base) Try() error {
	panic("implement me")
}

func (b *Base) Catch(err error) error {
	return err
}

func (b *Base) Finally() {
}

func (b *Base) Init(runtime connector.Runtime, desc string, result *ending.ModuleResult) {
	b.Runtime = runtime
	b.Desc = desc
	b.Result = result
}
