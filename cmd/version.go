package cmd

import (
	"fmt"
	"github.com/pixiake/kubekey/pkg/util"
	"github.com/spf13/cobra"
)

func NewCmdVersion() *cobra.Command {
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Get Version Information",
		Args:  cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(util.VERSION)
		},
	}

	return versionCmd
}
