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

// ======================================================================================
//                                     init os
// ======================================================================================

type InitOSOptions struct {
	CommonOptions
}

func (o *InitOSOptions) Flags() cliflag.NamedFlagSets {
	fss := o.CommonOptions.Flags()
	return fss
}

func (o InitOSOptions) Complete(cmd *cobra.Command, args []string) (*kkcorev1.Pipeline, *kkcorev1.Config, *kkcorev1.Inventory, error) {
	pipeline := &kkcorev1.Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "init-os-",
			Namespace:    o.Namespace,
			Annotations: map[string]string{
				kkcorev1.BuiltinsProjectAnnotation: "",
			},
		},
	}

	// complete playbook. now only support one playbook
	if len(args) == 1 {
		o.Playbook = args[0]
	} else {
		return nil, nil, nil, fmt.Errorf("%s\nSee '%s -h' for help and examples", cmd.Use, cmd.CommandPath())
	}

	pipeline.Spec = kkcorev1.PipelineSpec{
		Playbook: o.Playbook,
		Debug:    o.Debug,
	}
	config, inventory, err := o.completeRef(pipeline)
	if err != nil {
		return nil, nil, nil, err
	}

	return pipeline, config, inventory, nil
}

func NewInitOSOptions() *InitOSOptions {
	// set default value
	return &InitOSOptions{CommonOptions: newCommonOptions()}
}

// ======================================================================================
//                                    init registry
// ======================================================================================

type InitRegistryOptions struct {
	CommonOptions
}

func (o *InitRegistryOptions) Flags() cliflag.NamedFlagSets {
	fss := o.CommonOptions.Flags()
	return fss
}

func (o InitRegistryOptions) Complete(cmd *cobra.Command, args []string) (*kkcorev1.Pipeline, *kkcorev1.Config, *kkcorev1.Inventory, error) {
	pipeline := &kkcorev1.Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "init-registry-",
			Namespace:    o.Namespace,
			Annotations: map[string]string{
				kkcorev1.BuiltinsProjectAnnotation: "",
			},
		},
	}

	// complete playbook. now only support one playbook
	if len(args) == 1 {
		o.Playbook = args[0]
	} else {
		return nil, nil, nil, fmt.Errorf("%s\nSee '%s -h' for help and examples", cmd.Use, cmd.CommandPath())
	}

	pipeline.Spec = kkcorev1.PipelineSpec{
		Playbook: o.Playbook,
		Debug:    o.Debug,
	}
	config, inventory, err := o.completeRef(pipeline)
	if err != nil {
		return nil, nil, nil, err
	}

	return pipeline, config, inventory, nil
}

func NewInitRegistryOptions() *InitRegistryOptions {
	// set default value
	return &InitRegistryOptions{CommonOptions: newCommonOptions()}
}
