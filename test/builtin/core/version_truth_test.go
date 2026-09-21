//go:build builtin
// +build builtin

/*
Copyright 2026 The KubeSphere Authors.

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

// Package core_test. Every CNI / CRI version of KubeKey is written down in four
// separate places and nothing keeps them in sync:
//
//	roles/defaults/templates/manifests.yaml          per-minor $default_<name>_version
//	                                                 accumulators, consumed by
//	                                                 `kk artifact` / `kk pull`
//	roles/defaults/vars/v1.<minor>.yaml              the runtime source of truth,
//	                                                 read by `kk create` / `kk upgrade`
//	builtin/core/defaults/config/v1.<minor>.yaml     the example values printed by
//	                                                 `kk create config`
//	roles/defaults/defaults/main/10-download.yaml    the image manifest keys, i.e.
//	                                                 what the mirror is asked to carry
//
// A drift between them fails silently. The packaging template renders
// `{{ range index $.download.images "<repo>" . }}`, which yields an empty section
// (no error at all) when the accumulated version has no key in the image
// manifest, and the packaging and install paths read two different variable trees
// (`download.cni.*` filled from manifests.yaml via include_vars vs `cni.*` filled
// from vars/), so an artifact can ship a chart / image version that the install
// step never asks for.
//
// This test pins the three derived copies to the runtime truth:
//
//	manifests.yaml  ==  vars/                     exact string, every minor, every
//	                                              tracked field
//	manifests.yaml  ⊂   keys of 10-download.yaml  CNI fields only
//	config/v1.<minor>.yaml  ==  vars/             create-config example values
package core_test

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"

	core "github.com/kubesphere/kubekey/v4/builtin/core"
)

// versionTruthField describes one `$default_<name>_version` accumulator of
// roles/defaults/templates/manifests.yaml and where its runtime truth lives in
// roles/defaults/vars/v1.<minor>.yaml.
type versionTruthField struct {
	// name is the <name> part of $default_<name>_version.
	name string
	// path is the nested field path inside roles/defaults/vars/v1.<minor>.yaml.
	path []string
	// repo is the download.images key the version must exist under. It is empty
	// for components that are not carried by the image manifest (helm, etcd and
	// the container runtime binaries come from artifact URLs, not from the
	// mirror).
	repo string
	// absentIn lists the Kubernetes minors whose manifest block legitimately
	// omits the field. 1.23 still ships the in-tree dockershim, so cri-dockerd is
	// not part of its offline payload even though vars/v1.23.yaml carries a
	// version for it.
	absentIn []int
}

// trackedVersionFields is the registry of every per-minor version accumulator in
// the packaging template. TestVersionTruthConsistency reports a failure when the
// template grows a $default_<name>_version that is not registered here, so a new
// component cannot silently escape the cross-check.
var trackedVersionFields = []versionTruthField{
	{name: "helm", path: []string{"kubernetes", "helm_version"}},
	{name: "etcd", path: []string{"etcd", "etcd_version"}},
	{name: "crictl", path: []string{"cri", "crictl_version"}},
	{name: "docker", path: []string{"cri", "docker_version"}},
	{name: "cridockerd", path: []string{"cri", "cridockerd_version"}, absentIn: []int{23}},
	{name: "containerd", path: []string{"cri", "containerd_version"}},
	{name: "runc", path: []string{"cri", "runc_version"}},
	{name: "calico", path: []string{"cni", "calico_version"}, repo: "projectcalico/tigera-operator"},
	{name: "cilium", path: []string{"cni", "cilium_version"}, repo: "cilium/cilium"},
	{name: "flannel", path: []string{"cni", "flannel_version"}, repo: "flannel/flannel"},
	{name: "kubeovn", path: []string{"cni", "kubeovn_version"}, repo: "kubeovn/kube-ovn"},
}

var (
	// manifestsMinorRe matches the minor selector of a manifest block, e.g.
	// `{{- else if slice (. | splitList ".") 0 2 | join "." | eq "v1.29" }}`.
	manifestsMinorRe = regexp.MustCompile(`join "\." \| eq "v1\.(\d+)"`)
	// manifestsAppendRe matches one accumulated default, e.g.
	// `{{- $default_cilium_version = append $default_cilium_version "1.16.19" }}`.
	manifestsAppendRe = regexp.MustCompile(`\$default_(\w+)_version = append \$default_\w+_version "([^"]+)"`)
	// manifestsOutputRe marks the end of the version blocks. Everything after the
	// rendered `download:` document carries a second, unrelated set of `eq "v1.XX"`
	// branches (pause / CoreDNS image tags) that must not be parsed as versions.
	manifestsOutputRe = regexp.MustCompile(`^download:$`)
	// configVersionReFormat matches a CNI version line of a create-config
	// template, commented out or not: `#cilium_version: 1.20.1` as well as the
	// active `calico_version: v3.32.2`.
	configVersionReFormat = `^\s*#?\s*%s_version:\s*(\S+)\s*$`
)

// parseManifestVersionBlocks returns, per Kubernetes minor, the versions appended
// to every $default_<name>_version accumulator of the offline artifact template,
// together with the minors in file order.
func parseManifestVersionBlocks(t *testing.T) (map[int]map[string][]string, []int) {
	t.Helper()

	raw, err := core.BuiltinPlaybook.ReadFile("roles/defaults/templates/manifests.yaml")
	require.NoError(t, err, "manifests template must be embedded")

	blocks := make(map[int]map[string][]string)
	var order []int
	minor := 0

	for _, line := range strings.Split(string(raw), "\n") {
		if manifestsOutputRe.MatchString(line) {
			break
		}
		if m := manifestsMinorRe.FindStringSubmatch(line); m != nil {
			minor, err = strconv.Atoi(m[1])
			require.NoError(t, err)
			if _, seen := blocks[minor]; !seen {
				blocks[minor] = make(map[string][]string)
				order = append(order, minor)
			}
			continue
		}
		if m := manifestsAppendRe.FindStringSubmatch(line); m != nil {
			require.NotZero(t, minor, "found %q before any minor block", strings.TrimSpace(line))
			blocks[minor][m[1]] = append(blocks[minor][m[1]], m[2])
		}
	}

	require.NotEmpty(t, blocks, "no minor version block parsed out of manifests.yaml")
	return blocks, order
}

// varsVersion returns the runtime truth of one version field for one minor.
func varsVersion(t *testing.T, minor int, path ...string) (string, bool) {
	t.Helper()

	raw, err := core.BuiltinPlaybook.ReadFile(fmt.Sprintf("roles/defaults/vars/v1.%d.yaml", minor))
	require.NoError(t, err, "vars overlay must be embedded for minor %d", minor)

	var doc map[string]interface{}
	require.NoError(t, yaml.Unmarshal(raw, &doc), "vars/v1.%d.yaml must be valid YAML", minor)

	v, found, err := unstructured.NestedString(doc, path...)
	require.NoError(t, err)
	return v, found
}

// downloadImageKeys returns the version keys declared for every repository of
// download.images, i.e. what the mirror is asked to carry.
func downloadImageKeys(t *testing.T) map[string]map[string]bool {
	t.Helper()

	raw, err := core.BuiltinPlaybook.ReadFile("roles/defaults/defaults/main/10-download.yaml")
	require.NoError(t, err, "image manifest must be embedded")

	var doc map[string]interface{}
	require.NoError(t, yaml.Unmarshal(raw, &doc))

	images, found, err := unstructured.NestedMap(doc, "download", "images")
	require.NoError(t, err)
	require.True(t, found, "download.images must be present in the image manifest")

	keys := make(map[string]map[string]bool, len(images))
	for repo, versions := range images {
		list, ok := versions.(map[string]interface{})
		if !ok {
			continue
		}
		set := make(map[string]bool, len(list))
		for v := range list {
			set[v] = true
		}
		keys[repo] = set
	}
	return keys
}

// configDocVersion returns the CNI version a `kk create config` template shows for
// one minor, whether the line is commented out (an example value) or active.
func configDocVersion(t *testing.T, minor int, cni string) (string, bool) {
	t.Helper()

	raw, err := core.Defaults.ReadFile(fmt.Sprintf("defaults/config/v1.%d.yaml", minor))
	require.NoError(t, err, "create-config template must be embedded for minor %d", minor)

	re := regexp.MustCompile(fmt.Sprintf(configVersionReFormat, regexp.QuoteMeta(cni)))
	for _, line := range strings.Split(string(raw), "\n") {
		if m := re.FindStringSubmatch(line); m != nil {
			return m[1], true
		}
	}
	return "", false
}

// TestVersionTruthConsistency pins the packaging defaults, the create-config
// examples and the image manifest keys to the runtime version truth in vars/.
func TestVersionTruthConsistency(t *testing.T) {
	byName := make(map[string]versionTruthField, len(trackedVersionFields))
	for _, f := range trackedVersionFields {
		byName[f.name] = f
	}

	blocks, order := parseManifestVersionBlocks(t)
	keys := downloadImageKeys(t)

	t.Run("manifests_defaults_match_vars", func(t *testing.T) {
		// Report unregistered accumulators first, so a run lists all of them at
		// once instead of stopping at the first one.
		for _, minor := range order {
			for name := range blocks[minor] {
				if _, ok := byName[name]; !ok {
					t.Errorf("manifests.yaml v1.%d sets $default_%s_version, which is missing from trackedVersionFields",
						minor, name)
				}
			}
		}

		for _, minor := range order {
			block := blocks[minor]
			for name, versions := range block {
				field, ok := byName[name]
				if !ok {
					continue
				}
				want, found := varsVersion(t, minor, field.path...)
				require.True(t, found,
					"vars/v1.%d.yaml has no %s while manifests.yaml sets $default_%s_version",
					minor, strings.Join(field.path, "."), name)
				for _, got := range versions {
					assert.Equal(t, want, got,
						"v1.%d: manifests.yaml $default_%s_version = %q but vars/ %s = %q",
						minor, name, got, strings.Join(field.path, "."), want)
				}
			}
			for _, field := range trackedVersionFields {
				if slices.Contains(field.absentIn, minor) {
					continue
				}
				assert.NotEmpty(t, block[field.name],
					"v1.%d: manifests.yaml sets no $default_%s_version, expected it to mirror vars/ %s",
					minor, field.name, strings.Join(field.path, "."))
			}
		}
	})

	t.Run("cni_defaults_exist_in_download_manifest", func(t *testing.T) {
		for _, minor := range order {
			for _, field := range trackedVersionFields {
				if field.repo == "" {
					continue
				}
				repoKeys := keys[field.repo]
				require.NotEmpty(t, repoKeys, "download.images carries no %q key", field.repo)
				for _, v := range blocks[minor][field.name] {
					assert.True(t, repoKeys[v],
						"v1.%d: defaults %s to %q, but download.images[%q] has no such version key",
						minor, field.name, v, field.repo)
				}
			}
		}
	})

	t.Run("create_config_docs_match_vars", func(t *testing.T) {
		for _, minor := range order {
			for _, field := range trackedVersionFields {
				if slices.Contains(field.absentIn, minor) {
					continue
				}
				doc, found := configDocVersion(t, minor, field.name)
				require.True(t, found,
					"defaults/config/v1.%d.yaml has no %s_version line", minor, field.name)
				want, _ := varsVersion(t, minor, field.path...)
				assert.Equal(t, want, doc,
					"v1.%d: defaults/config shows %s_version: %s but vars/ sets %s",
					minor, field.name, doc, want)
			}
		}
	})
}
