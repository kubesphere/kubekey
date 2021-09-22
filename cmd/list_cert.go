package cmd

import (
	"github.com/kubesphere/kubekey/pkg/pipelines"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"github.com/spf13/cobra"
)

var listClusterCertsCmd = &cobra.Command{
	Use:   "check-expiration",
	Short: "Check certificates expiration for a Kubernetes cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		arg := common.Argument{
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
