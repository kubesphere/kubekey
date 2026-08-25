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
	"github.com/kubesphere/kubekey/v4/builtin/core"
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

// defaultKubeUpgradePath maps each Kubernetes minor to the highest recommended
// patch version used as the intermediate-hop target during multi-minor auto-step
// upgrades (kubeadm allows only a single-minor bump at a time). It is the Go-side
// default (replacing the former defaults/main YAML), seeded into the config under
// cluster_require.kube_upgrade_path and overridable per-minor via the config file
// or --set (e.g. `--set cluster_require.kube_upgrade_path.v1.24=v1.24.10`).
var defaultKubeUpgradePath = core.KubeUpgradePath{
	23: "v1.23.17",
	24: "v1.24.17",
	25: "v1.25.16",
	26: "v1.26.15",
	27: "v1.27.16",
	28: "v1.28.15",
	29: "v1.29.15",
	30: "v1.30.14",
	31: "v1.31.14",
	32: "v1.32.13",
	33: "v1.33.7",
	34: "v1.34.3",
}

// UpgradeClusterOptions contains options for upgrading a Kubernetes cluster
type UpgradeClusterOptions struct {
	options.CommonOptions
	// Kubernetes version which the cluster will upgrade to.
	Kubernetes string
	// UpgradeAllComponents indicates whether to upgrade all related components (etcd, cni, cri, storage_class).
	// If false, only kubelet/kubeadm will be upgraded (unless individual components are enabled via --set).
	UpgradeAllComponents bool
	// OnlyComponent, when non-empty, restricts the upgrade to a single component
	// (e.g. "etcd"). It is preset by the top-level `kk upgrade <component>` subcommands
	// and must NOT be set together with a regular `kk upgrade cluster` invocation.
	OnlyComponent string
}

// Flags returns the flag sets for UpgradeClusterOptions
func (o *UpgradeClusterOptions) Flags() cliflag.NamedFlagSets {
	fss := o.CommonOptions.Flags()
	kfs := fss.FlagSet("config")
	// Add a flag for specifying the target Kubernetes version
	kfs.StringVar(&o.Kubernetes, "with-kubernetes", o.Kubernetes, "Specify the target version of kubernetes to upgrade to. If not set, the version from config will be used.")
	kfs.BoolVar(&o.UpgradeAllComponents, "all", o.UpgradeAllComponents, "Upgrade all related components, including etcd, cni, cri and storage_class. If not set, only kubelet/kubeadm will be upgraded (unless individual components are enabled via --set).")

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
	// Seed the Go-defined kube_upgrade_path defaults into the config so the path
	// is a normal, config-readable value that a user can override per-minor via
	// config file or --set. Only absent keys are filled, so any explicit override
	// (config file / --set already applied by CommonOptions.Complete) wins over the
	// Go default.
	if err := core.MergeKubeUpgradePathDefaults(o.Config.Value(), defaultKubeUpgradePath); err != nil {
		return errors.Wrap(err, "merge kube_upgrade_path defaults into config")
	}

	// If with-kubernetes is specified, set kube_version in config
	if o.Kubernetes != "" {
		if err := unstructured.SetNestedField(o.Config.Value(), o.Kubernetes, "kubernetes", "kube_version"); err != nil {
			return errors.Wrapf(err, "failed to set %q to config", "kubernetes.kube_version")
		}
	}

	// Components that have a wired upgrade path in upgrade_cluster.yaml.
	// dns / image_registry / nfs are intentionally excluded: their upgrade roles
	// are not implemented yet, so exposing them would create dead switches.
	upgradeComponents := []string{"etcd", "cri", "cni", "storage_class"}

	if o.OnlyComponent != "" {
		// Standalone component upgrade (e.g. `kk upgrade etcd`): upgrade ONLY that
		// component and leave the Kubernetes control plane/worker nodes untouched.
		if err := unstructured.SetNestedField(o.Config.Value(), false, "upgrade", "kubernetes"); err != nil {
			return errors.Wrapf(err, "failed to set %q to config", "upgrade.kubernetes")
		}
		if err := unstructured.SetNestedField(o.Config.Value(), true, "upgrade", o.OnlyComponent); err != nil {
			return errors.Wrapf(err, "failed to set %q to config", "upgrade."+o.OnlyComponent)
		}
		// Other components keep their values (default false from the defaults role,
		// or explicitly overridden via --set).
	} else {
		// `kk upgrade cluster` (optionally with --all / --set): Kubernetes is always
		// the base of a cluster upgrade.
		if err := unstructured.SetNestedField(o.Config.Value(), true, "upgrade", "kubernetes"); err != nil {
			return errors.Wrapf(err, "failed to set %q to config", "upgrade.kubernetes")
		}

		if o.UpgradeAllComponents {
			// --all: upgrade every wired component.
			for _, comp := range upgradeComponents {
				if err := unstructured.SetNestedField(o.Config.Value(), true, "upgrade", comp); err != nil {
					return errors.Wrapf(err, "failed to set %q to config", "upgrade."+comp)
				}
			}
		} else {
			// Preserve values injected via --set (e.g. --set upgrade.etcd=true):
			// only fill a default when the component key is absent from the config.
			for _, comp := range upgradeComponents {
				if _, ok, _ := unstructured.NestedFieldNoCopy(o.Config.Value(), "upgrade", comp); !ok {
					if err := unstructured.SetNestedField(o.Config.Value(), false, "upgrade", comp); err != nil {
						return errors.Wrapf(err, "failed to set %q to config", "upgrade."+comp)
					}
				}
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
