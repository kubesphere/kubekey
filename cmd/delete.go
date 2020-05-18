package cmd

import (
	"github.com/kubesphere/kubekey/pkg/delete"
	"github.com/kubesphere/kubekey/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewCmdResetCluster() *cobra.Command {
	var (
		clusterCfgFile string
		verbose        bool
	)
	var clusterCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete Kubernetes Cluster",
		Run: func(cmd *cobra.Command, args []string) {
			logger := util.InitLogger(verbose)
			resetCluster(clusterCfgFile, logger, verbose)
		},
	}

	clusterCmd.Flags().StringVarP(&clusterCfgFile, "config", "f", "", "")
	clusterCmd.Flags().BoolVarP(&verbose, "debug", "", true, "")
	return clusterCmd
}

func resetCluster(clusterCfgFile string, logger *log.Logger, verbose bool) {
	delete.ResetCluster(clusterCfgFile, logger, verbose)
}
