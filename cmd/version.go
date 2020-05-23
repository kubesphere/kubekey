package cmd

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/spf13/cobra"
)

func NewCmdVersion() *cobra.Command {
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Get version information of KubeKey, default Kubernetes and default KubeSphere",
		Args:  cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(util.VERSION)
		},
	}

	return versionCmd
}
