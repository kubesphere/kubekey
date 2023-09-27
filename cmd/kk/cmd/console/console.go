package console

import (
	"github.com/spf13/cobra"

	"github.com/kubesphere/kubekey/v3/cmd/kk/cmd/options"
)

type ConsoleOptions struct {
	CommonOptions *options.CommonOptions
}

func NewConsoleOptions() *ConsoleOptions {
	return &ConsoleOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}
func NewCmdConsole() *cobra.Command {
	o := NewConsoleOptions()
	cmd := &cobra.Command{
		Use:   "console",
		Short: "Start a web console of kubekey",
	}

	o.CommonOptions.AddCommonFlag(cmd)
	cmd.AddCommand(NewCmdConsoleStart())
	return cmd
}
