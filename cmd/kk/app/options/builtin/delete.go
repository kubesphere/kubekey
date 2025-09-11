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

	"github.com/cockroachdb/errors"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	cliflag "k8s.io/component-base/cli/flag"

	"github.com/kubesphere/kubekey/v4/cmd/kk/app/options"
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
	o.CommonOptions.GetInventoryFunc = getInventory

	return o
}

// DeleteClusterOptions contains options for deleting a Kubernetes cluster
type DeleteClusterOptions struct {
	options.CommonOptions
	// Kubernetes version which the cluster will install.
	Kubernetes string
}

// Flags returns the flag sets for DeleteClusterOptions
func (o *DeleteClusterOptions) Flags() cliflag.NamedFlagSets {
	fss := o.CommonOptions.Flags()
	kfs := fss.FlagSet("config")
	// Add a flag for specifying the Kubernetes version
	kfs.StringVar(&o.Kubernetes, "with-kubernetes", o.Kubernetes, fmt.Sprintf("Specify a supported version of kubernetes. default is %s", o.Kubernetes))

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
	if _, ok, _ := unstructured.NestedFieldNoCopy(o.CommonOptions.Config.Value(), "kubernetes", "kube_version"); !ok {
		if err := unstructured.SetNestedField(o.CommonOptions.Config.Value(), o.Kubernetes, "kubernetes", "kube_version"); err != nil {
			return errors.Wrapf(err, "failed to set %q to config", "kube_version")
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
	o.CommonOptions.GetInventoryFunc = getInventory

	return o
}

// DeleteNodesOptions contains options for deleting Kubernetes cluster nodes
type DeleteNodesOptions struct {
	options.CommonOptions
	// Kubernetes version which the cluster will install.
	Kubernetes string
}

// Flags returns the flag sets for DeleteNodesOptions
func (o *DeleteNodesOptions) Flags() cliflag.NamedFlagSets {
	fss := o.CommonOptions.Flags()
	kfs := fss.FlagSet("config")
	// Add a flag for specifying the Kubernetes version
	kfs.StringVar(&o.Kubernetes, "with-kubernetes", o.Kubernetes, fmt.Sprintf("Specify a supported version of kubernetes. default is %s", o.Kubernetes))

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
	// If kube_version is not set in config, set it to the specified Kubernetes version
	if _, ok, _ := unstructured.NestedFieldNoCopy(o.CommonOptions.Config.Value(), "kubernetes", "kube_version"); !ok {
		if err := unstructured.SetNestedField(o.CommonOptions.Config.Value(), o.Kubernetes, "kubernetes", "kube_version"); err != nil {
			return errors.Wrapf(err, "failed to set %q to config", "kube_version")
		}
	}
	// Set the list of nodes to be deleted in the config
	if err := unstructured.SetNestedStringSlice(o.CommonOptions.Config.Value(), nodes, "delete_nodes"); err != nil {
		return errors.Wrapf(err, "failed to set %q to config", "delete_nodes")
	}

	return nil
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
	o.CommonOptions.GetInventoryFunc = getInventory

	return o
}

// DeleteRegistryOptions contains options for deleting an image_registry created by kubekey
type DeleteRegistryOptions struct {
	options.CommonOptions
	// kubernetes version which the config will install.
	Kubernetes string
}

// Flags returns the flag sets for DeleteImageRegistryOptions
func (o *DeleteRegistryOptions) Flags() cliflag.NamedFlagSets {
	fss := o.CommonOptions.Flags()
	kfs := fss.FlagSet("config")
	kfs.StringVar(&o.Kubernetes, "with-kubernetes", o.Kubernetes, fmt.Sprintf("Specify a supported version of kubernetes. default is %s", o.Kubernetes))

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

	// Complete config specific to delete image registry
	return playbook, nil
}
