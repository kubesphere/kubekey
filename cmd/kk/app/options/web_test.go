package options

import (
	"path/filepath"
	"testing"
)

func TestNewKubeKeyWebOptionsDefaults(t *testing.T) {
	opts := NewKubeKeyWebOptions()
	wantHostCheckPlaybookPath, err := filepath.Abs(filepath.Join(defaultWebInstallerPath, defaultHostCheckPlaybookRelPath))
	if err != nil {
		t.Fatalf("filepath.Abs() failed: %v", err)
	}

	if opts.SchemaPath != filepath.Join(defaultWebInstallerPath, "schema") {
		t.Fatalf("SchemaPath = %q, want %q", opts.SchemaPath, filepath.Join(defaultWebInstallerPath, "schema"))
	}
	if opts.UIPath != filepath.Join(defaultWebInstallerPath, "dist") {
		t.Fatalf("UIPath = %q, want %q", opts.UIPath, filepath.Join(defaultWebInstallerPath, "dist"))
	}
	if opts.HostCheckPlaybookPath != filepath.Join(defaultWebInstallerPath, defaultHostCheckPlaybookRelPath) {
		t.Fatalf("HostCheckPlaybookPath = %q, want %q", opts.HostCheckPlaybookPath, filepath.Join(defaultWebInstallerPath, defaultHostCheckPlaybookRelPath))
	}
	if got := opts.ResolvedHostCheckPlaybookPath(); got != wantHostCheckPlaybookPath {
		t.Fatalf("ResolvedHostCheckPlaybookPath() = %q, want %q", got, wantHostCheckPlaybookPath)
	}
}

func TestResolvedHostCheckPlaybookPathIsIndependentFromSchemaPath(t *testing.T) {
	opts := &KubeKeyWebOptions{
		SchemaPath:            filepath.Join("custom", "bundle", "schema"),
		HostCheckPlaybookPath: filepath.Join("custom", "bundle", "host_check.yaml"),
	}
	wantHostCheckPlaybookPath, err := filepath.Abs(filepath.Join("custom", "bundle", "host_check.yaml"))
	if err != nil {
		t.Fatalf("filepath.Abs() failed: %v", err)
	}

	if got := opts.ResolvedHostCheckPlaybookPath(); got != wantHostCheckPlaybookPath {
		t.Fatalf("ResolvedHostCheckPlaybookPath() = %q, want %q", got, wantHostCheckPlaybookPath)
	}
}
