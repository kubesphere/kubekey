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

package connector

import (
	"context"
	"strings"
	"testing"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
)

// TestNewSSHConnector_PrivateKeyPriority tests the private key parameter extraction
// and priority logic in the newSSHConnector function
func TestNewSSHConnector_PrivateKeyPriority(t *testing.T) {
	testCases := []struct {
		name                   string
		hostVars               map[string]any
		expectedPrivateKey     string
		expectedKeyContent     string
		expectedUseDefaultKeys bool
		description            string
	}{
		{
			name: "custom private_key without content",
			hostVars: map[string]any{
				_const.VariableConnector: map[string]any{
					_const.VariableConnectorPrivateKey: "/custom/.ssh/cluster-access",
				},
			},
			expectedPrivateKey:     "/custom/.ssh/cluster-access",
			expectedKeyContent:     "",
			expectedUseDefaultKeys: false,
			description:            "When only private_key is set, it should be preserved",
		},
		{
			name: "only private_key_content set",
			hostVars: map[string]any{
				_const.VariableConnector: map[string]any{
					_const.VariableConnectorPrivateKeyContent: "-----BEGIN OPENSSH PRIVATE KEY-----\ntest content\n-----END OPENSSH PRIVATE KEY-----",
				},
			},
			expectedPrivateKey:     "",
			expectedKeyContent:     "-----BEGIN OPENSSH PRIVATE KEY-----\ntest content\n-----END OPENSSH PRIVATE KEY-----",
			expectedUseDefaultKeys: false,
			description:            "When only private_key_content is set, default ~/.ssh keys should not be loaded",
		},
		{
			name: "both private_key and private_key_content set",
			hostVars: map[string]any{
				_const.VariableConnector: map[string]any{
					_const.VariableConnectorPrivateKey:        "/custom/.ssh/cluster-access",
					_const.VariableConnectorPrivateKeyContent: "-----BEGIN OPENSSH PRIVATE KEY-----\ntest content\n-----END OPENSSH PRIVATE KEY-----",
				},
			},
			expectedPrivateKey:     "/custom/.ssh/cluster-access",
			expectedKeyContent:     "-----BEGIN OPENSSH PRIVATE KEY-----\ntest content\n-----END OPENSSH PRIVATE KEY-----",
			expectedUseDefaultKeys: false,
			description:            "When both are set, both should be preserved (content takes priority in Init())",
		},
		{
			name:                   "neither private_key nor private_key_content set",
			hostVars:               map[string]any{},
			expectedPrivateKey:     "",
			expectedKeyContent:     "",
			expectedUseDefaultKeys: true,
			description:            "When neither is set, Init() should load all default ~/.ssh private keys",
		},
		{
			name: "empty connector variable",
			hostVars: map[string]any{
				_const.VariableConnector: map[string]any{},
			},
			expectedPrivateKey:     "",
			expectedKeyContent:     "",
			expectedUseDefaultKeys: true,
			description:            "When connector exists but keys are not set, Init() should load all default ~/.ssh private keys",
		},
		{
			name: "custom key path",
			hostVars: map[string]any{
				_const.VariableConnector: map[string]any{
					_const.VariableConnectorPrivateKey: "~/.ssh/cluster-access",
				},
			},
			expectedPrivateKey:     "~/.ssh/cluster-access",
			expectedKeyContent:     "",
			expectedUseDefaultKeys: false,
			description:            "Custom key paths should be preserved",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			connector := newSSHConnector("/tmp/workdir", "test-host", tc.hostVars)

			if connector.PrivateKey != tc.expectedPrivateKey {
				t.Errorf("%s\nExpected PrivateKey: %q\nGot: %q",
					tc.description, tc.expectedPrivateKey, connector.PrivateKey)
			}

			if connector.PrivateKeyContent != tc.expectedKeyContent {
				t.Errorf("%s\nExpected PrivateKeyContent: %q\nGot: %q",
					tc.description, tc.expectedKeyContent, connector.PrivateKeyContent)
			}

			if connector.useDefaultPrivateKeys != tc.expectedUseDefaultKeys {
				t.Errorf("%s\nExpected useDefaultPrivateKeys: %t\nGot: %t",
					tc.description, tc.expectedUseDefaultKeys, connector.useDefaultPrivateKeys)
			}
		})
	}
}

// TestNewSSHConnector_DefaultParameters tests that other default parameters
// are set correctly when not provided
func TestNewSSHConnector_DefaultParameters(t *testing.T) {
	testCases := []struct {
		name         string
		hostVars     map[string]any
		expectedHost string
		expectedPort int
		expectedUser string
	}{
		{
			name:         "all defaults",
			hostVars:     map[string]any{},
			expectedHost: "test-host",
			expectedPort: defaultSSHPort,
			expectedUser: defaultSSHUser,
		},
		{
			name: "custom host, port, user",
			hostVars: map[string]any{
				_const.VariableConnector: map[string]any{
					_const.VariableConnectorHost: "custom-host",
					_const.VariableConnectorPort: 2222,
					_const.VariableConnectorUser: "ubuntu",
				},
			},
			expectedHost: "custom-host",
			expectedPort: 2222,
			expectedUser: "ubuntu",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			connector := newSSHConnector("/tmp/workdir", "test-host", tc.hostVars)

			if connector.Host != tc.expectedHost {
				t.Errorf("Expected Host: %q, Got: %q", tc.expectedHost, connector.Host)
			}

			if connector.Port != tc.expectedPort {
				t.Errorf("Expected Port: %d, Got: %d", tc.expectedPort, connector.Port)
			}

			if connector.User != tc.expectedUser {
				t.Errorf("Expected User: %q, Got: %q", tc.expectedUser, connector.User)
			}
		})
	}
}

func TestNewSSHConnector_GatherFactsCacheUsesInventoryHost(t *testing.T) {
	connector := newSSHConnector("/tmp/workdir", "test-host", map[string]any{})

	if connector.gatherFacts.inventoryName != "test-host" {
		t.Fatalf("gatherFacts.inventoryName = %q, want %q", connector.gatherFacts.inventoryName, "test-host")
	}
}

// TestSSHConnector_InitValidation tests the Init() method validation logic
// Note: Full integration testing with actual SSH connections would require
// a mock SSH server, which is beyond the scope of unit tests. These tests
// verify the validation logic without making actual connections.
func TestSSHConnector_InitValidation(t *testing.T) {
	testCases := []struct {
		name        string
		connector   *sshConnector
		shouldError bool
		errorMsg    string
	}{
		{
			name: "no host set",
			connector: &sshConnector{
				Host: "",
			},
			shouldError: true,
			errorMsg:    "host is not set",
		},
		{
			name: "no authentication methods",
			connector: &sshConnector{
				Host:              "test-host",
				Port:              22,
				User:              "root",
				Password:          "",
				PrivateKey:        "",
				PrivateKeyContent: "",
			},
			shouldError: true,
			errorMsg:    "no authentication method available",
		},
		{
			name: "explicit private key path that doesn't exist (should fail)",
			connector: &sshConnector{
				Host:              "test-host",
				Port:              22,
				User:              "root",
				Password:          "test-password",
				PrivateKey:        "/tmp/custom/nonexistent/key.pem",
				PrivateKeyContent: "",
			},
			shouldError: true,
			errorMsg:    "private key file not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.connector.Init(context.TODO())

			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error containing %q, but got no error", tc.errorMsg)
				} else if !strings.Contains(err.Error(), tc.errorMsg) {
					t.Errorf("Expected error containing %q, but got: %v", tc.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}
			}
		})
	}
}
