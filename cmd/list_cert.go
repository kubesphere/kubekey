package cmd

import (
	"github.com/kubesphere/kubekey/pkg/cert"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/spf13/cobra"
)

var listClusterCertsCmd = &cobra.Command{
	Use:   "list",
	Short: "list a cluster certs",
	Run: func(cmd *cobra.Command, args []string) {
		logger := util.InitLogger(opt.Verbose)
		cert.ListCluster(opt.ClusterCfgFile, logger, opt.Verbose)
	},
}

func init() {
	certsCmd.AddCommand(listClusterCertsCmd)

	listClusterCertsCmd.Flags().StringVarP(&opt.ClusterCfgFile, "filename", "f", "", "Path to a configuration file")
}
