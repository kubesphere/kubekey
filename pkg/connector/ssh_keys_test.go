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
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

func testPrivateKeyPEM(t *testing.T) []byte {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatalf("generate private key: %v", err)
	}
	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
}

func writeTestPrivateKey(t *testing.T, dir, name string) string {
	t.Helper()
	keyPath := filepath.Join(dir, name)
	if err := os.WriteFile(keyPath, testPrivateKeyPEM(t), 0o600); err != nil {
		t.Fatalf("write private key: %v", err)
	}
	return keyPath
}

func TestDefaultPrivateKeyPaths(t *testing.T) {
	t.Run("loads all parsable private keys with preferred order", func(t *testing.T) {
		homeDir := t.TempDir()
		sshDir := filepath.Join(homeDir, ".ssh")
		if err := os.MkdirAll(sshDir, 0o700); err != nil {
			t.Fatalf("create ssh dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(sshDir, "known_hosts"), []byte("not a private key"), 0o600); err != nil {
			t.Fatalf("write known_hosts: %v", err)
		}
		clusterAccess := writeTestPrivateKey(t, sshDir, "cluster-access")
		idRSA := writeTestPrivateKey(t, sshDir, "id_rsa")
		idEd25519 := writeTestPrivateKey(t, sshDir, "id_ed25519")

		got := defaultPrivateKeyPaths(homeDir)
		want := []string{idEd25519, idRSA, clusterAccess}
		if len(got) != len(want) {
			t.Fatalf("defaultPrivateKeyPaths() = %#v, want %#v", got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("defaultPrivateKeyPaths()[%d] = %q, want %q (full=%#v)", i, got[i], want[i], got)
			}
		}
	})

	t.Run("returns empty when no parsable private key exists", func(t *testing.T) {
		homeDir := t.TempDir()
		sshDir := filepath.Join(homeDir, ".ssh")
		if err := os.MkdirAll(sshDir, 0o700); err != nil {
			t.Fatalf("create ssh dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(sshDir, "config"), []byte("Host *"), 0o600); err != nil {
			t.Fatalf("write config: %v", err)
		}

		if got := defaultPrivateKeyPaths(homeDir); len(got) != 0 {
			t.Fatalf("defaultPrivateKeyPaths() = %#v, want empty", got)
		}
	})

	t.Run("returns empty when ssh dir does not exist", func(t *testing.T) {
		if got := defaultPrivateKeyPaths(t.TempDir()); len(got) != 0 {
			t.Fatalf("defaultPrivateKeyPaths() = %#v, want empty", got)
		}
	})
}

func TestDefaultPrivateKeySigners(t *testing.T) {
	homeDir := t.TempDir()
	sshDir := filepath.Join(homeDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		t.Fatalf("create ssh dir: %v", err)
	}
	writeTestPrivateKey(t, sshDir, "id_rsa")
	writeTestPrivateKey(t, sshDir, "cluster-access")

	signers, err := DefaultPrivateKeySigners(homeDir)
	if err != nil {
		t.Fatalf("DefaultPrivateKeySigners() error = %v", err)
	}
	if len(signers) != 2 {
		t.Fatalf("len(signers) = %d, want 2", len(signers))
	}
}
