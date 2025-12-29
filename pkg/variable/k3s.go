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

package variable

import (
	"strings"

	"k8s.io/klog/v2"
)

// IsK3sVersion checks if the given kubernetes version string contains "k3s"
func IsK3sVersion(version string) bool {
	return strings.Contains(strings.ToLower(version), "k3s")
}

// GetK3sVersion extracts the base k3s version without the "+k3s" suffix
func GetK3sVersion(version string) string {
	if !IsK3sVersion(version) {
		return version
	}

	// Handle versions like "v1.21.4-k3s" or "v1.21.4+k3s1"
	parts := strings.Split(version, "+")
	if len(parts) > 0 {
		baseVersion := parts[0]
		// Remove any "-k3s" suffix
		baseVersion = strings.TrimSuffix(baseVersion, "-k3s")
		return baseVersion
	}

	return version
}

// DetectAndSetK3sVariables detects k3s usage and sets appropriate variables
func DetectAndSetK3sVariables(vars map[string]any) map[string]any {
	kubernetes, ok := vars["kubernetes"].(map[string]any)
	if !ok {
		return vars
	}

	kubeVersion, ok := kubernetes["kube_version"].(string)
	if !ok {
		return vars
	}

	if IsK3sVersion(kubeVersion) {
		klog.InfoS("Detected k3s version", "version", kubeVersion)

		// Set k3s-specific variables
		kubernetes["is_k3s"] = true
		kubernetes["k3s_version"] = GetK3sVersion(kubeVersion)

		// For k3s, we don't use kubeadm
		kubernetes["use_kubeadm"] = false

		// k3s has built-in CNI management, but we can still install external CNIs
		// Set a flag to indicate CNI should be configured after k3s is ready
		if network, ok := vars["cni"].(map[string]any); ok {
			network["k3s_compatible"] = true
		}

		vars["kubernetes"] = kubernetes
		klog.InfoS("Configured k3s-specific variables", "k3s_version", kubernetes["k3s_version"])
	} else {
		// For non-k3s versions, ensure k3s flags are not set
		kubernetes["is_k3s"] = false
		kubernetes["use_kubeadm"] = true
		vars["kubernetes"] = kubernetes
	}

	return vars
}
