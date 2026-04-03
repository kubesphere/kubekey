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
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/cockroachdb/errors"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	cliflag "k8s.io/component-base/cli/flag"

	"github.com/kubesphere/kubekey/v4/cmd/kk/app/options"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

// ======================================================================================
//                                  delete cluster
// ======================================================================================

// NewDeleteClusterOptions creates a new DeleteClusterOptions with default values
func NewDeleteClusterOptions() *DeleteClusterOptions {
	// set default value for DeleteClusterOptions
	o := &DeleteClusterOptions{
		CommonOptions: options.NewCommonOptions(),
		Kubernetes:    defaultKubeVersion,
	}
	// Set the function to get the config for the specified Kubernetes version
	// Set the function to get the inventory
	o.GetInventoryFunc = getInventory

	return o
}

// DeleteClusterOptions contains options for deleting a Kubernetes cluster
type DeleteClusterOptions struct {
	options.CommonOptions
	// Kubernetes version which the cluster will install.
	Kubernetes          string
	DeleteAllComponents bool
	DeleteData          bool
}

// Flags returns the flag sets for DeleteClusterOptions
func (o *DeleteClusterOptions) Flags() cliflag.NamedFlagSets {
	fss := o.CommonOptions.Flags()
	kfs := fss.FlagSet("config")
	// Add a flag for specifying the Kubernetes version
	kfs.StringVar(&o.Kubernetes, "with-kubernetes", o.Kubernetes, fmt.Sprintf("Specify a supported version of kubernetes. default is %s", o.Kubernetes))
	kfs.BoolVar(&o.DeleteAllComponents, "all", o.DeleteAllComponents, "Delete all cluster components, including cri, etcd, dns, and the image registry.")
	kfs.BoolVar(&o.DeleteData, "with-data", o.DeleteData, "Also delete data directories (harbor data, registry data, etc.). Use with caution.")

	return fss
}

// Complete validates and completes the DeleteClusterOptions configuration
func (o *DeleteClusterOptions) Complete(cmd *cobra.Command, args []string) (*kkcorev1.Playbook, error) {
	// Initialize playbook metadata for deleting a cluster
	playbook := &kkcorev1.Playbook{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "delete-cluster-",
			Namespace:    o.Namespace,
			Annotations: map[string]string{
				kkcorev1.BuiltinsProjectAnnotation: "",
			},
		},
	}

	// Validate playbook arguments: must have exactly one argument (the playbook)
	if len(args) != 1 {
		return nil, errors.Errorf("%s\nSee '%s -h' for help and examples", cmd.Use, cmd.CommandPath())
	}
	o.Playbook = args[0]

	// Set playbook specification
	playbook.Spec = kkcorev1.PlaybookSpec{
		Playbook: o.Playbook,
	}

	// Complete common options (e.g., config, inventory)
	if err := o.CommonOptions.Complete(playbook); err != nil {
		return nil, err
	}

	// Complete config specific to delete cluster
	return playbook, o.completeConfig()
}

// completeConfig updates the configuration with container manager settings
func (o *DeleteClusterOptions) completeConfig() error {
	// If kube_version is not set in config, set it to the specified Kubernetes version
	if _, ok, _ := unstructured.NestedFieldNoCopy(o.Config.Value(), "kubernetes", "kube_version"); !ok {
		if err := unstructured.SetNestedField(o.Config.Value(), o.Kubernetes, "kubernetes", "kube_version"); err != nil {
			return errors.Wrapf(err, "failed to set %q to config", "kube_version")
		}
	}
	if o.DeleteAllComponents {
		if err := unstructured.SetNestedField(o.Config.Value(), true, "delete", "cri"); err != nil {
			return errors.Wrapf(err, "failed to set %q to config", "delete_cri")
		}
		if err := unstructured.SetNestedField(o.Config.Value(), true, "delete", "etcd"); err != nil {
			return errors.Wrapf(err, "failed to set %q to config", "delete_etcd")
		}
		if err := unstructured.SetNestedField(o.Config.Value(), true, "delete", "dns"); err != nil {
			return errors.Wrapf(err, "failed to set %q to config", "delete_dns")
		}
		if err := unstructured.SetNestedField(o.Config.Value(), true, "delete", "image_registry"); err != nil {
			return errors.Wrapf(err, "failed to set %q to config", "delete_image_registry")
		}
		if o.DeleteData {
			if err := unstructured.SetNestedField(o.Config.Value(), true, "delete", "data"); err != nil {
				return errors.Wrapf(err, "failed to set %q to config", "delete_data")
			}
		}
	}

	return nil
}

// ======================================================================================
//                                  delete nodes
// ======================================================================================

// NewDeleteNodesOptions creates a new DeleteNodesOptions with default values
func NewDeleteNodesOptions() *DeleteNodesOptions {
	// set default value for DeleteNodesOptions
	o := &DeleteNodesOptions{
		CommonOptions: options.NewCommonOptions(),
		Kubernetes:    defaultKubeVersion,
	}
	// Set the function to get the inventory
	o.GetInventoryFunc = getInventory

	return o
}

// DeleteNodesOptions contains options for deleting Kubernetes cluster nodes
type DeleteNodesOptions struct {
	options.CommonOptions
	// Kubernetes version which the cluster will install.
	Kubernetes          string
	DeleteAllComponents bool
	DeleteData          bool
	// Override indicates whether to override the inventory file after successful execution.
	// When set to true, the inventory.yaml file will be updated.
	Override bool
	// deleteNodes stores the nodes to be deleted for later inventory update
	deleteNodes []string
}

// Flags returns the flag sets for DeleteNodesOptions
func (o *DeleteNodesOptions) Flags() cliflag.NamedFlagSets {
	fss := o.CommonOptions.Flags()
	kfs := fss.FlagSet("config")
	// Add a flag for specifying the Kubernetes version
	kfs.StringVar(&o.Kubernetes, "with-kubernetes", o.Kubernetes, fmt.Sprintf("Specify a supported version of kubernetes. default is %s", o.Kubernetes))
	kfs.BoolVar(&o.DeleteAllComponents, "all", o.DeleteAllComponents, "Delete all cluster components, including cri, etcd, dns, and the image registry.")
	kfs.BoolVar(&o.DeleteData, "with-data", o.DeleteData, "Also delete data directories (harbor data, registry data, etc.). Use with caution.")
	kfs.BoolVar(&o.Override, "override", o.Override, "Override the inventory file after successful execution")

	return fss
}

// Complete validates and completes the DeleteNodesOptions configuration
func (o *DeleteNodesOptions) Complete(cmd *cobra.Command, args []string) (*kkcorev1.Playbook, error) {
	// Initialize playbook metadata for deleting nodes
	playbook := &kkcorev1.Playbook{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "delete-nodes-",
			Namespace:    o.Namespace,
			Annotations: map[string]string{
				kkcorev1.BuiltinsProjectAnnotation: "",
			},
		},
	}

	// Validate playbook arguments: must have at least one argument (nodes + playbook)
	if len(args) < 1 {
		return nil, errors.Errorf("%s\nSee '%s -h' for help and examples", cmd.Use, cmd.CommandPath())
	}
	// All arguments except the last are node names, the last is the playbook
	nodes := args[:len(args)-1]
	o.Playbook = args[len(args)-1]

	// Set playbook specification
	playbook.Spec = kkcorev1.PlaybookSpec{
		Playbook: o.Playbook,
	}

	// Complete common options (e.g., config, inventory)
	if err := o.CommonOptions.Complete(playbook); err != nil {
		return nil, err
	}

	// Complete config specific to delete nodes
	return playbook, o.completeConfig(nodes)
}

// completeConfig updates the configuration with container manager settings
func (o *DeleteNodesOptions) completeConfig(nodes []string) error {
	// Store nodes for later inventory update
	o.deleteNodes = nodes

	// If kube_version is not set in config, set it to the specified Kubernetes version
	if _, ok, _ := unstructured.NestedFieldNoCopy(o.Config.Value(), "kubernetes", "kube_version"); !ok {
		if err := unstructured.SetNestedField(o.Config.Value(), o.Kubernetes, "kubernetes", "kube_version"); err != nil {
			return errors.Wrapf(err, "failed to set %q to config", "kube_version")
		}
	}
	// Set the list of nodes to be deleted in the config
	if err := unstructured.SetNestedStringSlice(o.Config.Value(), nodes, "delete_nodes"); err != nil {
		return errors.Wrapf(err, "failed to set %q to config", "delete_nodes")
	}

	if o.DeleteAllComponents {
		if err := unstructured.SetNestedField(o.Config.Value(), true, "delete", "cri"); err != nil {
			return errors.Wrapf(err, "failed to set %q to config", "delete_cri")
		}
		if err := unstructured.SetNestedField(o.Config.Value(), true, "delete", "etcd"); err != nil {
			return errors.Wrapf(err, "failed to set %q to config", "delete_etcd")
		}
		if err := unstructured.SetNestedField(o.Config.Value(), true, "delete", "dns"); err != nil {
			return errors.Wrapf(err, "failed to set %q to config", "delete_dns")
		}
		if err := unstructured.SetNestedField(o.Config.Value(), true, "delete", "image_registry"); err != nil {
			return errors.Wrapf(err, "failed to set %q to config", "delete_image_registry")
		}
		if o.DeleteData {
			if err := unstructured.SetNestedField(o.Config.Value(), true, "delete", "data"); err != nil {
				return errors.Wrapf(err, "failed to set %q to config", "delete_data")
			}
		}
	}

	return nil
}

// OverrideInventory updates the inventory.yaml file after successful execution.
// This should be called only when the Run method succeeds and Override flag is set.
func (o *DeleteNodesOptions) OverrideInventory() error {
	// Only update inventory file when --override flag is set
	if !o.Override || o.InventoryFile == "" || len(o.deleteNodes) == 0 {
		return nil
	}

	return o.removeNodesFromInventoryFile(o.deleteNodes)
}

// removeNodesFromInventoryFile removes nodes from inventory groups (kube_control_plane, kube_worker, etcd)
// and updates the inventory.yaml file while preserving comments and formatting
func (o *DeleteNodesOptions) removeNodesFromInventoryFile(nodes []string) error {
	// Read original file content
	content, err := os.ReadFile(o.InventoryFile)
	if err != nil {
		return errors.Wrapf(err, "failed to read inventory file %s", o.InventoryFile)
	}
	lines := strings.Split(string(content), "\n")

	// Get groups from inventory
	groups := variable.ConvertGroup(*o.Inventory)

	// Find which groups contain the nodes to be deleted
	deleteGroupHosts := make(map[string][]string)

	// Check if etcd should be deleted from config
	deleteEtcd, _, _ := unstructured.NestedBool(o.Config.Value(), "delete", "etcd")

	for _, node := range nodes {
		// Check if node exists in inventory
		if !slices.Contains(groups[_const.VariableGroupsAll], node) {
			return errors.Errorf("%q is not defined in inventory.", node)
		}
		// Check each group for the node
		for _, groupName := range []string{defaultGroupControlPlane, defaultGroupWorker, defaultGroupEtcd} {
			// Only delete from etcd group if delete.etcd is true in config
			if groupName == defaultGroupEtcd && !deleteEtcd {
				continue
			}
			if slices.Contains(groups[groupName], node) {
				if _, ok := deleteGroupHosts[groupName]; !ok {
					deleteGroupHosts[groupName] = []string{}
				}
				deleteGroupHosts[groupName] = append(deleteGroupHosts[groupName], node)
				// Remove node from the group in memory
				group := o.Inventory.Spec.Groups[groupName]
				group.Hosts = slices.DeleteFunc(group.Hosts, func(h string) bool {
					return h == node
				})
				o.Inventory.Spec.Groups[groupName] = group
			}
		}
	}

	if len(deleteGroupHosts) == 0 {
		return nil
	}

	// Find and remove lines containing the nodes from groups
	linesToRemove := o.findNodesToRemove(lines, deleteGroupHosts)
	if len(linesToRemove) == 0 {
		return nil
	}

	// Sort lines to remove in descending order to avoid index shifting
	slices.SortFunc(linesToRemove, func(a, b int) int {
		return b - a
	})

	// Remove lines
	for _, lineNum := range linesToRemove {
		if lineNum >= 0 && lineNum < len(lines) {
			lines = append(lines[:lineNum], lines[lineNum+1:]...)
		}
	}

	// Write back
	output := strings.Join(lines, "\n")
	return errors.Wrapf(os.WriteFile(o.InventoryFile, []byte(output), _const.PermFilePublic),
		"failed to write inventory file %s", o.InventoryFile)
}

// findNodesToRemove finds the line numbers of nodes that should be removed from groups
func (o *DeleteNodesOptions) findNodesToRemove(lines []string, deleteGroupHosts map[string][]string) []int {
	var linesToRemove []int
	currentGroup := ""
	inHosts := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check if we're entering a group
		for groupName := range deleteGroupHosts {
			if trimmed == groupName+":" {
				currentGroup = groupName
				inHosts = false
				break
			}
		}

		// Check if we're in the hosts section of a group
		if currentGroup != "" && strings.TrimSpace(trimmed) == "hosts:" {
			inHosts = true
			continue
		}

		// Check if this line contains a node to remove
		if inHosts && currentGroup != "" {
			if nodes, ok := deleteGroupHosts[currentGroup]; ok {
				// Check if this line is a list item with a node to remove
				for _, node := range nodes {
					// Match patterns like "- node1" or "        - node1"
					if strings.TrimSpace(trimmed) == "- "+node {
						linesToRemove = append(linesToRemove, i)
						break
					}
				}
			}
		}

		// If we encounter a new top-level key, reset the group context
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") && trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			if currentGroup != "" && !strings.HasPrefix(trimmed, currentGroup) {
				// Check if this is a new group or section
				isGroup := false
				for groupName := range deleteGroupHosts {
					if strings.HasPrefix(trimmed, groupName) {
						isGroup = true
						break
					}
				}
				if !isGroup && !strings.HasPrefix(trimmed, "hosts") {
					currentGroup = ""
					inHosts = false
				}
			}
		}
	}

	return linesToRemove
}

// ======================================================================================
//                                  delete registry
// ======================================================================================

// NewDeleteRegistryOptions creates a new DeleteRegistryOptions with default values
func NewDeleteRegistryOptions() *DeleteRegistryOptions {
	// set default value for DeleteImageRegistryOptions
	o := &DeleteRegistryOptions{
		CommonOptions: options.NewCommonOptions(),
	}
	// Set the function to get the inventory
	o.GetInventoryFunc = getInventory

	return o
}

// DeleteRegistryOptions contains options for deleting an image_registry created by kubekey
type DeleteRegistryOptions struct {
	options.CommonOptions
	// kubernetes version which the config will install.
	Kubernetes string
	DeleteData bool
}

// Flags returns the flag sets for DeleteImageRegistryOptions
func (o *DeleteRegistryOptions) Flags() cliflag.NamedFlagSets {
	fss := o.CommonOptions.Flags()
	kfs := fss.FlagSet("config")
	kfs.StringVar(&o.Kubernetes, "with-kubernetes", o.Kubernetes, fmt.Sprintf("Specify a supported version of kubernetes. default is %s", o.Kubernetes))
	kfs.BoolVar(&o.DeleteData, "with-data", o.DeleteData, "Also delete data directories (harbor data, registry data, etc.). Use with caution.")

	return fss
}

// Complete validates and completes the DeleteImageRegistryOptions configuration
func (o *DeleteRegistryOptions) Complete(cmd *cobra.Command, args []string) (*kkcorev1.Playbook, error) {
	// Initialize playbook metadata for deleting image registry
	playbook := &kkcorev1.Playbook{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "delete-imageregistry-",
			Namespace:    o.Namespace,
			Annotations: map[string]string{
				kkcorev1.BuiltinsProjectAnnotation: "",
			},
		},
	}

	// Validate playbook arguments: must have exactly one argument (the playbook)
	if len(args) != 1 {
		return nil, errors.Errorf("%s\nSee '%s -h' for help and examples", cmd.Use, cmd.CommandPath())
	}
	o.Playbook = args[0]

	// Set playbook specification
	playbook.Spec = kkcorev1.PlaybookSpec{
		Playbook: o.Playbook,
	}

	// Complete common options (e.g., config, inventory)
	if err := o.CommonOptions.Complete(playbook); err != nil {
		return nil, err
	}

	// Set delete data option if specified
	if o.DeleteData {
		if err := unstructured.SetNestedField(o.Config.Value(), true, "delete", "data"); err != nil {
			return nil, errors.Wrapf(err, "failed to set %q to config", "delete_data")
		}
	}

	// Complete config specific to delete image registry
	return playbook, nil
}
