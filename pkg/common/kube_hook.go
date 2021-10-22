package common

import (
	"github.com/kubesphere/kubekey/pkg/core/hook"
)

type UpdateCRStatusHook struct {
	hook.Base
}

func (u *UpdateCRStatusHook) Try() error {
	// todo: update cr status
	return nil
}
