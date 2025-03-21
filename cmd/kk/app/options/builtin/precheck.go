//go:build builtin
// +build builtin

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

package builtin

import (
	"github.com/cockroachdb/errors"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cliflag "k8s.io/component-base/cli/flag"

	"github.com/kubesphere/kubekey/v4/cmd/kk/app/options"
)

// NewPreCheckOptions for newPreCheckCommand
func NewPreCheckOptions() *PreCheckOptions {
	// set default value
	return &PreCheckOptions{CommonOptions: options.NewCommonOptions()}
}

// PreCheckOptions for NewPreCheckOptions
type PreCheckOptions struct {
	options.CommonOptions
}

// Flags add to newPreCheckCommand
func (o *PreCheckOptions) Flags() cliflag.NamedFlagSets {
	return o.CommonOptions.Flags()
}

// Complete options. create Playbook, Config and Inventory
func (o *PreCheckOptions) Complete(cmd *cobra.Command, args []string) (*kkcorev1.Playbook, error) {
	playbook := &kkcorev1.Playbook{
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
		return nil, errors.Errorf("%s\nSee '%s -h' for help and examples", cmd.Use, cmd.CommandPath())
	} else if len(args) == 1 {
		o.Playbook = args[0]
	} else {
		tags = args[:len(args)-1]
		o.Playbook = args[len(args)-1]
	}

	playbook.Spec = kkcorev1.PlaybookSpec{
		Playbook: o.Playbook,
		Debug:    o.Debug,
		Tags:     tags,
	}
	if err := completeInventory(o.CommonOptions.InventoryFile, o.CommonOptions.Inventory); err != nil {
		return nil, errors.WithStack(err)
	}
	if err := o.CommonOptions.Complete(playbook); err != nil {
		return nil, errors.WithStack(err)
	}

	return playbook, nil
}
