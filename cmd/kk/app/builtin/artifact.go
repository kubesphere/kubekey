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
	"github.com/spf13/cobra"

	"github.com/kubesphere/kubekey/v4/cmd/kk/app/options/builtin"
)

// NewArtifactCommand creates a new cobra.Command for managing KubeKey offline installation packages.
// It adds subcommands for exporting artifacts and managing images.
func NewArtifactCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "artifact",
		Short: "Manage a KubeKey offline installation package",
	}
	cmd.AddCommand(newArtifactExportCommand())
	cmd.AddCommand(newArtifactImagesCommand())

	return cmd
}

func newArtifactExportCommand() *cobra.Command {
	o := builtin.NewArtifactExportOptions()

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export a KubeKey offline installation package",
		RunE: func(cmd *cobra.Command, _ []string) error {
			playbook, err := o.Complete(cmd, []string{"playbooks/artifact_export.yaml"})
			if err != nil {
				return err
			}

			return o.CommonOptions.Run(cmd.Context(), playbook)
		},
	}
	flags := cmd.Flags()
	for _, f := range o.Flags().FlagSets {
		flags.AddFlagSet(f)
	}

	return cmd
}

func newArtifactImagesCommand() *cobra.Command {
	o := builtin.NewArtifactImagesOptions()

	cmd := &cobra.Command{
		Use:   "images",
		Short: "push images to a registry from an artifact",
		RunE: func(cmd *cobra.Command, _ []string) error {
			playbook, err := o.Complete(cmd, []string{"playbooks/artifact_images.yaml"})
			if err != nil {
				return err
			}

			return o.CommonOptions.Run(cmd.Context(), playbook)
		},
	}
	flags := cmd.Flags()
	for _, f := range o.Flags().FlagSets {
		flags.AddFlagSet(f)
	}

	return cmd
}
