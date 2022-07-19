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

package upgrade

import (
	"fmt"

	"github.com/kubesphere/kubekey/cmd/ctl/options"
	"github.com/kubesphere/kubekey/cmd/ctl/util"
	"github.com/kubesphere/kubekey/pkg/alpha"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/version/kubesphere"
	"github.com/spf13/cobra"
)

type UpgradePrecheckOptions struct {
	CommonOptions    *options.CommonOptions
	ClusterCfgFile   string
	Kubernetes       string
	EnableKubeSphere bool
	KubeSphere       string
}

func NewUpgradePrecheckOptions() *UpgradePrecheckOptions {
	return &UpgradePrecheckOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

// NewCmdUpgrade creates a new upgrade command
func NewCmdUpgradePrecheck() *cobra.Command {
	o := NewUpgradePrecheckOptions()
	cmd := &cobra.Command{
		Use:   "precheck",
		Short: "Precheck the nodes and cluster before the upgrade cluster",
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Complete(cmd, args))
			util.CheckErr(o.Run())
		},
	}

	o.CommonOptions.AddCommonFlag(cmd)
	o.AddFlags(cmd)
	return cmd
}

func (o *UpgradePrecheckOptions) Complete(cmd *cobra.Command, args []string) error {
	var ksVersion string
	if o.EnableKubeSphere && len(args) > 0 {
		ksVersion = args[0]
	} else {
		ksVersion = kubesphere.Latest().Version
	}
	o.KubeSphere = ksVersion
	return nil
}

func (o *UpgradePrecheckOptions) Run() error {
	arg := common.Argument{
		FilePath:          o.ClusterCfgFile,
		KubernetesVersion: o.Kubernetes,
		KsEnable:          o.EnableKubeSphere,
		KsVersion:         o.KubeSphere,
		Debug:             o.CommonOptions.Verbose,
	}
	return alpha.UpgradePrecheck(arg)
}

func (o *UpgradePrecheckOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.ClusterCfgFile, "filename", "f", "", "Path to a configuration file")
	cmd.Flags().StringVarP(&o.Kubernetes, "with-kubernetes", "", "", "Specify a supported version of kubernetes")
	cmd.Flags().BoolVarP(&o.EnableKubeSphere, "with-kubesphere", "", false, fmt.Sprintf("Deploy a specific version of kubesphere (default %s)", kubesphere.Latest().Version))
}
