package cmd

import (
	"github.com/pixiake/kubekey/pkg/config"
	"github.com/pixiake/kubekey/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewCmdResetCluster() *cobra.Command {
	var (
		clusterCfgFile string
	)
	var clusterCmd = &cobra.Command{
		Use:   "reset",
		Short: "Reset Kubernetes Cluster",
		Run: func(cmd *cobra.Command, args []string) {
			logger := util.InitLogger(true)
			resetCluster(clusterCfgFile, logger)
		},
	}

	clusterCmd.Flags().StringVarP(&clusterCfgFile, "config", "f", "", "")
	return clusterCmd
}

func resetCluster(clusterCfgFile string, logger *log.Logger) {
	config.ParseClusterCfg(clusterCfgFile, logger)
}
