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

	"github.com/kubesphere/kubekey/cmd/ctl/options"
	"github.com/kubesphere/kubekey/cmd/ctl/util"
	"github.com/kubesphere/kubekey/pkg/alpha/binary"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/version/kubernetes"
	"github.com/spf13/cobra"
)

type CreateBinaryOptions struct {
	CommonOptions  *options.CommonOptions
	ClusterCfgFile string
	Kubernetes     string
	DownloadCmd    string
}

func NewCreateBinaryOptions() *CreateBinaryOptions {
	return &CreateBinaryOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

// NewCmdCreateBinary creates a new artifact import command
func NewCmdCreateBinary() *cobra.Command {
	o := NewCreateBinaryOptions()
	cmd := &cobra.Command{
		Use:   "binary",
		Short: "Download the binaries on the local",
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

func (o *CreateBinaryOptions) Run() error {
	arg := common.Argument{
		FilePath:          o.ClusterCfgFile,
		KubernetesVersion: o.Kubernetes,
		Debug:             o.CommonOptions.Verbose,
	}
	return binary.CreateBinary(arg, o.DownloadCmd)
}

func (o *CreateBinaryOptions) AddFlags(cmd *cobra.Command) {
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
