package create

import (
	"github.com/spf13/cobra"
)

func NewCmdCreate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a cluster or a cluster configuration file",
	}
	cmd.AddCommand(NewCmdCreateCfg())
	cmd.AddCommand(NewCmdCreateCluster())
	cmd.AddCommand(NewCmdCreateEtcd())
	return cmd
}
