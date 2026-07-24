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
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cockroachdb/errors"
	"golang.org/x/crypto/ssh"
	"k8s.io/klog/v2"
)

var preferredSSHPrivateKeyNames = []string{
	"id_ed25519",
	"id_ecdsa",
	"id_rsa",
	"id_dsa",
}

var sshPrivateKeySkipFiles = map[string]struct{}{
	"authorized_keys":  {},
	"authorized_keys2": {},
	"config":           {},
	"environment":      {},
	"known_hosts":      {},
}

func isSSHPrivateKeyCandidate(name string) bool {
	if strings.HasSuffix(name, ".pub") {
		return false
	}
	if _, skip := sshPrivateKeySkipFiles[name]; skip {
		return false
	}
	return true
}

func defaultPrivateKeyPaths(homeDir string) []string {
	sshDir := filepath.Join(homeDir, ".ssh")
	entries, err := os.ReadDir(sshDir)
	if err != nil {
		klog.V(4).InfoS("failed to read ssh dir for default private keys", "dir", sshDir, "error", err)
		return nil
	}

	preferred := make(map[string]string, len(preferredSSHPrivateKeyNames))
	others := make([]string, 0)

	for _, entry := range entries {
		if entry.IsDir() || !isSSHPrivateKeyCandidate(entry.Name()) {
			continue
		}

		keyPath := filepath.Join(sshDir, entry.Name())
		if !isParsablePrivateKeyFile(keyPath) {
			continue
		}

		name := entry.Name()
		if isPreferredSSHPrivateKeyName(name) {
			preferred[name] = keyPath
			continue
		}
		others = append(others, keyPath)
	}

	sort.Strings(others)

	paths := make([]string, 0, len(preferred)+len(others))
	for _, name := range preferredSSHPrivateKeyNames {
		if keyPath, ok := preferred[name]; ok {
			paths = append(paths, keyPath)
		}
	}
	paths = append(paths, others...)
	return paths
}

func isPreferredSSHPrivateKeyName(name string) bool {
	for _, preferred := range preferredSSHPrivateKeyNames {
		if name == preferred {
			return true
		}
	}
	return false
}

func isParsablePrivateKeyFile(keyPath string) bool {
	key, err := os.ReadFile(keyPath)
	if err != nil {
		klog.V(4).InfoS("failed to read ssh private key candidate", "path", keyPath, "error", err)
		return false
	}
	if _, err := ssh.ParsePrivateKey(key); err != nil {
		klog.V(4).InfoS("skip non-private-key ssh file", "path", keyPath, "error", err)
		return false
	}
	return true
}

func privateKeySignersFromPaths(paths []string) ([]ssh.Signer, error) {
	signers := make([]ssh.Signer, 0, len(paths))
	for _, keyPath := range paths {
		key, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read private key %q", keyPath)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse private key %q", keyPath)
		}
		signers = append(signers, signer)
	}
	return signers, nil
}

func privateKeySignerFromFile(keyPath string) (ssh.Signer, error) {
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read private key %q", keyPath)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse private key %q", keyPath)
	}
	return signer, nil
}

// DefaultPrivateKeySigners loads all parsable private keys from ~/.ssh under homeDir.
func DefaultPrivateKeySigners(homeDir string) ([]ssh.Signer, error) {
	return privateKeySignersFromPaths(defaultPrivateKeyPaths(homeDir))
}

func userHomeDir() (string, error) {
	if currentUser, err := user.Current(); err == nil && currentUser.HomeDir != "" {
		return currentUser.HomeDir, nil
	}
	return os.UserHomeDir()
}
