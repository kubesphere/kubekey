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

package result

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

// TestResultArgsModule tests module execution with various argument scenarios.
func TestResultArgsModule(t *testing.T) {
	// Result module execution tests - minimal since it requires actual connector
	t.Run("module exists", func(t *testing.T) {
		require.NotNil(t, ModuleResult)
	})
}

// TestResultArgsParse tests argument parsing edge cases.
func TestResultArgsParse(t *testing.T) {
	testcases := []struct {
		name          string
		rawArgs       []byte
		expectParseOk bool
		description   string
	}{
		{
			name:          "valid yaml with single key-value",
			rawArgs:       []byte(`key: value`),
			expectParseOk: true,
			description:   "When args contain valid YAML, should parse successfully",
		},
		{
			name:          "valid yaml with multiple key-values",
			rawArgs:       []byte("key1: value1\nkey2: value2"),
			expectParseOk: true,
			description:   "When args contain multiple key-values, should parse successfully",
		},
		{
			name:          "valid yaml with nested object",
			rawArgs:       []byte("nested:\n  key: value"),
			expectParseOk: true,
			description:   "When args contain nested object, should parse successfully",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			raw := runtime.RawExtension{Raw: tc.rawArgs}
			arg, err := variable.Extension2String(nil, raw)
			if tc.expectParseOk {
				require.NoError(t, err, tc.description)
				require.NotEmpty(t, arg, tc.description)
			} else {
				require.Error(t, err, tc.description)
			}
		})
	}
}

// TestResultArgsTemplate tests template variable resolution in arguments.
func TestResultArgsTemplate(t *testing.T) {
	testcases := []struct {
		name          string
		rawArgs       []byte
		vars          map[string]any
		expectParseOk bool
		description   string
	}{
		{
			name:          "template: args with template variable",
			rawArgs:       []byte(`key: "{{ .value }}"`),
			vars:          map[string]any{"value": "resolved"},
			expectParseOk: true,
			description:   "When args contain template, should resolve with vars",
		},
		{
			name:          "template: args with numeric template",
			rawArgs:       []byte(`count: "{{ .num }}"`),
			vars:          map[string]any{"num": 42},
			expectParseOk: true,
			description:   "When args contain numeric template, should resolve",
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

// TestResultModule tests the actual functionality of the result module.
func TestResultModule(t *testing.T) {
	testcases := []struct {
		name         string
		hosts        []string
		initialVars  map[string]any
		args         map[string]any
		expectStdout string
		expectError  bool
		description  string
	}{
		{
			name:         "set single result variable",
			hosts:        []string{"node1"},
			initialVars:  nil,
			args:         map[string]any{"app_version": "1.0.0"},
			expectStdout: internal.StdoutSuccess,
			expectError:  false,
			description:  "Should set result variable successfully",
		},
		{
			name:        "set multiple result variables",
			hosts:       []string{"node1"},
			initialVars: nil,
			args: map[string]any{
				"db_host": "localhost",
				"db_port": 5432,
			},
			expectStdout: internal.StdoutSuccess,
			expectError:  false,
			description:  "Should set multiple result variables successfully",
		},
		{
			name:        "set nested result variable",
			hosts:       []string{"node1"},
			initialVars: nil,
			args: map[string]any{
				"config": map[string]any{
					"key": "value",
				},
			},
			expectStdout: internal.StdoutSuccess,
			expectError:  false,
			description:  "Should set nested result variable successfully",
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

			stdout, stderr, err := ModuleResult(ctx, opt)
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
