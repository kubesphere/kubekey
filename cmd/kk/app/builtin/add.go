//go:build builtin
// +build builtin

/*
Copyright 2025 The KubeSphere Authors.

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

// NewAddCommand creates a new cobra command for adding nodes to a Kubernetes cluster.
// It adds the "nodes" subcommand to enable adding worker nodes.
func NewAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add nodes to kubernetes cluster",
	}
	cmd.AddCommand(newAddNodeCommand())

	return cmd
}

// newAddNodeCommand creates a new cobra command for adding worker nodes to an existing cluster.
// It uses the AddNodeOptions to handle configuration and execution of the add nodes operation.
func newAddNodeCommand() *cobra.Command {
	o := builtin.NewAddNodeOptions()

	cmd := &cobra.Command{
		Use:     "nodes",
		Aliases: []string{"node"},
		Short:   "Add nodes to the cluster according to the new nodes information from the specified configuration file",
		Long: `There are two ways to add nodes to the cluster:

1. kk add nodes
   Requires all nodes to be pre-defined in inventory.yaml with their assigned groups 
   (kube_control_plane, kube_worker, etcd). This will add all nodes listed in the 
   inventory that are not yet installed in the cluster.

2. kk add nodes --control-plane node1,node2 --worker node1,node2 --etcd node1,node2
   Only requires node connection details (hostname/IP) to be defined in inventory.yaml.
   The nodes will be automatically assigned to the specified groups via command-line flags.
   The --override flag will update the inventory.yaml with the new group assignments.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Complete the configuration and create a playbook for adding nodes
			playbook, err := o.Complete(cmd, []string{"playbooks/add_nodes.yaml"})
			if err != nil {
				return err
			}

			// Execute the playbook to add the nodes
			if err := o.Run(cmd.Context(), playbook); err != nil {
				return err
			}

			// Update inventory file only after successful execution
			return o.OverrideInventory()
		},
	}
	flags := cmd.Flags()
	for _, f := range o.Flags().FlagSets {
		flags.AddFlagSet(f)
	}

	return cmd
}
