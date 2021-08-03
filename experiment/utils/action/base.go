package action

import "github.com/kubesphere/kubekey/experiment/utils/config"

type BaseAction struct {
	Manager *config.Manager
}

func (b *BaseAction) Init(mgr *config.Manager) {
	b.Manager = mgr
}
