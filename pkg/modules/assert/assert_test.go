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

package assert

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kubesphere/kubekey/v4/pkg/modules/internal"
)

// createRawArgs creates a runtime.RawExtension from a map (simulating JSON/YAML parsing)
func createRawArgs(data map[string]any) runtime.RawExtension {
	raw, _ := json.Marshal(data)
	return runtime.RawExtension{Raw: raw}
}

// TestAssertArgsModule tests module execution with various argument scenarios.
func TestAssertArgsModule(t *testing.T) {
	testcases := []struct {
		name         string
		opt          internal.ExecOptions
		expectStdout string
		expectError  bool
		description  string
	}{
		{
			name: "missing that",
			opt: internal.ExecOptions{
				Host:     "node1",
				Variable: internal.NewTestVariable([]string{"node1"}, nil),
				Args:     createRawArgs(map[string]any{"success_msg": "success"}),
			},
			expectStdout: internal.StdoutFailed,
			expectError:  true,
			description:  "When that is missing, should return failed",
		},
		{
			name: "empty that",
			opt: internal.ExecOptions{
				Host:     "node1",
				Variable: internal.NewTestVariable([]string{"node1"}, nil),
				Args:     createRawArgs(map[string]any{"that": []string{}}),
			},
			expectStdout: internal.StdoutSuccess,
			expectError:  false,
			description:  "When that is empty, returns success (no conditions to evaluate)",
		},
		{
			name: "condition evaluates to false",
			opt: internal.ExecOptions{
				Host:     "node1",
				Variable: internal.NewTestVariable([]string{"node1"}, nil),
				Args:     createRawArgs(map[string]any{"that": []string{"false"}}),
			},
			expectStdout: internal.StdoutFailed,
			expectError:  true,
			description:  "When condition evaluates to false, should return failed",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			stdout, stderr, err := ModuleAssert(ctx, tc.opt)
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

// TestAssertArgsParse tests argument parsing edge cases.
func TestAssertArgsParse(t *testing.T) {
	testcases := []struct {
		name             string
		args             map[string]any
		expectParseError bool
		expectThatCount  int
		description      string
	}{
		{
			name:             "valid args with string array that",
			args:             map[string]any{"that": []string{"test1", "test2"}},
			expectParseError: false,
			expectThatCount:  2,
			description:      "When that is a string array, should parse successfully",
		},
		{
			name:             "valid args with success_msg and fail_msg",
			args:             map[string]any{"that": []string{"test"}, "success_msg": "ok", "fail_msg": "error"},
			expectParseError: false,
			expectThatCount:  1,
			description:      "When success_msg and fail_msg are set, should parse successfully",
		},
		{
			name:             "valid args with msg only",
			args:             map[string]any{"that": []string{"test"}, "msg": "failed"},
			expectParseError: false,
			expectThatCount:  1,
			description:      "When only msg is set, should parse successfully",
		},
		{
			name:             "invalid args (missing that)",
			args:             map[string]any{"success_msg": "ok"},
			expectParseError: true,
			expectThatCount:  0,
			description:      "When that is missing, should return error",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			raw := createRawArgs(tc.args)
			aa, err := newAssertArgs(ctx, raw, nil)
			if tc.expectParseError {
				require.Error(t, err, tc.description)
			} else {
				require.NoError(t, err, tc.description)
				require.Len(t, aa.that, tc.expectThatCount, tc.description)
			}
		})
	}
}

// TestAssertModule tests the actual functionality of the assert module.
func TestAssertModule(t *testing.T) {
	testcases := []struct {
		name         string
		hosts        []string
		args         map[string]any
		expectStdout string
		expectError  bool
		description  string
	}{
		{
			name:         "default success message when condition is true",
			hosts:        []string{"node1"},
			args:         map[string]any{"that": []string{"true"}},
			expectStdout: internal.StdoutSuccess,
			expectError:  false,
			description:  "Should use default success message when condition is true",
		},
		{
			name:         "custom success_msg",
			hosts:        []string{"node1"},
			args:         map[string]any{"that": []string{"true"}, "success_msg": "custom success"},
			expectStdout: "custom success",
			expectError:  false,
			description:  "Should use custom success_msg when specified",
		},
		{
			name:         "custom fail_msg when condition is false",
			hosts:        []string{"node1"},
			args:         map[string]any{"that": []string{"false"}, "fail_msg": "custom failure"},
			expectStdout: internal.StdoutFailed,
			expectError:  true,
			description:  "Should use custom fail_msg when condition is false",
		},
		{
			name:         "msg as fallback when fail_msg is not set",
			hosts:        []string{"node1"},
			args:         map[string]any{"that": []string{"false"}, "msg": "fallback message"},
			expectStdout: internal.StdoutFailed,
			expectError:  true,
			description:  "Should use msg when fail_msg is not set",
		},
		{
			name:         "fail_msg has higher priority than msg",
			hosts:        []string{"node1"},
			args:         map[string]any{"that": []string{"false"}, "fail_msg": "high priority", "msg": "low priority"},
			expectStdout: internal.StdoutFailed,
			expectError:  true,
			description:  "fail_msg should take priority over msg",
		},
		{
			name:         "multiple conditions all true",
			hosts:        []string{"node1"},
			args:         map[string]any{"that": []string{"true", "true", "true"}},
			expectStdout: internal.StdoutSuccess,
			expectError:  false,
			description:  "Should pass when all conditions are true",
		},
		{
			name:         "multiple conditions one false",
			hosts:        []string{"node1"},
			args:         map[string]any{"that": []string{"true", "false", "true"}},
			expectStdout: internal.StdoutFailed,
			expectError:  true,
			description:  "Should fail when any condition is false",
		},
		{
			name:         "default success_msg value",
			hosts:        []string{"node1"},
			args:         map[string]any{"that": []string{"true"}},
			expectStdout: internal.StdoutSuccess,
			expectError:  false,
			description:  "Should use default 'success' message",
		},
		{
			name:         "default fail_msg value",
			hosts:        []string{"node1"},
			args:         map[string]any{"that": []string{"false"}},
			expectStdout: internal.StdoutFailed,
			expectError:  true,
			description:  "Should use default 'failed' message",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			testVar := internal.NewTestVariable(tc.hosts, nil)

			opt := internal.ExecOptions{
				Host:     tc.hosts[0],
				Variable: testVar,
				Args:     createRawArgs(tc.args),
			}

			stdout, stderr, err := ModuleAssert(ctx, opt)
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
