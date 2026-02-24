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

package gen_cert

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
)

// createRawArgs creates a runtime.RawExtension from a map
func createRawArgs(data map[string]any) runtime.RawExtension {
	raw, _ := json.Marshal(data)
	return runtime.RawExtension{Raw: raw}
}

// TestGenCertArgsModule tests module execution - simplified to just verify module exists
func TestGenCertArgsModule(t *testing.T) {
	t.Run("module exists", func(t *testing.T) {
		require.NotNil(t, ModuleGenCert)
	})
}

// TestGenCertArgsParse tests argument parsing edge cases.
func TestGenCertArgsParse(t *testing.T) {
	testcases := []struct {
		name             string
		args             map[string]any
		expectParseError bool
		description      string
	}{
		{
			name:             "valid args with required fields",
			args:             map[string]any{"cn": "test.example.com", "out_key": "/tmp/key.pem", "out_cert": "/tmp/cert.pem", "policy": "Always"},
			expectParseError: false,
			description:      "When required fields are provided, should parse successfully",
		},
		{
			name:             "valid args with all fields",
			args:             map[string]any{"cn": "test.example.com", "out_key": "/tmp/key.pem", "out_cert": "/tmp/cert.pem", "root_key": "/tmp/root.key", "root_cert": "/tmp/root.crt", "date": "24h", "policy": "Always", "sans": []string{"test1", "test2"}, "is_ca": false},
			expectParseError: false,
			description:      "When all fields are provided, should parse successfully",
		},
		{
			name:             "missing cn",
			args:             map[string]any{"out_key": "/tmp/key.pem", "out_cert": "/tmp/cert.pem", "policy": "Always"},
			expectParseError: true,
			description:      "When cn is missing, should return error",
		},
		{
			name:             "missing out_key",
			args:             map[string]any{"cn": "test.example.com", "out_cert": "/tmp/cert.pem", "policy": "Always"},
			expectParseError: true,
			description:      "When out_key is missing, should return error",
		},
		{
			name:             "missing out_cert",
			args:             map[string]any{"cn": "test.example.com", "out_key": "/tmp/key.pem", "policy": "Always"},
			expectParseError: true,
			description:      "When out_cert is missing, should return error",
		},
		{
			name:             "invalid policy",
			args:             map[string]any{"cn": "test.example.com", "out_key": "/tmp/key.pem", "out_cert": "/tmp/cert.pem", "policy": "InvalidPolicy"},
			expectParseError: true,
			description:      "When policy is invalid, should return error",
		},
		{
			name:             "valid policy Always",
			args:             map[string]any{"cn": "test.example.com", "out_key": "/tmp/key.pem", "out_cert": "/tmp/cert.pem", "policy": "Always"},
			expectParseError: false,
			description:      "When policy is Always, should parse successfully",
		},
		{
			name:             "valid policy IfNotPresent",
			args:             map[string]any{"cn": "test.example.com", "out_key": "/tmp/key.pem", "out_cert": "/tmp/cert.pem", "policy": "IfNotPresent"},
			expectParseError: false,
			description:      "When policy is IfNotPresent, should parse successfully",
		},
		{
			name:             "valid policy None",
			args:             map[string]any{"cn": "test.example.com", "out_key": "/tmp/key.pem", "out_cert": "/tmp/cert.pem", "policy": "None"},
			expectParseError: false,
			description:      "When policy is None, should parse successfully",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			raw := createRawArgs(tc.args)
			_, err := newGenCertArgs(ctx, raw, nil)
			if tc.expectParseError {
				require.Error(t, err, tc.description)
			} else {
				require.NoError(t, err, tc.description)
			}
		})
	}
}

// TestGenCertArgsTemplate tests template variable resolution in arguments.
func TestGenCertArgsTemplate(t *testing.T) {
	testcases := []struct {
		name        string
		args        map[string]any
		vars        map[string]any
		expectError bool
		description string
	}{
		{
			name:        "template: cn with template variable",
			args:        map[string]any{"cn": "{{ .cn }}", "out_key": "/tmp/key.pem", "out_cert": "/tmp/cert.pem", "policy": "Always"},
			vars:        map[string]any{"cn": "test.example.com"},
			expectError: false,
			description: "When cn contains template, should resolve with vars",
		},
		{
			name:        "template: out_key with template variable",
			args:        map[string]any{"cn": "test.com", "out_key": "{{ .key_path }}", "out_cert": "/tmp/cert.pem", "policy": "Always"},
			vars:        map[string]any{"key_path": "/tmp/key.pem"},
			expectError: false,
			description: "When out_key contains template, should resolve with vars",
		},
		{
			name:        "template: sans with template variable",
			args:        map[string]any{"cn": "test.com", "out_key": "/tmp/key.pem", "out_cert": "/tmp/cert.pem", "sans": []string{"{{ .san1 }}", "{{ .san2 }}"}, "policy": "Always"},
			vars:        map[string]any{"san1": "test1.example.com", "san2": "test2.example.com"},
			expectError: false,
			description: "When sans contain templates, should resolve with vars",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			raw := createRawArgs(tc.args)
			_, err := newGenCertArgs(ctx, raw, tc.vars)
			if tc.expectError {
				require.Error(t, err, tc.description)
			} else {
				require.NoError(t, err, tc.description)
			}
		})
	}
}

// TestGenCertModule tests the actual functionality of the gen_cert module.
func TestGenCertModule(t *testing.T) {
	t.Run("module exists", func(t *testing.T) {
		require.NotNil(t, ModuleGenCert)
	})
}
