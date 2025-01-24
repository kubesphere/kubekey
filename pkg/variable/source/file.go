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
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
	"k8s.io/klog/v2"
)

const (
	prefixYAML = "# Generate by variable\n"
)

var _ Source = &fileSource{}

// NewFileSource returns a new fileSource.
func NewFileSource(path string) (Source, error) {
	if _, err := os.Stat(path); err != nil {
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			klog.V(4).ErrorS(err, "create source path error", "path", path)

			return nil, err
		}
	}

	return &fileSource{path: path}, nil
}

type fileSource struct {
	path string
}

func (f *fileSource) Read() (map[string]map[string]any, error) {
	de, err := os.ReadDir(f.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read dir %s, error: %w", f.path, err)
	}

	result := make(map[string]map[string]any)
	for _, entry := range de {
		if entry.IsDir() {
			continue
		}
		// only read json data
		if strings.HasSuffix(entry.Name(), ".yaml") {
			fdata, err := os.ReadFile(filepath.Join(f.path, entry.Name()))
			if err != nil {
				return nil, fmt.Errorf("failed to read file %q error: %w", entry.Name(), err)
			}
			if bytes.HasPrefix(fdata, []byte(prefixYAML)) {
				var data map[string]any
				if err := yaml.Unmarshal(fdata, data); err != nil {
					return nil, fmt.Errorf("failed to unmarshal file %q error: %w", entry.Name(), err)
				}
				result[strings.TrimSuffix(entry.Name(), ".yaml")] = data
			}
		}
	}

	return result, nil
}

func (f *fileSource) Write(data map[string]any, host string) error {
	filename := host + ".yaml"
	file, err := os.Create(filepath.Join(f.path, filename))
	if err != nil {
		return fmt.Errorf("failed to create file %q error: %w", filename, err)
	}
	defer file.Close()
	// convert to yaml file
	fdata, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal file %q error: %w", filename, err)
	}
	if _, err := file.Write(append([]byte(prefixYAML), fdata...)); err != nil {
		return fmt.Errorf("failed to write file %q error: %w", filename, err)
	}

	return nil
}
