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
	"os"

	"github.com/kubesphere/kubekey/v3/pkg/util/hash"
)

// FileSystem is a filesystem implementation
type FileSystem struct {
}

// NewFileSystem returns a new CAPKK local filesystem implementation
func NewFileSystem() Interface {
	return FileSystem{}
}

// Stat returns the FileInfo for the given path
func (f FileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

// MkdirAll the same as os.MkdirAll().
func (f FileSystem) MkdirAll(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

// MD5Sum returns the file MD5 sum for the given local path.
func (f FileSystem) MD5Sum(localPath string) string {
	md5, err := hash.FileMD5(localPath)
	if err != nil {
		return ""
	}
	return md5
}

// SHA256Sum returns the file SHA256 sum for the given local path.
func (f FileSystem) SHA256Sum(localPath string) string {
	sha256, err := hash.FileSHA256(localPath)
	if err != nil {
		return ""
	}
	return sha256
}

// MkLocalTmpDir creates a temporary directory and returns the path
func (f FileSystem) MkLocalTmpDir() (string, error) {
	tempDir, err := os.MkdirTemp(DefaultLocalTmpDir, ".Tmp-")
	if err != nil {
		return "", err
	}
	return tempDir, os.MkdirAll(tempDir, os.ModePerm)
}

// MkLocalTmpFile creates a temporary file and returns the path.
func (f FileSystem) MkLocalTmpFile(dir, pattern string) (string, error) {
	file, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", err
	}
	_ = file.Close()
	return file.Name(), nil
}

// RemoveAll the same as os.RemoveAll().
func (f FileSystem) RemoveAll(path ...string) error {
	for _, fi := range path {
		err := os.RemoveAll(fi)
		if err != nil {
			return fmt.Errorf("failed to remove file %s, %v", fi, err)
		}
	}
	return nil
}
