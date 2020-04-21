package create

import (
	"github.com/spf13/cobra"
)

func NewCmdCreate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create Cluster-Info Config or Kubernetes Cluster",
	}
	cmd.AddCommand(NewCmdCreateCfg())
	cmd.AddCommand(NewCmdCreateCluster())
	cmd.AddCommand(NewCmdCreateEtcd())
	return cmd
}
