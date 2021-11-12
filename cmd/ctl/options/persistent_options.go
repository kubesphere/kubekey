package options

import (
	"github.com/spf13/cobra"
)

type CommonOptions struct {
	InCluster        bool
	Verbose          bool
	SkipConfirmCheck bool
	IgnoreErr        bool
}

func NewCommonOptions() *CommonOptions {
	return &CommonOptions{}
}

func (o *CommonOptions) AddCommonFlag(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&o.InCluster, "in-cluster", false, "Running inside the cluster")
	cmd.Flags().BoolVar(&o.Verbose, "debug", true, "Print detailed information")
	cmd.Flags().BoolVarP(&o.SkipConfirmCheck, "yes", "y", false, "Skip confirm check")
	cmd.Flags().BoolVar(&o.IgnoreErr, "ignore-err", false, "Ignore the error message, remove the host which reported error and force to continue")
}
