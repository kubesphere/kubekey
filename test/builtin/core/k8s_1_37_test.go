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

// Package core_test. These tests lock in the K8s v1.37 support surface added for
// kk create/upgrade with K8s 1.37: the create-config template, the per-minor
// version overlay used by the upgrade role, the cluster_require compatibility
// matrix (etcd floor + CNI allowed versions), the Go upgrade-path default, and
// the manifests download-image template branches.
package core_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"

	core "github.com/kubesphere/kubekey/v4/builtin/core"
)

// assertNestedString fetches a nested string field and asserts it equals want.
func assertNestedString(t *testing.T, m map[string]interface{}, want string, fields ...string) {
	t.Helper()
	got, found, err := unstructured.NestedString(m, fields...)
	require.NoError(t, err)
	require.True(t, found, "field %v not present in overlay", fields)
	assert.Equal(t, want, got, "field %v mismatch", fields)
}

// TestK8s137ConfigTemplate locks in the `kk create config --with-kubernetes v1.37.0`
// template's active component defaults. The template still carries the
// `{{ .kubernetes.kube_version }}` placeholder (substituted at render time), so
// the assertion checks the placeholder survives and the pinned component tags
// match kubeadm v1.37.0 constants.
func TestK8s137ConfigTemplate(t *testing.T) {
	raw, err := core.Defaults.ReadFile("defaults/config/v1.37.yaml")
	require.NoError(t, err, "v1.37 config template must be embedded")
	content := string(raw)

	// kube_version stays a render-time placeholder.
	assert.Contains(t, content, "kube_version: {{ .kubernetes.kube_version }}")
	// Active component defaults (the lines that are NOT commented out).
	assert.Contains(t, content, "etcd_version: v3.7.0", "1.37 stacked etcd default must be v3.7.0")
	assert.Contains(t, content, "crictl_version: v1.37.0", "crictl must track the kube minor (v1.37.0)")
	assert.Contains(t, content, "tag: v1.14.6", "CoreDNS image tag must be v1.14.6 (kubeadm 1.37)")
	assert.Contains(t, content, `tag: "3.10.2"`, "pause sandbox tag must be 3.10.2")
	assert.Contains(t, content, "calico_version: v3.32.2", "Calico forward-compat default for 1.37 is v3.32.2")
	assert.Contains(t, content, "helm_version: v3.18.5")
}

// TestK8s137VarsOverlay locks in the per-minor version overlay returned by
// core.UpgradeVersionOverlay for minor 37 — the exact values the upgrade role
// forces onto each intermediate hop when stepping through 1.37.
func TestK8s137VarsOverlay(t *testing.T) {
	ov, err := core.UpgradeVersionOverlay(core.BuiltinPlaybook, 37)
	require.NoError(t, err)
	require.NotNil(t, ov, "v1.37 overlay must exist")

	assertNestedString(t, ov, "v3.7.0", "etcd", "etcd_version")
	assertNestedString(t, ov, "v1.37.0", "cri", "crictl_version")
	assertNestedString(t, ov, "v2.3.4", "cri", "containerd_version")
	assertNestedString(t, ov, "v1.2.6", "cri", "runc_version")
	assertNestedString(t, ov, "v1.9.1", "cni", "cni_plugins_version")
	assertNestedString(t, ov, "v3.32.2", "cni", "calico_version")
	assertNestedString(t, ov, "1.20.1", "cni", "cilium_version")
	assertNestedString(t, ov, "v3.18.5", "kubernetes", "helm_version")
	assertNestedString(t, ov, "v1.14.6", "dns", "coredns", "image", "tag")
	assertNestedString(t, ov, "3.10.2", "kubernetes", "sandbox_image", "tag")
}

// TestK8s137ClusterRequireMatrix locks in the cluster_require compatibility
// matrix entries for 1.37: the etcd minimum floor and the CNI allowed-version
// rows. The etcd floor must be v3.5.24-0 (kubeadm MinExternalEtcdVersion for
// 1.37), NOT v3.7.0-0 (which is the stacked/SupportedEtcdVersion default).
func TestK8s137ClusterRequireMatrix(t *testing.T) {
	raw, err := core.BuiltinPlaybook.ReadFile("roles/defaults/defaults/main/01-cluster_require.yaml")
	require.NoError(t, err, "cluster_require must be embedded")

	var doc map[string]interface{}
	require.NoError(t, yaml.Unmarshal(raw, &doc))
	cr, found, err := unstructured.NestedMap(doc, "cluster_require")
	require.NoError(t, err)
	require.True(t, found, "cluster_require block must be present")

	// etcd_min_versions: v1.37 floor is the external-etcd minimum, v3.5.24-0.
	etcdFloor, found, err := unstructured.NestedString(cr, "etcd_min_versions", "v1.37")
	require.NoError(t, err)
	require.True(t, found, "etcd_min_versions must carry a v1.37 entry")
	assert.Equal(t, "v3.5.24-0", etcdFloor,
		"1.37 etcd floor must be the external-etcd minimum v3.5.24-0, not the stacked default v3.7.0-0")

	// CNI allowed-version rows must each include v1.37.
	calico, _, _ := unstructured.NestedStringSlice(cr, "calico_allowed_versions", "v3.32")
	assert.Contains(t, calico, "v1.37", "Calico v3.32 must cover 1.37")
	cilium, _, _ := unstructured.NestedStringSlice(cr, "cilium_allowed_versions", "1.20")
	assert.Contains(t, cilium, "v1.37", "Cilium 1.20 must cover 1.37")
	ko15, _, _ := unstructured.NestedStringSlice(cr, "kubeovn_allowed_versions", "v1.15")
	assert.Contains(t, ko15, "v1.37", "Kube-OVN v1.15 must cover 1.37")
	ko16, _, _ := unstructured.NestedStringSlice(cr, "kubeovn_allowed_versions", "v1.16")
	assert.Contains(t, ko16, "v1.37", "Kube-OVN v1.16 must cover 1.37")
}

// TestK8s137UpgradePath locks in the Go CLI default for minor 37 and that the
// default is seeded into an otherwise-empty config via MergeKubeUpgradePathDefaults.
func TestK8s137UpgradePath(t *testing.T) {
	// Mirror kept in sync with cmd/kk/app/options/builtin/upgrade.go.
	require.Equal(t, "v1.37.0", defaultPath[37], "Go default path must map 37 -> v1.37.0")

	cfg := map[string]interface{}{}
	require.NoError(t, core.MergeKubeUpgradePathDefaults(cfg, defaultPath))
	eff, err := core.EffectiveKubeUpgradePath(cfg)
	require.NoError(t, err)
	assert.Equal(t, "v1.37.0", eff[37], "merged config must carry the v1.37.0 default")
}

// TestK8s137ManifestsBranches locks in the manifests download-image template's
// two v1.37 branches (component version list + pause/CoreDNS image tags) so a
// regression that drops the 1.37 branch is caught.
func TestK8s137ManifestsBranches(t *testing.T) {
	raw, err := core.BuiltinPlaybook.ReadFile("roles/defaults/templates/manifests.yaml")
	require.NoError(t, err, "manifests template must be embedded")
	content := string(raw)

	// Component version branch for v1.37.
	assert.Contains(t, content, `slice (. | splitList ".") 0 2 | join "." | eq "v1.37"`, "manifests must branch on v1.37")
	assert.Contains(t, content, `append $default_etcd_version "v3.7.0"`, "1.37 etcd in manifests must be v3.7.0")
	assert.Contains(t, content, `append $default_crictl_version "v1.37.0"`, "1.37 crictl in manifests must be v1.37.0")
	assert.Contains(t, content, `append $default_containerd_version "v2.3.4"`, "1.37 containerd in manifests must be v2.3.4")
	assert.Contains(t, content, `append $default_runc_version "v1.2.6"`, "1.37 runc in manifests must be v1.2.6")

	// pause / CoreDNS image branch for v1.37.
	assert.Contains(t, content, "kubernetes/pause:3.10.2", "1.37 pause image must be 3.10.2")
	assert.Contains(t, content, "coredns/coredns:v1.14.6", "1.37 CoreDNS image must be v1.14.6")
}
