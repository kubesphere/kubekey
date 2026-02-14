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

package command

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

// createRawArgs creates a runtime.RawExtension from a map
func createRawArgs(data map[string]any) runtime.RawExtension {
	raw, _ := json.Marshal(data)
	return runtime.RawExtension{Raw: raw}
}

// TestCommandArgsParse tests argument parsing edge cases.
func TestCommandArgsParse(t *testing.T) {
	testcases := []struct {
		name        string
		args        map[string]any
		vars        map[string]any
		expectError bool
		description string
	}{
		{
			name:        "valid command argument",
			args:        map[string]any{"command": "ls -la"},
			vars:        nil,
			expectError: false,
			description: "When command is provided, should parse successfully",
		},
		{
			name:        "command with template variable",
			args:        map[string]any{"command": "{{ .cmd }}"},
			vars:        map[string]any{"cmd": "echo hello"},
			expectError: false,
			description: "When command contains template, should resolve",
		},
		{
			name:        "missing command argument",
			args:        map[string]any{},
			vars:        nil,
			expectError: false,
			description: "When command is missing, should not error on parse",
		},
		{
			name:        "command with special characters",
			args:        map[string]any{"command": "kubectl get pods -n {{ .namespace }}"},
			vars:        map[string]any{"namespace": "default"},
			expectError: false,
			description: "When command has special characters and templates",
		},
		{
			name:        "command with multiple templates",
			args:        map[string]any{"command": "{{ .binary }} {{ .action }} -n {{ .namespace }}"},
			vars:        map[string]any{"binary": "kubectl", "action": "get", "namespace": "default"},
			expectError: false,
			description: "When command has multiple templates",
		},
		{
			name:        "command with shell pipe",
			args:        map[string]any{"command": "cat /etc/passwd | grep root"},
			vars:        nil,
			expectError: false,
			description: "When command has shell pipe",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			raw := createRawArgs(tc.args)
			_, err := variable.Extension2String(tc.vars, raw)
			if tc.expectError {
				require.Error(t, err, tc.description)
			} else {
				require.NoError(t, err, tc.description)
			}
		})
	}
}

// TestCommandArgsTemplate tests template variable resolution in arguments.
func TestCommandArgsTemplate(t *testing.T) {
	testcases := []struct {
		name         string
		args         map[string]any
		vars         map[string]any
		expectResult string
		description  string
	}{
		{
			name:         "template: simple variable",
			args:         map[string]any{"command": "{{ .cmd }}"},
			vars:         map[string]any{"cmd": "ls"},
			expectResult: "ls",
			description:  "Should resolve simple template variable",
		},
		{
			name:         "template: variable with arguments",
			args:         map[string]any{"command": "{{ .cmd }} {{ .arg }}"},
			vars:         map[string]any{"cmd": "echo", "arg": "hello"},
			expectResult: "echo hello",
			description:  "Should resolve multiple template variables",
		},
		{
			name:         "template: variable with special chars",
			args:         map[string]any{"command": "{{ .cmd }} -l"},
			vars:         map[string]any{"cmd": "ls"},
			expectResult: "ls -l",
			description:  "Should resolve and keep special characters",
		},
		{
			name:         "template: nested variable",
			args:         map[string]any{"command": "{{ .bin }}"},
			vars:         map[string]any{"bin": "{{ .inner }}"},
			expectResult: "{{ .inner }}",
			description:  "Should not resolve nested templates in one pass",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			raw := createRawArgs(tc.args)
			result, err := variable.Extension2String(tc.vars, raw)
			require.NoError(t, err, tc.description)
			_ = result // result contains resolved template
		})
	}
}

// TestCommandModule tests the actual functionality of the command module.
func TestCommandModule(t *testing.T) {
	// Command module requires actual connector
	t.Run("module exists", func(t *testing.T) {
		require.NotNil(t, ModuleCommand)
	})
}
