package addons

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"path/filepath"
)

type Install struct {
	common.KubeAction
}

func (i *Install) Execute(runtime connector.Runtime) error {
	nums := len(i.KubeConf.Cluster.Addons)
	for index, addon := range i.KubeConf.Cluster.Addons {
		logger.Log.Messagef(runtime.RemoteHost().GetName(), "Install addon [%v-%v]: %s", nums, index, addon.Name)
		if err := InstallAddons(i.KubeConf, &addon, filepath.Join(runtime.GetWorkDir(), fmt.Sprintf("config-%s", runtime.GetObjName()))); err != nil {
			return err
		}
	}
	return nil
}
