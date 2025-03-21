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
	"github.com/spf13/cobra"

	"github.com/kubesphere/kubekey/v4/cmd/kk/app/options/builtin"
)

// NewInitCommand creates a new cobra.Command for initializing the installation environment.
// It sets up the "init" command with a short description and adds subcommands for initializing
// the operating system and the registry.
func NewInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initializes the installation environment",
	}
	cmd.AddCommand(newInitOSCommand())
	cmd.AddCommand(newInitRegistryCommand())

	return cmd
}

func newInitOSCommand() *cobra.Command {
	o := builtin.NewInitOSOptions()

	cmd := &cobra.Command{
		Use:   "os",
		Short: "Init operating system",
		RunE: func(cmd *cobra.Command, _ []string) error {
			playbook, err := o.Complete(cmd, []string{"playbooks/init_os.yaml"})
			if err != nil {
				return errors.WithStack(err)
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

func newInitRegistryCommand() *cobra.Command {
	o := builtin.NewInitRegistryOptions()

	cmd := &cobra.Command{
		Use:   "registry",
		Short: "Init a local image registry",
		RunE: func(cmd *cobra.Command, _ []string) error {
			playbook, err := o.Complete(cmd, []string{"playbooks/init_registry.yaml"})
			if err != nil {
				return errors.WithStack(err)
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
