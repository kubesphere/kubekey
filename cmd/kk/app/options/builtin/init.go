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
	"path/filepath"

	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cliflag "k8s.io/component-base/cli/flag"

	"github.com/kubesphere/kubekey/v4/cmd/kk/app/options"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
)

// ======================================================================================
//                                     init os
// ======================================================================================

// InitOSOptions for NewInitOSOptions
type InitOSOptions struct {
	options.CommonOptions
}

// NewInitOSOptions for newInitOSCommand
func NewInitOSOptions() *InitOSOptions {
	// set default value
	return &InitOSOptions{CommonOptions: options.NewCommonOptions()}
}

// Flags add to newInitOSCommand
func (o *InitOSOptions) Flags() cliflag.NamedFlagSets {
	fss := o.CommonOptions.Flags()

	return fss
}

// Complete options. create Pipeline, Config and Inventory
func (o *InitOSOptions) Complete(cmd *cobra.Command, args []string) (*kkcorev1.Pipeline, error) {
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
	if len(args) != 1 {
		return nil, fmt.Errorf("%s\nSee '%s -h' for help and examples", cmd.Use, cmd.CommandPath())
	}
	o.Playbook = args[0]

	pipeline.Spec = kkcorev1.PipelineSpec{
		Playbook: o.Playbook,
		Debug:    o.Debug,
	}

	if err := o.CommonOptions.Complete(pipeline); err != nil {
		return nil, err
	}

	return pipeline, o.complateConfig()
}

func (o *InitOSOptions) complateConfig() error {
	if workdir, err := o.CommonOptions.Config.GetValue("workdir"); err != nil {
		// workdir should set by CommonOptions
		return err
	} else {
		wd, ok := workdir.(string)
		if !ok {
			return fmt.Errorf("workdir should be string value")
		}
		// set binary dir if not set
		if _, err := o.CommonOptions.Config.GetValue(_const.BinaryDir); err != nil {
			// not found set default
			if err := o.CommonOptions.Config.SetValue(_const.BinaryDir, filepath.Join(wd, "kubekey")); err != nil {
				return err
			}
		}
	}

	return nil
}

// ======================================================================================
//                                    init registry
// ======================================================================================

// InitRegistryOptions for NewInitRegistryOptions
type InitRegistryOptions struct {
	options.CommonOptions
}

// NewInitRegistryOptions for newInitRegistryCommand
func NewInitRegistryOptions() *InitRegistryOptions {
	// set default value
	return &InitRegistryOptions{CommonOptions: options.NewCommonOptions()}
}

// Flags add to newInitRegistryCommand
func (o *InitRegistryOptions) Flags() cliflag.NamedFlagSets {
	fss := o.CommonOptions.Flags()

	return fss
}

// Complete options. create Pipeline, Config and Inventory
func (o *InitRegistryOptions) Complete(cmd *cobra.Command, args []string) (*kkcorev1.Pipeline, error) {
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
	if len(args) != 1 {
		return nil, fmt.Errorf("%s\nSee '%s -h' for help and examples", cmd.Use, cmd.CommandPath())
	}
	o.Playbook = args[0]

	pipeline.Spec = kkcorev1.PipelineSpec{
		Playbook: o.Playbook,
		Debug:    o.Debug,
	}
	if err := o.CommonOptions.Complete(pipeline); err != nil {
		return nil, err
	}

	return pipeline, nil
}
