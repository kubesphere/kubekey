package prepare

import (
	"github.com/kubesphere/kubekey/experiment/utils/cache"
	"github.com/kubesphere/kubekey/experiment/utils/config"
)

type Prepare interface {
	PreCheck() (bool, error)
	Init(mgr *config.Manager, pool *cache.Cache)
}
