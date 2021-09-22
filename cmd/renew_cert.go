package cmd

import (
	"github.com/kubesphere/kubekey/pkg/pipelines"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"github.com/spf13/cobra"
)

var renewClusterCertsCmd = &cobra.Command{
	Use:   "renew",
	Short: "renew a cluster certs",
	RunE: func(cmd *cobra.Command, args []string) error {
		arg := common.Argument{
			FilePath: opt.ClusterCfgFile,
			Debug:    opt.Verbose,
		}
		return pipelines.RenewCerts(arg)
	},
}

func init() {
	certsCmd.AddCommand(renewClusterCertsCmd)

	renewClusterCertsCmd.Flags().StringVarP(&opt.ClusterCfgFile, "filename", "f", "", "Path to a configuration file")
}
