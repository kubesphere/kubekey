package cmd

import "github.com/spf13/cobra"

var certCmd = &cobra.Command{
	Use:   "cert",
	Short: "cluster cert",
}

func init() {
	rootCmd.AddCommand(certCmd)
}
