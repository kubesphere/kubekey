package tmpl

import (
	"bytes"
	"testing"
	"text/template"
)

func render(t *testing.T, tpl string, data any) (string, error) {
	t.Helper()
	tmpl := template.New("probe").Funcs(funcMap())
	tmpl, err := tmpl.Parse(tpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	return buf.String(), err
}

func TestProbeHas(t *testing.T) {
	list := []any{"v1.23", "v1.24", "v1.25"}
	cases := []string{
		`{{ has "v1.24" .list }}`,
		`{{ has .list "v1.24" }}`,
		`{{ .list | has "v1.24" }}`,
		`{{ "v1.24" | has .list }}`,
	}
	for _, c := range cases {
		out, err := render(t, c, map[string]any{"list": list})
		t.Logf("tpl=%-40q => out=%q err=%v", c, out, err)
	}
}

func TestCalicoFull(t *testing.T) {
	ctx := map[string]any{
		"cluster_require": map[string]any{
			"calico_allowed_versions": map[string]any{
				"v3.26": []any{"v1.23", "v1.24", "v1.25", "v1.26", "v1.27", "v1.28"},
			},
		},
		"cni":        map[string]any{"calico_version": "v3.26.5"},
		"kubernetes": map[string]any{"kube_version": "v1.24.17"},
	}
	// Candidate corrected forms
	forms := map[string]string{
		"A_index_first_pipe":      `index .cluster_require.calico_allowed_versions (slice (.cni.calico_version | splitList ".") 0 2 | join ".") | has (slice (.kubernetes.kube_version | splitList ".") 0 2 | join ".")`,
		"B_has_needle_first":      `has (slice (.kubernetes.kube_version | splitList ".") 0 2 | join ".") (index .cluster_require.calico_allowed_versions (slice (.cni.calico_version | splitList ".") 0 2 | join "."))`,
		"C_index_piped_as_needle": `(index .cluster_require.calico_allowed_versions (slice (.cni.calico_version | splitList ".") 0 2 | join ".")) | has (slice (.kubernetes.kube_version | splitList ".") 0 2 | join ".")`,
	}
	for name, form := range forms {
		out, err := render(t, `{{ `+form+` }}`, ctx)
		t.Logf("%-26s => out=%q err=%v", name, out, err)
	}
}
