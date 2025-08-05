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
	"fmt"

	"github.com/cockroachdb/errors"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	cliflag "k8s.io/component-base/cli/flag"

	"github.com/kubesphere/kubekey/v4/cmd/kk/app/options"
)

// NewPreCheckOptions for newPreCheckCommand
func NewPreCheckOptions() *PreCheckOptions {
	// set default value
	o := &PreCheckOptions{
		CommonOptions: options.NewCommonOptions(),
		Kubernetes:    defaultKubeVersion,
	}
	o.CommonOptions.GetConfigFunc = func() (*kkcorev1.Config, error) {
		data, err := getConfig(o.Kubernetes)
		if err != nil {
			return nil, err
		}
		config := &kkcorev1.Config{}
		return config, errors.Wrapf(yaml.Unmarshal(data, config), "failed to unmarshal local configFile for kube_version: %q.", o.Kubernetes)
	}
	o.CommonOptions.GetInventoryFunc = getInventory

	return o
}

// PreCheckOptions for NewPreCheckOptions
type PreCheckOptions struct {
	options.CommonOptions
	// kubernetes version which the cluster will install.
	Kubernetes string
}

// Flags add to newPreCheckCommand
func (o *PreCheckOptions) Flags() cliflag.NamedFlagSets {
	fss := o.CommonOptions.Flags()
	kfs := fss.FlagSet("config")
	kfs.StringVar(&o.Kubernetes, "with-kubernetes", o.Kubernetes, fmt.Sprintf("Specify a supported version of kubernetes. default is %s", o.Kubernetes))

	return fss
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
		Tags:     tags,
	}

	return playbook, o.CommonOptions.Complete(playbook)
}
