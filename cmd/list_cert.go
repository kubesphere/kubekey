package cmd

import (
	"github.com/kubesphere/kubekey/pkg/cert"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/spf13/cobra"
)

var listClusterCertCmd = &cobra.Command{
	Use:   "list",
	Short: "list a cluster cert",
	Run: func(cmd *cobra.Command, args []string) {
		logger := util.InitLogger(opt.Verbose)
		cert.ListCluster(opt.ClusterCfgFile, logger, opt.Verbose)
	},
}

func init() {
	certCmd.AddCommand(listClusterCertCmd)

	listClusterCertCmd.Flags().StringVarP(&opt.ClusterCfgFile, "filename", "f", "", "Path to a configuration file")
}
