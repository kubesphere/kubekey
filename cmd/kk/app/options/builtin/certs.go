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

// NewCertsRenewOptions for newCertsRenewCommand
func NewCertsRenewOptions() *CertsRenewOptions {
	// set default value
	return &CertsRenewOptions{CommonOptions: options.NewCommonOptions()}
}

// CertsRenewOptions for NewCertsRenewOptions
type CertsRenewOptions struct {
	options.CommonOptions
}

// Flags add to newCertsRenewCommand
func (o *CertsRenewOptions) Flags() cliflag.NamedFlagSets {
	fss := o.CommonOptions.Flags()

	return fss
}

// Complete options. create Pipeline, Config and Inventory
func (o *CertsRenewOptions) Complete(cmd *cobra.Command, args []string) (*kkcorev1.Pipeline, error) {
	pipeline := &kkcorev1.Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "certs-renew-",
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
		Tags:     []string{"certs"},
	}
	if err := o.CommonOptions.Complete(pipeline); err != nil {
		return nil, err
	}

	return pipeline, nil
}
