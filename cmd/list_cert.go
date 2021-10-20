package cmd

import (
	common2 "github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/pipelines"
	"github.com/spf13/cobra"
)

var listClusterCertsCmd = &cobra.Command{
	Use:   "check-expiration",
	Short: "Check certificates expiration for a Kubernetes cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		arg := common2.Argument{
			FilePath: opt.ClusterCfgFile,
			Debug:    opt.Verbose,
		}
		return pipelines.CheckCerts(arg)
	},
}

func init() {
	certsCmd.AddCommand(listClusterCertsCmd)

	listClusterCertsCmd.Flags().StringVarP(&opt.ClusterCfgFile, "filename", "f", "", "Path to a configuration file")
}
