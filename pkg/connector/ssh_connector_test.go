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
	"bufio"
	"context"
	"os"
	"os/exec"
	"path/filepath"
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

// TestExecuteCommand_SudoPasswordPromptDeliversPassword reproduces the bug
// reported in https://github.com/kubesphere/kubekey/issues/2412 : when sudo
// on the remote host requires a password (NOPASSWD is not configured), the
// interactive password prompt must be read from sudo's own stdin, and the
// password written by ExecuteCommand's read/write loop must reach it.
//
// Before the buildSudoCommand fix, the script was piped into sudo's stdin via
// a heredoc ("sudo -E <shell> << 'EOF' ... EOF"), so sudo consumed lines of
// the script itself as bogus password attempts and always failed, exactly
// matching the "Sorry, try again" / "3 incorrect password attempts" output in
// the issue -- regardless of whether the real password was correct.
//
// This test drives the exact command built by buildSudoCommand through a
// local subprocess (no real SSH/sudo involved) using a fake "sudo" on PATH
// that mimics the real prompt/read behavior, and reuses the same byte-by-byte
// prompt-detection loop ExecuteCommand uses over the SSH session's stdio.
func TestExecuteCommand_SudoPasswordPromptDeliversPassword(t *testing.T) {
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not available")
	}

	const (
		user     = "testuser"
		password = "s3cret"
	)

	binDir := t.TempDir()
	fakeSudo := filepath.Join(binDir, "sudo")
	fakeSudoScript := `#!/bin/bash
shift # drop -E
attempts=0
while [ $attempts -lt 3 ]; do
  printf '[sudo] password for ` + user + `: '
  IFS= read -r pass
  if [ "$pass" = "` + password + `" ]; then
    exec "$@"
  fi
  echo ""
  echo "Sorry, try again."
  attempts=$((attempts+1))
done
echo "sudo: 3 incorrect password attempts" >&2
exit 1
`
	if err := os.WriteFile(fakeSudo, []byte(fakeSudoScript), 0o755); err != nil { //nolint:gosec // test fixture, needs to be executable
		t.Fatalf("write fake sudo: %v", err)
	}

	script := "echo IT_WORKED"
	cmd := buildSudoCommand(user, "bash", script)

	c := exec.Command("bash", "-c", cmd)
	c.Env = append(os.Environ(), "PATH="+binDir+":"+os.Getenv("PATH"))

	in, err := c.StdinPipe()
	if err != nil {
		t.Fatalf("stdin pipe: %v", err)
	}
	out, err := c.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	if err := c.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}

	// Same byte-by-byte prompt-detection loop as sshConnector.ExecuteCommand.
	var output []byte
	line := ""
	r := bufio.NewReader(out)
	for {
		b, err := r.ReadByte()
		if err != nil {
			break
		}
		output = append(output, b)
		if b == '\n' {
			line = ""
			continue
		}
		line += string(b)
		if (strings.HasPrefix(line, "[sudo] password for ") || strings.HasPrefix(line, "Password")) && strings.HasSuffix(line, ": ") {
			if _, err := in.Write([]byte(password + "\n")); err != nil {
				break
			}
		}
	}

	waitErr := c.Wait()
	outStr := string(output)

	if waitErr != nil {
		t.Fatalf("expected sudo to accept the password and run the script, got error: %v\noutput:\n%s", waitErr, outStr)
	}
	if !strings.Contains(outStr, "IT_WORKED") {
		t.Fatalf("expected script output %q in output, got:\n%s", "IT_WORKED", outStr)
	}
	if strings.Contains(outStr, "incorrect password") {
		t.Fatalf("sudo reported incorrect password attempts, the write to stdin did not reach it:\n%s", outStr)
	}
}

// TestBuildSudoCommand_PreservesSpecialCharacters ensures the quoting used to
// pass the script via "sudo ... -c \"$(cat <<'EOF' ... EOF)\"" round-trips
// shell metacharacters (quotes, backticks, "$(...)") in the script literally,
// without the outer command line re-interpreting or executing them.
func TestBuildSudoCommand_PreservesSpecialCharacters(t *testing.T) {
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not available")
	}

	script := "echo \"hello \\\"world\\\"\"\n" +
		"echo 'literal: $(id) and `whoami`'"

	cmd := buildSudoCommand("root", "bash", script)
	// Use the real "sudo" -> here none is needed since target user matches
	// (bypass by aliasing sudo to a passthrough), so this test only exercises
	// quoting, not the password prompt path.
	binDir := t.TempDir()
	passthroughSudo := filepath.Join(binDir, "sudo")
	if err := os.WriteFile(passthroughSudo, []byte("#!/bin/bash\nshift\nexec \"$@\"\n"), 0o755); err != nil { //nolint:gosec // test fixture, needs to be executable
		t.Fatalf("write passthrough sudo: %v", err)
	}

	c := exec.Command("bash", "-c", cmd)
	c.Env = append(os.Environ(), "PATH="+binDir+":"+os.Getenv("PATH"))
	output, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("run: %v\noutput:\n%s", err, output)
	}

	want := "hello \"world\"\nliteral: $(id) and `whoami`\n"
	if string(output) != want {
		t.Fatalf("output mismatch:\ngot:  %q\nwant: %q", output, want)
	}
}
