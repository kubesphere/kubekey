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
	//
	Etcd string
}

// Flags adds flags for configuring AddNodeOptions to the specified FlagSet
func (o *AddNodeOptions) Flags() cliflag.NamedFlagSets {
	fss := o.CommonOptions.Flags()
	kfs := fss.FlagSet("config")
	kfs.StringVar(&o.Kubernetes, "with-kubernetes", o.Kubernetes, fmt.Sprintf("Specify a supported version of kubernetes. default is %s", o.Kubernetes))
	kfs.StringVar(&o.ControlPlane, "kube_control_plane", o.ControlPlane, "Which nodes will be installed as control-plane. Multiple nodes are supported, separated by commas (e.g., node1, node2, ...)")
	kfs.StringVar(&o.Worker, "kube_worker", o.Worker, "Which nodes will be installed as workers. Multiple nodes are supported, separated by commas (e.g., node1, node2, ...)")
	kfs.StringVar(&o.Worker, "etcd", o.Worker, "Which nodes will be installed as workers. Multiple nodes are supported, separated by commas (e.g., node1, node2, ...)")

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

// updateInventoryFile updates the inventory file with new nodes while preserving comments
func (o *AddNodeOptions) updateInventoryFile(addGroupHosts map[string][]string, existingGroups map[string][]string) error {
	data, err := os.ReadFile(o.InventoryFile)
	if err != nil {
		return errors.Wrapf(err, "failed to read inventory file %s", o.InventoryFile)
	}

	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return errors.Wrap(err, "failed to unmarshal inventory file")
	}

	// Find spec.groups and update hosts
	if len(root.Content) > 0 {
		o.updateGroupsInNode(root.Content[0], addGroupHosts, existingGroups)
	}

	// Write back with preserved formatting
	out, err := yaml.Marshal(&root)
	if err != nil {
		return errors.Wrap(err, "failed to marshal updated inventory")
	}

	return errors.Wrapf(os.WriteFile(o.InventoryFile, out, _const.PermFilePublic), "failed to write inventory file %s", o.InventoryFile)
}

// updateGroupsInNode recursively finds and updates the groups section in YAML node
func (o *AddNodeOptions) updateGroupsInNode(node *yaml.Node, addGroupHosts map[string][]string, existingGroups map[string][]string) {
	if node.Kind != yaml.MappingNode {
		return
	}

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		if keyNode.Value == "spec" && valueNode.Kind == yaml.MappingNode {
			// Look for groups inside spec
			for j := 0; j < len(valueNode.Content); j += 2 {
				specKey := valueNode.Content[j]
				specValue := valueNode.Content[j+1]
				if specKey.Value == "groups" && specValue.Kind == yaml.MappingNode {
					o.updateGroupHosts(specValue, addGroupHosts, existingGroups)
					return
				}
			}
		}
	}
}

// updateGroupHosts updates the hosts list for each group in the YAML node
func (o *AddNodeOptions) updateGroupHosts(groupsNode *yaml.Node, addGroupHosts map[string][]string, existingGroups map[string][]string) {
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
				// Add new nodes that don't already exist
				existingNodes := existingGroups[groupName]
				for _, node := range nodesToAdd {
					if !slices.Contains(existingNodes, node) {
						newNode := &yaml.Node{
							Kind:  yaml.ScalarNode,
							Tag:   "!!str",
							Value: node,
						}
						hostValue.Content = append(hostValue.Content, newNode)
					}
				}
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

	addGroupHosts := make(map[string][]string)
	groups := variable.ConvertGroup(*o.Inventory)

	// add nodes to groups
	if err := o.addNodesToGroup(o.ControlPlane, defaultGroupControlPlane, groups, addGroupHosts); err != nil {
		return err
	}
	if err := o.addNodesToGroup(o.Worker, defaultGroupWorker, groups, addGroupHosts); err != nil {
		return err
	}
	if err := o.addNodesToGroup(o.Etcd, defaultGroupEtcd, groups, addGroupHosts); err != nil {
		return err
	}

	if o.InventoryFile != "" && len(addGroupHosts) > 0 {
		if err := o.updateInventoryFile(addGroupHosts, groups); err != nil {
			return err
		}
	}

	// collect unique addNodes
	addNodesSet := make(map[string]struct{})
	for _, nodes := range addGroupHosts {
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
