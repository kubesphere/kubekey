package phase

import (
	"github.com/spf13/cobra"

	"github.com/kubesphere/kubekey/v3/cmd/kk/cmd/options"
	"github.com/kubesphere/kubekey/v3/cmd/kk/cmd/util"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/phase/addons"
)

type ApplyAddonsOptions struct {
	CommonOptions  *options.CommonOptions
	ClusterCfgFile string
}

func NewApplyAddonsOptions() *ApplyAddonsOptions {
	return &ApplyAddonsOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

func NewCmdApplyAddons() *cobra.Command {
	o := NewApplyAddonsOptions()
	cmd := &cobra.Command{
		Use:   "addons",
		Short: "Apply cluster addons",
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Run())
		},
	}

	o.CommonOptions.AddCommonFlag(cmd)
	o.AddFlags(cmd)
	return cmd
}

func (o *ApplyAddonsOptions) Run() error {
	arg := common.Argument{
		FilePath: o.ClusterCfgFile,
		Debug:    o.CommonOptions.Verbose,
	}
	return addons.ApplyClusterAddons(arg)
}

func (o *ApplyAddonsOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.ClusterCfgFile, "filename", "f", "", "Path to a configuration file")
}
