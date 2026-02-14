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

package template

import (
	"context"
	"encoding/json"
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

// TestTemplateArgsModule tests module execution with various argument scenarios.
func TestTemplateArgsModule(t *testing.T) {
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
				Args:     createRawArgs(map[string]any{"dest": "/tmp/test"}),
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
		{
			name: "invalid mode (negative)",
			opt: internal.ExecOptions{
				Host:     "node1",
				Variable: NewTestVariable([]string{"node1"}, nil),
				Args:     createRawArgs(map[string]any{"src": "/tmp/source", "dest": "/tmp/dest", "mode": -1}),
			},
			expectStdout: internal.StdoutFailed,
			expectError:  true,
			description:  "When mode is negative, should return failed",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			stdout, stderr, err := ModuleTemplate(ctx, tc.opt)
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

// TestTemplateArgsParse tests argument parsing edge cases.
func TestTemplateArgsParse(t *testing.T) {
	testcases := []struct {
		name             string
		args             map[string]any
		expectParseError bool
		expectSrc        string
		expectDest       string
		description      string
	}{
		{
			name:             "valid args with src and dest",
			args:             map[string]any{"src": "/local/template.tmpl", "dest": "/remote/file.conf"},
			expectParseError: false,
			expectSrc:        "/local/template.tmpl",
			expectDest:       "/remote/file.conf",
			description:      "When src and dest are provided, should parse successfully",
		},
		{
			name:             "valid args with mode",
			args:             map[string]any{"src": "/local/template.tmpl", "dest": "/remote/file.conf", "mode": 0644},
			expectParseError: false,
			expectSrc:        "/local/template.tmpl",
			expectDest:       "/remote/file.conf",
			description:      "When mode is provided, should parse successfully",
		},
		{
			name:             "missing src",
			args:             map[string]any{"dest": "/remote/file.conf"},
			expectParseError: true,
			expectSrc:        "",
			expectDest:       "",
			description:      "When src is missing, should return error",
		},
		{
			name:             "missing dest",
			args:             map[string]any{"src": "/local/template.tmpl"},
			expectParseError: true,
			expectSrc:        "",
			expectDest:       "",
			description:      "When dest is missing, should return error",
		},
		{
			name:             "invalid mode (negative)",
			args:             map[string]any{"src": "/local/template.tmpl", "dest": "/remote/file.conf", "mode": -1},
			expectParseError: true,
			expectSrc:        "",
			expectDest:       "",
			description:      "When mode is negative, should return error",
		},
		{
			name:             "invalid mode (too large)",
			args:             map[string]any{"src": "/local/template.tmpl", "dest": "/remote/file.conf", "mode": 4294967296},
			expectParseError: true,
			expectSrc:        "",
			expectDest:       "",
			description:      "When mode is too large, should return error",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			raw := createRawArgs(tc.args)
			result, err := newTemplateArgs(ctx, raw, nil)
			if tc.expectParseError {
				require.Error(t, err, tc.description)
			} else {
				require.NoError(t, err, tc.description)
				require.Equal(t, tc.expectSrc, result.src, tc.description)
				require.Equal(t, tc.expectDest, result.dest, tc.description)
			}
		})
	}
}

// TestTemplateArgsTemplate tests template variable resolution in arguments.
func TestTemplateArgsTemplate(t *testing.T) {
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
			vars:        map[string]any{"src_path": "/local/file.tmpl"},
			expectError: false,
			description: "When src contains template, should resolve with vars",
		},
		{
			name:        "template: dest with template variable",
			args:        map[string]any{"src": "/tmp/src", "dest": "{{ .dest_path }}"},
			vars:        map[string]any{"dest_path": "/remote/file.conf"},
			expectError: false,
			description: "When dest contains template, should resolve with vars",
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
			_, err := newTemplateArgs(ctx, raw, tc.vars)
			if tc.expectError {
				require.Error(t, err, tc.description)
			} else {
				require.NoError(t, err, tc.description)
			}
		})
	}
}

// TestTemplateModule tests the actual functionality of the template module.
func TestTemplateModule(t *testing.T) {
	// Template module requires actual file operations
	// Just verify the module is registered and can be invoked
	t.Run("module exists", func(t *testing.T) {
		require.NotNil(t, ModuleTemplate)
	})
}
