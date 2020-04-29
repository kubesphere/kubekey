package cmd

import (
	"github.com/pixiake/kubekey/cmd/create"
	"github.com/spf13/cobra"
)

var Verbose bool

func NewKubekeyCommand() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "kk",
		Short: "Kubernetes Deploy Tool",
		Long:  "Deploy a Kubernetes Cluster Flexibly and Easily .",
	}

	rootCmd.AddCommand(create.NewCmdCreate())
	rootCmd.AddCommand(NewCmdScaleCluster())
	rootCmd.AddCommand(NewCmdVersion())
	rootCmd.AddCommand(NewCmdResetCluster())
	return rootCmd
}
