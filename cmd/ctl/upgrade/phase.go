package upgrade

import (
	"github.com/spf13/cobra"
)

func NewPhaseCommand() *cobra.Command {
	cmds := &cobra.Command{
		Use:   "phase",
		Short: "KubeKey upgrade phase",
		Long:  `This is the upgrade phase run cmd`,
	}

	cmds.AddCommand(NewCmdUpgradePrecheck())
	return cmds
}
