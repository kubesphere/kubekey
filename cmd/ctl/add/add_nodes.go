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

package add

import (
	"github.com/kubesphere/kubekey/cmd/ctl/options"
	"github.com/kubesphere/kubekey/cmd/ctl/util"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/pipelines"
	"github.com/spf13/cobra"
)

type AddNodesOptions struct {
	CommonOptions    *options.CommonOptions
	ClusterCfgFile   string
	SkipPullImages   bool
	ContainerManager string
	DownloadCmd      string
}

func NewAddNodesOptions() *AddNodesOptions {
	return &AddNodesOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

// NewCmdAddNodes creates a new add nodes command
func NewCmdAddNodes() *cobra.Command {
	o := NewAddNodesOptions()
	cmd := &cobra.Command{
		Use:   "nodes",
		Short: "Add nodes to the cluster according to the new nodes information from the specified configuration file",
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Run())
		},
	}

	o.CommonOptions.AddCommonFlag(cmd)
	o.AddFlags(cmd)
	return cmd
}

func (o *AddNodesOptions) Run() error {
	arg := common.Argument{
		FilePath:         o.ClusterCfgFile,
		KsEnable:         false,
		Debug:            o.CommonOptions.Verbose,
		IgnoreErr:        o.CommonOptions.IgnoreErr,
		SkipConfirmCheck: o.CommonOptions.SkipConfirmCheck,
		SkipPullImages:   o.SkipPullImages,
		InCluster:        o.CommonOptions.InCluster,
		ContainerManager: o.ContainerManager,
	}
	return pipelines.AddNodes(arg, o.DownloadCmd)
}

func (o *AddNodesOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.ClusterCfgFile, "filename", "f", "", "Path to a configuration file")
	cmd.Flags().BoolVarP(&o.SkipPullImages, "skip-pull-images", "", false, "Skip pre pull images")
	cmd.Flags().StringVarP(&o.ContainerManager, "container-manager", "", "docker", "Container manager: docker, crio, containerd and isula.")
	cmd.Flags().StringVarP(&o.DownloadCmd, "download-cmd", "", "curl -L -o %s %s",
		`The user defined command to download the necessary binary files. The first param '%s' is output path, the second param '%s', is the URL`)
}
