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

package core

import (
	"fmt"
	"strconv"
	"strings"

	"sigs.k8s.io/yaml"
)

// clusterRequirePath is the embedded builtin defaults file that defines the
// upgrade path. It is embedded via the `BuiltinPlaybook` FS (//go:embed playbooks roles).
const clusterRequirePath = "roles/defaults/defaults/main/01-cluster_require.yaml"

// KubeUpgradePath maps a Kubernetes minor version (e.g. 1.23) to the highest
// recommended patch version (e.g. "v1.23.17") for that minor, as defined in the
// builtin cluster_require defaults. kk consumes it to split a multi-minor
// upgrade into a sequence of single-minor steps (kubeadm only allows upgrading
// one minor at a time).
type KubeUpgradePath map[int]string

// ParseKubeUpgradePath reads the embedded builtin defaults and returns the
// minor -> highest-patch upgrade path. The data lives in
// roles/defaults/defaults/main/01-cluster_require.yaml under
// `cluster_require.kube_upgrade_path`.
func ParseKubeUpgradePath() (KubeUpgradePath, error) {
	data, err := BuiltinPlaybook.ReadFile(clusterRequirePath)
	if err != nil {
		return nil, fmt.Errorf("read embedded cluster_require: %w", err)
	}
	var doc struct {
		ClusterRequire struct {
			KubeUpgradePath map[string]string `json:"kube_upgrade_path"`
		} `json:"cluster_require"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse cluster_require: %w", err)
	}
	path := make(KubeUpgradePath, len(doc.ClusterRequire.KubeUpgradePath))
	for minorStr, patch := range doc.ClusterRequire.KubeUpgradePath {
		minor, err := parseMinor(minorStr)
		if err != nil {
			return nil, fmt.Errorf("invalid minor key %q in kube_upgrade_path: %w", minorStr, err)
		}
		path[minor] = patch
	}
	return path, nil
}

// ComputeUpgradeSteps returns the ordered list of target versions the cluster
// must pass through to upgrade from currentMinor to targetMinor.
//
// Intermediate minors (currentMinor+1 .. targetMinor-1) each use the highest
// recommended patch from the upgrade path; the final target minor uses the
// caller-supplied requestedTarget so a specific patch can be pinned.
//
// It returns an error for downgrades, or when an intermediate minor is missing
// from the path (which would make auto-stepping unsafe).
func ComputeUpgradeSteps(currentMinor, targetMinor int, requestedTarget string, path KubeUpgradePath) ([]string, error) {
	if targetMinor < currentMinor {
		return nil, fmt.Errorf("downgrade from v1.%d to v1.%d is not supported", currentMinor, targetMinor)
	}
	if targetMinor == currentMinor {
		// Patch-only upgrade: go straight to the requested version.
		return []string{requestedTarget}, nil
	}
	steps := make([]string, 0, targetMinor-currentMinor)
	for m := currentMinor + 1; m < targetMinor; m++ {
		patch, ok := path[m]
		if !ok {
			return nil, fmt.Errorf("kube_upgrade_path has no entry for minor v1.%d; cannot auto-step", m)
		}
		steps = append(steps, patch)
	}
	steps = append(steps, requestedTarget)
	return steps, nil
}

// parseMinor extracts the minor component from a version string such as
// "v1.23", "1.23" or "v1.23.17". It returns the integer minor (e.g. 23).
func parseMinor(s string) (int, error) {
	s = strings.TrimPrefix(s, "v")
	parts := strings.SplitN(s, ".", 3)
	if len(parts) < 2 {
		return 0, fmt.Errorf("not a major.minor version: %q", s)
	}
	return strconv.Atoi(parts[1])
}

// MinorOfVersion extracts the minor component from a Kubernetes version string.
// The discovery API may return suffixes (e.g. "34+"), which are ignored.
func MinorOfVersion(s string) (int, error) {
	s = strings.TrimPrefix(s, "v")
	parts := strings.SplitN(s, ".", 3)
	if len(parts) < 2 {
		return 0, fmt.Errorf("not a major.minor version: %q", s)
	}
	minorStr := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, parts[1])
	if minorStr == "" {
		return 0, fmt.Errorf("invalid minor in %q", s)
	}
	return strconv.Atoi(minorStr)
}
