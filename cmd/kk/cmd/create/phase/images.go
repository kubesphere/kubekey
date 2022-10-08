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

	"github.com/kubesphere/kubekey/cmd/kk/cmd/options"
	"github.com/kubesphere/kubekey/cmd/kk/cmd/util"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/phase/images"
)

type CreateImagesOptions struct {
	CommonOptions    *options.CommonOptions
	ClusterCfgFile   string
	Kubernetes       string
	ContainerManager string
}

func NewCreateImagesOptions() *CreateImagesOptions {
	return &CreateImagesOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

// NewCmdCreateImages creates a new Images command
func NewCmdCreateImages() *cobra.Command {
	o := NewCreateImagesOptions()
	cmd := &cobra.Command{
		Use:   "images",
		Short: "Down the container and pull the images before creating your cluster",
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Validate(cmd, args))
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

func (o *CreateImagesOptions) Validate(_ *cobra.Command, _ []string) error {
	switch o.ContainerManager {
	case common.Docker, common.Conatinerd, common.Crio, common.Isula:
	default:
		return fmt.Errorf("unsupport container runtime [%s]", o.ContainerManager)
	}
	return nil
}

func (o *CreateImagesOptions) Run() error {
	arg := common.Argument{
		FilePath:          o.ClusterCfgFile,
		KubernetesVersion: o.Kubernetes,
		ContainerManager:  o.ContainerManager,
		Debug:             o.CommonOptions.Verbose,
	}
	return images.CreateImages(arg)
}

func (o *CreateImagesOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.ClusterCfgFile, "filename", "f", "", "Path to a configuration file")
	cmd.Flags().StringVarP(&o.Kubernetes, "with-kubernetes", "", "", "Specify a supported version of kubernetes")
	cmd.Flags().StringVarP(&o.ContainerManager, "container-manager", "", "docker", "Container runtime: docker, crio, containerd and isula.")
}
