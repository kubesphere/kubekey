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

// Package core_test holds L2 logic tests that validate the `when:` expressions
// copied from the builtin roles/playbooks (under builtin/core/roles and
// builtin/core/playbooks) using kk's own template engine. These are intentionally
// kept OUT of builtin/core so the embedded-source package only carries L1 unit
// tests (e.g. upgrade_path_test.go) next to their source files.
package core_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubesphere/kubekey/v4/pkg/converter/tmpl"
)

// These tests evaluate the ACTUAL `when:` expressions copied from the builtin
// roles/playbooks, using kk's own template engine (tmpl.ParseBool). They lock in
// the scenario-distinction and component-gating logic so a regression (e.g. the
// old `.upgrade | empty` predicate that broke `kk create`/`add nodes`) is caught.

// ----------------------------------------------------------------------------
// TC-I01: precheck/kubernetes scenario predicate (P0 fix verification)
//
//	install branch:  when: not (.upgrade.kubernetes | default false)
//	upgrade branch:  when: .upgrade.kubernetes | default false
//
// ----------------------------------------------------------------------------
func TestPrecheckKubernetesScenarioPredicate(t *testing.T) {
	installExpr := `{{ not (.upgrade.kubernetes | default false) }}`
	upgradeExpr := `{{ .upgrade.kubernetes | default false }}`

	tests := []struct {
		name     string
		ctx      map[string]any
		wantInst bool
		wantUpgr bool
	}{
		{
			// create / add nodes / precheck on existing cluster: no upgrade marker
			name:     "install scenario (no upgrade.kubernetes)",
			ctx:      map[string]any{},
			wantInst: true,
			wantUpgr: false,
		},
		{
			// kk upgrade cluster
			name:     "cluster upgrade",
			ctx:      map[string]any{"upgrade": map[string]any{"kubernetes": true}},
			wantInst: false,
			wantUpgr: true,
		},
		{
			// kk upgrade etcd: Kubernetes base must NOT be upgraded
			name:     "component-only etcd",
			ctx:      map[string]any{"upgrade": map[string]any{"kubernetes": false, "etcd": true}},
			wantInst: true,
			wantUpgr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst, err := tmpl.ParseBool(tt.ctx, installExpr)
			require.NoError(t, err)
			upgr, err := tmpl.ParseBool(tt.ctx, upgradeExpr)
			require.NoError(t, err)
			assert.Equal(t, tt.wantInst, inst, "安装分支判定错误")
			assert.Equal(t, tt.wantUpgr, upgr, "升级分支判定错误")
		})
	}
}

// ----------------------------------------------------------------------------
// TC-I02: precheck/etcd upgrade gate  when: .upgrade.etcd
// ----------------------------------------------------------------------------
func TestPrecheckEtcdUpgradeGate(t *testing.T) {
	expr := `{{ .upgrade.etcd }}`

	tests := []struct {
		name string
		ctx  map[string]any
		want bool
	}{
		{"no upgrade section", map[string]any{}, false},
		{"upgrade.etcd=false", map[string]any{"upgrade": map[string]any{"etcd": false}}, false},
		{"upgrade.etcd=true", map[string]any{"upgrade": map[string]any{"etcd": true}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tmpl.ParseBool(tt.ctx, expr)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ----------------------------------------------------------------------------
// TC-I03: precheck/cri upgrade gates
//
//	validate target containerd when upgrading cri:  when: .upgrade.cri
//	validate installed containerd when NOT upgrading cri: when: .upgrade.cri | not
//
// ----------------------------------------------------------------------------
func TestPrecheckCriUpgradeGates(t *testing.T) {
	targetExpr := `{{ .upgrade.cri }}`
	installedExpr := `{{ .upgrade.cri | not }}`

	tests := []struct {
		name     string
		ctx      map[string]any
		wantTgt  bool
		wantInst bool
	}{
		{"no upgrade section", map[string]any{}, false, true},
		{"upgrade.cri=true", map[string]any{"upgrade": map[string]any{"cri": true}}, true, false},
		{"upgrade.cri=false", map[string]any{"upgrade": map[string]any{"cri": false}}, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tgt, err := tmpl.ParseBool(tt.ctx, targetExpr)
			require.NoError(t, err)
			inst, err := tmpl.ParseBool(tt.ctx, installedExpr)
			require.NoError(t, err)
			assert.Equal(t, tt.wantTgt, tgt, "升级 cri 时校验目标版本分支错误")
			assert.Equal(t, tt.wantInst, inst, "不升级 cri 时校验已装版本分支错误")
		})
	}
}

// ----------------------------------------------------------------------------
// TC-I05: upgrade_cluster.yaml component gates
//
//	upgrade-kubernetes role:  when: .upgrade.kubernetes
//	cni role:                 when: .upgrade.cni
//	storageclass role:        when: .upgrade.storage_class
//	etcd external:            when: and (.etcd.deployment_type | eq "external") (.upgrade.etcd)
//
// ----------------------------------------------------------------------------
func TestUpgradeClusterPlaybookGates(t *testing.T) {
	kubernetesGate := `{{ .upgrade.kubernetes }}`
	cniGate := `{{ .upgrade.cni }}`
	criGate := `{{ .upgrade.cri }}`
	// cri/cridockerd is referenced directly in upgrade_cluster.yaml for Kubernetes-only
	// upgrades on docker + Kubernetes >= v1.24.0. The `.upgrade.cri | not` clause keeps it
	// mutually exclusive with the cri role dependency, so it never runs twice.
	criDockerdUpgradeGate := `{{ and (.cri.container_manager | eq "docker") (.kubernetes.kube_version | semverCompare ">=v1.24.0") (.upgrade.cri | not) }}`
	scGate := `{{ .upgrade.storage_class }}`
	etcdExternalGate := `{{ and (.etcd.deployment_type | eq "external") (.upgrade.etcd) }}`

	t.Run("kubernetes gate", func(t *testing.T) {
		cluster := map[string]any{"upgrade": map[string]any{"kubernetes": true}}
		onlyEtcd := map[string]any{"upgrade": map[string]any{"kubernetes": false, "etcd": true}}

		got, err := tmpl.ParseBool(cluster, kubernetesGate)
		require.NoError(t, err)
		assert.True(t, got, "cluster 升级应启用 upgrade-kubernetes role")
		got, err = tmpl.ParseBool(onlyEtcd, kubernetesGate)
		require.NoError(t, err)
		assert.False(t, got, "组件级升级不应启用 upgrade-kubernetes role")
	})

	t.Run("cni gate", func(t *testing.T) {
		ctx := map[string]any{"upgrade": map[string]any{"cni": true}}
		got, err := tmpl.ParseBool(ctx, cniGate)
		require.NoError(t, err)
		assert.True(t, got)
		got, err = tmpl.ParseBool(map[string]any{"upgrade": map[string]any{"cni": false}}, cniGate)
		require.NoError(t, err)
		assert.False(t, got)
	})

	t.Run("cri gate", func(t *testing.T) {
		// The cri role runs ONLY when upgrade.cri is set. A Kubernetes-only upgrade
		// must never touch the Docker daemon / containerd, so the gate is now plain.
		got, err := tmpl.ParseBool(map[string]any{"upgrade": map[string]any{"cri": true}}, criGate)
		require.NoError(t, err)
		assert.True(t, got, "upgrade.cri=true 应启用 cri role")

		containerdOnly := map[string]any{
			"upgrade":    map[string]any{"cri": false},
			"cri":        map[string]any{"container_manager": "containerd"},
			"kubernetes": map[string]any{"kube_version": "v1.24.0"},
		}
		got, err = tmpl.ParseBool(containerdOnly, criGate)
		require.NoError(t, err)
		assert.False(t, got, "containerd + Kubernetes-only 升级应跳过 cri role（不重启运行时）")

		dockerCross124 := map[string]any{
			"upgrade":    map[string]any{"cri": false},
			"cri":        map[string]any{"container_manager": "docker"},
			"kubernetes": map[string]any{"kube_version": "v1.24.0"},
		}
		got, err = tmpl.ParseBool(dockerCross124, criGate)
		require.NoError(t, err)
		assert.False(t, got, "docker + 目标>=1.24 的 Kubernetes-only 升级由 cri/cridockerd 直接处理，cri role 仍跳过")
	})

	t.Run("cri cridockerd direct gate", func(t *testing.T) {
		// Kubernetes-only upgrade, docker + target >=1.24: cri-dockerd must be installed
		// (dockershim removed in 1.24) WITHOUT touching the Docker daemon.
		dockerCross124 := map[string]any{
			"upgrade":    map[string]any{"cri": false},
			"cri":        map[string]any{"container_manager": "docker"},
			"kubernetes": map[string]any{"kube_version": "v1.24.0"},
		}
		got, err := tmpl.ParseBool(dockerCross124, criDockerdUpgradeGate)
		require.NoError(t, err)
		assert.True(t, got, "docker + 目标>=1.24 的 Kubernetes-only 升级应直接启用 cri/cridockerd")

		// Docker + target >=1.24 but upgrade.cri=true: handled by the cri role dependency,
		// the direct reference must NOT fire (mutually exclusive, no double run).
		dockerCross124UpgradeCri := map[string]any{
			"upgrade":    map[string]any{"cri": true},
			"cri":        map[string]any{"container_manager": "docker"},
			"kubernetes": map[string]any{"kube_version": "v1.24.0"},
		}
		got, err = tmpl.ParseBool(dockerCross124UpgradeCri, criDockerdUpgradeGate)
		require.NoError(t, err)
		assert.False(t, got, "upgrade.cri=true 时 cri/cridockerd 由 cri 依赖处理，直接引用不应触发")

		// Containerd runtime never needs cri-dockerd.
		containerdOnly := map[string]any{
			"upgrade":    map[string]any{"cri": false},
			"cri":        map[string]any{"container_manager": "containerd"},
			"kubernetes": map[string]any{"kube_version": "v1.24.0"},
		}
		got, err = tmpl.ParseBool(containerdOnly, criDockerdUpgradeGate)
		require.NoError(t, err)
		assert.False(t, got, "containerd 运行时不需要 cri-dockerd")

		// Docker but target <1.24: dockershim still present, cri-dockerd not needed.
		dockerPre124 := map[string]any{
			"upgrade":    map[string]any{"cri": false},
			"cri":        map[string]any{"container_manager": "docker"},
			"kubernetes": map[string]any{"kube_version": "v1.23.0"},
		}
		got, err = tmpl.ParseBool(dockerPre124, criDockerdUpgradeGate)
		require.NoError(t, err)
		assert.False(t, got, "docker + 目标<1.24 不需要 cri-dockerd")
	})

	t.Run("storage_class gate", func(t *testing.T) {
		ctx := map[string]any{"upgrade": map[string]any{"storage_class": true}}
		got, err := tmpl.ParseBool(ctx, scGate)
		require.NoError(t, err)
		assert.True(t, got)
	})

	t.Run("etcd external gate", func(t *testing.T) {
		externalUp := map[string]any{
			"etcd":    map[string]any{"deployment_type": "external"},
			"upgrade": map[string]any{"etcd": true},
		}
		externalNoUp := map[string]any{
			"etcd":    map[string]any{"deployment_type": "external"},
			"upgrade": map[string]any{"etcd": false},
		}
		stackedUp := map[string]any{
			"etcd":    map[string]any{"deployment_type": "internal"},
			"upgrade": map[string]any{"etcd": true},
		}
		got, err := tmpl.ParseBool(externalUp, etcdExternalGate)
		require.NoError(t, err)
		assert.True(t, got, "external etcd + upgrade.etcd 应触发外部 etcd 升级")
		got, err = tmpl.ParseBool(externalNoUp, etcdExternalGate)
		require.NoError(t, err)
		assert.False(t, got, "external etcd 但未启用 upgrade.etcd 不应触发")
		got, err = tmpl.ParseBool(stackedUp, etcdExternalGate)
		require.NoError(t, err)
		assert.False(t, got, "stacked etcd 不应走外部 etcd 升级链路")
	})
}

// ----------------------------------------------------------------------------
// TC-I06: precheck/kubernetes upgrade-path assertion correctness
//
//	when: .upgrade.kubernetes | default false
//	assert: installed | trimPrefix "Kubernetes " | semverCompare (printf "<%s" target)
//
// Verifies the assertion passes for a valid upgrade and fails for same/downgrade.
// NOTE: this only validates precheck. kk still does NOT auto-step minor versions
// (kube_upgrade_path is defined but has zero consumers), so `kk upgrade cluster
// --with-kubernetes v1.34.3` from v1.23.17 will pass precheck but kubeadm apply
// v1.34.3 will be rejected by kubeadm itself — that gap is tracked separately.
// ----------------------------------------------------------------------------
func TestPrecheckUpgradePathAssertion(t *testing.T) {
	expr := `{{ .kubernetes_install_version | default "" | trimPrefix "Kubernetes " | semverCompare (printf "<%s" .kubernetes.kube_version) }}`

	tests := []struct {
		name     string
		ctx      map[string]any
		wantPass bool
	}{
		{
			name:     "valid upgrade 1.23.17 -> 1.34.3",
			ctx:      map[string]any{"kubernetes_install_version": "v1.23.17", "kubernetes": map[string]any{"kube_version": "v1.34.3"}},
			wantPass: true,
		},
		{
			name:     "same version rejected",
			ctx:      map[string]any{"kubernetes_install_version": "v1.34.3", "kubernetes": map[string]any{"kube_version": "v1.34.3"}},
			wantPass: false,
		},
		{
			name:     "downgrade rejected",
			ctx:      map[string]any{"kubernetes_install_version": "v1.34.3", "kubernetes": map[string]any{"kube_version": "v1.33.0"}},
			wantPass: false,
		},
		{
			name:     "kubelet version string with Kubernetes prefix",
			ctx:      map[string]any{"kubernetes_install_version": "Kubernetes v1.28.15", "kubernetes": map[string]any{"kube_version": "v1.29.15"}},
			wantPass: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tmpl.ParseBool(tt.ctx, expr)
			require.NoError(t, err)
			assert.Equal(t, tt.wantPass, got, "升级路径断言结果不符合预期")
		})
	}
}
