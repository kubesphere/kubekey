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

package core

import (
	"fmt"
	"io/fs"
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// KubeUpgradePath maps a Kubernetes minor version (e.g. 1.23) to the highest
// recommended patch version (e.g. "v1.23.17") for that minor. kk consumes it to
// split a multi-minor upgrade into a sequence of single-minor steps (kubeadm only
// allows upgrading one minor at a time).
//
// The default mapping is a Go constant defined at the CLI layer
// (cmd/kk/app/options/builtin/upgrade.go) and seeded into the config under
// cluster_require.kube_upgrade_path by MergeKubeUpgradePathDefaults. It can be
// overridden per-minor via the config file or --set.
type KubeUpgradePath map[int]string

// MergeKubeUpgradePathDefaults seeds the supplied default minor->patch mapping
// into cfg's cluster_require.kube_upgrade_path for every minor the user has NOT
// already set (via config file or --set). After the merge the path is an ordinary
// config value — readable by Go and by playbook templates — and overridable
// per-minor via --set (e.g. `--set cluster_require.kube_upgrade_path.v1.24=v1.24.10`)
// or the config file. Precedence is preserved because only absent keys are filled:
// Go default < config file == --set.
func MergeKubeUpgradePathDefaults(cfg map[string]interface{}, def KubeUpgradePath) error {
	cur, _, err := unstructured.NestedStringMap(cfg, "cluster_require", "kube_upgrade_path")
	if err != nil {
		return fmt.Errorf("read cluster_require.kube_upgrade_path from config: %w", err)
	}
	merged := make(map[string]string, len(def)+len(cur))
	for k, v := range cur {
		merged[k] = v
	}
	for minor, patch := range def {
		key := fmt.Sprintf("v1.%d", minor)
		if _, ok := merged[key]; !ok {
			merged[key] = patch
		}
	}
	return unstructured.SetNestedStringMap(cfg, merged, "cluster_require", "kube_upgrade_path")
}

// EffectiveKubeUpgradePath returns the live upgrade path carried in the config's
// cluster_require.kube_upgrade_path, which MergeKubeUpgradePathDefaults has already
// seeded with the Go defaults and which a user may have overridden via config file
// or --set.
func EffectiveKubeUpgradePath(cfg map[string]interface{}) (KubeUpgradePath, error) {
	m, _, err := unstructured.NestedStringMap(cfg, "cluster_require", "kube_upgrade_path")
	if err != nil {
		return nil, fmt.Errorf("read cluster_require.kube_upgrade_path from config: %w", err)
	}
	path := make(KubeUpgradePath, len(m))
	for minorStr, patch := range m {
		minor, err := parseMinor(minorStr)
		if err != nil {
			return nil, fmt.Errorf("invalid minor key %q in kube_upgrade_path: %w", minorStr, err)
		}
		path[minor] = patch
	}
	return path, nil
}

// ComputeUpgradeSteps returns the ordered list of target versions the cluster
// must pass through to upgrade from currentMinor to targetMinor.
//
// Intermediate minors (currentMinor+1 .. targetMinor-1) each use the highest
// recommended patch from the upgrade path; the final target minor uses the
// caller-supplied requestedTarget so a specific patch can be pinned.
//
// It returns an error for downgrades, or when an intermediate minor is missing
// from the path (which would make auto-stepping unsafe).
func ComputeUpgradeSteps(currentMinor, targetMinor int, requestedTarget string, path KubeUpgradePath) ([]string, error) {
	if targetMinor < currentMinor {
		return nil, fmt.Errorf("downgrade from v1.%d to v1.%d is not supported", currentMinor, targetMinor)
	}
	if targetMinor == currentMinor {
		// Patch-only upgrade: go straight to the requested version.
		return []string{requestedTarget}, nil
	}
	steps := make([]string, 0, targetMinor-currentMinor)
	for m := currentMinor + 1; m < targetMinor; m++ {
		patch, ok := path[m]
		if !ok {
			return nil, fmt.Errorf("kube_upgrade_path has no entry for minor v1.%d; cannot auto-step", m)
		}
		steps = append(steps, patch)
	}
	steps = append(steps, requestedTarget)
	return steps, nil
}

// upgradeVersionFieldPaths lists the nested config paths that carry a per-minor
// "version info" value. They mirror the version pins defined in
// roles/defaults/vars/v1.{m}.yaml (helm/sandbox/etcd/cri/cni/dns/storage/registry
// versions). This is the ALLOW-LIST used to construct an explicit intermediate-hop
// version overlay from that embedded vars file: for each intermediate hop we read
// the value of every listed path from roles/defaults/vars/v1.{hop}.yaml and SET it
// into the per-hop config, overriding the terminal config. It is deliberately NOT
// a blind deep-merge of the whole vars file, because that file also carries
// non-version keys (e.g. cri.container_manager) which must keep coming from the
// terminal config.
//
// The FINAL hop (and any single-hop upgrade) keeps the terminal config's own
// component version fields untouched, honoring explicit user pins for the end state.
//
// Non-version fields (cri.container_manager, image_registry.type/ha_vip, zone, ssh,
// network, storage.enabled, inventory, custom_labels, ...) are intentionally absent:
// they always come from the terminal config.
var upgradeVersionFieldPaths = [][]string{
	{"kubernetes", "helm_version"},
	{"kubernetes", "sandbox_image", "tag"},
	{"kubernetes", "control_plane_endpoint", "kube_vip", "image", "tag"},
	{"kubernetes", "control_plane_endpoint", "haproxy", "image", "tag"},

	{"etcd", "etcd_version"},

	{"image_registry", "keepalived_version"},
	{"image_registry", "harbor_version"},
	{"image_registry", "dockercompose_version"},
	{"image_registry", "docker_registry_version"},

	{"cri", "crictl_version"},
	{"cri", "docker_version"},
	{"cri", "cridockerd_version"},
	{"cri", "containerd_version"},
	{"cri", "runc_version"},

	{"cni", "cni_plugins_version"},
	{"cni", "calico_version"},
	{"cni", "cilium_version"},
	{"cni", "flannel_version"},
	{"cni", "kubeovn_version"},
	{"cni", "hybridnet_version"},
	{"cni", "multus", "image", "tag"},
	{"cni", "spiderpool_version"},

	{"storage_class", "localpv_provisioner_version"},
	{"storage_class", "nfs_provisioner_version"},

	{"dns", "coredns", "image", "tag"},
	{"dns", "nodelocaldns", "image", "tag"},
}

// UpgradeVersionOverlay reads the per-minor version overlay for an intermediate
// minor (e.g. 28) from the embedded roles/defaults/vars/v1.{m}.yaml. Only the
// allow-listed version paths are picked out; non-version keys in that file are
// ignored. Reading the embedded vars file here makes the intermediate version
// values an explicit, inspectable part of the per-hop config (rather than an
// implicit runtime include_vars), which is the "deliberate intermediate version
// file overriding config.yaml" model.
func UpgradeVersionOverlay(vfs fs.FS, minor int) (map[string]interface{}, error) {
	data, err := fs.ReadFile(vfs, fmt.Sprintf("roles/defaults/vars/v1.%d.yaml", minor))
	if err != nil {
		return nil, fmt.Errorf("read embedded roles/defaults/vars/v1.%d.yaml: %w", minor, err)
	}
	var vars map[string]interface{}
	if err := yaml.Unmarshal(data, &vars); err != nil {
		return nil, fmt.Errorf("parse embedded roles/defaults/vars/v1.%d.yaml: %w", minor, err)
	}

	overlay := make(map[string]interface{})
	for _, path := range upgradeVersionFieldPaths {
		v, found, err := unstructured.NestedFieldNoCopy(vars, path...)
		if err != nil {
			return nil, fmt.Errorf("read version field %q from v1.%d.yaml: %w", strings.Join(path, "."), minor, err)
		}
		if !found {
			continue // field may be absent (e.g. commented out) in some minors
		}
		if err := unstructured.SetNestedField(overlay, v, path...); err != nil {
			return nil, fmt.Errorf("set overlay field %q for v1.%d: %w", strings.Join(path, "."), minor, err)
		}
	}
	return overlay, nil
}

// BuildUpgradeStepConfig builds a standalone config map for a single-minor upgrade
// step, derived from the terminal config.
//
// The terminal config is the end-state configuration. Every hop of the auto-step
// upgrade reuses its non-version fields (cri.container_manager, image_registry,
// zone, ssh, network, storage, ...) verbatim, because those describe the target
// cluster topology/behaviour, not a per-minor version.
//
// What differs per hop is the version info:
//   - kube_version is always forced to this hop's target version.
//   - For an intermediate hop (isFinal=false), the allow-listed version fields are
//     explicitly overlaid from roles/defaults/vars/v1.{hop}.yaml via
//     UpgradeVersionOverlay, so etcd/helm/sandbox/crictl/containerd/calico/dns take
//     the correct monotonically-increasing intermediate values (e.g. etcd
//     3.5.6 -> 3.6.5). This replaces the old buggy behaviour where reusing the
//     terminal config on the first hop jumped components straight to the terminal
//     value.
//   - For the final hop (or a single-hop upgrade, isFinal=true), the terminal
//     config's own version fields are kept unchanged, honoring any explicit
//     component versions the user pinned for the end state.
func BuildUpgradeStepConfig(vfs fs.FS, base map[string]interface{}, kubeVersion string, isFinal bool) (map[string]interface{}, error) {
	cfg := runtime.DeepCopyJSON(base)
	if err := unstructured.SetNestedField(cfg, kubeVersion, "kubernetes", "kube_version"); err != nil {
		return nil, fmt.Errorf("set step kube_version %q: %w", kubeVersion, err)
	}
	if isFinal {
		return cfg, nil
	}
	minor, err := MinorOfVersion(kubeVersion)
	if err != nil {
		return nil, fmt.Errorf("parse intermediate kube_version %q: %w", kubeVersion, err)
	}
	overlay, err := UpgradeVersionOverlay(vfs, minor)
	if err != nil {
		return nil, err
	}
	cfg = combineConfigMaps(cfg, overlay)
	return cfg, nil
}

// CompareVersionNumbers is the exported form of compareVersionNumbers, used by the
// test package. See compareVersionNumbers for semantics.
func CompareVersionNumbers(a, b string) int {
	return compareVersionNumbers(a, b)
}

// compareVersionNumbers orders two version strings by semver semantics, using the
// same Masterminds/semver library that backs the template pipeline's semverCompare
// function (pkg/converter/tmpl -> sprig.TxtFuncMap -> semverCompare). It accepts
// the heterogeneous per-minor version strings (leading "v", e.g. "v3.5.15";
// single/two-part forms like "3.6"; prerelease suffixes like "2.9.6-alpine") and
// returns -1, 0, or +1. If either side does not parse as semver, it falls back to
// plain string ordering so the monotonically-upgrade guard never silently passes a
// genuinely lower value.
func compareVersionNumbers(a, b string) int {
	va, aerr := semver.NewVersion(a)
	vb, berr := semver.NewVersion(b)
	if aerr == nil && berr == nil {
		return va.Compare(vb)
	}
	// Fallback: lexical ordering on unparseable values (alphabetical, so "v3" <
	// "v10" is slightly off, but such values should not appear in the version set).
	return strings.Compare(a, b)
}

// ValidateTerminalVersions checks that every per-minor version field the user
// explicitly pinned in the terminal config is not lower than the value required by
// any intermediate hop of an auto-step upgrade.
//
// Component upgrades (etcd, helm, sandbox, containerd, calico, ...) are
// monotonic: the auto-step path only ever advances them, and kubeadm/etcd upgrades
// refuse to downgrade. If a field pinned in the terminal config is lower than the
// value an intermediate hop already reached, the final hop would have to roll the
// component back — which fails. (This mirrors the concern that reusing the terminal
// config on intermediate hops caused the first-hop etcd over-jump.)
//
// steps is the ordered single-minor hop list from ComputeUpgradeSteps; the last
// element is the final/target version and is NOT compared (it is the terminal hop
// itself). Intermediate overlay values are read from the embedded
// roles/defaults/vars/v1.{m}.yaml. Only fields actually present in the terminal
// config are validated; an unpinned field falls back to include_vars defaults at
// every hop and is inherently monotonic.
func ValidateTerminalVersions(vfs fs.FS, base map[string]interface{}, steps []string) error {
	if len(steps) <= 1 {
		return nil // single hop / patch-only: no intermediate to conflict with
	}

	var problems []string
	for _, path := range upgradeVersionFieldPaths {
		termVal, found, err := unstructured.NestedString(base, path...)
		if err != nil {
			return fmt.Errorf("read terminal version field %q: %w", strings.Join(path, "."), err)
		}
		if !found || termVal == "" {
			continue // not pinned by the user -> no conflict possible
		}

		// Compare the pinned terminal value against every intermediate hop's overlay.
		for _, step := range steps[:len(steps)-1] {
			minor, err := MinorOfVersion(step)
			if err != nil {
				return fmt.Errorf("parse intermediate step %q: %w", step, err)
			}
			overlay, err := UpgradeVersionOverlay(vfs, minor)
			if err != nil {
				return err
			}
			hopVal, found, err := unstructured.NestedString(overlay, path...)
			if err != nil {
				return fmt.Errorf("read intermediate version field %q for v1.%d: %w", strings.Join(path, "."), minor, err)
			}
			if !found || hopVal == "" {
				continue
			}
			if compareVersionNumbers(termVal, hopVal) < 0 {
				problems = append(problems, fmt.Sprintf(
					"%s: terminal %q is lower than intermediate hop (v1.%d) %q",
					strings.Join(path, "."), termVal, minor, hopVal))
			}
		}
	}

	if len(problems) > 0 {
		return fmt.Errorf("terminal config pins component versions lower than intermediate hops (components cannot be downgraded):\n  - %s",
			strings.Join(problems, "\n  - "))
	}
	return nil
}

// combineConfigMaps deep-merges src over base: nested maps are merged recursively,
// leaf values from src win. Used to lay per-hop version overlay over the terminal
// config without disturbing its non-version fields.
func combineConfigMaps(base map[string]interface{}, src map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(base)+len(src))
	for k, v := range base {
		out[k] = v
	}
	for k, sv := range src {
		if bv, ok := out[k]; ok {
			if bm, bok := bv.(map[string]interface{}); bok {
				if sm, sok := sv.(map[string]interface{}); sok {
					out[k] = combineConfigMaps(bm, sm)
					continue
				}
			}
		}
		out[k] = sv
	}
	return out
}

// parseMinor extracts the minor component from a version string such as
// "v1.23", "1.23" or "v1.23.17". It returns the integer minor (e.g. 23).
func parseMinor(s string) (int, error) {
	s = strings.TrimPrefix(s, "v")
	parts := strings.SplitN(s, ".", 3)
	if len(parts) < 2 {
		return 0, fmt.Errorf("not a major.minor version: %q", s)
	}
	return strconv.Atoi(parts[1])
}

// MinorOfVersion extracts the minor component from a Kubernetes version string.
// The discovery API may return suffixes (e.g. "34+"), which are ignored.
func MinorOfVersion(s string) (int, error) {
	s = strings.TrimPrefix(s, "v")
	parts := strings.SplitN(s, ".", 3)
	if len(parts) < 2 {
		return 0, fmt.Errorf("not a major.minor version: %q", s)
	}
	minorStr := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, parts[1])
	if minorStr == "" {
		return 0, fmt.Errorf("invalid minor in %q", s)
	}
	return strconv.Atoi(minorStr)
}
