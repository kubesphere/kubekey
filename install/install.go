package install

import (
	kubekeyapi "github.com/pixiake/kubekey/apis/v1alpha1"
)

func CreateCluster(clusterCfgFile string, addons string, pkg string) {
	kubekeyapi.GetClusterCfg(clusterCfgFile)
}
