package clusterinfo

import (
	"github.com/spf13/cobra"
)

// NewCmdClusterInfo creates a new clusterinfo command
func NewCmdClusterInfo() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "cluster-info",
		Short: "display cluster information",
	}

	cmd.AddCommand(NewCmdClusterInfoDump())

	return cmd
}
