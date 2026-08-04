//go:build builtin
// +build builtin

/*
Copyright 2026 The KubeSphere Authors.

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

	"github.com/spf13/cobra"

	"github.com/kubesphere/kubekey/v4/cmd/kk/app/options/builtin"
)

// NewUpgradeCommand creates a new upgrade command that allows upgrading a cluster.
// It provides subcommands for upgrading the cluster and individual components.
func NewUpgradeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade a Kubernetes cluster",
	}
	cmd.AddCommand(newUpgradeClusterCommand())
	// Standalone component upgrades: upgrade ONLY the named component, leaving the
	// Kubernetes control plane and other components untouched.
	cmd.AddCommand(newUpgradeComponentCommand("etcd", "etcd"))
	cmd.AddCommand(newUpgradeComponentCommand("cri", "cri"))
	cmd.AddCommand(newUpgradeComponentCommand("cni", "cni"))
	cmd.AddCommand(newUpgradeComponentCommand("storageclass", "storage_class"))

	return cmd
}

// newUpgradeClusterCommand creates a new command for upgrading a Kubernetes cluster.
// It uses the upgrade_cluster.yaml playbook to:
// - Upgrade Kubernetes control plane components using kubeadm upgrade
// - Upgrade kubelet on all nodes
// - Optionally upgrade cri, cni, storageclass and etcd when --all is set
func newUpgradeClusterCommand() *cobra.Command {
	// Initialize options for upgrading a cluster
	o := builtin.NewUpgradeClusterOptions()

	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Upgrade a Kubernetes cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Complete the configuration and create a playbook for upgrading the cluster
			playbook, err := o.Complete(cmd, []string{"playbooks/upgrade_cluster.yaml"})
			if err != nil {
				return err
			}

			// Execute the playbook to upgrade the cluster
			return o.Run(cmd.Context(), playbook)
		},
	}
	// Add all relevant flag sets to the command
	flags := cmd.Flags()
	for _, f := range o.Flags().FlagSets {
		flags.AddFlagSet(f)
	}

	return cmd
}

// newUpgradeComponentCommand creates a command that upgrades only a single component
// (e.g. `kk upgrade etcd`), leaving the Kubernetes control plane and other components
// untouched. It reuses the same upgrade_cluster.yaml playbook but presets OnlyComponent
// so that completeConfig disables the Kubernetes base and enables just this component.
func newUpgradeComponentCommand(name, compKey string) *cobra.Command {
	o := builtin.NewUpgradeClusterOptions()
	o.OnlyComponent = compKey

	cmd := &cobra.Command{
		Use:   name,
		Short: fmt.Sprintf("Upgrade only the %s component", name),
		RunE: func(cmd *cobra.Command, args []string) error {
			playbook, err := o.Complete(cmd, []string{"playbooks/upgrade_cluster.yaml"})
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
	// Component-only upgrades never touch the Kubernetes control plane, so the
	// cluster-level flags are meaningless (and --all would be silently ignored
	// because OnlyComponent takes precedence). Hide them to avoid confusion.
	_ = flags.MarkHidden("all")
	_ = flags.MarkHidden("with-kubernetes")

	return cmd
}
