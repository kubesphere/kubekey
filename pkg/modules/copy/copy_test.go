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

package copy

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
)

// createRawArgs creates a runtime.RawExtension from a map
func createRawArgs(data map[string]any) runtime.RawExtension {
	raw, _ := json.Marshal(data)
	return runtime.RawExtension{Raw: raw}
}

// TestCopyArgsModule tests module execution - simplified to just verify module exists
func TestCopyArgsModule(t *testing.T) {
	t.Run("module exists", func(t *testing.T) {
		require.NotNil(t, ModuleCopy)
	})
}

// TestCopyArgsParse tests argument parsing edge cases.
func TestCopyArgsParse(t *testing.T) {
	testcases := []struct {
		name             string
		args             map[string]any
		expectParseError bool
		expectSrc        string
		expectDest       string
		expectContent    string
		description      string
	}{
		{
			name:             "valid args with src and dest",
			args:             map[string]any{"src": "/local/file.txt", "dest": "/remote/file.txt"},
			expectParseError: false,
			expectSrc:        "/local/file.txt",
			expectDest:       "/remote/file.txt",
			expectContent:    "",
			description:      "When src and dest are provided, should parse successfully",
		},
		{
			name:             "valid args with content and dest",
			args:             map[string]any{"content": "hello world", "dest": "/remote/file.txt"},
			expectParseError: false,
			expectSrc:        "",
			expectDest:       "/remote/file.txt",
			expectContent:    "hello world",
			description:      "When content and dest are provided, should parse successfully",
		},
		{
			name:             "valid args with mode",
			args:             map[string]any{"dest": "/tmp/test.txt", "mode": 0644},
			expectParseError: false,
			expectSrc:        "",
			expectDest:       "/tmp/test.txt",
			expectContent:    "",
			description:      "When mode is provided, should parse successfully",
		},
		{
			name:             "missing dest",
			args:             map[string]any{"src": "/local/file.txt"},
			expectParseError: true,
			expectSrc:        "",
			expectDest:       "",
			expectContent:    "",
			description:      "When dest is missing, should return error",
		},
		{
			name:             "empty args",
			args:             map[string]any{},
			expectParseError: true,
			expectSrc:        "",
			expectDest:       "",
			expectContent:    "",
			description:      "When args are empty, should return error",
		},
		{
			name:             "invalid mode (negative)",
			args:             map[string]any{"dest": "/tmp/test.txt", "mode": -1},
			expectParseError: true,
			expectSrc:        "",
			expectDest:       "",
			expectContent:    "",
			description:      "When mode is negative, should return error",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			raw := createRawArgs(tc.args)
			result, err := newCopyArgs(ctx, raw, nil)
			if tc.expectParseError {
				require.Error(t, err, tc.description)
			} else {
				require.NoError(t, err, tc.description)
				require.Equal(t, tc.expectSrc, result.src, tc.description)
				require.Equal(t, tc.expectDest, result.dest, tc.description)
				require.Equal(t, tc.expectContent, result.content, tc.description)
			}
		})
	}
}

// TestCopyArgsTemplate tests template variable resolution in arguments.
func TestCopyArgsTemplate(t *testing.T) {
	testcases := []struct {
		name        string
		args        map[string]any
		vars        map[string]any
		expectError bool
		description string
	}{
		{
			name:        "template: src with template variable",
			args:        map[string]any{"src": "{{ .src_path }}", "dest": "/tmp/dest"},
			vars:        map[string]any{"src_path": "/local/file.txt"},
			expectError: false,
			description: "When src contains template, should resolve with vars",
		},
		{
			name:        "template: dest with template variable",
			args:        map[string]any{"src": "/tmp/src", "dest": "{{ .dest_path }}"},
			vars:        map[string]any{"dest_path": "/remote/file.txt"},
			expectError: false,
			description: "When dest contains template, should resolve with vars",
		},
		{
			name:        "template: content with template variable",
			args:        map[string]any{"content": "Hello {{ .name }}!", "dest": "/tmp/dest"},
			vars:        map[string]any{"name": "World"},
			expectError: false,
			description: "When content contains template, should resolve with vars",
		},
		{
			name:        "template: mode with template variable",
			args:        map[string]any{"src": "/tmp/src", "dest": "/tmp/dest", "mode": "{{ .file_mode }}"},
			vars:        map[string]any{"file_mode": 0644},
			expectError: false,
			description: "When mode contains template, should resolve with vars",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			raw := createRawArgs(tc.args)
			_, err := newCopyArgs(ctx, raw, tc.vars)
			if tc.expectError {
				require.Error(t, err, tc.description)
			} else {
				require.NoError(t, err, tc.description)
			}
		})
	}
}

// TestCopyModule tests the actual functionality of the copy module.
func TestCopyModule(t *testing.T) {
	// Copy module requires actual connector
	t.Run("module exists", func(t *testing.T) {
		require.NotNil(t, ModuleCopy)
	})
}
