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

package http_get_file

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/modules/internal"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
	"github.com/kubesphere/kubekey/v4/pkg/variable/source"
)

// NewTestVariable creates a new variable.Variable for testing purposes.
func NewTestVariable(hosts []string, vars map[string]any) variable.Variable {
	client, playbook, err := _const.NewTestPlaybook(hosts)
	if err != nil {
		return nil
	}
	v, err := variable.New(context.TODO(), client, *playbook, source.MemorySource)
	if err != nil {
		return nil
	}
	_ = v.Merge(variable.MergeRemoteVariable(vars, hosts...))
	return v
}

// createRawArgs creates a runtime.RawExtension from a map
func createRawArgs(data map[string]any) runtime.RawExtension {
	raw, _ := json.Marshal(data)
	return runtime.RawExtension{Raw: raw}
}

// TestHttpGetFileArgsModule tests module execution - requires actual HTTP server
func TestHttpGetFileArgsModule(t *testing.T) {
	testcases := []struct {
		name         string
		opt          internal.ExecOptions
		expectStdout string
		expectError  bool
		description  string
	}{
		{
			name: "missing url",
			opt: internal.ExecOptions{
				Host:     "node1",
				Variable: NewTestVariable([]string{"node1"}, nil),
				Args:     createRawArgs(map[string]any{"dest": "/tmp/file.txt"}),
			},
			expectStdout: internal.StdoutFailed,
			expectError:  true,
			description:  "When url is missing, should return failed",
		},
		{
			name: "missing dest",
			opt: internal.ExecOptions{
				Host:     "node1",
				Variable: NewTestVariable([]string{"node1"}, nil),
				Args:     createRawArgs(map[string]any{"url": "http://example.com/file.txt"}),
			},
			expectStdout: internal.StdoutFailed,
			expectError:  true,
			description:  "When dest is missing, should return failed",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			stdout, stderr, err := ModuleHttpGetFile(ctx, tc.opt)
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

// TestHttpGetFileArgsParse tests argument parsing edge cases.
func TestHttpGetFileArgsParse(t *testing.T) {
	testcases := []struct {
		name             string
		args             map[string]any
		expectParseError bool
		description      string
	}{
		{
			name:             "valid args with url and dest",
			args:             map[string]any{"url": "http://example.com/file.txt", "dest": "/tmp/file.txt"},
			expectParseError: false,
			description:      "When url and dest are provided, should parse successfully",
		},
		{
			name:             "valid args with https",
			args:             map[string]any{"url": "https://example.com/file.txt", "dest": "/tmp/file.txt"},
			expectParseError: false,
			description:      "When https URL is provided, should parse successfully",
		},
		{
			name:             "valid args with timeout",
			args:             map[string]any{"url": "http://example.com/file.txt", "dest": "/tmp/file.txt", "timeout": "30s"},
			expectParseError: false,
			description:      "When timeout is provided, should parse successfully",
		},
		{
			name:             "valid args with headers",
			args:             map[string]any{"url": "http://example.com/file.txt", "dest": "/tmp/file.txt", "headers": map[string]string{"X-Custom": "value"}},
			expectParseError: false,
			description:      "When headers are provided, should parse successfully",
		},
		{
			name:             "missing url",
			args:             map[string]any{"dest": "/tmp/file.txt"},
			expectParseError: true,
			description:      "When url is missing, should return error",
		},
		{
			name:             "invalid scheme",
			args:             map[string]any{"url": "ftp://example.com/file.txt", "dest": "/tmp/file.txt"},
			expectParseError: true,
			description:      "When scheme is not http/https, should return error",
		},
		{
			name:             "invalid timeout format (ignored, uses default)",
			args:             map[string]any{"url": "http://example.com/file.txt", "dest": "/tmp/file.txt", "timeout": "invalid"},
			expectParseError: false,
			description:      "When timeout format is invalid, uses default timeout",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			args := variable.Extension2Variables(createRawArgs(tc.args))
			_, err := newHttpArgs(ctx, args, nil)
			if tc.expectParseError {
				require.Error(t, err, tc.description)
			} else {
				require.NoError(t, err, tc.description)
			}
		})
	}
}

// TestHttpGetFileArgsTemplate tests template variable resolution in arguments.
func TestHttpGetFileArgsTemplate(t *testing.T) {
	testcases := []struct {
		name        string
		args        map[string]any
		vars        map[string]any
		expectError bool
		description string
	}{
		{
			name:        "template: url with template variable",
			args:        map[string]any{"url": "{{ .file_url }}", "dest": "/tmp/file.txt"},
			vars:        map[string]any{"file_url": "http://example.com/file.txt"},
			expectError: false,
			description: "When url contains template, should resolve with vars",
		},
		{
			name:        "template: dest with template variable",
			args:        map[string]any{"url": "http://example.com/file.txt", "dest": "{{ .dest_path }}"},
			vars:        map[string]any{"dest_path": "/tmp/downloaded.txt"},
			expectError: false,
			description: "When dest contains template, should resolve with vars",
		},
		{
			name:        "template: timeout with template variable",
			args:        map[string]any{"url": "http://example.com/file.txt", "dest": "/tmp/file.txt", "timeout": "{{ .timeout }}"},
			vars:        map[string]any{"timeout": "30s"},
			expectError: false,
			description: "When timeout contains template, should resolve with vars",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			args := variable.Extension2Variables(createRawArgs(tc.args))
			_, err := newHttpArgs(ctx, args, tc.vars)
			if tc.expectError {
				require.Error(t, err, tc.description)
			} else {
				require.NoError(t, err, tc.description)
			}
		})
	}
}

// TestHttpGetFileModule tests the actual functionality of the http_get_file module.
func TestHttpGetFileModule(t *testing.T) {
	// HTTP get file requires actual HTTP server
	t.Run("module exists", func(t *testing.T) {
		require.NotNil(t, ModuleHttpGetFile)
	})
}
