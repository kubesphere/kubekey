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

package options

import (
	"fmt"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cliflag "k8s.io/component-base/cli/flag"

	kkcorev1 "github.com/kubesphere/kubekey/v4/pkg/apis/core/v1"
)

// NewCreateClusterOptions for newCreateClusterCommand
func NewCreateClusterOptions() *CreateClusterOptions {
	// set default value
	return &CreateClusterOptions{commonOptions: newCommonOptions()}
}

// CreateClusterOptions for NewCreateClusterOptions
type CreateClusterOptions struct {
	commonOptions
	// kubernetes version which the cluster will install.
	Kubernetes string
	// ContainerRuntime for kubernetes. Such as docker, containerd etc.
	ContainerManager string
}

// Flags add to newCreateClusterCommand
func (o *CreateClusterOptions) Flags() cliflag.NamedFlagSets {
	fss := o.commonOptions.flags()
	kfs := fss.FlagSet("config")
	kfs.StringVar(&o.Kubernetes, "with-kubernetes", "", "Specify a supported version of kubernetes")
	kfs.StringVar(&o.ContainerManager, "container-manager", "", "Container runtime: docker, crio, containerd and isula.")

	return fss
}

// Complete options. create Pipeline, Config and Inventory
func (o *CreateClusterOptions) Complete(cmd *cobra.Command, args []string) (*kkcorev1.Pipeline, *kkcorev1.Config, *kkcorev1.Inventory, error) {
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
		return nil, nil, nil, fmt.Errorf("%s\nSee '%s -h' for help and examples", cmd.Use, cmd.CommandPath())
	}
	o.Playbook = args[0]

	pipeline.Spec = kkcorev1.PipelineSpec{
		Playbook: o.Playbook,
		Debug:    o.Debug,
	}

	config, inventory, err := o.completeRef(pipeline)
	if err != nil {
		return nil, nil, nil, err
	}

	if o.Kubernetes != "" {
		// override kube_version in config
		if err := config.SetValue("kube_version", o.Kubernetes); err != nil {
			return nil, nil, nil, err
		}
	}
	if o.ContainerManager != "" {
		// override container_manager in config
		if err := config.SetValue("cri.container_manager", o.ContainerManager); err != nil {
			return nil, nil, nil, err
		}
	}

	return pipeline, config, inventory, nil
}
