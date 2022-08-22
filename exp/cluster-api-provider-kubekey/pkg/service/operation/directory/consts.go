/*
 Copyright 2022 The KubeSphere Authors.

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

package directory

const (
	// TmpDir represents the tmp directory of the remote instance
	TmpDir = "/tmp/kubekey"
	// BinDir represents the bin directory of the remote instance
	BinDir = "/usr/local/bin"
	// KubeConfigDir represents the normal kubernetes data directory of the remote instance
	KubeConfigDir = "/etc/kubernetes"
	// KubeCertDir represents the normal kubernetes cert directory of the remote instance
	KubeCertDir = "/etc/kubernetes/pki"
	// KubeManifestDir represents the normal kubernetes manifest directory of the remote instance
	KubeManifestDir = "/etc/kubernetes/manifests"
	// KubeScriptDir represents the kubernetes manage tools scripts directory of the remote instance
	KubeScriptDir = "/usr/local/bin/kube-scripts"
	// KubeletFlexvolumesPluginsDir represents the kubernetes kubelet plugin volume directory of the remote instance
	KubeletFlexvolumesPluginsDir = "/usr/libexec/kubernetes/kubelet-plugins/volume/exec"
)
