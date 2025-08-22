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
	// kubernetes version which the cluster will install.
	Kubernetes string
}

// NewArtifactExportOptions for newArtifactExportCommand
func NewArtifactExportOptions() *ArtifactExportOptions {
	// set default value
	o := &ArtifactExportOptions{CommonOptions: options.NewCommonOptions()}
	o.CommonOptions.GetInventoryFunc = getInventory

	return o
}

// Flags add to newArtifactExportCommand
func (o *ArtifactExportOptions) Flags() cliflag.NamedFlagSets {
	fss := o.CommonOptions.Flags()
	kfs := fss.FlagSet("config")
	kfs.StringVar(&o.Kubernetes, "with-kubernetes", o.Kubernetes, fmt.Sprintf("Specify a supported version of kubernetes. default is %s", o.Kubernetes))

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
		SkipTags: []string{"certs"},
	}
	if err := o.CommonOptions.Complete(playbook); err != nil {
		return nil, err
	}

	return playbook, nil
}

// ======================================================================================
//                                   artifact image
// ======================================================================================

// ArtifactImagesOptions for NewArtifactImagesOptions
type ArtifactImagesOptions struct {
	options.CommonOptions
	// kubernetes version which the cluster will install.
	Kubernetes string
	Push       bool
	Pull       bool
}

// NewArtifactImagesOptions for newArtifactImagesCommand
func NewArtifactImagesOptions() *ArtifactImagesOptions {
	o := &ArtifactImagesOptions{CommonOptions: options.NewCommonOptions()}
	o.CommonOptions.GetInventoryFunc = getInventory

	return o
}

// Flags add to newArtifactImagesCommand
func (o *ArtifactImagesOptions) Flags() cliflag.NamedFlagSets {
	fss := o.CommonOptions.Flags()
	kfs := fss.FlagSet("config")
	kfs.StringVar(&o.Kubernetes, "with-kubernetes", o.Kubernetes, fmt.Sprintf("Specify a supported version of kubernetes. default is %s", o.Kubernetes))
	kfs.BoolVar(&o.Push, "push", o.Push, "Push image to image registry")
	kfs.BoolVar(&o.Pull, "pull", o.Pull, "Pull image to binary dir")

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

	var tags []string
	if o.Push {
		tags = append(tags, "push")
	}
	if o.Pull {
		tags = append(tags, "pull")
	}

	playbook.Spec = kkcorev1.PlaybookSpec{
		Playbook: o.Playbook,
		Tags:     tags,
	}

	if err := o.CommonOptions.Complete(playbook); err != nil {
		return nil, errors.WithStack(err)
	}

	return playbook, nil
}
