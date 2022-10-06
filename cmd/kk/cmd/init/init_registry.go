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

package init

import (
	"github.com/spf13/cobra"

	"github.com/kubesphere/kubekey/cmd/kk/cmd/options"
	"github.com/kubesphere/kubekey/cmd/kk/cmd/util"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/pipelines"
)

type InitRegistryOptions struct {
	CommonOptions  *options.CommonOptions
	ClusterCfgFile string
	DownloadCmd    string
	Artifact       string
}

func NewInitRegistryOptions() *InitRegistryOptions {
	return &InitRegistryOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

// NewCmdInitRegistry creates a new init os command
func NewCmdInitRegistry() *cobra.Command {
	o := NewInitRegistryOptions()
	cmd := &cobra.Command{
		Use:   "registry",
		Short: "Init a local image registry",
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Complete(cmd, args))
			util.CheckErr(o.Run())
		},
	}

	o.CommonOptions.AddCommonFlag(cmd)
	o.AddFlags(cmd)
	return cmd
}

func (o *InitRegistryOptions) Complete(_ *cobra.Command, _ []string) error {
	return nil
}

func (o *InitRegistryOptions) Run() error {
	arg := common.Argument{
		FilePath: o.ClusterCfgFile,
		Debug:    o.CommonOptions.Verbose,
		Artifact: o.Artifact,
	}
	return pipelines.InitRegistry(arg, o.DownloadCmd)
}

func (o *InitRegistryOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.ClusterCfgFile, "filename", "f", "", "Path to a configuration file")
	cmd.Flags().StringVarP(&o.DownloadCmd, "download-cmd", "", "curl -L -o %s %s",
		`The user defined command to download the necessary files. The first param '%s' is output path, the second param '%s', is the URL`)
	cmd.Flags().StringVarP(&o.Artifact, "artifact", "a", "", "Path to a KubeKey artifact")
}
