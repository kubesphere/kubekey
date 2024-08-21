/*
Copyright 2023 The KubeSphere Authors.

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

// NewPreCheckOptions for newPreCheckCommand
func NewPreCheckOptions() *PreCheckOptions {
	// set default value
	return &PreCheckOptions{commonOptions: newCommonOptions()}
}

// PreCheckOptions for NewPreCheckOptions
type PreCheckOptions struct {
	commonOptions
}

// Flags add to newPreCheckCommand
func (o *PreCheckOptions) Flags() cliflag.NamedFlagSets {
	return o.commonOptions.flags()
}

// Complete options. create Pipeline, Config and Inventory
func (o *PreCheckOptions) Complete(cmd *cobra.Command, args []string) (*kkcorev1.Pipeline, *kkcorev1.Config, *kkcorev1.Inventory, error) {
	pipeline := &kkcorev1.Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "precheck-",
			Namespace:    o.Namespace,
			Annotations: map[string]string{
				kkcorev1.BuiltinsProjectAnnotation: "",
			},
		},
	}

	// complete playbook. now only support one playbook
	var tags []string
	if len(args) < 1 {
		return nil, nil, nil, fmt.Errorf("%s\nSee '%s -h' for help and examples", cmd.Use, cmd.CommandPath())
	} else if len(args) == 1 {
		o.Playbook = args[0]
	} else {
		tags = args[:len(args)-1]
		o.Playbook = args[len(args)-1]
	}

	pipeline.Spec = kkcorev1.PipelineSpec{
		Playbook: o.Playbook,
		Debug:    o.Debug,
		Tags:     tags,
	}
	config, inventory, err := o.completeRef(pipeline)
	if err != nil {
		return nil, nil, nil, err
	}

	return pipeline, config, inventory, nil
}
