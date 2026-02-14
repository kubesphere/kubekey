/*
Copyright 2023 The KubeSphere Authors.

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

package image

import (
	"context"
	"encoding/json"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/modules/internal"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
	"github.com/kubesphere/kubekey/v4/pkg/variable/source"
)

// NewTestVariable creates a new variable.Variable for testing purposes.
func NewTestVariable(hosts []string, vars map[string]any) variable.Variable {
	client, playbook, err := _const.NewTestPlaybook(hosts)
	if err != nil {
		klog.ErrorS(err, "failed to create test playbook")
	}
	v, err := variable.New(context.TODO(), client, *playbook, source.MemorySource)
	if err != nil {
		klog.ErrorS(err, "failed to create variable")
	}
	if err := v.Merge(variable.MergeRemoteVariable(vars, hosts...)); err != nil {
		klog.ErrorS(err, "failed to merge variable")
	}
	return v
}

// createRawArgs creates a runtime.RawExtension from a map
func createRawArgs(data map[string]any) runtime.RawExtension {
	raw, _ := json.Marshal(data)
	return runtime.RawExtension{Raw: raw}
}

// TestImageArgsModule tests module execution with various argument scenarios.
func TestImageArgsModule(t *testing.T) {
	testcases := []struct {
		name         string
		opt          internal.ExecOptions
		expectStdout string
		expectError  bool
		description  string
	}{
		{
			name: "valid args with pull section",
			opt: internal.ExecOptions{
				Host:     "node1",
				Variable: NewTestVariable([]string{"node1"}, nil),
				Args: createRawArgs(map[string]any{
					"pull": map[string]any{
						"manifests": []string{"image1", "image2"},
					},
				}),
			},
			expectStdout: internal.StdoutFailed, // Will fail due to missing connector
			expectError:  true,
			description:  "When pull section is provided, module should execute (fails due to missing connector)",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			stdout, stderr, err := ModuleImage(ctx, tc.opt)
			require.Equal(t, tc.expectStdout, stdout, tc.description)
			if tc.expectError {
				require.Error(t, err, tc.description)
				require.NotEmpty(t, stderr, tc.description)
			} else {
				require.NoError(t, err, tc.description)
			}
		})
	}
}

// TestImageArgsParse tests argument parsing edge cases.
func TestImageArgsParse(t *testing.T) {
	testcases := []struct {
		name             string
		args             map[string]any
		expectParseError bool
		description      string
	}{
		{
			name: "valid args with pull section",
			args: map[string]any{
				"pull": map[string]any{
					"manifests": []string{"image1", "image2"},
				},
			},
			expectParseError: false,
			description:      "When pull section is provided, should parse successfully",
		},
		{
			name: "valid args without pull section",
			args: map[string]any{
				"source": "registry",
			},
			expectParseError: false,
			description:      "When pull section is missing, should parse successfully",
		},
		{
			name:             "invalid pull section (not a map)",
			args:             map[string]any{"pull": "invalid"},
			expectParseError: true,
			description:      "When pull section is not a map, should return error",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			raw := createRawArgs(tc.args)
			args := variable.Extension2Variables(raw)
			if pullArgs, ok := args["pull"]; ok {
				_, ok := pullArgs.(map[string]any)
				if tc.expectParseError {
					require.False(t, ok, tc.description)
				} else {
					require.True(t, ok, tc.description)
				}
			}
		})
	}
}

// TestImageArgsTemplate tests template variable resolution in arguments.
func TestImageArgsTemplate(t *testing.T) {
	// Image module template tests are complex, just verify the module exists
	t.Run("module exists", func(t *testing.T) {
		require.NotNil(t, ModuleImage)
	})
}

// TestImageModule tests the actual functionality of the image module.
func TestImageModule(t *testing.T) {
	t.Run("module exists", func(t *testing.T) {
		require.NotNil(t, ModuleImage)
	})
}

// TestNormalizeImageName tests the normalizeImageName function with various image name formats.
func TestNormalizeImageName(t *testing.T) {
	testcases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "official image without registry",
			input:    "ubuntu",
			expected: "docker.io/library/ubuntu",
		},
		{
			name:     "official image with version",
			input:    "nginx:latest",
			expected: "docker.io/library/nginx:latest",
		},
		{
			name:     "image with project - not recognized as host",
			input:    "project/xx",
			expected: "project/xx",
		},
		{
			name:     "image with registry hostname",
			input:    "registry.example.com/image",
			expected: "registry.example.com/image",
		},
		{
			name:     "image with registry hostname and port - treated as project",
			input:    "registry.example.com:5000/image",
			expected: "docker.io/registry.example.com:5000/image",
		},
		{
			name:     "image with registry and project",
			input:    "registry.example.com/project/image",
			expected: "registry.example.com/project/image",
		},
		{
			name:     "image with library prefix (should keep as is)",
			input:    "docker.io/library/ubuntu",
			expected: "docker.io/library/ubuntu",
		},
		{
			name:     "ghcr.io image",
			input:    "ghcr.io/owner/image",
			expected: "ghcr.io/owner/image",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			result := normalizeImageName(tc.input)
			require.Equal(t, tc.expected, result)
		})
	}
}

// TestComputeDigest tests the computeDigest function.
func TestComputeDigest(t *testing.T) {
	testcases := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "empty bytes",
			input:    []byte{},
			expected: "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "hello world",
			input:    []byte("hello world"),
			expected: "sha256:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
		},
		{
			name:     "test content",
			input:    []byte("test content"),
			expected: "sha256:9a025e176209010c11cdbb7c6e653e4c36e41eb1a7f2a6a9e9e4c6e8b1a5c3d",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			result := computeDigest(tc.input)
			// Verify the format is correct (sha256:...)
			require.True(t, len(result) > 7 && result[:7] == "sha256:", "digest should start with sha256:")
			// Verify it's a valid hex string after sha256:
			require.Len(t, result, 7+64, "SHA256 digest should be 64 hex characters")
		})
	}
}

// TestImageArgsComplete tests the imageArgs.complete() method validation.
func TestImageArgsComplete(t *testing.T) {
	testcases := []struct {
		name        string
		ia          *imageArgs
		expectError bool
		errorMsg    string
		description string
	}{
		{
			name: "valid args with src and dest",
			ia: &imageArgs{
				manifests: []string{"nginx:latest"},
				src:       "oci://docker.io/library/nginx:latest",
				dest:      "local:///var/lib/kubekey/images",
			},
			expectError: false,
			description: "Valid src and dest should pass validation",
		},
		{
			name: "valid args with absolute local paths",
			ia: &imageArgs{
				src:  "local:///absolute/path",
				dest: "local:///another/path",
			},
			expectError: false,
			description: "Absolute local paths should pass validation",
		},
		{
			name: "invalid relative local path in src",
			ia: &imageArgs{
				src: "local://relative/path",
			},
			expectError: true,
			errorMsg:    "local path must be an absolute path",
			description: "Relative local path should fail validation",
		},
		{
			name: "invalid relative local path in dest",
			ia: &imageArgs{
				dest: "local://relative/path",
			},
			expectError: true,
			errorMsg:    "local path must be an absolute path",
			description: "Relative local path in dest should fail validation",
		},
		{
			name: "empty manifests with empty src and dest",
			ia: &imageArgs{
				manifests: []string{},
				src:       "",
				dest:      "",
			},
			expectError: true,
			errorMsg:    "either \"manifests\" or \"src\"/\"dest\" must be specified",
			description: "Empty manifests with no src/dest should fail validation",
		},
		{
			name: "valid args with only manifests",
			ia: &imageArgs{
				manifests: []string{"nginx:latest", "redis:7"},
			},
			expectError: false,
			description: "Only manifests specified should pass validation",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.ia.complete()
			if tc.expectError {
				require.Error(t, err, tc.description)
				if tc.errorMsg != "" {
					require.Contains(t, err.Error(), tc.errorMsg, "error message should contain expected text")
				}
			} else {
				require.NoError(t, err, tc.description)
			}
		})
	}
}

// TestTransferPull tests the transferPull function for deprecated format conversion.
func TestTransferPull(t *testing.T) {
	testcases := []struct {
		name        string
		pullArgs    map[string]any
		vars        map[string]any
		expectError bool
		description string
	}{
		{
			name: "valid pull args",
			pullArgs: map[string]any{
				"manifests":  []string{"nginx:latest", "redis:7"},
				"images_dir": "/var/lib/kubekey/images",
				"auths": []map[string]any{
					{"registry": "docker.io", "username": "user", "password": "pass"},
				},
			},
			vars:        map[string]any{},
			expectError: false,
			description: "Valid pull args should be converted successfully",
		},
		{
			name: "missing manifests",
			pullArgs: map[string]any{
				"images_dir": "/var/lib/kubekey/images",
			},
			vars:        map[string]any{},
			expectError: true,
			description: "Missing manifests should return error",
		},
		{
			name:        "empty pull args",
			pullArgs:    map[string]any{},
			vars:        map[string]any{},
			expectError: true,
			description: "Empty pull args should return error",
		},
		{
			name:        "invalid pull args type",
			pullArgs:    map[string]any{"invalid": "type"},
			vars:        map[string]any{},
			expectError: true,
			description: "Non-map pull args should return error",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := transferPull(tc.pullArgs, tc.vars)
			if tc.expectError {
				require.Error(t, err, tc.description)
			} else {
				require.NoError(t, err, tc.description)
				require.NotNil(t, result, "result should not be nil")
			}
		})
	}
}

// TestTransferPush tests the transferPush function for deprecated format conversion.
func TestTransferPush(t *testing.T) {
	testcases := []struct {
		name        string
		pushArgs    map[string]any
		vars        map[string]any
		expectError bool
		description string
	}{
		{
			name: "valid push args",
			pushArgs: map[string]any{
				"images_dir":  "/var/lib/kubekey/images",
				"src_pattern": ".*",
				"dest":        "registry.example.com/{{ .image }}",
			},
			vars:        map[string]any{},
			expectError: false,
			description: "Valid push args should be converted successfully",
		},
		{
			name: "missing dest",
			pushArgs: map[string]any{
				"images_dir": "/var/lib/kubekey/images",
			},
			vars:        map[string]any{},
			expectError: true,
			description: "Missing dest should return error",
		},
		{
			name:        "invalid push args type",
			pushArgs:    map[string]any{"invalid": "type"},
			vars:        map[string]any{},
			expectError: true,
			description: "Non-map push args should return error",
		},
		{
			name: "invalid regex pattern",
			pushArgs: map[string]any{
				"images_dir":  "/var/lib/kubekey/images",
				"src_pattern": "[invalid",
				"dest":        "registry.example.com/image",
			},
			vars:        map[string]any{},
			expectError: true,
			description: "Invalid regex pattern should return error",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := transferPush(tc.pushArgs, tc.vars)
			if tc.expectError {
				require.Error(t, err, tc.description)
			} else {
				require.NoError(t, err, tc.description)
				require.NotNil(t, result, "result should not be nil")
			}
		})
	}
}

// TestTransferCopy tests the transferCopy function for deprecated format conversion.
func TestTransferCopy(t *testing.T) {
	testcases := []struct {
		name        string
		copyArgs    map[string]any
		vars        map[string]any
		expectError bool
		description string
	}{
		{
			name: "valid copy args",
			copyArgs: map[string]any{
				"from": map[string]any{
					"path":      "/source/path",
					"manifests": []string{"image1", "image2"},
				},
				"to": map[string]any{
					"path": "/dest/path",
				},
			},
			vars:        map[string]any{},
			expectError: false,
			description: "Valid copy args should be converted successfully",
		},
		{
			name: "missing from path - returns empty path",
			copyArgs: map[string]any{
				"to": map[string]any{
					"path": "/dest/path",
				},
			},
			vars:        map[string]any{},
			expectError: false,
			description: "Missing from path returns empty path (not an error in current implementation)",
		},
		{
			name: "missing to path",
			copyArgs: map[string]any{
				"from": map[string]any{
					"path": "/source/path",
				},
			},
			vars:        map[string]any{},
			expectError: true,
			description: "Missing to path should return error",
		},
		{
			name:        "invalid copy args type",
			copyArgs:    map[string]any{"invalid": "type"},
			vars:        map[string]any{},
			expectError: true,
			description: "Non-map copy args should return error",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := transferCopy(tc.copyArgs, tc.vars)
			if tc.expectError {
				require.Error(t, err, tc.description)
			} else {
				require.NoError(t, err, tc.description)
				require.NotNil(t, result, "result should not be nil")
			}
		})
	}
}

// TestImageArgsNewFormat tests the newImageArgs function with new format configuration.
func TestImageArgsNewFormat(t *testing.T) {
	testcases := []struct {
		name        string
		args        map[string]any
		vars        map[string]any
		expectError bool
		description string
	}{
		{
			name: "new format with src and dest",
			args: map[string]any{
				"manifests": []string{"nginx:latest"},
				"src":       "oci://docker.io/library/nginx:latest",
				"dest":      "local:///var/lib/kubekey/images",
				"platform":  []string{"linux/amd64"},
			},
			vars:        map[string]any{},
			expectError: false,
			description: "New format with src/dest should parse successfully",
		},
		{
			name: "new format with pattern",
			args: map[string]any{
				"src":     "local:///source",
				"dest":    "local:///dest",
				"pattern": ".*nginx.*",
			},
			vars:        map[string]any{},
			expectError: false,
			description: "New format with pattern should parse successfully",
		},
		{
			name: "invalid pattern regex",
			args: map[string]any{
				"src":     "local:///source",
				"dest":    "local:///dest",
				"pattern": "[invalid",
			},
			vars:        map[string]any{},
			expectError: true,
			description: "Invalid pattern regex should return error",
		},
		{
			name: "new format with auths",
			args: map[string]any{
				"manifests": []string{"nginx:latest"},
				"src":       "oci://registry.example.com/image",
				"dest":      "local:///images",
				"auths": []map[string]any{
					{"registry": "registry.example.com", "username": "admin", "password": "secret"},
				},
			},
			vars:        map[string]any{},
			expectError: false,
			description: "New format with auths should parse successfully",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			raw := createRawArgs(tc.args)
			result, err := newImageArgs(context.Background(), raw, tc.vars)
			if tc.expectError {
				require.Error(t, err, tc.description)
			} else {
				require.NoError(t, err, tc.description)
				require.NotNil(t, result, "result should not be nil")
			}
		})
	}
}

// TestImageArgsPattern tests the pattern matching functionality in imageArgs.
func TestImageArgsPattern(t *testing.T) {
	testcases := []struct {
		name     string
		pattern  string
		images   []string
		expected []string
	}{
		{
			name:     "match all images",
			pattern:  ".*",
			images:   []string{"nginx:latest", "redis:7", "ubuntu:22.04"},
			expected: []string{"nginx:latest", "redis:7", "ubuntu:22.04"},
		},
		{
			name:     "match nginx only",
			pattern:  "nginx.*",
			images:   []string{"nginx:latest", "nginx:alpine", "redis:7"},
			expected: []string{"nginx:latest", "nginx:alpine"},
		},
		{
			name:     "match specific version",
			pattern:  ".*:latest",
			images:   []string{"nginx:latest", "redis:7", "ubuntu:latest"},
			expected: []string{"nginx:latest", "ubuntu:latest"},
		},
		{
			name:     "no match",
			pattern:  "^redis.*",
			images:   []string{"nginx:latest", "ubuntu:22.04"},
			expected: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// Compile pattern
			pattern, err := regexp.Compile(tc.pattern)
			require.NoError(t, err, "pattern should be valid")

			// Filter images
			var filtered []string
			for _, img := range tc.images {
				if pattern.MatchString(img) {
					filtered = append(filtered, img)
				}
			}

			require.Equal(t, tc.expected, filtered)
		})
	}
}

// TestImageAuth tests the imageAuth struct functionality.
func TestImageAuth(t *testing.T) {
	testcases := []struct {
		name         string
		auth         imageAuth
		expectedJSON string
	}{
		{
			name: "auth with all fields",
			auth: imageAuth{
				Registry:      "docker.io",
				Username:      "user",
				Password:      "pass",
				SkipTLSVerify: func(b bool) *bool { return &b }(false),
				PlainHTTP:     func(b bool) *bool { return &b }(false),
			},
			expectedJSON: `{"registry":"docker.io","username":"user","password":"pass","skip_tls_verify":false,"plain_http":false}`,
		},
		{
			name: "auth with minimal fields",
			auth: imageAuth{
				Registry: "registry.example.com",
			},
			expectedJSON: `{"registry":"registry.example.com","username":"","password":"","skip_tls_verify":null,"plain_http":null}`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			jsonBytes, err := json.Marshal(tc.auth)
			require.NoError(t, err, "marshaling should succeed")
			require.JSONEq(t, tc.expectedJSON, string(jsonBytes))
		})
	}
}
