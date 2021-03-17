package cmd

import (
	"github.com/kubesphere/kubekey/pkg/delete"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/spf13/cobra"
)

var deleteClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Delete a cluster",
	Run: func(cmd *cobra.Command, args []string) {
		logger := util.InitLogger(opt.Verbose)
		delete.ResetCluster(opt.ClusterCfgFile, logger, opt.Verbose)
	},
}

func init() {
	deleteCmd.AddCommand(deleteClusterCmd)

	deleteClusterCmd.Flags().StringVarP(&opt.ClusterCfgFile, "filename", "f", "", "Path to a configuration file")
}
