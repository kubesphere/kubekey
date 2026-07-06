package options

import (
	"path/filepath"
	"testing"
)

func TestNewKubeKeyWebOptionsDefaults(t *testing.T) {
	opts := NewKubeKeyWebOptions()

	if opts.SchemaPath != filepath.Join(defaultWebInstallerPath, "schema") {
		t.Fatalf("SchemaPath = %q, want %q", opts.SchemaPath, filepath.Join(defaultWebInstallerPath, "schema"))
	}
	if opts.UIPath != filepath.Join(defaultWebInstallerPath, "dist") {
		t.Fatalf("UIPath = %q, want %q", opts.UIPath, filepath.Join(defaultWebInstallerPath, "dist"))
	}
	if got := opts.HostCheckPlaybookPath(); got != filepath.Join(defaultWebInstallerPath, "host_check.yaml") {
		t.Fatalf("HostCheckPlaybookPath() = %q, want %q", got, filepath.Join(defaultWebInstallerPath, "host_check.yaml"))
	}
}

func TestHostCheckPlaybookPathUsesSchemaRoot(t *testing.T) {
	opts := &KubeKeyWebOptions{
		SchemaPath: filepath.Join("custom", "bundle", "schema"),
	}

	if got := opts.HostCheckPlaybookPath(); got != filepath.Join("custom", "bundle", "host_check.yaml") {
		t.Fatalf("HostCheckPlaybookPath() = %q, want %q", got, filepath.Join("custom", "bundle", "host_check.yaml"))
	}
}
