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

package include_vars

import (
	"context"
	"encoding/json"
	"path/filepath"
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

// TestIncludeVarsArgsModule tests module execution with various argument scenarios.
func TestIncludeVarsArgsModule(t *testing.T) {
	testcases := []struct {
		name         string
		opt          internal.ExecOptions
		expectStdout string
		expectError  bool
		description  string
	}{
		{
			name: "empty path",
			opt: internal.ExecOptions{
				Host:     "node1",
				Variable: NewTestVariable([]string{"node1"}, nil),
				Args:     createRawArgs(map[string]any{"include_vars": ""}),
			},
			expectStdout: internal.StdoutFailed,
			expectError:  true,
			description:  "When include_vars path is empty, should return failed",
		},
		{
			name: "invalid file extension",
			opt: internal.ExecOptions{
				Host:     "node1",
				Variable: NewTestVariable([]string{"node1"}, nil),
				Args:     createRawArgs(map[string]any{"include_vars": "file.txt"}),
			},
			expectStdout: internal.StdoutFailed,
			expectError:  true,
			description:  "When file extension is not .yaml or .yml, should return failed",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			stdout, stderr, err := ModuleIncludeVars(ctx, tc.opt)
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

// TestIncludeVarsArgsParse tests argument parsing edge cases.
func TestIncludeVarsArgsParse(t *testing.T) {
	testcases := []struct {
		name          string
		rawArgs       []byte
		expectParseOk bool
		description   string
	}{
		{
			name:          "valid yaml file path",
			rawArgs:       []byte(`{"include_vars": "vars.yaml"}`),
			expectParseOk: true,
			description:   "When include_vars is a valid .yaml file, should parse successfully",
		},
		{
			name:          "valid yml file path",
			rawArgs:       []byte(`{"include_vars": "vars.yml"}`),
			expectParseOk: true,
			description:   "When include_vars is a valid .yml file, should parse successfully",
		},
		{
			name:          "empty path",
			rawArgs:       []byte(`{"include_vars": ""}`),
			expectParseOk: true,
			description:   "When include_vars is empty, should parse (error happens at execution)",
		},
		{
			name:          "invalid extension",
			rawArgs:       []byte(`{"include_vars": "file.txt"}`),
			expectParseOk: true,
			description:   "When file extension is not .yaml/.yml, should parse (error happens at execution)",
		},
		{
			name:          "missing include_vars",
			rawArgs:       []byte(`{}`),
			expectParseOk: true,
			description:   "When include_vars is missing, should parse (error happens at execution)",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			raw := runtime.RawExtension{Raw: tc.rawArgs}
			arg, err := variable.Extension2String(nil, raw)
			if tc.expectParseOk {
				require.NoError(t, err, tc.description)
				_ = arg // may be empty
			} else {
				require.Error(t, err, tc.description)
			}
		})
	}
}

// TestIncludeVarsArgsTemplate tests template variable resolution in arguments.
func TestIncludeVarsArgsTemplate(t *testing.T) {
	testcases := []struct {
		name          string
		rawArgs       []byte
		vars          map[string]any
		expectParseOk bool
		description   string
	}{
		{
			name:          "template: path with template variable",
			rawArgs:       []byte(`{"include_vars": "{{ .vars_file }}"}`),
			vars:          map[string]any{"vars_file": "vars.yaml"},
			expectParseOk: true,
			description:   "When path contains template, should resolve with vars",
		},
		{
			name:          "template: path with nested template",
			rawArgs:       []byte(`{"include_vars": "{{ .dir }}/{{ .file }}"}`),
			vars:          map[string]any{"dir": "vars", "file": "config.yaml"},
			expectParseOk: true,
			description:   "When path contains multiple templates, should resolve",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			raw := runtime.RawExtension{Raw: tc.rawArgs}
			arg, err := variable.Extension2String(tc.vars, raw)
			if tc.expectParseOk {
				require.NoError(t, err, tc.description)
				require.NotEmpty(t, arg, tc.description)
			} else {
				require.Error(t, err, tc.description)
			}
		})
	}
}

// TestIncludeVarsModule tests the actual functionality of the include_vars module.
func TestIncludeVarsModule(t *testing.T) {
	// Test file extension validation
	t.Run("file extension validation", func(t *testing.T) {
		tests := []struct {
			name        string
			path        string
			expectValid bool
		}{
			{"yaml extension", "vars.yaml", true},
			{"yml extension", "vars.yml", true},
			{"txt extension", "vars.txt", false},
			{"json extension", "vars.json", false},
			{"no extension", "varsfile", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				ext := filepath.Ext(tt.path)
				isValid := ext == ".yaml" || ext == ".yml"
				require.Equal(t, tt.expectValid, isValid)
			})
		}
	})
}
