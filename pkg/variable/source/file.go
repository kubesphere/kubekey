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
	"os"
	"path/filepath"
	"strings"

	"k8s.io/klog/v2"
)

type fileSource struct {
	path string
}

func (f *fileSource) Read() (map[string][]byte, error) {
	de, err := os.ReadDir(f.path)
	if err != nil {
		klog.Errorf("read dir %s error %v", f.path, err)
		return nil, err
	}
	var result map[string][]byte
	for _, entry := range de {
		if entry.IsDir() {
			continue
		}
		if result == nil {
			result = make(map[string][]byte)
		}
		// only read json data
		if strings.HasSuffix(entry.Name(), ".json") {
			data, err := os.ReadFile(filepath.Join(f.path, entry.Name()))
			if err != nil {
				return nil, err
			}
			result[entry.Name()] = data
		}
	}

	return result, nil
}

func (f *fileSource) Write(data []byte, filename string) error {
	file, err := os.Create(filepath.Join(f.path, filename))
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.Write(data); err != nil {
		return err
	}
	return nil
}
