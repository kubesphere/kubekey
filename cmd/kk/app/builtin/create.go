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

// NewCreateCommand creates a new cobra command for creating a cluster or a cluster configuration file.
// It returns a pointer to the created cobra.Command.
func NewCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a cluster or a cluster configuration file",
	}
	cmd.AddCommand(newCreateClusterCommand())
	cmd.AddCommand(newCreateConfigCommand())

	return cmd
}

func newCreateClusterCommand() *cobra.Command {
	o := builtin.NewCreateClusterOptions()

	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Create a Kubernetes or KubeSphere cluster",
		RunE: func(cmd *cobra.Command, _ []string) error {
			playbook, err := o.Complete(cmd, []string{"playbooks/create_cluster.yaml"})
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

func newCreateConfigCommand() *cobra.Command {
	o := builtin.NewCreateConfigOptions()

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Create a Kubernetes or KubeSphere cluster",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return o.Run()
		},
	}
	flags := cmd.Flags()
	for _, f := range o.Flags().FlagSets {
		flags.AddFlagSet(f)
	}

	return cmd
}
