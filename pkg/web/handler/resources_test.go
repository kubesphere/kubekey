package handler

import (
	"path/filepath"
	"testing"
)

func TestResolveSchemaPlaybookPath(t *testing.T) {
	tests := []struct {
		name         string
		schemaRoot   string
		playbookPath string
		want         string
	}{
		{
			name:         "schema root rewrites bundled web installer prefix",
			schemaRoot:   filepath.Join("tmp", "web-installer-out", "schema"),
			playbookPath: filepath.ToSlash(filepath.Join("web-installer", "kubernetes", "playbooks", "precheck.yaml")),
			want:         filepath.Join("tmp", "web-installer-out", "kubernetes", "playbooks", "precheck.yaml"),
		},
		{
			name:         "schema root resolves relative path without prefix",
			schemaRoot:   filepath.Join("tmp", "web-installer-out", "schema"),
			playbookPath: filepath.ToSlash(filepath.Join("kubernetes", "playbooks", "precheck.yaml")),
			want:         filepath.Join("tmp", "web-installer-out", "kubernetes", "playbooks", "precheck.yaml"),
		},
		{
			name:         "empty schema root keeps relative path unchanged",
			schemaRoot:   "",
			playbookPath: filepath.ToSlash(filepath.Join("web-installer", "kubernetes", "playbooks", "precheck.yaml")),
			want:         filepath.ToSlash(filepath.Join("web-installer", "kubernetes", "playbooks", "precheck.yaml")),
		},
		{
			name:         "absolute path is already resolved",
			schemaRoot:   filepath.Join("tmp", "web-installer-out", "schema"),
			playbookPath: filepath.Join(string(filepath.Separator), "opt", "web-installer", "kubernetes", "playbooks", "precheck.yaml"),
			want:         filepath.Join(string(filepath.Separator), "opt", "web-installer", "kubernetes", "playbooks", "precheck.yaml"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveSchemaPlaybookPath(tt.schemaRoot, tt.playbookPath)
			if got != tt.want {
				t.Fatalf("resolveSchemaPlaybookPath() = %q, want %q", got, tt.want)
			}
		})
	}
}
