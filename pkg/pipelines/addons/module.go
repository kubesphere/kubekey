package addons

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"path/filepath"
)

type AddonsModule struct {
	common.KubeModule
	Skip bool
}

func (a *AddonsModule) IsSkip() bool {
	return a.Skip
}

func (a *AddonsModule) Init() {
	a.Name = "AddonsModule"
	a.Desc = "install addons"
}

func (a *AddonsModule) Run() error {
	nums := len(a.KubeConf.Cluster.Addons)
	for index, addon := range a.KubeConf.Cluster.Addons {
		logger.Log.Messagef(common.LocalHost, "Install addon [%v-%v]: %s", nums, index, addon.Name)
		if err := InstallAddons(a.KubeConf, &addon, filepath.Join(a.Runtime.GetWorkDir(), fmt.Sprintf("config-%s", a.Runtime.GetObjName()))); err != nil {
			return err
		}
	}
	return nil
}
