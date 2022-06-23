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

	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/util/filesystem"
)

type Local struct {
	clusterName string
	basePath    string
	fs          filesystem.Interface
}

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

func (l *Local) ClusterRootFsDir() string {
	return filepath.Join(l.basePath, l.clusterName)
}

func (l *Local) HostRootFsDir(host string) string {
	return filepath.Join(l.basePath, l.clusterName, host)
}

func (l *Local) Fs() filesystem.Interface {
	return l.fs
}
