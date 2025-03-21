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
	"github.com/cockroachdb/errors"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cliflag "k8s.io/component-base/cli/flag"

	"github.com/kubesphere/kubekey/v4/cmd/kk/app/options"
)

// ======================================================================================
//                                    artifact export
// ======================================================================================

// ArtifactExportOptions for NewArtifactExportOptions
type ArtifactExportOptions struct {
	options.CommonOptions
}

// NewArtifactExportOptions for newArtifactExportCommand
func NewArtifactExportOptions() *ArtifactExportOptions {
	// set default value
	return &ArtifactExportOptions{CommonOptions: options.NewCommonOptions()}
}

// Flags add to newArtifactExportCommand
func (o *ArtifactExportOptions) Flags() cliflag.NamedFlagSets {
	fss := o.CommonOptions.Flags()

	return fss
}

// Complete options. create Playbook, Config and Inventory
func (o *ArtifactExportOptions) Complete(cmd *cobra.Command, args []string) (*kkcorev1.Playbook, error) {
	playbook := &kkcorev1.Playbook{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "artifact-export-",
			Namespace:    o.Namespace,
			Annotations: map[string]string{
				kkcorev1.BuiltinsProjectAnnotation: "",
			},
		},
	}

	// complete playbook. now only support one playbook
	if len(args) != 1 {
		return nil, errors.Errorf("%s\nSee '%s -h' for help and examples", cmd.Use, cmd.CommandPath())
	}
	o.Playbook = args[0]

	playbook.Spec = kkcorev1.PlaybookSpec{
		Playbook: o.Playbook,
		Debug:    o.Debug,
		SkipTags: []string{"certs"},
	}
	if err := completeInventory(o.CommonOptions.InventoryFile, o.CommonOptions.Inventory); err != nil {
		return nil, errors.Wrap(err, "failed to get local inventory. Please set it by \"--inventory\"")
	}
	if err := o.CommonOptions.Complete(playbook); err != nil {
		return nil, errors.WithStack(err)
	}

	return playbook, nil
}

// ======================================================================================
//                                   artifact image
// ======================================================================================

// ArtifactImagesOptions for NewArtifactImagesOptions
type ArtifactImagesOptions struct {
	options.CommonOptions
}

// NewArtifactImagesOptions for newArtifactImagesCommand
func NewArtifactImagesOptions() *ArtifactImagesOptions {
	// set default value
	return &ArtifactImagesOptions{CommonOptions: options.NewCommonOptions()}
}

// Flags add to newArtifactImagesCommand
func (o *ArtifactImagesOptions) Flags() cliflag.NamedFlagSets {
	fss := o.CommonOptions.Flags()

	return fss
}

// Complete options. create Playbook, Config and Inventory
func (o *ArtifactImagesOptions) Complete(cmd *cobra.Command, args []string) (*kkcorev1.Playbook, error) {
	playbook := &kkcorev1.Playbook{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "artifact-images-",
			Namespace:    o.Namespace,
			Annotations: map[string]string{
				kkcorev1.BuiltinsProjectAnnotation: "",
			},
		},
	}

	// complete playbook. now only support one playbook
	if len(args) != 1 {
		return nil, errors.Errorf("%s\nSee '%s -h' for help and examples", cmd.Use, cmd.CommandPath())
	}
	o.Playbook = args[0]

	playbook.Spec = kkcorev1.PlaybookSpec{
		Playbook: o.Playbook,
		Debug:    o.Debug,
		Tags:     []string{"only_image"},
	}

	if err := o.CommonOptions.Complete(playbook); err != nil {
		return nil, errors.WithStack(err)
	}

	return playbook, nil
}
