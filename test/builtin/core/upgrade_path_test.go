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

// Package core_test holds L2 logic tests for the builtin/core upgrade-path
// machinery. upgrade_path_test.go validates the Go-defined kube_upgrade_path
// default (seeded into the config by MergeKubeUpgradePathDefaults) and the
// auto-step splitting logic via the package's exported API.
package core_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	core "github.com/kubesphere/kubekey/v4/builtin/core"
)

// defaultPath mirrors the Go CLI default (cmd/kk/app/options/builtin/upgrade.go ->
// defaultKubeUpgradePath) used as the seed by MergeKubeUpgradePathDefaults.
var defaultPath = core.KubeUpgradePath{
	23: "v1.23.17", 24: "v1.24.17", 25: "v1.25.16", 26: "v1.26.15",
	27: "v1.27.16", 28: "v1.28.15", 29: "v1.29.15", 30: "v1.30.14",
	31: "v1.31.14", 32: "v1.32.13", 33: "v1.33.13", 34: "v1.34.11",
	35: "v1.35.8", 36: "v1.36.4", 37: "v1.37.0",
}

// TestMergeAndEffectiveKubeUpgradePath locks in the "Go default into config"
// contract: MergeKubeUpgradePathDefaults seeds the full default path into an
// otherwise-empty config (so Go and templates read it as an ordinary value),
// preserves a partial per-minor override set in config, and
// EffectiveKubeUpgradePath resolves the live path respecting those overrides.
func TestMergeAndEffectiveKubeUpgradePath(t *testing.T) {
	// 1. Empty config -> merge seeds the whole default path.
	cfg := map[string]interface{}{}
	require.NoError(t, core.MergeKubeUpgradePathDefaults(cfg, defaultPath))
	eff, err := core.EffectiveKubeUpgradePath(cfg)
	require.NoError(t, err)
	assert.Equal(t, "v1.24.17", eff[24], "v1.24 default should be seeded")
	assert.Equal(t, "v1.34.11", eff[34], "v1.34 default should be seeded")
	assert.Equal(t, "v1.36.4", eff[36], "v1.36 default should be seeded")
	assert.Equal(t, "v1.37.0", eff[37], "v1.37 default should be seeded")
	assert.Len(t, eff, len(defaultPath), "empty config should carry the full default path")

	// 2. Partial override set in config -> merge keeps it, fills the rest.
	over := map[string]interface{}{
		"cluster_require": map[string]interface{}{
			"kube_upgrade_path": map[string]interface{}{
				"v1.24": "v1.24.99",
				"v1.31": "v1.31.99",
			},
		},
	}
	require.NoError(t, core.MergeKubeUpgradePathDefaults(over, defaultPath))
	eff2, err := core.EffectiveKubeUpgradePath(over)
	require.NoError(t, err)
	assert.Equal(t, "v1.34.11", eff2[34], "unset minor should fall back to the default")
	assert.Equal(t, "v1.24.99", eff2[24], "config override of v1.24 must win")
	assert.Equal(t, "v1.31.99", eff2[31], "config override of v1.31 must win")
	assert.Len(t, eff2, len(defaultPath), "merged config must still cover every minor")
}

// TestComputeUpgradeSteps locks in the auto-step splitting logic.
func TestComputeUpgradeSteps(t *testing.T) {
	tests := []struct {
		name            string
		currentMinor    int
		targetMinor     int
		requestedTarget string
		want            []string
		wantErr         bool
	}{
		{
			// The headline scenario: v1.23.17 -> v1.36.4 must be split into 13
			// single-minor hops, each intermediate using the path's highest patch
			// and the final hop using the user-requested version.
			name:            "multi-minor 1.23 -> 1.36",
			currentMinor:    23,
			targetMinor:     36,
			requestedTarget: "v1.36.4",
			want: []string{
				"v1.24.17", "v1.25.16", "v1.26.15", "v1.27.16", "v1.28.15",
				"v1.29.15", "v1.30.14", "v1.31.14", "v1.32.13", "v1.33.13",
				"v1.34.11", "v1.35.8", "v1.36.4",
			},
		},
		{
			// Adjacent minor: a single hop, no intermediate steps.
			name:            "adjacent minor 1.23 -> 1.24",
			currentMinor:    23,
			targetMinor:     24,
			requestedTarget: "v1.24.17",
			want:            []string{"v1.24.17"},
		},
		{
			// Same minor: patch-only upgrade goes straight to the requested patch.
			name:            "patch-only 1.34.x -> 1.34.5",
			currentMinor:    34,
			targetMinor:     34,
			requestedTarget: "v1.34.5",
			want:            []string{"v1.34.5"},
		},
		{
			// Downgrade must be rejected.
			name:            "downgrade rejected",
			currentMinor:    34,
			targetMinor:     33,
			requestedTarget: "v1.33.13",
			wantErr:         true,
		},
		{
			// An intermediate minor missing from the path makes auto-stepping unsafe.
			name:            "gap in path rejected",
			currentMinor:    23,
			targetMinor:     40,
			requestedTarget: "v1.40.0",
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := core.ComputeUpgradeSteps(tt.currentMinor, tt.targetMinor, tt.requestedTarget, defaultPath)
			if tt.wantErr {
				assert.Error(t, err, "expected an error")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestMinorOfVersion handles the discovery API's quirky minor strings.
func TestMinorOfVersion(t *testing.T) {
	tests := []struct {
		in   string
		want int
	}{
		{"v1.34.3", 34},
		{"1.23", 23},
		{"v1.34+", 34}, // discovery sometimes appends "+"
		{"v1.33.7", 33},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := core.MinorOfVersion(tt.in)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// terminalCfg is a representative terminal (end-state) config: non-version fields
// (cri.container_manager, zone, image_registry.type/ha_vip) mixed with per-minor
// version fields (etcd, helm, sandbox, containerd, calico, dns) plus an unrelated
// custom_labels block that must never be touched.
func terminalCfg() map[string]interface{} {
	return map[string]interface{}{
		"zone": "cn",
		"custom_labels": map[string]interface{}{
			"env": "prod",
		},
		"kubernetes": map[string]interface{}{
			"kube_version": "v1.36.4",
			"helm_version": "v3.18.5",
			"sandbox_image": map[string]interface{}{
				"tag": "3.10.2",
			},
			"control_plane_endpoint": map[string]interface{}{
				"type": "local",
				"kube_vip": map[string]interface{}{
					"image": map[string]interface{}{"tag": "v0.7.2"},
				},
				"haproxy": map[string]interface{}{
					"image": map[string]interface{}{"tag": "2.9.6-alpine"},
				},
			},
		},
		"etcd": map[string]interface{}{
			"etcd_version": "v3.6.8",
		},
		"image_registry": map[string]interface{}{
			"type":           "harbor",
			"ha_vip":         "10.0.0.10",
			"harbor_version": "v2.10.2",
		},
		"cri": map[string]interface{}{
			"container_manager":  "containerd",
			"containerd_version": "v1.7.13",
			"crictl_version":     "v1.36.0",
		},
		"cni": map[string]interface{}{
			"type":           "calico",
			"calico_version": "v3.32.1",
		},
		"dns": map[string]interface{}{
			"coredns": map[string]interface{}{
				"image": map[string]interface{}{"tag": "v1.14.2"},
			},
		},
		"storage_class": map[string]interface{}{
			"local": map[string]interface{}{"enabled": true, "default": true},
		},
	}
}

// TestBuildUpgradeStepConfig locks in the final upgrade model's per-hop config
// semantics:
//   - kube_version is always forced to the hop's target.
//   - Non-version fields (cri.container_manager, image_registry.type/ha_vip, zone,
//     custom_labels, storage.enabled) are always inherited verbatim from the
//     terminal config, on every hop.
//   - Intermediate hops (isFinal=false) have their version fields EXPLICITLY
//     overlaid from the embedded roles/defaults/vars/v1.{hop}.yaml (e.g. etcd
//     v3.5.15 for v1.28, not the terminal v3.6.5).
//   - The FINAL hop (isFinal=true) keeps the terminal config's own version fields.
func TestBuildUpgradeStepConfig(t *testing.T) {
	vfs := core.BuiltinPlaybook

	t.Run("intermediate hop overlays version fields from embedded vars, keeps non-version", func(t *testing.T) {
		got, err := core.BuildUpgradeStepConfig(vfs, terminalCfg(), "v1.28.15", false)
		require.NoError(t, err)

		// kube_version forced to this hop.
		v, _, _ := unstructured.NestedString(got, "kubernetes", "kube_version")
		assert.Equal(t, "v1.28.15", v)

		// Version fields explicitly overlaid from embedded v1.28.yaml (NOT terminal).
		v, _, _ = unstructured.NestedString(got, "etcd", "etcd_version")
		assert.Equal(t, "v3.5.15", v, "etcd.etcd_version must come from v1.28.yaml, not terminal v3.6.8")
		v, _, _ = unstructured.NestedString(got, "cni", "calico_version")
		assert.Equal(t, "v3.28.5", v, "cni.calico_version must come from v1.28.yaml, not terminal v3.32.1")
		v, _, _ = unstructured.NestedString(got, "kubernetes", "helm_version")
		assert.Equal(t, "v3.12.1", v, "kubernetes.helm_version must come from v1.28.yaml")
		// crictl follows kube_version: the v1.28 hop pins crictl to v1.28.0 (from
		// v1.28.yaml), NOT the terminal v1.36.0 — so a plain Kubernetes-only upgrade
		// keeps the CRI CLI tool in lockstep with each minor it passes through.
		v, _, _ = unstructured.NestedString(got, "cri", "crictl_version")
		assert.Equal(t, "v1.28.0", v, "cri.crictl_version must come from v1.28.yaml, not terminal v1.36.0")

		// Non-version terminal fields inherited.
		cm, _, _ := unstructured.NestedString(got, "cri", "container_manager")
		assert.Equal(t, "containerd", cm, "container_manager is a non-version field and must persist")
		typ, _, _ := unstructured.NestedString(got, "image_registry", "type")
		assert.Equal(t, "harbor", typ, "image_registry.type must persist")
		haVip, _, _ := unstructured.NestedString(got, "image_registry", "ha_vip")
		assert.Equal(t, "10.0.0.10", haVip, "image_registry.ha_vip must persist")
		zone, _, _ := unstructured.NestedString(got, "zone")
		assert.Equal(t, "cn", zone, "zone must persist")
		labels, _, _ := unstructured.NestedString(got, "custom_labels", "env")
		assert.Equal(t, "prod", labels, "custom_labels must persist")
		st, found, _ := unstructured.NestedBool(got, "storage_class", "local", "enabled")
		assert.True(t, found)
		assert.True(t, st, "storage_class.local.enabled must persist")
	})

	t.Run("final hop keeps terminal version fields", func(t *testing.T) {
		got, err := core.BuildUpgradeStepConfig(vfs, terminalCfg(), "v1.36.4", true)
		require.NoError(t, err)

		v, _, _ := unstructured.NestedString(got, "kubernetes", "kube_version")
		assert.Equal(t, "v1.36.4", v)
		ev, _, _ := unstructured.NestedString(got, "etcd", "etcd_version")
		assert.Equal(t, "v3.6.8", ev, "terminal etcd.etcd_version must be honored on the final hop")
		cv, _, _ := unstructured.NestedString(got, "cni", "calico_version")
		assert.Equal(t, "v3.32.1", cv, "terminal cni.calico_version must be honored on the final hop")
		cm, _, _ := unstructured.NestedString(got, "cri", "container_manager")
		assert.Equal(t, "containerd", cm)
	})

	t.Run("overlay uses separate file and does not pull non-version keys", func(t *testing.T) {
		ov, err := core.UpgradeVersionOverlay(vfs, 28)
		require.NoError(t, err)
		// Version fields present.
		_, found, _ := unstructured.NestedString(ov, "etcd", "etcd_version")
		assert.True(t, found, "etcd.etcd_version must be in the v1.28 overlay")
		// Non-version keys from the vars file must NOT be pulled into the overlay.
		_, found, _ = unstructured.NestedString(ov, "cri", "container_manager")
		assert.False(t, found, "cri.container_manager is non-version and must not be in the overlay")
	})

	t.Run("base config is not mutated", func(t *testing.T) {
		base := terminalCfg()
		_, err := core.BuildUpgradeStepConfig(vfs, base, "v1.28.15", false)
		require.NoError(t, err)
		// The original terminal config must keep its version fields and its kube_version.
		v, _, _ := unstructured.NestedString(base, "kubernetes", "kube_version")
		assert.Equal(t, "v1.36.4", v)
		ev, _, _ := unstructured.NestedString(base, "etcd", "etcd_version")
		assert.Equal(t, "v3.6.8", ev)
	})
}

// TestValidateTerminalVersions locks in the monotonic-upgrade guard: if a component
// version pinned in the terminal config is lower than any intermediate hop's value,
// the upgrade must error (components cannot be downgraded). Multi-minor steps are
// compared, non-pinned fields and single-hop runs pass.
func TestValidateTerminalVersions(t *testing.T) {
	vfs := core.BuiltinPlaybook
	multi := []string{"v1.24.17", "v1.25.16", "v1.26.15", "v1.27.16", "v1.28.15", "v1.36.4"}

	t.Run("terminal etcd lower than intermediate hop errors", func(t *testing.T) {
		// terminal etcd v3.5.6 is below v1.28's etcd v3.5.15.
		cfg := terminalCfg()
		_ = unstructured.SetNestedField(cfg, "v3.5.6", "etcd", "etcd_version")
		err := core.ValidateTerminalVersions(vfs, cfg, multi)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "etcd.etcd_version")
		assert.Contains(t, err.Error(), "v3.5.6")
	})

	t.Run("terminal etcd at or above all intermediates passes", func(t *testing.T) {
		cfg := terminalCfg() // terminal etcd v3.6.8 >= every intermediate
		require.NoError(t, core.ValidateTerminalVersions(vfs, cfg, multi))
	})

	t.Run("non-pinned version fields pass", func(t *testing.T) {
		// No etcd/component version pinned at all -> nothing to conflict.
		cfg := map[string]interface{}{
			"kubernetes": map[string]interface{}{"kube_version": "v1.36.4"},
			"cri":        map[string]interface{}{"container_manager": "containerd"},
		}
		require.NoError(t, core.ValidateTerminalVersions(vfs, cfg, multi))
	})

	t.Run("single-hop / patch-only skip", func(t *testing.T) {
		cfg := terminalCfg()
		_ = unstructured.SetNestedField(cfg, "v3.5.6", "etcd", "etcd_version")
		require.NoError(t, core.ValidateTerminalVersions(vfs, cfg, []string{"v1.34.3"}))
	})

	t.Run("terminal crictl lower than intermediate errors", func(t *testing.T) {
		cfg := terminalCfg()
		_ = unstructured.SetNestedField(cfg, "v1.28.0", "cri", "crictl_version")
		err := core.ValidateTerminalVersions(vfs, cfg, []string{"v1.30.14", "v1.34.3"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cri.crictl_version")
	})
}

func TestCompareVersionNumbers(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"v3.5.15", "v3.6.5", -1},
		{"v3.6.5", "v3.5.15", 1},
		{"v3.5.15", "v3.5.15", 0},
		{"3.10.1", "3.9", 1},
		{"v1.34.0", "v1.28.0", 1},
		{"v3.29.7", "v3.31.3", -1},
		{"2.9.6-alpine", "2.9.6-alpine", 0},
		{"v1.7.13", "v1.6.20", 1},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			got := core.CompareVersionNumbers(tt.a, tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}
