package delete

import (
	"github.com/kubesphere/kubekey/cmd/ctl/options"
	"github.com/kubesphere/kubekey/cmd/ctl/util"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/pipelines"
	"github.com/spf13/cobra"
)

type DeleteClusterOptions struct {
	CommonOptions  *options.CommonOptions
	ClusterCfgFile string
}

func NewDeleteClusterOptions() *DeleteClusterOptions {
	return &DeleteClusterOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

// NewCmdDeleteCluster creates a new delete cluster command
func NewCmdDeleteCluster() *cobra.Command {
	o := NewDeleteClusterOptions()
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Delete a cluster",
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Run())
		},
	}

	o.CommonOptions.AddCommonFlag(cmd)
	o.AddFlags(cmd)
	return cmd
}

func (o *DeleteClusterOptions) Run() error {
	arg := common.Argument{
		FilePath: o.ClusterCfgFile,
		Debug:    o.CommonOptions.Verbose,
	}
	return pipelines.DeleteCluster(arg)
}

func (o *DeleteClusterOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.ClusterCfgFile, "filename", "f", "", "Path to a configuration file")
}
