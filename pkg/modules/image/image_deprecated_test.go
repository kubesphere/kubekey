/*
Copyright 2024 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT ANY KIND, either WARRANTIES OR CONDITIONS OF express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package image

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestTransferPullArgs tests the transferPull function for parameter conversion.
func TestTransferPullArgs(t *testing.T) {
	testcases := []struct {
		name        string
		args        map[string]any
		vars        map[string]any
		expectError bool
		expect      *imageArgs
	}{
		{
			name: "valid pull args with manifests",
			args: map[string]any{
				"manifests":  []string{"nginx:latest", "redis:7"},
				"images_dir": "/var/lib/kubekey/images",
			},
			vars:        map[string]any{},
			expectError: false,
			expect: &imageArgs{
				manifests: []string{"nginx:latest", "redis:7"},
				src:       "oci://{{ .module.image.reference }}/{{ .module.image.reference.repository }}:{{ .module.image.reference.reference }}",
				dest:      "local:///var/lib/kubekey/images",
			},
		},
		{
			name: "pull args with platform",
			args: map[string]any{
				"manifests":  []string{"nginx:latest"},
				"images_dir": "/var/lib/kubekey/images",
				"platform":   []string{"linux/amd64", "linux/arm64"},
			},
			vars:        map[string]any{},
			expectError: false,
			expect: &imageArgs{
				manifests: []string{"nginx:latest"},
				platform:  []string{"linux/amd64", "linux/arm64"},
				src:       "oci://{{ .module.image.reference }}/{{ .module.image.reference.repository }}:{{ .module.image.reference.reference }}",
				dest:      "local:///var/lib/kubekey/images",
			},
		},
		{
			name: "missing manifests should error",
			args: map[string]any{
				"images_dir": "/var/lib/kubekey/images",
			},
			vars:        map[string]any{},
			expectError: true,
			expect:      nil,
		},
		{
			name: "pull args not a map",
			args: map[string]any{
				"pull": "not a map",
			},
			vars:        map[string]any{},
			expectError: true,
			expect:      nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := transferPull(tc.args, tc.vars)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expect.manifests, result.manifests, "manifests mismatch")
				require.Equal(t, tc.expect.src, result.src, "src mismatch")
				require.Equal(t, tc.expect.dest, result.dest, "dest mismatch")
				require.Equal(t, tc.expect.platform, result.platform, "platform mismatch")
			}
		})
	}
}

// TestTransferPushArgs tests the transferPush function for parameter conversion.
func TestTransferPushArgs(t *testing.T) {
	testcases := []struct {
		name        string
		args        map[string]any
		vars        map[string]any
		expectError bool
		expect      *imageArgs
	}{
		{
			name: "valid push args",
			args: map[string]any{
				"images_dir": "/var/lib/kubekey/images",
				"dest":       "registry.example.com/images",
			},
			vars:        map[string]any{},
			expectError: false,
			expect: &imageArgs{
				src:  "local:///var/lib/kubekey/images",
				dest: "oci://registry.example.com/images",
			},
		},
		{
			name: "push args with src_pattern",
			args: map[string]any{
				"images_dir":  "/var/lib/kubekey/images",
				"dest":        "registry.example.com/images",
				"src_pattern": "nginx.*",
			},
			vars:        map[string]any{},
			expectError: false,
			expect: &imageArgs{
				src:     "local:///var/lib/kubekey/images",
				dest:    "oci://registry.example.com/images",
				pattern: regexp.MustCompile("nginx.*"),
			},
		},
		{
			name: "missing dest should error",
			args: map[string]any{
				"images_dir": "/var/lib/kubekey/images",
			},
			vars:        map[string]any{},
			expectError: true,
			expect:      nil,
		},
		{
			name: "invalid src_pattern should error",
			args: map[string]any{
				"images_dir":  "/var/lib/kubekey/images",
				"dest":        "registry.example.com/images",
				"src_pattern": "[invalid",
			},
			vars:        map[string]any{},
			expectError: true,
			expect:      nil,
		},
		{
			name: "push args not a map",
			args: map[string]any{
				"push": "not a map",
			},
			vars:        map[string]any{},
			expectError: true,
			expect:      nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := transferPush(tc.args, tc.vars)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expect.src, result.src, "src mismatch")
				require.Equal(t, tc.expect.dest, result.dest, "dest mismatch")

				// Compare pattern
				if tc.expect.pattern == nil {
					require.Nil(t, result.pattern, "pattern should be nil")
				} else {
					require.NotNil(t, result.pattern, "pattern should not be nil")
					require.Equal(t, tc.expect.pattern.String(), result.pattern.String(), "pattern mismatch")
				}
			}
		})
	}
}

// TestTransferCopyArgs tests the transferCopy function for parameter conversion.
func TestTransferCopyArgs(t *testing.T) {
	testcases := []struct {
		name        string
		args        map[string]any
		vars        map[string]any
		expectError bool
		expect      *imageArgs
	}{
		{
			name: "valid copy args",
			args: map[string]any{
				"from": map[string]any{"path": "/source/images"},
				"to":   map[string]any{"path": "/dest/images"},
			},
			vars:        map[string]any{},
			expectError: false,
			expect: &imageArgs{
				src:  "local:///source/images",
				dest: "local:///dest/images",
			},
		},
		{
			name: "copy args with from manifests",
			args: map[string]any{
				"from": map[string]any{
					"path":      "/source/images",
					"manifests": []string{"nginx:latest", "redis:7"},
				},
				"to": map[string]any{"path": "/dest/images"},
			},
			vars:        map[string]any{},
			expectError: false,
			expect: &imageArgs{
				src:       "local:///source/images",
				dest:      "local:///dest/images",
				manifests: []string{"nginx:latest", "redis:7"},
			},
		},
		{
			name: "copy args with platform",
			args: map[string]any{
				"platform": []string{"linux/amd64"},
				"from":     map[string]any{"path": "/source/images"},
				"to":       map[string]any{"path": "/dest/images"},
			},
			vars:        map[string]any{},
			expectError: false,
			expect: &imageArgs{
				src:      "local:///source/images",
				dest:     "local:///dest/images",
				platform: []string{"linux/amd64"},
			},
		},
		{
			name: "copy args with to src_pattern",
			args: map[string]any{
				"from": map[string]any{"path": "/source/images"},
				"to": map[string]any{
					"path":        "/dest/images",
					"src_pattern": "nginx.*",
				},
			},
			vars:        map[string]any{},
			expectError: false,
			expect: &imageArgs{
				src:     "local:///source/images",
				dest:    "local:///dest/images",
				pattern: regexp.MustCompile("nginx.*"),
			},
		},
		{
			name: "missing to.path should error",
			args: map[string]any{
				"from": map[string]any{"path": "/source/images"},
				"to":   map[string]any{},
			},
			vars:        map[string]any{},
			expectError: true,
			expect:      nil,
		},
		{
			name: "invalid to.src_pattern should error",
			args: map[string]any{
				"from": map[string]any{"path": "/source/images"},
				"to": map[string]any{
					"path":        "/dest/images",
					"src_pattern": "[invalid",
				},
			},
			vars:        map[string]any{},
			expectError: true,
			expect:      nil,
		},
		{
			name: "copy args not a map",
			args: map[string]any{
				"copy": "not a map",
			},
			vars:        map[string]any{},
			expectError: true,
			expect:      nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := transferCopy(tc.args, tc.vars)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expect.src, result.src, "src mismatch")
				require.Equal(t, tc.expect.dest, result.dest, "dest mismatch")
				require.Equal(t, tc.expect.platform, result.platform, "platform mismatch")
				require.Equal(t, tc.expect.manifests, result.manifests, "manifests mismatch")

				// Compare pattern
				if tc.expect.pattern == nil {
					require.Nil(t, result.pattern, "pattern should be nil")
				} else {
					require.NotNil(t, result.pattern, "pattern should not be nil")
					require.Equal(t, tc.expect.pattern.String(), result.pattern.String(), "pattern mismatch")
				}
			}
		})
	}
}
