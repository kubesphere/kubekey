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

package source

import (
	"io/fs"
	"os"

	"k8s.io/klog/v2"
)

// Source is the source from which config is loaded.
type Source interface {
	Read() (map[string][]byte, error)
	Write(data []byte, filename string) error
	//Watch() (Watcher, error)
}

// Watcher watches a source for changes.
type Watcher interface {
	Next() ([]byte, error)
	Stop() error
}

// New returns a new source.
func New(path string) (Source, error) {
	if _, err := os.Stat(path); err != nil {
		if err := os.MkdirAll(path, fs.ModePerm); err != nil {
			klog.ErrorS(err, "create source path error", "path", path)
			return nil, err
		}
	}
	return &fileSource{path: path}, nil
}
