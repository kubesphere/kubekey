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

	"github.com/kubesphere/kubekey/v2/cmd/ctl/options"
	"github.com/kubesphere/kubekey/v2/cmd/ctl/util"
	"github.com/kubesphere/kubekey/v2/pkg/common"
	"github.com/kubesphere/kubekey/v2/pkg/phase/etcd"
)

type CreateEtcdOptions struct {
	CommonOptions  *options.CommonOptions
	ClusterCfgFile string
}

func NewCreateEtcdOptions() *CreateEtcdOptions {
	return &CreateEtcdOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

// NewCmdCreateEtcd creates a new install etcd command
func NewCmdCreateEtcd() *cobra.Command {
	o := NewCreateEtcdOptions()
	cmd := &cobra.Command{
		Use:   "etcd",
		Short: "Install the ETCD cluster on the master",
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Run())
		},
	}

	o.CommonOptions.AddCommonFlag(cmd)
	o.AddFlags(cmd)
	return cmd
}

func (o *CreateEtcdOptions) Run() error {
	arg := common.Argument{
		FilePath: o.ClusterCfgFile,
		Debug:    o.CommonOptions.Verbose,
	}
	return etcd.CreateEtcd(arg)
}

func (o *CreateEtcdOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.ClusterCfgFile, "filename", "f", "", "Path to a configuration file")
}
