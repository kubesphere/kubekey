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

// NewAddNodeOptions creates a new AddNodeOptions with default values
func NewAddNodeOptions() *AddNodeOptions {
	// set default value
	o := &AddNodeOptions{
		CommonOptions: options.NewCommonOptions(),
		Kubernetes:    defaultKubeVersion,
	}
	o.CommonOptions.GetInventoryFunc = getInventory

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
}

// Flags adds flags for configuring AddNodeOptions to the specified FlagSet
func (o *AddNodeOptions) Flags() cliflag.NamedFlagSets {
	fss := o.CommonOptions.Flags()
	kfs := fss.FlagSet("config")
	kfs.StringVar(&o.Kubernetes, "with-kubernetes", o.Kubernetes, fmt.Sprintf("Specify a supported version of kubernetes. default is %s", o.Kubernetes))
	kfs.StringVar(&o.ControlPlane, "control-plane", o.ControlPlane, "Which nodes will be installed as control-plane. Multiple nodes are supported, separated by commas (e.g., node1, node2, ...)")
	kfs.StringVar(&o.Worker, "worker", o.Worker, "Which nodes will be installed as workers. Multiple nodes are supported, separated by commas (e.g., node1, node2, ...)")

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

// complete updates the configuration with container manager and kubernetes version settings
func (o *AddNodeOptions) complete() error {
	if _, ok, _ := unstructured.NestedFieldNoCopy(o.CommonOptions.Config.Value(), "kube_version"); !ok {
		if err := unstructured.SetNestedField(o.CommonOptions.Config.Value(), o.Kubernetes, "kube_version"); err != nil {
			return errors.Wrapf(err, "failed to set %q to config", "kube_version")
		}
	}

	var addNodes []string
	groups := variable.ConvertGroup(*o.Inventory)
	// add nodes to control_plane group
	if o.ControlPlane != "" {
		for _, node := range strings.Split(o.ControlPlane, ",") {
			if !slices.Contains(groups[_const.VariableGroupsAll], node) {
				return errors.Errorf("%q is not defined in inventory.", node)
			}
			if !slices.Contains(groups[defaultGroupControlPlane], node) {
				group := o.Inventory.Spec.Groups[defaultGroupControlPlane]
				group.Hosts = append(group.Hosts, node)
				o.Inventory.Spec.Groups[defaultGroupControlPlane] = group
			}
			addNodes = append(addNodes, node)
		}
	}
	// add nodes to worker group
	if o.Worker != "" {
		for _, node := range strings.Split(o.ControlPlane, ",") {
			if !slices.Contains(groups[_const.VariableGroupsAll], node) {
				return errors.Errorf("%q is not defined in inventory.", node)
			}
			if !slices.Contains(groups[defaultGroupWorker], node) {
				group := o.Inventory.Spec.Groups[defaultGroupWorker]
				group.Hosts = append(group.Hosts, node)
				o.Inventory.Spec.Groups[defaultGroupControlPlane] = group
			}
			addNodes = append(addNodes, node)
		}
	}
	if err := unstructured.SetNestedStringSlice(o.CommonOptions.Config.Value(), addNodes, "add_nodes"); err != nil {
		return errors.Wrapf(err, "failed to set %q to config", "add_nodes")
	}

	return nil
}
