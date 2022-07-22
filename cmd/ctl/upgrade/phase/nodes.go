/*
Copyright 2022 The KubeSphere Authors.

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

package phase

import (
	"fmt"

	"github.com/kubesphere/kubekey/cmd/ctl/options"
	"github.com/kubesphere/kubekey/cmd/ctl/util"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/phase/nodes"
	"github.com/spf13/cobra"
)

type UpgradeNodesOptions struct {
	CommonOptions  *options.CommonOptions
	ClusterCfgFile string
	Kubernetes     string
}

func NewUpgradeNodesOptions() *UpgradeNodesOptions {
	return &UpgradeNodesOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

// NewCmdUpgrade creates a new upgrade command
func NewCmdUpgradeNodes() *cobra.Command {
	o := NewUpgradeNodesOptions()
	cmd := &cobra.Command{
		Use:   "nodes",
		Short: "Upgrade cluster on master nodes and worker nodes to the version you input",
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Run())
		},
	}
	o.CommonOptions.AddCommonFlag(cmd)
	o.AddFlags(cmd)

	if err := k8sCompletionSetting(cmd); err != nil {
		panic(fmt.Sprintf("Got error with the completion setting"))
	}
	return cmd
}

func (o *UpgradeNodesOptions) Run() error {
	arg := common.Argument{
		FilePath:          o.ClusterCfgFile,
		KubernetesVersion: o.Kubernetes,
		Debug:             o.CommonOptions.Verbose,
	}
	return nodes.UpgradeNodes(arg)
}

func (o *UpgradeNodesOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.ClusterCfgFile, "filename", "f", "", "Path to a configuration file")
	cmd.Flags().StringVarP(&o.Kubernetes, "with-kubernetes", "", "", "Specify a supported version of kubernetes")
}
