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
//                                    artifact export
// ======================================================================================

// ArtifactExportOptions for NewArtifactExportOptions
type ArtifactExportOptions struct {
	commonOptions
}

// Flags add to newArtifactExportCommand
func (o *ArtifactExportOptions) Flags() cliflag.NamedFlagSets {
	fss := o.commonOptions.flags()

	return fss
}

// Complete options. create Pipeline, Config and Inventory
func (o *ArtifactExportOptions) Complete(cmd *cobra.Command, args []string) (*kkcorev1.Pipeline, *kkcorev1.Config, *kkcorev1.Inventory, error) {
	pipeline := &kkcorev1.Pipeline{
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

	return pipeline, config, inventory, nil
}

// NewArtifactExportOptions for newArtifactExportCommand
func NewArtifactExportOptions() *ArtifactExportOptions {
	// set default value
	return &ArtifactExportOptions{commonOptions: newCommonOptions()}
}

// ======================================================================================
//                                   artifact image
// ======================================================================================

// ArtifactImagesOptions for NewArtifactImagesOptions
type ArtifactImagesOptions struct {
	commonOptions
}

// Flags add to newArtifactImagesCommand
func (o *ArtifactImagesOptions) Flags() cliflag.NamedFlagSets {
	fss := o.commonOptions.flags()

	return fss
}

// Complete options. create Pipeline, Config and Inventory
func (o *ArtifactImagesOptions) Complete(cmd *cobra.Command, args []string) (*kkcorev1.Pipeline, *kkcorev1.Config, *kkcorev1.Inventory, error) {
	pipeline := &kkcorev1.Pipeline{
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
		return nil, nil, nil, fmt.Errorf("%s\nSee '%s -h' for help and examples", cmd.Use, cmd.CommandPath())
	}
	o.Playbook = args[0]

	pipeline.Spec = kkcorev1.PipelineSpec{
		Playbook: o.Playbook,
		Debug:    o.Debug,
		Tags:     []string{"only_image"},
	}

	config, inventory, err := o.completeRef(pipeline)
	if err != nil {
		return nil, nil, nil, err
	}

	return pipeline, config, inventory, nil
}

// NewArtifactImagesOptions for newArtifactImagesCommand
func NewArtifactImagesOptions() *ArtifactImagesOptions {
	// set default value
	return &ArtifactImagesOptions{commonOptions: newCommonOptions()}
}
