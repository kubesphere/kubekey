package prepare

import "github.com/kubesphere/kubekey/pkg/core/config"

type FastPrepare struct {
	BasePrepare
	Inject func() (bool, error)
}

func (b *FastPrepare) PreCheck(runtime *config.Runtime) (bool, error) {
	return b.Inject()
}
