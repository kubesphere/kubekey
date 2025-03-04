//go:build builtin
// +build builtin

/*
Copyright 2024 The KubeSphere Authors.

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

package builtin

import (
	"fmt"

	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cliflag "k8s.io/component-base/cli/flag"

	"github.com/kubesphere/kubekey/v4/cmd/kk/app/options"
)

const (
	defaultKubeVersion      = "v1.23.15"
	defaultContainerManager = "containerd"
)

// NewCreateClusterOptions for newCreateClusterCommand
func NewCreateClusterOptions() *CreateClusterOptions {
	// set default value
	return &CreateClusterOptions{
		CommonOptions:    options.NewCommonOptions(),
		Kubernetes:       defaultKubeVersion,
		ContainerManager: defaultContainerManager,
	}
}

// CreateClusterOptions for NewCreateClusterOptions
type CreateClusterOptions struct {
	options.CommonOptions
	// kubernetes version which the cluster will install.
	Kubernetes string
	// ContainerRuntime for kubernetes. Such as docker, containerd etc.
	ContainerManager string
}

// Flags add to newCreateClusterCommand
func (o *CreateClusterOptions) Flags() cliflag.NamedFlagSets {
	fss := o.CommonOptions.Flags()
	kfs := fss.FlagSet("config")
	kfs.StringVar(&o.Kubernetes, "with-kubernetes", o.Kubernetes, fmt.Sprintf("Specify a supported version of kubernetes. default is %s", o.Kubernetes))
	kfs.StringVar(&o.ContainerManager, "container-manager", o.ContainerManager, fmt.Sprintf("Container runtime: docker, crio, containerd and isula. default is %s", o.ContainerManager))

	return fss
}

// Complete options. create Pipeline, Config and Inventory
func (o *CreateClusterOptions) Complete(cmd *cobra.Command, args []string) (*kkcorev1.Pipeline, error) {
	pipeline := &kkcorev1.Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "create-cluster-",
			Namespace:    o.Namespace,
			Annotations: map[string]string{
				kkcorev1.BuiltinsProjectAnnotation: "",
			},
		},
	}

	// complete playbook. now only support one playbook
	if len(args) != 1 {
		return nil, fmt.Errorf("%s\nSee '%s -h' for help and examples", cmd.Use, cmd.CommandPath())
	}
	o.Playbook = args[0]

	pipeline.Spec = kkcorev1.PipelineSpec{
		Playbook: o.Playbook,
		Debug:    o.Debug,
	}
	// override kube_version in config
	if err := completeConfig(o.Kubernetes, o.CommonOptions.ConfigFile, o.CommonOptions.Config); err != nil {
		return nil, fmt.Errorf("cannot get config for kube_version: %s. error is %w. Please set it by \"--config\"", o.Kubernetes, err)
	}
	if err := completeInventory(o.CommonOptions.InventoryFile, o.CommonOptions.Inventory); err != nil {
		return nil, fmt.Errorf("cannot get local inventory. error is %w. Please set it by \"--inventory\"", err)
	}
	if err := o.CommonOptions.Complete(pipeline); err != nil {
		return nil, err
	}

	return pipeline, o.completeConfig()
}

func (o *CreateClusterOptions) completeConfig() error {
	if o.ContainerManager != "" {
		// override container_manager in config
		if err := o.CommonOptions.Config.SetValue("cri.container_manager", o.ContainerManager); err != nil {
			return err
		}
	}

	if err := o.CommonOptions.Config.SetValue("kube_version", o.Kubernetes); err != nil {
		return err
	}

	return nil
}
