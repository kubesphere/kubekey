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
	"github.com/spf13/cobra"

	"github.com/kubesphere/kubekey/cmd/kk/cmd/options"
	"github.com/kubesphere/kubekey/cmd/kk/cmd/util"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/phase/os"
)

type ConfigOSOptions struct {
	CommonOptions   *options.CommonOptions
	ClusterCfgFile  string
	InstallPackages bool
}

func NewConfigOSOptions() *ConfigOSOptions {
	return &ConfigOSOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

// NewCmdConfigOS creates a new init os command
func NewCmdConfigOS() *cobra.Command {
	o := NewConfigOSOptions()
	cmd := &cobra.Command{
		Use:   "os",
		Short: "Init the os configure",
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Run())
		},
	}

	o.CommonOptions.AddCommonFlag(cmd)
	o.AddFlags(cmd)
	return cmd
}

func (o *ConfigOSOptions) Run() error {
	arg := common.Argument{
		FilePath:        o.ClusterCfgFile,
		Debug:           o.CommonOptions.Verbose,
		InstallPackages: o.InstallPackages,
	}
	return os.ConfigOS(arg)
}

func (o *ConfigOSOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.ClusterCfgFile, "filename", "f", "", "Path to a configuration file")
	cmd.Flags().BoolVarP(&o.InstallPackages, "with-packages", "", false, "install operation system packages by artifact")
}
