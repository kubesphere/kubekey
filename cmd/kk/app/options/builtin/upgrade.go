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
	"github.com/cockroachdb/errors"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	cliflag "k8s.io/component-base/cli/flag"

	"github.com/kubesphere/kubekey/v4/cmd/kk/app/options"
)

// ======================================================================================
//                                  upgrade cluster
// ======================================================================================

// NewUpgradeClusterOptions creates a new UpgradeClusterOptions with default values
func NewUpgradeClusterOptions() *UpgradeClusterOptions {
	// set default value for UpgradeClusterOptions
	o := &UpgradeClusterOptions{
		CommonOptions: options.NewCommonOptions(),
		Kubernetes:    "",
	}
	// Set the function to get the inventory
	o.GetInventoryFunc = getInventory

	return o
}

// UpgradeClusterOptions contains options for upgrading a Kubernetes cluster
type UpgradeClusterOptions struct {
	options.CommonOptions
	// Kubernetes version which the cluster will upgrade to.
	Kubernetes string
	// UpgradeAllComponents indicates whether to upgrade all related components (etcd, cni, cri, helm, etc.)
	// If false, only kubelet will be upgraded.
	UpgradeAllComponents bool
}

// Flags returns the flag sets for UpgradeClusterOptions
func (o *UpgradeClusterOptions) Flags() cliflag.NamedFlagSets {
	fss := o.CommonOptions.Flags()
	kfs := fss.FlagSet("config")
	// Add a flag for specifying the target Kubernetes version
	kfs.StringVar(&o.Kubernetes, "with-kubernetes", o.Kubernetes, "Specify the target version of kubernetes to upgrade to. If not set, the version from config will be used.")
	kfs.BoolVar(&o.UpgradeAllComponents, "all", o.UpgradeAllComponents, "Upgrade all related components, including etcd, cni, cri, helm, etc. If not set, only kubelet will be upgraded.")

	return fss
}

// Complete validates and completes the UpgradeClusterOptions configuration
func (o *UpgradeClusterOptions) Complete(cmd *cobra.Command, args []string) (*kkcorev1.Playbook, error) {
	// Initialize playbook metadata for upgrading a cluster
	playbook := &kkcorev1.Playbook{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "upgrade-cluster-",
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

	// Complete config specific to upgrade cluster
	return playbook, o.completeConfig()
}

// completeConfig updates the configuration with upgrade settings
func (o *UpgradeClusterOptions) completeConfig() error {
	// If with-kubernetes is specified, set kube_version in config
	if o.Kubernetes != "" {
		if err := unstructured.SetNestedField(o.Config.Value(), o.Kubernetes, "kubernetes", "kube_version"); err != nil {
			return errors.Wrapf(err, "failed to set %q to config", "kubernetes.kube_version")
		}
	}

	// Set upgrade.all flag in config
	if err := unstructured.SetNestedField(o.Config.Value(), o.UpgradeAllComponents, "upgrade", "all"); err != nil {
		return errors.Wrapf(err, "failed to set %q to config", "upgrade.all")
	}

	// Set individual component upgrade flags.
	// If --all is set, all components are marked for upgrade.
	// If not set, components default to false (only kubelet/kubeadm will be upgraded).
	// Users can still override individual components via --set (e.g., --set upgrade.cri=true).
	upgradeComponents := []string{"etcd", "cri", "cni", "storage_class", "dns", "image_registry", "nfs"}
	for _, comp := range upgradeComponents {
		if _, ok, _ := unstructured.NestedFieldNoCopy(o.Config.Value(), "upgrade", comp); !ok {
			if err := unstructured.SetNestedField(o.Config.Value(), o.UpgradeAllComponents, "upgrade", comp); err != nil {
				return errors.Wrapf(err, "failed to set %q to config", "upgrade."+comp)
			}
		}
	}

	if o.Artifact != "" { // change default value to false
		if _, ok, _ := unstructured.NestedFieldNoCopy(o.Config.Value(), "download", "fetch"); !ok {
			if err := unstructured.SetNestedField(o.Config.Value(), false, "download", "fetch"); err != nil {
				return errors.Wrapf(err, "failed to set %q to config", "download.fetch")
			}
		}
	}

	return nil
}
