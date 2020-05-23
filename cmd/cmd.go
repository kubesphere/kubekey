package cmd

import (
	"github.com/kubesphere/kubekey/cmd/create"
	"github.com/spf13/cobra"
)

var Verbose bool

func NewKubekeyCommand() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "kk",
		Short: "Kubernetes/KubeSphere Deploy Tool",
		Long:  "Deploy a Kubernetes or KubeSphere cluster efficiently, flexibly and easily. There are three scenarios to use KubeKey. \n" +
		"1. Install Kubernetes only \n" + 
		"2. Install Kubernetes and KubeSphere together in one command \n" +  
		"3. Install Kubernetes first, then deploy KubeSphere on it using https://github.com/kubesphere/ks-installer",
	}

	rootCmd.AddCommand(create.NewCmdCreate())
	rootCmd.AddCommand(NewCmdScaleCluster())
	rootCmd.AddCommand(NewCmdVersion())
	rootCmd.AddCommand(NewCmdResetCluster())
	return rootCmd
}
