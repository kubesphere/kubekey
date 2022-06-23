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
	"fmt"
	"io/ioutil"
	"os"

	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/util/hash"
)

type FileSystem struct {
}

func NewFileSystem() Interface {
	return FileSystem{}
}

func (f FileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (f FileSystem) MkdirAll(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

func (f FileSystem) MD5Sum(localPath string) string {
	md5, err := hash.FileMD5(localPath)
	if err != nil {
		return ""
	}
	return md5
}

func (f FileSystem) MkLocalTmpDir() (string, error) {
	tempDir, err := ioutil.TempDir(DefaultLocalTmpDir, ".Tmp-")
	if err != nil {
		return "", err
	}
	return tempDir, os.MkdirAll(tempDir, os.ModePerm)
}

func (f FileSystem) RemoveAll(path ...string) error {
	for _, fi := range path {
		err := os.RemoveAll(fi)
		if err != nil {
			return fmt.Errorf("failed to remove file %s, %v", fi, err)
		}
	}
	return nil
}
