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
	"testing"

	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// newUpgradeTestOptions builds an UpgradeClusterOptions with a fresh, empty
// Config so tests can exercise completeConfig() without loading files.
func newUpgradeTestOptions() *UpgradeClusterOptions {
	o := NewUpgradeClusterOptions()
	o.Config = &kkcorev1.Config{}
	return o
}

func upgradeField(t *testing.T, o *UpgradeClusterOptions, keys ...string) (any, bool) {
	t.Helper()
	v, ok, err := unstructured.NestedFieldNoCopy(o.Config.Value(), append([]string{"upgrade"}, keys...)...)
	require.NoError(t, err)
	return v, ok
}

// ----------------------------------------------------------------------------
// TC-U01: NewUpgradeClusterOptions defaults
// ----------------------------------------------------------------------------
func TestNewUpgradeClusterOptionsDefaults(t *testing.T) {
	o := NewUpgradeClusterOptions()
	assert.Equal(t, "", o.Kubernetes, "Kubernetes 默认应为空")
	assert.False(t, o.UpgradeAllComponents, "UpgradeAllComponents 默认应为 false")
	assert.Equal(t, "", o.OnlyComponent, "OnlyComponent 默认应为空")
}

// ----------------------------------------------------------------------------
// TC-U02: Flags registration
// ----------------------------------------------------------------------------
func TestUpgradeClusterOptionsFlags(t *testing.T) {
	o := NewUpgradeClusterOptions()
	fss := o.Flags()
	fs := (&fss).FlagSet("config")

	assert.NotNil(t, fs.Lookup("with-kubernetes"), "--with-kubernetes 应已注册")
	assert.NotNil(t, fs.Lookup("all"), "--all 应已注册")
}

// ----------------------------------------------------------------------------
// TC-U03: kk upgrade cluster (no flags) -> only Kubernetes base is upgraded
// ----------------------------------------------------------------------------
func TestCompleteConfigClusterUpgradeDefault(t *testing.T) {
	o := newUpgradeTestOptions()
	require.NoError(t, unstructured.SetNestedField(o.Config.Value(), "v1.23.17", "kubernetes", "kube_version"))

	require.NoError(t, o.completeConfig())

	kubernetes, ok := upgradeField(t, o, "kubernetes")
	require.True(t, ok, "upgrade.kubernetes 必须被写入")
	assert.True(t, kubernetes.(bool), "kk upgrade cluster 必须启用 Kubernetes 升级")

	for _, comp := range []string{"etcd", "cri", "cni", "storage_class"} {
		v, ok := upgradeField(t, o, comp)
		require.True(t, ok, "组件 %s 必须有默认值", comp)
		assert.False(t, v.(bool), "未指定 --all 时组件 %s 不应升级", comp)
	}

	// kube_version 未被 --with-kubernetes 覆盖，应保持原值
	kv, _, err := unstructured.NestedString(o.Config.Value(), "kubernetes", "kube_version")
	require.NoError(t, err)
	assert.Equal(t, "v1.23.17", kv, "未指定 --with-kubernetes 时 kube_version 不应改变")
}

// ----------------------------------------------------------------------------
// TC-U04: kk upgrade cluster --all -> every wired component upgraded
// ----------------------------------------------------------------------------
func TestCompleteConfigClusterUpgradeAll(t *testing.T) {
	o := newUpgradeTestOptions()
	o.UpgradeAllComponents = true

	require.NoError(t, o.completeConfig())

	kubernetes, _ := upgradeField(t, o, "kubernetes")
	assert.True(t, kubernetes.(bool), "cluster 升级必须启用 Kubernetes")

	for _, comp := range []string{"etcd", "cri", "cni", "storage_class"} {
		v, ok := upgradeField(t, o, comp)
		require.True(t, ok, "组件 %s 必须被 --all 启用", comp)
		assert.True(t, v.(bool), "--all 时组件 %s 应升级", comp)
	}

	// dns / image_registry / nfs 有意未接入升级链路，--all 不应误启用
	for _, excluded := range []string{"dns", "image_registry", "nfs"} {
		_, ok := upgradeField(t, o, excluded)
		assert.False(t, ok, "未接入的组件 %s 不应被 --all 启用（保持缺省/不写入）", excluded)
	}
}

// ----------------------------------------------------------------------------
// TC-U05: kk upgrade cluster --with-kubernetes v1.34.3 -> overrides config
// ----------------------------------------------------------------------------
func TestCompleteConfigClusterUpgradeWithKubernetes(t *testing.T) {
	o := newUpgradeTestOptions()
	require.NoError(t, unstructured.SetNestedField(o.Config.Value(), "v1.23.17", "kubernetes", "kube_version"))
	o.Kubernetes = "v1.34.3"

	require.NoError(t, o.completeConfig())

	kv, _, err := unstructured.NestedString(o.Config.Value(), "kubernetes", "kube_version")
	require.NoError(t, err)
	assert.Equal(t, "v1.34.3", kv, "--with-kubernetes 应覆盖 config 中的 kube_version")

	kubernetes, _ := upgradeField(t, o, "kubernetes")
	assert.True(t, kubernetes.(bool))
}

// ----------------------------------------------------------------------------
// TC-U06: kk upgrade cluster with config pre-set upgrade.etcd=true (--set 等价)
// -> preserved, other components default false
// ----------------------------------------------------------------------------
func TestCompleteConfigClusterUpgradeSetOverride(t *testing.T) {
	o := newUpgradeTestOptions()
	require.NoError(t, unstructured.SetNestedField(o.Config.Value(), true, "upgrade", "etcd"))

	require.NoError(t, o.completeConfig())

	etcd, _ := upgradeField(t, o, "etcd")
	assert.True(t, etcd.(bool), "config 中已置 true 的 upgrade.etcd 应被保留")

	for _, comp := range []string{"cri", "cni", "storage_class"} {
		v, _ := upgradeField(t, o, comp)
		assert.False(t, v.(bool), "未指定的组件 %s 应为 false", comp)
	}
	kubernetes, _ := upgradeField(t, o, "kubernetes")
	assert.True(t, kubernetes.(bool))
}

// ----------------------------------------------------------------------------
// TC-U07~U10: kk upgrade <component> (OnlyComponent) -> ONLY that component,
// Kubernetes base disabled, other components untouched (not written).
// ----------------------------------------------------------------------------
func TestCompleteConfigComponentOnlyEtcd(t *testing.T) {
	o := newUpgradeTestOptions()
	o.OnlyComponent = "etcd"

	require.NoError(t, o.completeConfig())

	kubernetes, _ := upgradeField(t, o, "kubernetes")
	assert.False(t, kubernetes.(bool), "组件级升级不应触碰 Kubernetes 控制平面")

	etcd, _ := upgradeField(t, o, "etcd")
	assert.True(t, etcd.(bool), "kk upgrade etcd 应启用 upgrade.etcd")

	for _, untouched := range []string{"cri", "cni", "storage_class"} {
		_, ok := upgradeField(t, o, untouched)
		assert.False(t, ok, "组件级升级不应写入其他组件 %s", untouched)
	}
}

func TestCompleteConfigComponentOnlyCri(t *testing.T) {
	o := newUpgradeTestOptions()
	o.OnlyComponent = "cri"

	require.NoError(t, o.completeConfig())

	kubernetes, _ := upgradeField(t, o, "kubernetes")
	assert.False(t, kubernetes.(bool))
	cri, _ := upgradeField(t, o, "cri")
	assert.True(t, cri.(bool))
	_, ok := upgradeField(t, o, "etcd")
	assert.False(t, ok, "组件级升级不应写入 etcd")
}

func TestCompleteConfigComponentOnlyCni(t *testing.T) {
	o := newUpgradeTestOptions()
	o.OnlyComponent = "cni"

	require.NoError(t, o.completeConfig())

	kubernetes, _ := upgradeField(t, o, "kubernetes")
	assert.False(t, kubernetes.(bool))
	cni, _ := upgradeField(t, o, "cni")
	assert.True(t, cni.(bool))
}

func TestCompleteConfigComponentOnlyStorageClass(t *testing.T) {
	o := newUpgradeTestOptions()
	o.OnlyComponent = "storage_class"

	require.NoError(t, o.completeConfig())

	kubernetes, _ := upgradeField(t, o, "kubernetes")
	assert.False(t, kubernetes.(bool))
	sc, _ := upgradeField(t, o, "storage_class")
	assert.True(t, sc.(bool))
}

// ----------------------------------------------------------------------------
// TC-U11: Artifact mode disables download.fetch only when not already set
// ----------------------------------------------------------------------------
func TestCompleteConfigArtifactDisablesFetch(t *testing.T) {
	t.Run("fetch unset -> set false", func(t *testing.T) {
		o := newUpgradeTestOptions()
		o.Artifact = "/tmp/artifact"
		require.NoError(t, o.completeConfig())

		v, ok, err := unstructured.NestedBool(o.Config.Value(), "download", "fetch")
		require.NoError(t, err)
		require.True(t, ok)
		assert.False(t, v, "指定 --artifact 且 fetch 未设置时应改为 false")
	})

	t.Run("fetch already true -> preserved", func(t *testing.T) {
		o := newUpgradeTestOptions()
		o.Artifact = "/tmp/artifact"
		require.NoError(t, unstructured.SetNestedField(o.Config.Value(), true, "download", "fetch"))
		require.NoError(t, o.completeConfig())

		v, _, err := unstructured.NestedBool(o.Config.Value(), "download", "fetch")
		require.NoError(t, err)
		assert.True(t, v, "已有 download.fetch=true 时应被保留")
	})
}
