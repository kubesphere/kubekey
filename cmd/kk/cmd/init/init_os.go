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

package init

import (
	"github.com/spf13/cobra"

	"github.com/kubesphere/kubekey/cmd/kk/cmd/options"
	"github.com/kubesphere/kubekey/cmd/kk/cmd/util"
	"github.com/kubesphere/kubekey/cmd/kk/internal/common"
	"github.com/kubesphere/kubekey/cmd/kk/internal/pipelines"
)

type InitOsOptions struct {
	CommonOptions  *options.CommonOptions
	ClusterCfgFile string
	Artifact       string
}

func NewInitOsOptions() *InitOsOptions {
	return &InitOsOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

// NewCmdInitOs creates a new init os command
func NewCmdInitOs() *cobra.Command {
	o := NewInitOsOptions()
	cmd := &cobra.Command{
		Use:   "os",
		Short: "Init operating system",
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Run())
		},
	}

	o.CommonOptions.AddCommonFlag(cmd)
	o.AddFlags(cmd)
	return cmd
}

func (o *InitOsOptions) Run() error {
	arg := common.Argument{
		FilePath: o.ClusterCfgFile,
		Debug:    o.CommonOptions.Verbose,
		Artifact: o.Artifact,
	}
	return pipelines.InitDependencies(arg)
}

func (o *InitOsOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.ClusterCfgFile, "filename", "f", "", "Path to a configuration file")
	cmd.Flags().StringVarP(&o.Artifact, "artifact", "a", "", "Path to a KubeKey artifact")
}
