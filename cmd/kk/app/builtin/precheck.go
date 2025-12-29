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
	"github.com/spf13/cobra"

	"github.com/kubesphere/kubekey/v4/cmd/kk/app/options/builtin"
)

// NewPreCheckCommand creates a new cobra.Command for performing pre-checks on nodes
// to determine their eligibility for cluster deployment. The command supports
// specifying tags to check specific items such as etcd, os, network, cri, and nfs.
//
// Returns:
//
//	*cobra.Command: The cobra command configured for pre-checks.
func NewPreCheckCommand() *cobra.Command {
	o := builtin.NewPreCheckOptions()

	cmd := &cobra.Command{
		Use:   "precheck tags...",
		Short: "Check if the nodes is eligible for cluster deployment.",
		Long:  "the tags can specify check items. support: etcd, os, network, cri, nfs.",
		RunE: func(cmd *cobra.Command, args []string) error {
			playbook, err := o.Complete(cmd, append(args, "playbooks/precheck.yaml"))
			if err != nil {
				return err
			}

			return o.Run(cmd.Context(), playbook)
		},
	}
	flags := cmd.Flags()
	for _, f := range o.Flags().FlagSets {
		flags.AddFlagSet(f)
	}

	return cmd
}
