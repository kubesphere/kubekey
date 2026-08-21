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
// machinery. upgrade_path_test.go validates the embedded cluster_require
// defaults (kube_upgrade_path) via the package's exported API, so it is kept
// OUT of builtin/core and run as an external test package against the embedded
// BuiltinPlaybook.
package core_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	core "github.com/kubesphere/kubekey/v4/builtin/core"
)

// TestParseKubeUpgradePath verifies the embedded upgrade path is parsed into a
// minor -> highest-patch map covering v1.23 .. v1.34.
func TestParseKubeUpgradePath(t *testing.T) {
	path, err := core.ParseKubeUpgradePath()
	require.NoError(t, err)

	want := map[int]string{
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
	assert.Len(t, path, len(want), "upgrade path should cover 12 minors")
	for minor, patch := range want {
		got, ok := path[minor]
		require.True(t, ok, "missing minor v1.%d in upgrade path", minor)
		assert.Equal(t, patch, got, "wrong patch for minor v1.%d", minor)
	}
}

// TestComputeUpgradeSteps locks in the auto-step splitting logic.
func TestComputeUpgradeSteps(t *testing.T) {
	path := core.KubeUpgradePath{
		23: "v1.23.17", 24: "v1.24.17", 25: "v1.25.16", 26: "v1.26.15",
		27: "v1.27.16", 28: "v1.28.15", 29: "v1.29.15", 30: "v1.30.14",
		31: "v1.31.14", 32: "v1.32.13", 33: "v1.33.7", 34: "v1.34.3",
	}

	tests := []struct {
		name            string
		currentMinor    int
		targetMinor     int
		requestedTarget string
		want            []string
		wantErr         bool
	}{
		{
			// The headline scenario: v1.23.17 -> v1.34.3 must be split into 11
			// single-minor hops, each intermediate using the path's highest patch
			// and the final hop using the user-requested version.
			name:            "multi-minor 1.23 -> 1.34",
			currentMinor:    23,
			targetMinor:     34,
			requestedTarget: "v1.34.3",
			want: []string{
				"v1.24.17", "v1.25.16", "v1.26.15", "v1.27.16", "v1.28.15",
				"v1.29.15", "v1.30.14", "v1.31.14", "v1.32.13", "v1.33.7", "v1.34.3",
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
			requestedTarget: "v1.33.7",
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
			got, err := core.ComputeUpgradeSteps(tt.currentMinor, tt.targetMinor, tt.requestedTarget, path)
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
