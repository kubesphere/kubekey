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
	"os"
)

// Interface is an interface for filesystem operations
type Interface interface {
	// Stat returns the FileInfo structure describing the named file.
	Stat(name string) (os.FileInfo, error)
	// MkdirAll the same as os.MkdirAll().
	MkdirAll(path string) error
	// MD5Sum returns the file MD5 sum for the given local path.
	MD5Sum(localPath string) string
	// MkLocalTmpDir creates a temporary directory and returns the path.
	MkLocalTmpDir() (string, error)
	// RemoveAll the same as os.RemoveAll().
	RemoveAll(path ...string) error
}
