/*
Copyright 2020 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
