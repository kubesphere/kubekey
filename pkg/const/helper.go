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

package _const

import (
	"fmt"
	"os"
	"strings"

	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
)

// GetWorkdirFromConfig retrieves the working directory from the provided configuration.
// If the 'workdir' value is set in the configuration and is a string, it returns that value.
// If the 'workdir' value is not set or is not a string, it logs an informational message
// and attempts to get the current working directory of the process.
// If it fails to get the current working directory, it logs another informational message
// and returns a default directory path "/opt/kubekey".
func GetWorkdirFromConfig(config kkcorev1.Config) string {
	if wd, _, err := unstructured.NestedString(config.Value(), Workdir); err == nil {
		return wd
	}
	klog.Info("work_dir is not set use current dir.")
	wd, err := os.Getwd()
	if err != nil {
		klog.Info("failed to get current dir. use default: /root/kubekey")

		return "/opt/kubekey"
	}

	return wd
}

// Host2ProviderID converts a cluster name and host into a provider ID string.
// It returns a pointer to a string in the format "kk://<cluster_name>/<host>".
func Host2ProviderID(clusterName, host string) *string {
	return ptr.To(fmt.Sprintf("kk://%s/%s", clusterName, host))
}

// ProviderID2Host extracts the host name from a provider ID string.
// It takes a cluster name and provider ID pointer, and returns the host portion
// by trimming off the "kk://<cluster_name>/" prefix. If providerID is nil,
// returns an empty string.
func ProviderID2Host(clusterName string, providerID *string) string {
	return strings.TrimPrefix(ptr.Deref(providerID, ""), fmt.Sprintf("kk://%s/", clusterName))
}
