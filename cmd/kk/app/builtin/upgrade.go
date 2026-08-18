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
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"github.com/kubesphere/kubekey/v4/builtin/core"
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

			// Execute the playbook to upgrade the cluster. When the requested target
			// spans multiple minor versions, kk auto-steps the upgrade one minor at a
			// time (kubeadm only allows single-minor upgrades).
			return runClusterUpgrade(cmd.Context(), o, playbook)
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

// runClusterUpgrade drives the cluster upgrade. For a multi-minor target it
// splits the upgrade into single-minor steps (using the embedded
// kube_upgrade_path) and runs the upgrade playbook once per step. For an
// adjacent or patch-only target it performs a single run, matching the legacy
// behavior.
func runClusterUpgrade(ctx context.Context, o *builtin.UpgradeClusterOptions, base *kkcorev1.Playbook) error {
	targetVer, _, err := unstructured.NestedString(o.Config.Value(), "kubernetes", "kube_version")
	if err != nil {
		return errors.Wrap(err, "read target kubernetes version")
	}
	if targetVer == "" {
		return errors.New("target kubernetes version is empty; set --with-kubernetes or kubernetes.kube_version in config")
	}
	targetMinor, err := core.MinorOfVersion(targetVer)
	if err != nil {
		return err
	}

	// Detect the currently installed cluster minor. If detection fails (e.g. the
	// cluster is unreachable), fall back to the legacy single-run behavior so we
	// do not regress existing upgrade flows that rely on the playbook precheck.
	currentMinor, derr := detectCurrentClusterMinor()
	if derr != nil {
		fmt.Fprintf(os.Stderr, "warning: could not detect current cluster version (%v); performing a single upgrade run\n", derr)
		return o.Run(ctx, base)
	}
	if targetMinor <= currentMinor {
		// Patch-only (same minor) or already-at/above target: single run.
		return o.Run(ctx, base)
	}

	path, err := core.ParseKubeUpgradePath()
	if err != nil {
		return err
	}
	steps, err := core.ComputeUpgradeSteps(currentMinor, targetMinor, targetVer, path)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "kk will auto-step the cluster upgrade across %d minor versions: %s\n", len(steps), strings.Join(steps, " -> "))
	for i, step := range steps {
		if err := unstructured.SetNestedField(o.Config.Value(), step, "kubernetes", "kube_version"); err != nil {
			return errors.Wrapf(err, "set step version %q", step)
		}
		// Each step runs the upgrade playbook as a standalone Playbook object.
		// The variable pipeline (pkg/variable) reads the config from
		// playbook.Spec.Config (a value snapshot), NOT the live o.Config pointer.
		// We must therefore carry the per-step config (which has kube_version set
		// to the current step's target above) and the inventory reference onto the
		// per-step Playbook, otherwise the template context would be missing
		// kubernetes.kube_version and the defaults role would skip loading the
		// per-minor version pins (e.g. calico_version), failing CNI prechecks.
		pb := &kkcorev1.Playbook{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "upgrade-cluster-",
				Namespace:    o.Namespace,
				Annotations:  map[string]string{kkcorev1.BuiltinsProjectAnnotation: ""},
			},
			Spec: kkcorev1.PlaybookSpec{
				Playbook:     "playbooks/upgrade_cluster.yaml",
				Config:       *o.Config,
				InventoryRef: base.Spec.InventoryRef,
			},
		}
		fmt.Fprintf(os.Stderr, "[%d/%d] upgrading cluster to %s ...\n", i+1, len(steps), step)
		if err := o.Run(ctx, pb); err != nil {
			return errors.Wrapf(err, "upgrade step %d/%d (-> %s) failed", i+1, len(steps), step)
		}
	}
	return nil
}

// detectCurrentClusterMinor resolves the currently installed Kubernetes minor
// version from the target cluster, using the standard kubeconfig loading rules
// (KUBECONFIG env, then ~/.kube/config). It returns an error if the cluster is
// unreachable or the version cannot be parsed.
func detectCurrentClusterMinor() (int, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	cfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})
	restCfg, err := cfg.ClientConfig()
	if err != nil {
		return 0, errors.Wrap(err, "build kubeconfig client config")
	}
	cs, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return 0, errors.Wrap(err, "build kubernetes clientset")
	}
	vi, err := cs.Discovery().ServerVersion()
	if err != nil {
		return 0, errors.Wrap(err, "discover server version")
	}
	// client-go reports the minor as a bare number string (e.g. "23"), and
	// recent servers may append a "+" (e.g. "34+"). Parse it directly rather
	// than via MinorOfVersion, which expects a "v1.23" style string.
	minorStr := strings.TrimSuffix(vi.Minor, "+")
	minor, err := strconv.Atoi(minorStr)
	if err != nil {
		return 0, errors.Wrapf(err, "parse server minor %q", vi.Minor)
	}
	return minor, nil
}
