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

	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
)

// ======================================================================================
//                                    artifact export
// ======================================================================================

type ArtifactExportOptions struct {
	CommonOptions
}

func (o *ArtifactExportOptions) Flags() cliflag.NamedFlagSets {
	fss := o.CommonOptions.Flags()
	return fss
}

func (o ArtifactExportOptions) Complete(cmd *cobra.Command, args []string) (*kubekeyv1.Pipeline, *kubekeyv1.Config, *kubekeyv1.Inventory, error) {
	pipeline := &kubekeyv1.Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "artifact-export-",
			Namespace:    o.Namespace,
			Annotations: map[string]string{
				kubekeyv1.BuiltinsProjectAnnotation: "",
			},
		},
	}

	// complete playbook. now only support one playbook
	if len(args) == 1 {
		o.Playbook = args[0]
	} else {
		return nil, nil, nil, fmt.Errorf("%s\nSee '%s -h' for help and examples", cmd.Use, cmd.CommandPath())
	}

	pipeline.Spec = kubekeyv1.PipelineSpec{
		Playbook: o.Playbook,
		Debug:    o.Debug,
	}
	config, inventory, err := o.completeRef(pipeline)
	if err != nil {
		return nil, nil, nil, err
	}

	return pipeline, config, inventory, nil
}

func NewArtifactExportOptions() *ArtifactExportOptions {
	// set default value
	return &ArtifactExportOptions{CommonOptions: newCommonOptions()}
}

// ======================================================================================
//                                   artifact image
// ======================================================================================

type ArtifactImagesOptions struct {
	CommonOptions
}

func (o *ArtifactImagesOptions) Flags() cliflag.NamedFlagSets {
	fss := o.CommonOptions.Flags()
	return fss
}

func (o ArtifactImagesOptions) Complete(cmd *cobra.Command, args []string) (*kubekeyv1.Pipeline, *kubekeyv1.Config, *kubekeyv1.Inventory, error) {
	pipeline := &kubekeyv1.Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "artifact-images-",
			Namespace:    o.Namespace,
			Annotations: map[string]string{
				kubekeyv1.BuiltinsProjectAnnotation: "",
			},
		},
	}

	// complete playbook. now only support one playbook
	if len(args) == 1 {
		o.Playbook = args[0]
	} else {
		return nil, nil, nil, fmt.Errorf("%s\nSee '%s -h' for help and examples", cmd.Use, cmd.CommandPath())
	}

	pipeline.Spec = kubekeyv1.PipelineSpec{
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

func NewArtifactImagesOptions() *ArtifactImagesOptions {
	// set default value
	return &ArtifactImagesOptions{CommonOptions: newCommonOptions()}
}
