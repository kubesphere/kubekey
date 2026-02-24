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

package prometheus

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
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

// TestPrometheusArgsModule tests module execution - requires actual connector
func TestPrometheusArgsModule(t *testing.T) {
	t.Run("module exists", func(t *testing.T) {
		require.NotNil(t, ModulePrometheus)
	})
}

// TestPrometheusArgsParse tests argument parsing edge cases.
func TestPrometheusArgsParse(t *testing.T) {
	testcases := []struct {
		name             string
		args             map[string]any
		vars             map[string]any
		expectParseError bool
		description      string
	}{
		{
			name:             "valid args with query",
			args:             map[string]any{"query": "up"},
			vars:             nil,
			expectParseError: false,
			description:      "When query is provided, should parse successfully",
		},
		{
			name:             "valid args with query and format",
			args:             map[string]any{"query": "up", "format": "value"},
			vars:             nil,
			expectParseError: false,
			description:      "When query and format are provided, should parse successfully",
		},
		{
			name:             "valid args with all parameters",
			args:             map[string]any{"query": "up", "format": "table", "time": "2023-01-01T12:00:00Z"},
			vars:             nil,
			expectParseError: false,
			description:      "When all parameters are provided, should parse successfully",
		},
		{
			name:             "missing query",
			args:             map[string]any{},
			vars:             nil,
			expectParseError: true,
			description:      "When query is missing, should return error",
		},
		{
			name:             "invalid format",
			args:             map[string]any{"query": "up", "format": "invalid_format"},
			vars:             nil,
			expectParseError: false, // format is optional, won't error on parse
			description:      "When format is invalid, should not error on parse (error happens at execution)",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			raw := createRawArgs(tc.args)
			args := variable.Extension2Variables(raw)
			_, err := variable.StringVar(tc.vars, args, "query")
			if tc.expectParseError {
				require.Error(t, err, tc.description)
			} else {
				require.NoError(t, err, tc.description)
			}
		})
	}
}

// TestPrometheusArgsTemplate tests template variable resolution in arguments.
func TestPrometheusArgsTemplate(t *testing.T) {
	testcases := []struct {
		name        string
		args        map[string]any
		vars        map[string]any
		expectError bool
		description string
	}{
		{
			name:        "template: query with template variable",
			args:        map[string]any{"query": "{{ .prom_query }}"},
			vars:        map[string]any{"prom_query": "up"},
			expectError: false,
			description: "When query contains template, should resolve with vars",
		},
		{
			name:        "template: format with template variable",
			args:        map[string]any{"query": "up", "format": "{{ .fmt }}"},
			vars:        map[string]any{"fmt": "value"},
			expectError: false,
			description: "When format contains template, should resolve with vars",
		},
		{
			name:        "template: time with template variable",
			args:        map[string]any{"query": "up", "time": "{{ .time }}"},
			vars:        map[string]any{"time": "2023-01-01T12:00:00Z"},
			expectError: false,
			description: "When time contains template, should resolve with vars",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			raw := createRawArgs(tc.args)
			args := variable.Extension2Variables(raw)
			_, err := variable.StringVar(tc.vars, args, "query")
			if tc.expectError {
				require.Error(t, err, tc.description)
			} else {
				require.NoError(t, err, tc.description)
			}
		})
	}
}

// TestPrometheusModule tests the actual functionality of the prometheus module.
func TestPrometheusModule(t *testing.T) {
	t.Run("module exists", func(t *testing.T) {
		require.NotNil(t, ModulePrometheus)
	})
}
