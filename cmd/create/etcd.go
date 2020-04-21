package create

import (
	kubekeyapi "github.com/pixiake/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/pixiake/kubekey/pkg/util"

	log "github.com/sirupsen/logrus"
	//"github.com/pixiake/kubekey/cluster"
	//"github.com/pixiake/kubekey/cluster/etcd"
	"github.com/spf13/cobra"
)

func NewCmdCreateEtcd() *cobra.Command {
	var (
		clusterCfgFile string
	)
	var clusterCmd = &cobra.Command{
		Use:   "etcd",
		Short: "Manage Etcd Cluster",
		Run: func(cmd *cobra.Command, args []string) {
			logger := util.InitLogger(true)
			createEtcdCluster(clusterCfgFile, logger)
		},
	}

	clusterCmd.Flags().StringVarP(&clusterCfgFile, "cluster-info", "", "", "")
	return clusterCmd
}

func createEtcdCluster(clusterCfgFile string, logger *log.Logger) {
	kubekeyapi.LoadClusterCfg(clusterCfgFile, logger)
	//allNodes, etcdNodes, _, _, _ := cfg.GroupHosts()
	//etcd.EtcdPrepare(etcdNodes)
	//etcd.GenEtcdFiles(cfg, allNodes, etcdNodes)
	//etcd.GetEtcdCtl(etcdNodes, "amd64")
	//etcd.SetupEtcd(etcdNodes)

}
