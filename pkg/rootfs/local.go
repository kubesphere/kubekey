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
	"path/filepath"

	"github.com/kubesphere/kubekey/pkg/util/filesystem"
)

// Local is a rootfs for local.
type Local struct {
	clusterName string
	basePath    string
	fs          filesystem.Interface
}

// NewLocalRootFs returns a new Local implementation of rootfs interface.
func NewLocalRootFs(clusterName, basePath string) Interface {
	if basePath == "" {
		basePath = DefaultLocalRootFsDir
	}
	return &Local{
		clusterName: clusterName,
		basePath:    basePath,
		fs:          filesystem.NewFileSystem(),
	}
}

// ClusterRootFsDir returns the rootfs directory for the cluster.
func (l *Local) ClusterRootFsDir() string {
	return filepath.Join(l.basePath, l.clusterName)
}

// HostRootFsDir returns the rootfs directory for the host.
func (l *Local) HostRootFsDir(host string) string {
	return filepath.Join(l.basePath, l.clusterName, host)
}

// Fs returns the filesystem interface.
func (l *Local) Fs() filesystem.Interface {
	return l.fs
}
