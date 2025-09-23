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
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	cliflag "k8s.io/component-base/cli/flag"

	"github.com/kubesphere/kubekey/v4/cmd/kk/app/options"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

// NewAddNodeOptions creates a new AddNodeOptions with default values
func NewAddNodeOptions() *AddNodeOptions {
	// set default value
	o := &AddNodeOptions{
		CommonOptions: options.NewCommonOptions(),
		Kubernetes:    defaultKubeVersion,
	}
	o.GetInventoryFunc = getInventory

	return o
}

// AddNodeOptions contains options for adding nodes to a cluster
type AddNodeOptions struct {
	options.CommonOptions
	// kubernetes version which the cluster will install.
	Kubernetes string
	// ControlPlane nodes which will be added.
	ControlPlane string
	// Worker nodes which will to be added.
	Worker string
	// Etcd nodes which will be added.
	Etcd string
	// Override indicates whether to override the inventory file after successful execution.
	// When set to true, the inventory.yaml file will be updated.
	Override bool
	// addGroupHosts stores the nodes to be added to each group for later inventory update
	addGroupHosts map[string][]string
}

// Flags adds flags for configuring AddNodeOptions to the specified FlagSet
func (o *AddNodeOptions) Flags() cliflag.NamedFlagSets {
	fss := o.CommonOptions.Flags()
	kfs := fss.FlagSet("config")
	kfs.StringVar(&o.Kubernetes, "with-kubernetes", o.Kubernetes, fmt.Sprintf("Specify a supported version of kubernetes. default is %s", o.Kubernetes))
	kfs.StringVar(&o.ControlPlane, "control-plane", o.ControlPlane, "Which nodes will be installed as control-plane. Multiple nodes are supported, separated by commas (e.g., node1, node2, ...)")
	kfs.StringVar(&o.Worker, "worker", o.Worker, "Which nodes will be installed as workers. Multiple nodes are supported, separated by commas (e.g., node1, node2, ...)")
	kfs.StringVar(&o.Etcd, "etcd", o.Etcd, "Which nodes will be installed as etcd. Multiple nodes are supported, separated by commas (e.g., node1, node2, ...)")
	kfs.BoolVar(&o.Override, "override", o.Override, "Override the inventory file after successful execution")

	return fss
}

// Complete validates and completes the AddNodeOptions configuration.
// It creates and returns a Playbook object based on the options.
func (o *AddNodeOptions) Complete(cmd *cobra.Command, args []string) (*kkcorev1.Playbook, error) {
	playbook := &kkcorev1.Playbook{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "add-nodes-",
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
	}
	// override kube_version in config
	if err := o.CommonOptions.Complete(playbook); err != nil {
		return nil, err
	}

	return playbook, o.complete()
}

// addNodesToGroup adds nodes to a specific inventory group and returns the added nodes
func (o *AddNodeOptions) addNodesToGroup(nodeList string, groupName string, groups map[string][]string, addGroupHosts map[string][]string) error {
	if nodeList == "" {
		return nil
	}

	var nodes []string
	for _, node := range strings.Split(nodeList, ",") {
		if !slices.Contains(groups[_const.VariableGroupsAll], node) {
			return errors.Errorf("%q is not defined in inventory.", node)
		}
		if !slices.Contains(groups[groupName], node) {
			group := o.Inventory.Spec.Groups[groupName]
			group.Hosts = append(group.Hosts, node)
			o.Inventory.Spec.Groups[groupName] = group
		}
		nodes = append(nodes, node)
	}
	if len(nodes) > 0 {
		addGroupHosts[groupName] = nodes
	}
	return nil
}

// updateInventoryFile updates the inventory file with new nodes while preserving comments and formatting
func (o *AddNodeOptions) updateInventoryFile(addGroupHosts map[string][]string, existingGroups map[string][]string) error {
	// Read original file content
	content, err := os.ReadFile(o.InventoryFile)
	if err != nil {
		return errors.Wrapf(err, "failed to read inventory file %s", o.InventoryFile)
	}
	lines := strings.Split(string(content), "\n")

	// Parse to get line numbers for each hosts list
	var root yaml.Node
	if err := yaml.Unmarshal(content, &root); err != nil {
		return errors.Wrap(err, "failed to unmarshal inventory file")
	}

	// Find insert positions for each group
	insertions := o.findInsertPositions(&root, addGroupHosts, lines)
	if len(insertions) == 0 {
		return nil
	}

	// Sort insertions by line number in descending order to insert from bottom to top
	sortInsertions(insertions)

	// Insert new lines
	for _, ins := range insertions {
		newLines := make([]string, len(ins.nodes))
		for i, node := range ins.nodes {
			newLines[i] = ins.indent + "- " + node
		}
		// Insert after the line
		pos := ins.lineNum // Convert to 0-based index
		if pos < len(lines) {
			lines = append(lines[:pos+1], append(newLines, lines[pos+1:]...)...)
		} else {
			lines = append(lines, newLines...)
		}
	}

	// Write back
	output := strings.Join(lines, "\n")
	return errors.Wrapf(os.WriteFile(o.InventoryFile, []byte(output), _const.PermFilePublic),
		"failed to write inventory file %s", o.InventoryFile)
}

// insertion holds information about where to insert new nodes
type insertion struct {
	groupName string
	lineNum   int // 0-based line number where to insert (after this line)
	indent    string
	nodes     []string
}

// findInsertPositions finds the line numbers where new nodes should be inserted
func (o *AddNodeOptions) findInsertPositions(root *yaml.Node, addGroupHosts map[string][]string, lines []string) []insertion {
	var insertions []insertion
	if len(root.Content) == 0 {
		return insertions
	}
	o.findInsertionsInNode(root.Content[0], addGroupHosts, lines, &insertions)
	return insertions
}

// findInsertionsInNode recursively searches for groups and their hosts
func (o *AddNodeOptions) findInsertionsInNode(node *yaml.Node, addGroupHosts map[string][]string, lines []string, insertions *[]insertion) {
	if node.Kind != yaml.MappingNode {
		return
	}

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		if keyNode.Value == "spec" && valueNode.Kind == yaml.MappingNode {
			for j := 0; j < len(valueNode.Content); j += 2 {
				specKey := valueNode.Content[j]
				specValue := valueNode.Content[j+1]
				if specKey.Value == "groups" && specValue.Kind == yaml.MappingNode {
					o.findHostsInGroups(specValue, addGroupHosts, lines, insertions)
					return
				}
			}
		}
	}
}

// getIndentFromLine extracts the leading whitespace from a line
func getIndentFromLine(line string) string {
	for i, ch := range line {
		if ch != ' ' && ch != '\t' {
			return line[:i]
		}
	}
	return line
}

// findHostsInGroups finds hosts lists within groups
func (o *AddNodeOptions) findHostsInGroups(groupsNode *yaml.Node, addGroupHosts map[string][]string, lines []string, insertions *[]insertion) {
	for i := 0; i < len(groupsNode.Content); i += 2 {
		groupKey := groupsNode.Content[i]
		groupValue := groupsNode.Content[i+1]

		groupName := groupKey.Value
		nodesToAdd, ok := addGroupHosts[groupName]
		if !ok || len(nodesToAdd) == 0 {
			continue
		}

		if groupValue.Kind != yaml.MappingNode {
			continue
		}

		// Find hosts key in group
		for j := 0; j < len(groupValue.Content); j += 2 {
			hostKey := groupValue.Content[j]
			hostValue := groupValue.Content[j+1]

			if hostKey.Value == "hosts" && hostValue.Kind == yaml.SequenceNode {
				// Read existing hosts from file content
				existingNodes := make(map[string]struct{})
				for _, node := range hostValue.Content {
					if node.Kind == yaml.ScalarNode {
						existingNodes[node.Value] = struct{}{}
					}
				}

				var newNodes []string
				for _, node := range nodesToAdd {
					if _, exists := existingNodes[node]; !exists {
						newNodes = append(newNodes, node)
					}
				}

				if len(newNodes) > 0 {
					var lineIdx int
					var indent string

					if len(hostValue.Content) > 0 {
						// Get the last node in the hosts list
						lastNode := hostValue.Content[len(hostValue.Content)-1]
						lineIdx = lastNode.Line - 1 // Convert to 0-based
					} else {
						// Empty hosts list, insert after hosts: line
						lineIdx = hostKey.Line - 1 // Convert to 0-based
					}

					// Get indent from the actual line in the file
					if lineIdx >= 0 && lineIdx < len(lines) {
						baseIndent := getIndentFromLine(lines[lineIdx])
						if len(hostValue.Content) > 0 {
							indent = baseIndent // Use same indent as existing items
						} else {
							// Add extra indentation for new list items (2 more spaces)
							indent = baseIndent + "  "
						}
					} else {
						indent = "        " // fallback: 8 spaces
					}

					*insertions = append(*insertions, insertion{
						groupName: groupName,
						lineNum:   lineIdx,
						indent:    indent,
						nodes:     newNodes,
					})
				}
			}
		}
	}
}

// sortInsertions sorts insertions by line number in descending order
func sortInsertions(insertions []insertion) {
	for i := 0; i < len(insertions)-1; i++ {
		for j := i + 1; j < len(insertions); j++ {
			if insertions[i].lineNum < insertions[j].lineNum {
				insertions[i], insertions[j] = insertions[j], insertions[i]
			}
		}
	}
}

// complete updates the configuration with container manager and kubernetes version settings
func (o *AddNodeOptions) complete() error {
	if _, ok, _ := unstructured.NestedFieldNoCopy(o.Config.Value(), "kubernetes", "kube_version"); !ok {
		if err := unstructured.SetNestedField(o.Config.Value(), o.Kubernetes, "kubernetes", "kube_version"); err != nil {
			return errors.Wrapf(err, "failed to set %q to config", "kube_version")
		}
	}

	o.addGroupHosts = make(map[string][]string)
	groups := variable.ConvertGroup(*o.Inventory)

	// add nodes to groups
	if err := o.addNodesToGroup(o.ControlPlane, defaultGroupControlPlane, groups, o.addGroupHosts); err != nil {
		return err
	}
	if err := o.addNodesToGroup(o.Worker, defaultGroupWorker, groups, o.addGroupHosts); err != nil {
		return err
	}
	if err := o.addNodesToGroup(o.Etcd, defaultGroupEtcd, groups, o.addGroupHosts); err != nil {
		return err
	}

	// collect unique addNodes
	addNodesSet := make(map[string]struct{})
	for _, nodes := range o.addGroupHosts {
		for _, node := range nodes {
			addNodesSet[node] = struct{}{}
		}
	}
	addNodes := make([]string, 0, len(addNodesSet))
	for node := range addNodesSet {
		addNodes = append(addNodes, node)
	}

	if err := unstructured.SetNestedStringSlice(o.Config.Value(), addNodes, "add_nodes"); err != nil {
		return errors.Wrapf(err, "failed to set %q to config", "add_nodes")
	}

	return nil
}

// OverrideInventory updates the inventory.yaml file after successful execution.
// This should be called only when the Run method succeeds and Override flag is set.
func (o *AddNodeOptions) OverrideInventory() error {
	// Only update inventory file when --override flag is set
	if !o.Override || o.InventoryFile == "" || len(o.addGroupHosts) == 0 {
		return nil
	}

	groups := variable.ConvertGroup(*o.Inventory)
	return o.updateInventoryFile(o.addGroupHosts, groups)
}
