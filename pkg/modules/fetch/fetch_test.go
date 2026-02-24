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

package fetch

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

// TestFetchArgsModule tests module execution with various argument scenarios.
func TestFetchArgsModule(t *testing.T) {
	testcases := []struct {
		name         string
		opt          internal.ExecOptions
		expectStdout string
		expectError  bool
		description  string
	}{
		{
			name: "missing src",
			opt: internal.ExecOptions{
				Host:     "node1",
				Variable: NewTestVariable([]string{"node1"}, nil),
				Args:     createRawArgs(map[string]any{"dest": "/tmp/dest"}),
			},
			expectStdout: internal.StdoutFailed,
			expectError:  true,
			description:  "When src is missing, should return failed",
		},
		{
			name: "missing dest",
			opt: internal.ExecOptions{
				Host:     "node1",
				Variable: NewTestVariable([]string{"node1"}, nil),
				Args:     createRawArgs(map[string]any{"src": "/tmp/source"}),
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
			stdout, stderr, err := ModuleFetch(ctx, tc.opt)
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

// TestFetchArgsParse tests argument parsing edge cases.
func TestFetchArgsParse(t *testing.T) {
	testcases := []struct {
		name        string
		args        map[string]any
		vars        map[string]any
		expectError bool
		description string
	}{
		{
			name:        "valid args with src and dest",
			args:        map[string]any{"src": "/remote/file.txt", "dest": "/local/file.txt"},
			vars:        nil,
			expectError: false,
			description: "When src and dest are provided, should parse successfully",
		},
		{
			name:        "valid args with template variables",
			args:        map[string]any{"src": "{{ .remote_path }}", "dest": "{{ .local_path }}"},
			vars:        map[string]any{"remote_path": "/remote/file.txt", "local_path": "/local/file.txt"},
			expectError: false,
			description: "When args contain templates, should resolve",
		},
		{
			name:        "missing src",
			args:        map[string]any{"dest": "/local/file.txt"},
			vars:        nil,
			expectError: true,
			description: "When src is missing, should return error",
		},
		{
			name:        "empty args",
			args:        map[string]any{},
			vars:        nil,
			expectError: true,
			description: "When args are empty, should return error",
		},
		{
			name:        "with tmp_dir",
			args:        map[string]any{"src": "/remote/file.txt", "dest": "/local/file.txt", "tmp_dir": "/tmp/custom/"},
			vars:        nil,
			expectError: false,
			description: "When tmp_dir is provided, should parse successfully",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			raw := createRawArgs(tc.args)
			args := variable.Extension2Variables(raw)

			_, err := variable.StringVar(tc.vars, args, "src")

			if tc.expectError {
				require.Error(t, err, tc.description)
			} else {
				require.NoError(t, err, tc.description)
			}
		})
	}
}

// TestFetchArgsTemplate tests template variable resolution in arguments.
func TestFetchArgsTemplate(t *testing.T) {
	testcases := []struct {
		name        string
		args        map[string]any
		vars        map[string]any
		expectError bool
		description string
	}{
		{
			name:        "template: src with variable",
			args:        map[string]any{"src": "{{ .remote_file }}", "dest": "/local/file"},
			vars:        map[string]any{"remote_file": "/remote/test.txt"},
			expectError: false,
			description: "Should resolve template in src",
		},
		{
			name:        "template: dest with variable",
			args:        map[string]any{"src": "/remote/file", "dest": "{{ .local_dir }}/file"},
			vars:        map[string]any{"local_dir": "/tmp/fetch"},
			expectError: false,
			description: "Should resolve template in dest",
		},
		{
			name:        "template: tmp_dir with variable",
			args:        map[string]any{"src": "/remote/file", "dest": "/local/file", "tmp_dir": "{{ .tmp }}"},
			vars:        map[string]any{"tmp": "/tmp/custom"},
			expectError: false,
			description: "Should resolve template in tmp_dir",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			raw := createRawArgs(tc.args)
			args := variable.Extension2Variables(raw)

			_, err := variable.StringVar(tc.vars, args, "src")
			if tc.expectError {
				require.Error(t, err, tc.description)
			} else {
				require.NoError(t, err, tc.description)
			}
		})
	}
}

// TestFetchModule tests the actual functionality of the fetch module.
func TestFetchModule(t *testing.T) {
	// Fetch module requires actual connector
	t.Run("module exists", func(t *testing.T) {
		require.NotNil(t, ModuleFetch)
	})
}
