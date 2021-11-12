package add

import (
	"github.com/kubesphere/kubekey/cmd/ctl/options"
	"github.com/spf13/cobra"
)

type AddOptions struct {
	CommonOptions *options.CommonOptions
}

func NewAddOptions() *AddOptions {
	return &AddOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

// NewCmdAdd creates a new add command
func NewCmdAdd() *cobra.Command {
	o := NewAddOptions()
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add nodes to kubernetes cluster",
	}

	o.CommonOptions.AddCommonFlag(cmd)

	cmd.AddCommand(NewCmdAddNodes())
	return cmd
}
