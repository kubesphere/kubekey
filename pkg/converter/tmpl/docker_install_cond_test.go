package tmpl

import "testing"

// dockerInstallCond is the install/provision gate used by the Docker roles. It
// evaluates to true (do the install/config work) whenever the client-side docker
// version probe differs from the pinned cri.docker_version. Mirror of
// builtin/core/roles/cri/docker/tasks/main.yaml and
// builtin/capkk/roles/install/cri/tasks/install_docker.yaml (they use .cri.docker_version
// and bare .docker_version respectively).
const dockerInstallCond = `or (.docker_install_version.error | empty | not) (.docker_install_version.stdout | ne .cri.docker_version)`

func TestDockerInstallCond(t *testing.T) {
	cases := []struct {
		name    string
		stdout  string
		err     string
		target  string // .cri.docker_version
		want    bool
	}{
		{"matching -> no install", "24.0.9", "", "24.0.9", false},
		{"mismatch -> install", "25.0.5", "", "24.0.9", true},
		{"not installed (empty stdout) -> install", "", "", "24.0.9", true},
		{"probe error -> install", "", "command failed", "24.0.9", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := map[string]any{
				"docker_install_version": map[string]any{"stdout": tc.stdout, "error": tc.err},
				"cri":                    map[string]any{"docker_version": tc.target},
			}
			b, err := ParseBool(v, "{{ "+dockerInstallCond+" }}")
			if err != nil {
				t.Fatalf("ParseBool error: %v", err)
			}
			if b != tc.want {
				t.Fatalf("got %v, want %v (stdout=%q target=%q)", b, tc.want, tc.stdout, tc.target)
			}
		})
	}
}
