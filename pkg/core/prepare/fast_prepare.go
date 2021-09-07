package prepare

import (
	"github.com/kubesphere/kubekey/pkg/core/connector"
)

type FastPrepare struct {
	BasePrepare
	Inject func(runtime connector.Runtime) (bool, error)
}

func (b *FastPrepare) PreCheck(runtime connector.Runtime) (bool, error) {
	return b.Inject(runtime)
}
