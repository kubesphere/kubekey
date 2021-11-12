package delete

import (
	"github.com/kubesphere/kubekey/cmd/ctl/options"
	"github.com/kubesphere/kubekey/cmd/ctl/util"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/pipelines"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"strings"
)

type DeleteNodeOptions struct {
	CommonOptions  *options.CommonOptions
	ClusterCfgFile string
	nodes          string
}

func NewDeleteNodeOptions() *DeleteNodeOptions {
	return &DeleteNodeOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

// NewCmdDeleteNode creates a new delete node command
func NewCmdDeleteNode() *cobra.Command {
	o := NewDeleteNodeOptions()
	cmd := &cobra.Command{
		Use:   "node",
		Short: "delete a node",
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Complete(cmd, args))
			util.CheckErr(o.Validate())
			util.CheckErr(o.Run())
		},
	}

	o.CommonOptions.AddCommonFlag(cmd)
	o.AddFlags(cmd)
	return cmd
}

func (o *DeleteNodeOptions) Complete(cmd *cobra.Command, args []string) error {
	o.nodes = strings.Join(args, "")
	return nil
}

func (o *DeleteNodeOptions) Validate() error {
	if o.nodes == "" {
		return errors.New("node can not be empty")
	}
	return nil
}

func (o *DeleteNodeOptions) Run() error {
	arg := common.Argument{
		FilePath: o.ClusterCfgFile,
		Debug:    o.CommonOptions.Verbose,
		NodeName: o.nodes,
	}
	return pipelines.DeleteNode(arg)
}

func (o *DeleteNodeOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.ClusterCfgFile, "filename", "f", "", "Path to a configuration file")

}
