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

package cert

import (
	"github.com/kubesphere/kubekey/cmd/ctl/options"
	"github.com/kubesphere/kubekey/cmd/ctl/util"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/pipelines"
	"github.com/spf13/cobra"
)

type CertRenewOptions struct {
	CommonOptions  *options.CommonOptions
	ClusterCfgFile string
}

func NewCertRenewOptions() *CertRenewOptions {
	return &CertRenewOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

// NewCmdCertRenew creates a new cert renew command
func NewCmdCertRenew() *cobra.Command {
	o := NewCertRenewOptions()
	cmd := &cobra.Command{
		Use:   "renew",
		Short: "renew a cluster certs",
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Run())
		},
	}

	o.CommonOptions.AddCommonFlag(cmd)
	o.AddFlags(cmd)
	return cmd
}

func (o *CertRenewOptions) Run() error {
	arg := common.Argument{
		FilePath: o.ClusterCfgFile,
		Debug:    o.CommonOptions.Verbose,
	}
	return pipelines.RenewCerts(arg)
}

func (o *CertRenewOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.ClusterCfgFile, "filename", "f", "", "Path to a configuration file")
}
