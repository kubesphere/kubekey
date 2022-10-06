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
	"github.com/spf13/cobra"

	"github.com/kubesphere/kubekey/cmd/kk/cmd/options"
	"github.com/kubesphere/kubekey/cmd/kk/cmd/util"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/pipelines"
)

type DeleteClusterOptions struct {
	CommonOptions  *options.CommonOptions
	ClusterCfgFile string
	Kubernetes     string
	DeleteCRI      bool
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
		FilePath:          o.ClusterCfgFile,
		Debug:             o.CommonOptions.Verbose,
		KubernetesVersion: o.Kubernetes,
		DeleteCRI:         o.DeleteCRI,
	}
	return pipelines.DeleteCluster(arg)
}

func (o *DeleteClusterOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.ClusterCfgFile, "filename", "f", "", "Path to a configuration file")
	cmd.Flags().StringVarP(&o.Kubernetes, "with-kubernetes", "", "", "Specify a supported version of kubernetes")
	cmd.Flags().BoolVarP(&o.DeleteCRI, "all", "A", false, "Delete total cri conficutation and data directories")
}
