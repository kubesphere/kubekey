package cmd

import (
	"github.com/kubesphere/kubekey/pkg/pipelines"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"strings"
)

var deleteNodeCmd = &cobra.Command{
	Use:   "node",
	Short: "delete a node",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("node can not be empty")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		arg := common.Argument{
			FilePath: opt.ClusterCfgFile,
			Debug:    opt.Verbose,
			NodeName: strings.Join(args, ""),
		}
		return pipelines.DeleteNode(arg)
	},
}

func init() {
	deleteCmd.AddCommand(deleteNodeCmd)
	deleteNodeCmd.Flags().StringVarP(&opt.ClusterCfgFile, "filename", "f", "", "Path to a configuration file")
}
