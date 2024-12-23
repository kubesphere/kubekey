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

// NewCertsCommand creates a new cobra command for managing cluster certificates.
// It returns a pointer to the created cobra.Command instance.
// The command has a subcommand for renewing certificates.
func NewCertsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "certs",
		Short: "cluster certs",
	}
	cmd.AddCommand(newCertsRenewCommand())

	return cmd
}

func newCertsRenewCommand() *cobra.Command {
	o := builtin.NewCertsRenewOptions()

	cmd := &cobra.Command{
		Use:   "renew",
		Short: "renew a cluster certs",
		RunE: func(cmd *cobra.Command, _ []string) error {
			pipeline, err := o.Complete(cmd, []string{"playbooks/certs_renew.yaml"})
			if err != nil {
				return err
			}

			return o.CommonOptions.Run(cmd.Context(), pipeline)
		},
	}
	flags := cmd.Flags()
	for _, f := range o.Flags().FlagSets {
		flags.AddFlagSet(f)
	}

	return cmd
}
