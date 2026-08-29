package tmpl

import (
	"testing"
)

// singleMinorGuard is the `that` assertion expression used by the
// "Kubernetes | Allow only a single-minor upgrade per invocation" task in
// builtin/core/roles/precheck/kubernetes/tasks/main.yaml. It is kept in sync
// with that playbook: any change to the guard must be mirrored here.
//
// The running version is the locally-installed kubelet version
// (.kubernetes_install_version.stdout, e.g. "Kubernetes v1.33.1"), normalized by
// stripping the "Kubernetes " prefix. It flattens each version to
// (major*1000+minor) and asserts targetFlat <= runningFlat + 1, so that both a
// two-minor leap within the same major and a major bump are rejected, while a
// patch-only (0 minor) or +1-minor step passes.
const singleMinorGuard = `{{- $run := .kubernetes_install_version.stdout | default "" | trimPrefix "Kubernetes " -}}
{{- $rmaj := index (splitList "." ($run | trimPrefix "v")) 0 | atoi -}}
{{- $rmin := index (splitList "." ($run | trimPrefix "v")) 1 | atoi -}}
{{- $tmaj := index (splitList "." (.kubernetes.kube_version | trimPrefix "v")) 0 | atoi -}}
{{- $tmin := index (splitList "." (.kubernetes.kube_version | trimPrefix "v")) 1 | atoi -}}
{{- $rf := ($rmaj | mul 1000) | add $rmin -}}
{{- $tf := ($tmaj | mul 1000) | add $tmin -}}
{{- ($rf | add 1) | le $tf -}}`

func TestSingleMinorUpgradeGuard(t *testing.T) {
	cases := []struct {
		name    string
		running string // .kubernetes_install_version.stdout ("Kubernetes vX.Y.Z")
		target  string // .kubernetes.kube_version
		want    bool
	}{
		{"same minor patch", "Kubernetes v1.33.1", "v1.33.5", true},
		{"single minor", "Kubernetes v1.33.5", "v1.34.3", true},
		{"leap two minors", "Kubernetes v1.23.5", "v1.25.0", false},
		{"leap many minors", "Kubernetes v1.23.5", "v1.34.3", false},
		{"cross major", "Kubernetes v1.23.5", "v2.0.0", false},
		{"kubelet up to date", "Kubernetes v1.34.3", "v1.34.3", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := map[string]any{
				"kubernetes_install_version": map[string]any{"stdout": tc.running},
				"kubernetes": map[string]any{
					"kube_version": tc.target,
				},
			}
			b, err := ParseBool(v, singleMinorGuard)
			if err != nil {
				t.Fatalf("ParseBool error: %v", err)
			}
			if b != tc.want {
				t.Fatalf("got %v, want %v (running=%s target=%s)", b, tc.want, tc.running, tc.target)
			}
		})
	}
}
