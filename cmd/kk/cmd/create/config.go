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

package create

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kubesphere/kubekey/v3/cmd/kk/cmd/options"
	"github.com/kubesphere/kubekey/v3/cmd/kk/cmd/util"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/config"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/version/kubesphere"
)

type CreateConfigOptions struct {
	CommonOptions    *options.CommonOptions
	Name             string
	ClusterCfgFile   string
	Kubernetes       string
	EnableKubeSphere bool
	KubeSphere       string
	FromCluster      bool
	KubeConfig       string
}

func NewCreateConfigOptions() *CreateConfigOptions {
	return &CreateConfigOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

// NewCmdCreateConfig creates a create config command
func NewCmdCreateConfig() *cobra.Command {
	o := NewCreateConfigOptions()
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Create cluster configuration file",
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Complete(cmd, args))
			util.CheckErr(o.Run())

		},
	}

	o.CommonOptions.AddCommonFlag(cmd)
	o.AddFlags(cmd)
	return cmd
}

func (o *CreateConfigOptions) Complete(cmd *cobra.Command, args []string) error {
	var ksVersion string
	if o.EnableKubeSphere && len(args) > 0 {
		ksVersion = args[0]
	} else {
		ksVersion = kubesphere.Latest().Version
	}
	o.KubeSphere = ksVersion
	return nil
}

func (o *CreateConfigOptions) Run() error {
	arg := common.Argument{
		FilePath:          o.ClusterCfgFile,
		KubernetesVersion: o.Kubernetes,
		KsEnable:          o.EnableKubeSphere,
		KsVersion:         o.KubeSphere,
		FromCluster:       o.FromCluster,
		KubeConfig:        o.KubeConfig,
	}

	return config.GenerateKubeKeyConfig(arg, o.Name)
}

func (o *CreateConfigOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Name, "name", "", "sample", "Specify a name of cluster object")
	cmd.Flags().StringVarP(&o.ClusterCfgFile, "filename", "f", "", "Specify a configuration file path")
	cmd.Flags().StringVarP(&o.Kubernetes, "with-kubernetes", "", "", "Specify a supported version of kubernetes")
	cmd.Flags().BoolVarP(&o.EnableKubeSphere, "with-kubesphere", "", false, fmt.Sprintf("Deploy a specific version of kubesphere (default %s)", kubesphere.Latest().Version))
	cmd.Flags().BoolVarP(&o.FromCluster, "from-cluster", "", false, "Create a configuration based on existing cluster")
	cmd.Flags().StringVarP(&o.KubeConfig, "kubeconfig", "", "", "Specify a kubeconfig file")
}
