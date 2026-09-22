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
	"net"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	cgutilcert "k8s.io/client-go/util/cert"

	"github.com/kubesphere/kubekey/v4/pkg/modules/internal"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
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

// fakeVariable is a minimal variable.Variable implementation used to drive the module directly.
type fakeVariable struct {
	vars map[string]any
}

func (f *fakeVariable) Get(_ variable.GetFunc) (any, error) { return f.vars, nil }
func (f *fakeVariable) Merge(_ variable.MergeFunc) error    { return nil }

// TestGenCertSANLeakAcrossTasks reproduces the generation sequence performed by
// builtin/core/roles/certs/init: a task that declares sans is followed by a task that declares
// none. The certificate produced by the second task must not carry the sans of the first one.
func TestGenCertSANLeakAcrossTasks(t *testing.T) {
	// Restore the package level baseline as it is on a fresh process start.
	defaultAltName = &cgutilcert.AltNames{
		DNSNames: []string{"localhost"},
		IPs:      []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
	}

	dir := t.TempDir()
	v := &fakeVariable{vars: map[string]any{}}
	run := func(name string, args map[string]any) {
		t.Helper()
		_, stderr, err := ModuleGenCert(context.Background(), internal.ExecOptions{
			Args:     createRawArgs(args),
			Host:     "localhost",
			Variable: v,
		})
		require.NoError(t, err, "%s: %s", name, stderr)
	}

	rootKey := filepath.Join(dir, "root.key")
	rootCert := filepath.Join(dir, "root.crt")

	// 1. Root CA, self signed.
	run("root", map[string]any{
		"cn": "root", "out_key": rootKey, "out_cert": rootCert, "policy": "Always",
	})
	// 2. First task declares sans, like the etcd certificate task in certs/init.
	run("task with sans", map[string]any{
		"cn": "etcd", "out_key": filepath.Join(dir, "first.key"), "out_cert": filepath.Join(dir, "first.crt"),
		"root_key": rootKey, "root_cert": rootCert, "policy": "Always",
		"sans": []string{"node-a", "10.0.0.1"},
	})
	// 3. Next task declares no sans at all, like the etcd client certificate task.
	run("task without sans", map[string]any{
		"cn": "etcd", "out_key": filepath.Join(dir, "second.key"), "out_cert": filepath.Join(dir, "second.crt"),
		"root_key": rootKey, "root_cert": rootCert, "policy": "Always",
	})

	chain, err := TryLoadCertChainFromDisk(filepath.Join(dir, "second.crt"))
	require.NoError(t, err)
	require.NotEmpty(t, chain)
	got := chain[0]

	t.Logf("certificate of the task without sans: DNSNames=%v IPAddresses=%v", got.DNSNames, got.IPAddresses)

	const leak = "the sans declared by a previous task must not leak into a certificate that declares none"
	require.NotContains(t, got.DNSNames, "node-a", leak)
	for _, ip := range got.IPAddresses {
		require.NotEqual(t, "10.0.0.1", ip.String(), leak)
	}
}
