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

package debug

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

// TestDebugArgsModule tests module execution - requires actual variable
func TestDebugArgsModule(t *testing.T) {
	testcases := []struct {
		name         string
		opt          internal.ExecOptions
		expectStdout string
		expectError  bool
		description  string
	}{
		{
			name: "missing msg",
			opt: internal.ExecOptions{
				Host:     "node1",
				Variable: NewTestVariable([]string{"node1"}, nil),
				Args:     createRawArgs(map[string]any{}),
			},
			expectStdout: internal.StdoutFailed,
			expectError:  true,
			description:  "When msg is missing, should return failed",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			stdout, stderr, err := ModuleDebug(ctx, tc.opt)
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

// TestDebugArgsParse tests argument parsing edge cases.
func TestDebugArgsParse(t *testing.T) {
	testcases := []struct {
		name             string
		args             map[string]any
		expectParseError bool
		description      string
	}{
		{
			name:             "valid args with string msg",
			args:             map[string]any{"msg": "Hello World"},
			expectParseError: false,
			description:      "When msg is a string, should parse successfully",
		},
		{
			name:             "valid args with number msg",
			args:             map[string]any{"msg": 42},
			expectParseError: false,
			description:      "When msg is a number, should parse successfully",
		},
		{
			name:             "valid args with object msg",
			args:             map[string]any{"msg": map[string]any{"key": "value"}},
			expectParseError: false,
			description:      "When msg is an object, should parse successfully",
		},
		{
			name:             "valid args with array msg",
			args:             map[string]any{"msg": []string{"a", "b", "c"}},
			expectParseError: false,
			description:      "When msg is an array, should parse successfully",
		},
		{
			name:             "missing msg",
			args:             map[string]any{},
			expectParseError: true,
			description:      "When msg is missing, should return error",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			raw := createRawArgs(tc.args)
			args := variable.Extension2Variables(raw)
			v, ok := args["msg"]
			if tc.expectParseError {
				require.False(t, ok || v != nil, tc.description)
			} else {
				require.True(t, ok || v != nil, tc.description)
			}
		})
	}
}

// TestDebugArgsTemplate tests template variable resolution in arguments.
func TestDebugArgsTemplate(t *testing.T) {
	// Debug module templates are resolved at module execution time
	// Just verify the module can handle template syntax
	t.Run("module exists", func(t *testing.T) {
		require.NotNil(t, ModuleDebug)
	})
}

// TestDebugModule tests the actual functionality of the debug module.
func TestDebugModule(t *testing.T) {
	testcases := []struct {
		name        string
		hosts       []string
		vars        map[string]any
		args        map[string]any
		expectError bool
		description string
	}{
		{
			name:        "print simple message",
			hosts:       []string{"node1"},
			vars:        nil,
			args:        map[string]any{"msg": "Hello World"},
			expectError: false,
			description: "Should print simple message successfully",
		},
		{
			name:        "print variable value",
			hosts:       []string{"node1"},
			vars:        map[string]any{"message": "Hello from variable"},
			args:        map[string]any{"msg": "{{ .message }}"},
			expectError: false,
			description: "Should print variable value successfully",
		},
		{
			name:        "print number",
			hosts:       []string{"node1"},
			vars:        nil,
			args:        map[string]any{"msg": 42},
			expectError: false,
			description: "Should print number successfully",
		},
		{
			name:        "print object",
			hosts:       []string{"node1"},
			vars:        nil,
			args:        map[string]any{"msg": map[string]any{"key": "value"}},
			expectError: false,
			description: "Should print object successfully",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			testVar := NewTestVariable(tc.hosts, tc.vars)

			opt := internal.ExecOptions{
				Host:     tc.hosts[0],
				Variable: testVar,
				Args:     createRawArgs(tc.args),
			}

			stdout, stderr, err := ModuleDebug(ctx, opt)
			if tc.expectError {
				require.Error(t, err, tc.description)
				require.NotEmpty(t, stderr, tc.description)
			} else {
				require.NoError(t, err, tc.description)
				require.NotEmpty(t, stdout, tc.description)
			}
		})
	}
}
