package certs

import (
	"github.com/kubesphere/kubekey/pkg/certs/templates"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"path/filepath"
)

type AutoRenewCertsEnabled struct {
	common.KubePrepare
	Not bool
}

func (a *AutoRenewCertsEnabled) PreCheck(runtime connector.Runtime) (bool, error) {
	exist, err := runtime.GetRunner().FileExist(filepath.Join("/etc/systemd/system/", templates.K8sCertsRenewService.Name()))
	if err != nil {
		return false, err
	}
	if exist {
		return !a.Not, nil
	} else {
		return a.Not, nil
	}
}
