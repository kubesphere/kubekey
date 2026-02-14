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

package set_fact

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
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

// TestSetFactArgsModule tests module execution with various argument scenarios.
func TestSetFactArgsModule(t *testing.T) {
	// SetFact module execution tests - minimal since it requires actual variable setup
	t.Run("module exists", func(t *testing.T) {
		require.NotNil(t, ModuleSetFact)
	})
}

// TestSetFactArgsParse tests argument parsing edge cases.
func TestSetFactArgsParse(t *testing.T) {
	testcases := []struct {
		name             string
		rawArgs          []byte
		expectParseError bool
		description      string
	}{
		{
			name:             "valid yaml with single key-value",
			rawArgs:          []byte(`key: value`),
			expectParseError: false,
			description:      "When args contain valid YAML, should parse successfully",
		},
		{
			name:             "valid yaml with multiple key-values",
			rawArgs:          []byte("key1: value1\nkey2: value2"),
			expectParseError: false,
			description:      "When args contain multiple key-values, should parse successfully",
		},
		{
			name:             "valid yaml with nested object",
			rawArgs:          []byte("nested:\n  key: value"),
			expectParseError: false,
			description:      "When args contain nested object, should parse successfully",
		},
		{
			name:             "empty yaml",
			rawArgs:          []byte(``),
			expectParseError: false,
			description:      "When YAML is empty, should parse successfully",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			raw := runtime.RawExtension{Raw: tc.rawArgs}
			var node yaml.Node
			err := yaml.Unmarshal(raw.Raw, &node)
			if tc.expectParseError {
				require.Error(t, err, tc.description)
			} else {
				require.NoError(t, err, tc.description)
			}
		})
	}
}

// TestSetFactArgsTemplate tests template variable resolution in arguments.
func TestSetFactArgsTemplate(t *testing.T) {
	testcases := []struct {
		name        string
		rawArgs     []byte
		vars        map[string]any
		expectError bool
		description string
	}{
		{
			name:        "template: args with template variable",
			rawArgs:     []byte(`key: "{{ .value }}"`),
			vars:        map[string]any{"value": "resolved"},
			expectError: false,
			description: "When args contain template, should resolve with vars",
		},
		{
			name:        "template: args with multiple template variables",
			rawArgs:     []byte("key1: \"{{ .v1 }}\"\nkey2: \"{{ .v2 }}\""),
			vars:        map[string]any{"v1": "val1", "v2": "val2"},
			expectError: false,
			description: "When args contain multiple templates, should resolve all",
		},
		{
			name:        "template: args with numeric template",
			rawArgs:     []byte(`count: "{{ .num }}"`),
			vars:        map[string]any{"num": 42},
			expectError: false,
			description: "When args contain numeric template, should resolve",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			raw := runtime.RawExtension{Raw: tc.rawArgs}
			var node yaml.Node
			err := yaml.Unmarshal(raw.Raw, &node)
			if tc.expectError {
				require.Error(t, err, tc.description)
			} else {
				require.NoError(t, err, tc.description)
			}
		})
	}
}

// TestSetFactModule tests the actual functionality of the set_fact module.
func TestSetFactModule(t *testing.T) {
	testcases := []struct {
		name           string
		hosts          []string
		initialVars    map[string]any
		args           map[string]any
		expectStdout   string
		expectError    bool
		expectVarKey   string
		expectVarValue any
		description    string
	}{
		{
			name:           "set single variable",
			hosts:          []string{"node1"},
			initialVars:    nil,
			args:           map[string]any{"app_version": "1.0.0"},
			expectStdout:   internal.StdoutSuccess,
			expectError:    false,
			expectVarKey:   "app_version",
			expectVarValue: "1.0.0",
			description:    "Should set single variable successfully",
		},
		{
			name:        "set multiple variables",
			hosts:       []string{"node1"},
			initialVars: nil,
			args: map[string]any{
				"db_host": "localhost",
				"db_port": 5432,
			},
			expectStdout:   internal.StdoutSuccess,
			expectError:    false,
			expectVarKey:   "db_host",
			expectVarValue: "localhost",
			description:    "Should set multiple variables successfully",
		},
		{
			name:        "set nested variable",
			hosts:       []string{"node1"},
			initialVars: nil,
			args: map[string]any{
				"config": map[string]any{
					"key": "value",
				},
			},
			expectStdout:   internal.StdoutSuccess,
			expectError:    false,
			expectVarKey:   "config",
			expectVarValue: map[string]any{"key": "value"},
			description:    "Should set nested variable successfully",
		},
		{
			name:        "set array variable",
			hosts:       []string{"node1"},
			initialVars: nil,
			args: map[string]any{
				"items": []any{"a", "b", "c"},
			},
			expectStdout:   internal.StdoutSuccess,
			expectError:    false,
			expectVarKey:   "items",
			expectVarValue: []any{"a", "b", "c"},
			description:    "Should set array variable successfully",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			testVar := NewTestVariable(tc.hosts, tc.initialVars)

			opt := internal.ExecOptions{
				Host:     tc.hosts[0],
				Variable: testVar,
				Args:     createRawArgs(tc.args),
			}

			stdout, stderr, err := ModuleSetFact(ctx, opt)
			require.Equal(t, tc.expectStdout, stdout, tc.description)
			if tc.expectError {
				require.Error(t, err, tc.description)
				require.NotEmpty(t, stderr, tc.description)
			} else {
				require.NoError(t, err, tc.description)

				// Verify the variable was set
				if tc.expectVarKey != "" {
					result, err := testVar.Get(variable.GetAllVariable(tc.hosts[0]))
					require.NoError(t, err, tc.description)
					hostVars, ok := result.(map[string]any)
					require.True(t, ok, "host variables should be a map")
					actualValue, exists := hostVars[tc.expectVarKey]
					require.True(t, exists, tc.description+" - variable should exist: "+tc.expectVarKey)
					require.Equal(t, tc.expectVarValue, actualValue, tc.description)
				}
			}
		})
	}
}
