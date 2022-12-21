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
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kubesphere/kubekey/v2/cmd/ctl/options"
	"github.com/kubesphere/kubekey/v2/cmd/ctl/util"
	"github.com/kubesphere/kubekey/v2/pkg/common"
	"github.com/kubesphere/kubekey/v2/pkg/pipelines"
)

type DeleteNodeOptions struct {
	CommonOptions  *options.CommonOptions
	ClusterCfgFile string
	nodeName       string
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
	o.nodeName = strings.Join(args, "")
	return nil
}

func (o *DeleteNodeOptions) Validate() error {
	if o.nodeName == "" {
		return errors.New("node name can not be empty")
	}
	return nil
}

func (o *DeleteNodeOptions) Run() error {
	arg := common.Argument{
		FilePath: o.ClusterCfgFile,
		Debug:    o.CommonOptions.Verbose,
		NodeName: o.nodeName,
	}
	return pipelines.DeleteNode(arg)
}

func (o *DeleteNodeOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.ClusterCfgFile, "filename", "f", "", "Path to a configuration file")

}
