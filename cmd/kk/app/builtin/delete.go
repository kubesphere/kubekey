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

// NewDeleteCommand creates a new delete command that allows deleting nodes or clusters.
// It provides subcommands for deleting either an entire cluster or individual nodes, as well as the image registry.
func NewDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete node or cluster",
	}
	// Add subcommands for cluster, nodes, and image registry deletion
	cmd.AddCommand(newDeleteClusterCommand())
	cmd.AddCommand(newDeleteNodesCommand())
	cmd.AddCommand(newDeleteImageRegistryCommand())

	return cmd
}

// newDeleteClusterCommand creates a new command for deleting a Kubernetes cluster.
// It uses the delete_cluster.yaml playbook to:
// - Uninstall Kubernetes components from all nodes
// - Clean up container runtime if specified
// - Remove DNS entries if specified
// - Clean up etcd if specified
func newDeleteClusterCommand() *cobra.Command {
	// Initialize options for deleting a cluster
	o := builtin.NewDeleteClusterOptions()

	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Delete a cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Complete the configuration and create a playbook for deleting the cluster
			playbook, err := o.Complete(cmd, []string{"playbooks/delete_cluster.yaml"})
			if err != nil {
				return err
			}

			// Execute the playbook to delete the cluster
			return o.CommonOptions.Run(cmd.Context(), playbook)
		},
	}
	// Add all relevant flag sets to the command
	flags := cmd.Flags()
	for _, f := range o.Flags().FlagSets {
		flags.AddFlagSet(f)
	}

	return cmd
}

// newDeleteNodesCommand creates a new command for deleting specific nodes from a cluster.
// It uses the delete_nodes.yaml playbook to:
// - Remove Kubernetes components from specified nodes
// - Clean up container runtime if specified
// - Remove DNS entries if specified
// - Clean up etcd if the node was part of etcd cluster
func newDeleteNodesCommand() *cobra.Command {
	// Initialize options for deleting nodes
	o := builtin.NewDeleteNodesOptions()

	cmd := &cobra.Command{
		Use:     "nodes {node1 node2 ...}",
		Aliases: []string{"node"},
		Short:   "Delete a cluster nodes",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Complete the configuration and create a playbook for deleting nodes
			// The playbook path is appended as the last argument
			playbook, err := o.Complete(cmd, append(args, "playbooks/delete_nodes.yaml"))
			if err != nil {
				return err
			}

			// Execute the playbook to delete the specified nodes
			return o.CommonOptions.Run(cmd.Context(), playbook)
		},
	}
	// Add all relevant flag sets to the command
	flags := cmd.Flags()
	for _, f := range o.Flags().FlagSets {
		flags.AddFlagSet(f)
	}

	return cmd
}

// newDeleteImageRegistryCommand creates a new command for deleting the image registry created by kubekey.
// It uses the delete_image_registry.yaml playbook to remove the image registry and optionally its container runtime.
func newDeleteImageRegistryCommand() *cobra.Command {
	// Initialize options for deleting the image registry
	o := builtin.NewDeleteImageRegistryOptions()

	cmd := &cobra.Command{
		Use:   "image_registry",
		Short: "Delete a image_registry which create by kubekey.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Complete the configuration and create a playbook for deleting the image registry
			playbook, err := o.Complete(cmd, []string{"playbooks/delete_image_registry.yaml"})
			if err != nil {
				return err
			}

			// Execute the playbook to delete the image registry
			return o.CommonOptions.Run(cmd.Context(), playbook)
		},
	}
	// Add all relevant flag sets to the command
	flags := cmd.Flags()
	for _, f := range o.Flags().FlagSets {
		flags.AddFlagSet(f)
	}

	return cmd
}
