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
	"github.com/kubesphere/kubekey/v2/pkg/phase/binary"
	"github.com/kubesphere/kubekey/v2/pkg/version/kubernetes"
)

type UpgradeBinaryOptions struct {
	CommonOptions  *options.CommonOptions
	ClusterCfgFile string
	Kubernetes     string
	DownloadCmd    string
}

func NewUpgradeBinaryOptions() *UpgradeBinaryOptions {
	return &UpgradeBinaryOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

// NewCmdUpgradeBinary creates a new artifact import command
func NewCmdUpgradeBinary() *cobra.Command {
	o := NewUpgradeBinaryOptions()
	cmd := &cobra.Command{
		Use:   "binary",
		Short: "Download the binary and synchronize kubernetes binaries",
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

func (o *UpgradeBinaryOptions) Run() error {
	arg := common.Argument{
		FilePath:          o.ClusterCfgFile,
		KubernetesVersion: o.Kubernetes,
		Debug:             o.CommonOptions.Verbose,
	}
	return binary.UpgradeBinary(arg, o.DownloadCmd)
}

func (o *UpgradeBinaryOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.ClusterCfgFile, "filename", "f", "", "Path to a configuration file")
	cmd.Flags().StringVarP(&o.Kubernetes, "with-kubernetes", "", "", "Specify a supported version of kubernetes")
	cmd.Flags().StringVarP(&o.DownloadCmd, "download-cmd", "", "curl -L -o %s %s",
		`The user defined command to download the necessary binary files. The first param '%s' is output path, the second param '%s', is the URL`)

}

func k8sCompletionSetting(cmd *cobra.Command) (err error) {
	err = cmd.RegisterFlagCompletionFunc("with-kubernetes", func(cmd *cobra.Command, args []string, toComplete string) (
		strings []string, directive cobra.ShellCompDirective) {
		return kubernetes.SupportedK8sVersionList(), cobra.ShellCompDirectiveNoFileComp
	})
	return
}
