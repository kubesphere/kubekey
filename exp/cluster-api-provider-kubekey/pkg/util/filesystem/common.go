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

package filesystem

import (
	"path/filepath"

	"github.com/mitchellh/go-homedir"
)

const (
	DefaultLocalTmpDir = "/var/lib/kubekey"
)

const (
	FileMode0755 = 0755
	FileMode0644 = 0644
)

func GetClusterWorkDir(clusterName string) string {
	return filepath.Join(HomeDir(), ".kubekey", clusterName)
}

func GetCommonWorkDir() string {
	return filepath.Join(HomeDir(), ".kubekey")
}

func HomeDir() string {
	home, err := homedir.Dir()
	if err != nil {
		return "/root"
	}
	return home
}
