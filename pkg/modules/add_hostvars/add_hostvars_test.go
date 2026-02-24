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

package add_hostvars

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kubesphere/kubekey/v4/pkg/modules/internal"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

// TestAddHostvarsArgsModule tests module execution with various argument scenarios.
func TestAddHostvarsArgsModule(t *testing.T) {
	testcases := []struct {
		name         string
		opt          internal.ExecOptions
		expectStdout string
		expectError  bool
		description  string
	}{
		{
			name: "missing hosts",
			opt: internal.ExecOptions{
				Host:     "node1",
				Variable: internal.NewTestVariable([]string{"node1"}, nil),
				Args: runtime.RawExtension{
					Raw: []byte(`
vars:
  foo: bar
`),
				},
			},
			expectStdout: internal.StdoutFailed,
			expectError:  true,
			description:  "When hosts is missing, should return failed",
		},
		{
			name: "missing vars",
			opt: internal.ExecOptions{
				Host:     "node1",
				Variable: internal.NewTestVariable([]string{"node1"}, nil),
				Args: runtime.RawExtension{
					Raw: []byte(`
hosts: node1
`),
				},
			},
			expectStdout: internal.StdoutFailed,
			expectError:  true,
			description:  "When vars is missing, should return failed",
		},
		{
			name: "invalid hosts type",
			opt: internal.ExecOptions{
				Host:     "node1",
				Variable: internal.NewTestVariable([]string{"node1"}, nil),
				Args: runtime.RawExtension{
					Raw: []byte(`
hosts:
  foo: bar
vars:
  a: b
`),
				},
			},
			expectStdout: internal.StdoutFailed,
			expectError:  true,
			description:  "When hosts type is invalid (not string or array), should return failed",
		},
		{
			name: "string hosts",
			opt: internal.ExecOptions{
				Host:     "node1",
				Variable: internal.NewTestVariable([]string{"node1"}, nil),
				Args: runtime.RawExtension{
					Raw: []byte(`
hosts: node1
vars:
  a: b
`),
				},
			},
			expectStdout: internal.StdoutSuccess,
			expectError:  false,
			description:  "When hosts is a string and vars is valid, should return success",
		},
		{
			name: "array hosts",
			opt: internal.ExecOptions{
				Host:     "node1",
				Variable: internal.NewTestVariable([]string{"node1", "node2"}, nil),
				Args: runtime.RawExtension{
					Raw: []byte(`
hosts:
  - node1
  - node2
vars:
  test_var: test_value
`),
				},
			},
			expectStdout: internal.StdoutSuccess,
			expectError:  false,
			description:  "When hosts is an array, should return success",
		},
		{
			name: "template in hosts",
			opt: internal.ExecOptions{
				Host:     "node1",
				Variable: internal.NewTestVariable([]string{"node1", "node2"}, map[string]any{"target_host": "node1"}),
				Args: runtime.RawExtension{
					Raw: []byte(`
hosts: "{{ .target_host }}"
vars:
  test_var: test_value
`),
				},
			},
			expectStdout: internal.StdoutSuccess,
			expectError:  false,
			description:  "When hosts uses template variable, should resolve correctly",
		},
		{
			name: "template in vars",
			opt: internal.ExecOptions{
				Host:     "node1",
				Variable: internal.NewTestVariable([]string{"node1"}, map[string]any{"base_dir": "/opt"}),
				Args: runtime.RawExtension{
					Raw: []byte(`
hosts: node1
vars:
  app_dir: "{{ .base_dir }}/app"
`),
				},
			},
			expectStdout: internal.StdoutSuccess,
			expectError:  false,
			description:  "When vars contains template variable, should resolve before adding",
		},
		{
			name: "complex template in vars",
			opt: internal.ExecOptions{
				Host:     "node1",
				Variable: internal.NewTestVariable([]string{"node1"}, map[string]any{"env": "prod", "region": "us-east"}),
				Args: runtime.RawExtension{
					Raw: []byte(`
hosts: node1
vars:
  env_tag: "{{ .env }}-{{ .region }}"
`),
				},
			},
			expectStdout: internal.StdoutSuccess,
			expectError:  false,
			description:  "When vars contains complex template expression, should resolve correctly",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			stdout, stderr, err := ModuleAddHostvars(ctx, tc.opt)
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

// TestAddHostvarsArgsParse tests argument parsing edge cases.
func TestAddHostvarsArgsParse(t *testing.T) {
	testcases := []struct {
		name             string
		rawArgs          []byte
		expectParseError bool
		description      string
	}{
		{
			name:             "valid args with string hosts",
			rawArgs:          []byte(`hosts: node1` + "\n" + `vars:` + "\n" + `  key: value`),
			expectParseError: false,
			description:      "When hosts is string and vars is mapping, should parse successfully",
		},
		{
			name:             "valid args with array hosts",
			rawArgs:          []byte(`hosts: [node1, node2]` + "\n" + `vars:` + "\n" + `  key: value`),
			expectParseError: false,
			description:      "When hosts is array and vars is mapping, should parse successfully",
		},
		{
			name:             "empty hosts",
			rawArgs:          []byte(`vars:` + "\n" + `  key: value`),
			expectParseError: true,
			description:      "When hosts is empty, should return error",
		},
		{
			name:             "empty vars",
			rawArgs:          []byte(`hosts: node1`),
			expectParseError: true,
			description:      "When vars is empty, should return error",
		},
		{
			name:             "invalid yaml format",
			rawArgs:          []byte(`{invalid: yaml}`),
			expectParseError: true,
			description:      "When YAML format is invalid, should return error",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			raw := runtime.RawExtension{Raw: tc.rawArgs}
			_, err := newAddHostvarsArgs(ctx, raw, nil)
			if tc.expectParseError {
				require.Error(t, err, tc.description)
			} else {
				require.NoError(t, err, tc.description)
			}
		})
	}
}

// TestAddHostvarsArgsTemplate tests template variable resolution in arguments.
func TestAddHostvarsArgsTemplate(t *testing.T) {
	testcases := []struct {
		name            string
		hosts           []string
		initialVars     map[string]any
		args            runtime.RawExtension
		expectStdout    string
		expectError     bool
		expectHostNames []string
		description     string
	}{
		{
			name:        "template: hosts with template variable",
			hosts:       []string{"node1", "node2"},
			initialVars: map[string]any{"target_host": "node1"},
			args: runtime.RawExtension{
				Raw: []byte(`
hosts: "{{ .target_host }}"
vars:
  test_var: test_value
`),
			},
			expectStdout:    internal.StdoutSuccess,
			expectError:     false,
			expectHostNames: []string{"node1"},
			description:     "When hosts uses template variable, should resolve correctly",
		},
		{
			name:        "template: vars with template variable",
			hosts:       []string{"node1"},
			initialVars: map[string]any{"base_dir": "/opt"},
			args: runtime.RawExtension{
				Raw: []byte(`
hosts: node1
vars:
  app_dir: "{{ .base_dir }}/app"
`),
			},
			expectStdout: internal.StdoutSuccess,
			expectError:  false,
			description:  "When vars contains template variable, should resolve before adding",
		},
		{
			name:        "template: vars with complex expression",
			hosts:       []string{"node1"},
			initialVars: map[string]any{"env": "prod", "region": "us-east"},
			args: runtime.RawExtension{
				Raw: []byte(`
hosts: node1
vars:
  env_tag: "{{ .env }}-{{ .region }}"
`),
			},
			expectStdout: internal.StdoutSuccess,
			expectError:  false,
			description:  "When vars contains complex template expression, should resolve correctly",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			testVar := internal.NewTestVariable(tc.hosts, tc.initialVars)

			opt := internal.ExecOptions{
				Host:     tc.hosts[0],
				Variable: testVar,
				Args:     tc.args,
			}

			stdout, stderr, err := ModuleAddHostvars(ctx, opt)
			require.Equal(t, tc.expectStdout, stdout, tc.description)
			if tc.expectError {
				require.Error(t, err, tc.description)
				require.NotEmpty(t, stderr, tc.description)
			} else {
				require.NoError(t, err, tc.description)

				if tc.expectHostNames != nil {
					for _, hostName := range tc.expectHostNames {
						result, err := testVar.Get(variable.GetAllVariable(hostName))
						require.NoError(t, err, tc.description)
						hostVars, ok := result.(map[string]any)
						require.True(t, ok, "host variables should be a map")
						_, exists := hostVars["test_var"]
						require.True(t, exists, tc.description+" - test_var should exist on "+hostName)
					}
				}
			}
		})
	}
}

// TestAddHostvarsModule tests the actual functionality of adding host variables.
func TestAddHostvarsModule(t *testing.T) {
	testcases := []struct {
		name           string
		hosts          []string
		initialVars    map[string]any
		args           runtime.RawExtension
		targetHost     string
		expectVarKey   string
		expectVarValue any
		description    string
	}{
		{
			name:        "add variable to single host",
			hosts:       []string{"node1"},
			initialVars: nil,
			args: runtime.RawExtension{
				Raw: []byte(`
hosts: node1
vars:
  custom_var: custom_value
`),
			},
			targetHost:     "node1",
			expectVarKey:   "custom_var",
			expectVarValue: "custom_value",
			description:    "Should add custom_var to node1 and be retrievable",
		},
		{
			name:        "add multiple variables to single host",
			hosts:       []string{"node1"},
			initialVars: nil,
			args: runtime.RawExtension{
				Raw: []byte(`
hosts: node1
vars:
  var1: value1
  var2: 123
  var3: true
`),
			},
			targetHost:     "node1",
			expectVarKey:   "var1",
			expectVarValue: "value1",
			description:    "Should add multiple variables and be retrievable",
		},
		{
			name:        "add nested map variable",
			hosts:       []string{"node1"},
			initialVars: nil,
			args: runtime.RawExtension{
				Raw: []byte(`
hosts: node1
vars:
  config:
    key1: val1
    key2: val2
`),
			},
			targetHost:     "node1",
			expectVarKey:   "config",
			expectVarValue: map[string]any{"key1": "val1", "key2": "val2"},
			description:    "Should add nested map variable and be retrievable",
		},
		{
			name:        "add variable to multiple hosts",
			hosts:       []string{"node1", "node2"},
			initialVars: nil,
			args: runtime.RawExtension{
				Raw: []byte(`
hosts:
  - node1
  - node2
vars:
  shared_var: shared_value
`),
			},
			targetHost:     "node1",
			expectVarKey:   "shared_var",
			expectVarValue: "shared_value",
			description:    "Should add shared variable to all specified hosts",
		},
		{
			name:        "merge with existing host variables",
			hosts:       []string{"node1"},
			initialVars: map[string]any{"existing_var": "existing_value"},
			args: runtime.RawExtension{
				Raw: []byte(`
hosts: node1
vars:
  new_var: new_value
`),
			},
			targetHost:     "node1",
			expectVarKey:   "new_var",
			expectVarValue: "new_value",
			description:    "Should merge new variable with existing variables",
		},
		{
			name:        "add variable with template resolution",
			hosts:       []string{"node1"},
			initialVars: map[string]any{"base_dir": "/opt"},
			args: runtime.RawExtension{
				Raw: []byte(`
hosts: node1
vars:
  app_dir: "{{ .base_dir }}/app"
`),
			},
			targetHost:     "node1",
			expectVarKey:   "app_dir",
			expectVarValue: "/opt/app",
			description:    "Should resolve template variable before adding",
		},
		{
			name:        "add array variable",
			hosts:       []string{"node1"},
			initialVars: nil,
			args: runtime.RawExtension{
				Raw: []byte(`
hosts: node1
vars:
  ports:
    - 8080
    - 8081
    - 8082
`),
			},
			targetHost:     "node1",
			expectVarKey:   "ports",
			expectVarValue: []any{8080, 8081, 8082},
			description:    "Should add array variable correctly",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			testVar := internal.NewTestVariable(tc.hosts, tc.initialVars)

			opt := internal.ExecOptions{
				Host:     tc.targetHost,
				Variable: testVar,
				Args:     tc.args,
			}

			stdout, _, err := ModuleAddHostvars(ctx, opt)
			require.NoError(t, err, tc.description)
			require.Equal(t, internal.StdoutSuccess, stdout, tc.description)

			result, err := testVar.Get(variable.GetAllVariable(tc.targetHost))
			require.NoError(t, err, tc.description)

			hostVars, ok := result.(map[string]any)
			require.True(t, ok, "host variables should be a map")

			actualValue, exists := hostVars[tc.expectVarKey]
			require.True(t, exists, tc.description+" - variable should exist: "+tc.expectVarKey)
			require.Equal(t, tc.expectVarValue, actualValue, tc.description)
		})
	}

	t.Run("multiple hosts isolation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		hosts := []string{"node1", "node2", "node3"}
		testVar := internal.NewTestVariable(hosts, nil)

		opt := internal.ExecOptions{
			Host:     "node1",
			Variable: testVar,
			Args: runtime.RawExtension{
				Raw: []byte(`
hosts:
  - node1
  - node2
vars:
  cluster_var: cluster_value
  node_specific: node1_node2
`),
			},
		}

		stdout, _, err := ModuleAddHostvars(ctx, opt)
		require.NoError(t, err)
		require.Equal(t, internal.StdoutSuccess, stdout)

		result1, err := testVar.Get(variable.GetAllVariable("node1"))
		require.NoError(t, err)
		hostVars1 := result1.(map[string]any)
		require.Equal(t, "cluster_value", hostVars1["cluster_var"])
		require.Equal(t, "node1_node2", hostVars1["node_specific"])

		result2, err := testVar.Get(variable.GetAllVariable("node2"))
		require.NoError(t, err)
		hostVars2 := result2.(map[string]any)
		require.Equal(t, "cluster_value", hostVars2["cluster_var"])
		require.Equal(t, "node1_node2", hostVars2["node_specific"])

		result3, err := testVar.Get(variable.GetAllVariable("node3"))
		require.NoError(t, err)
		hostVars3 := result3.(map[string]any)
		_, exists := hostVars3["cluster_var"]
		require.False(t, exists, "node3 should not have cluster_var")
	})
}
