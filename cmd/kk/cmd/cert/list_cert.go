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
	"github.com/spf13/cobra"

	"github.com/kubesphere/kubekey/cmd/kk/cmd/options"
	"github.com/kubesphere/kubekey/cmd/kk/cmd/util"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/pipelines"
)

type CertListOptions struct {
	CommonOptions  *options.CommonOptions
	ClusterCfgFile string
}

func NewCertListOptions() *CertListOptions {
	return &CertListOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

// NewCmdCertList creates a new cert list command
func NewCmdCertList() *cobra.Command {
	o := NewCertListOptions()
	cmd := &cobra.Command{
		Use:   "check-expiration",
		Short: "Check certificates expiration for a Kubernetes cluster",
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Run())
		},
	}

	o.CommonOptions.AddCommonFlag(cmd)
	o.AddFlags(cmd)
	return cmd
}

func (o *CertListOptions) Run() error {
	arg := common.Argument{
		FilePath: o.ClusterCfgFile,
		Debug:    o.CommonOptions.Verbose,
	}
	return pipelines.CheckCerts(arg)
}

func (o *CertListOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.ClusterCfgFile, "filename", "f", "", "Path to a configuration file")
}
