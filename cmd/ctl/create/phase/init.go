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

	"github.com/spf13/cobra"

	"github.com/kubesphere/kubekey/v2/cmd/ctl/options"
	"github.com/kubesphere/kubekey/v2/cmd/ctl/util"
	"github.com/kubesphere/kubekey/v2/pkg/common"
	"github.com/kubesphere/kubekey/v2/pkg/phase/kubernetes"
)

type CreateInitClusterOptions struct {
	CommonOptions  *options.CommonOptions
	ClusterCfgFile string
	Kubernetes     string
}

func NewCreateInitClusterOptions() *CreateInitClusterOptions {
	return &CreateInitClusterOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

// NewCmdUpgrade creates a new upgrade command
func NewCmdCreateInitCluster() *cobra.Command {
	o := NewCreateInitClusterOptions()
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Init the k8s cluster",
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

func (o *CreateInitClusterOptions) Run() error {
	arg := common.Argument{
		FilePath:          o.ClusterCfgFile,
		KubernetesVersion: o.Kubernetes,
		Debug:             o.CommonOptions.Verbose,
		Namespace:         o.CommonOptions.Namespace,
	}

	return kubernetes.CreateInitCluster(arg)
}

func (o *CreateInitClusterOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.ClusterCfgFile, "filename", "f", "", "Path to a configuration file")
	cmd.Flags().StringVarP(&o.Kubernetes, "with-kubernetes", "", "", "Specify a supported version of kubernetes")
}
