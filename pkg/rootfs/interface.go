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

package rootfs

import (
	"github.com/kubesphere/kubekey/v3/pkg/util/filesystem"
)

// Interface is the interface for rootfs.
type Interface interface {
	// ClusterRootFsDir returns the rootfs directory of the cluster.
	ClusterRootFsDir() string
	// HostRootFsDir returns the rootfs directory of the host.
	HostRootFsDir(host string) string
	// Fs returns the filesystem interface.
	Fs() filesystem.Interface
}
