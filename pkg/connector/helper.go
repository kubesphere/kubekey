/*
Copyright 2024 The KubeSphere Authors.

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

package connector

import (
	"bufio"
	"bytes"
	"strings"
)

// convertBytesToMap with split string, only convert line which contain split
func convertBytesToMap(bs []byte, split string) map[string]string {
	config := make(map[string]string)
	scanner := bufio.NewScanner(bytes.NewBuffer(bs))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, split, 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			config[key] = value
		}
	}

	return config
}

// convertBytesToSlice with split string. only convert line which contain split.
// group by empty line
func convertBytesToSlice(bs []byte, split string) []map[string]string {
	var config []map[string]string
	currentMap := make(map[string]string)

	scanner := bufio.NewScanner(bytes.NewBuffer(bs))
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if line != "" {
			parts := strings.SplitN(line, split, 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				currentMap[key] = value
			}
		} else if len(currentMap) > 0 {
			// If encountering an empty line, add the current map to config and create a new map
			config = append(config, currentMap)
			currentMap = make(map[string]string)
		}
	}

	// Add the last map if not already added
	if len(currentMap) > 0 {
		config = append(config, currentMap)
	}

	return config
}
