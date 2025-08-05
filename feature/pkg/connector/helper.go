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

// convertBytesToMap parses the given byte slice into a map[string]string using the provided split string.
// Only lines containing the split string are processed. Each such line is split into key and value at the first occurrence of the split string.
// Leading and trailing spaces are trimmed from both key and value.
// Example input (split = "="):
//
//	FOO=bar
//	BAZ = qux
//
// Result: map[string]string{"FOO": "bar", "BAZ": "qux"}
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

// convertBytesToSlice parses the given byte slice into a slice of map[string]string using the provided split string.
// Only lines containing the split string are processed. Each such line is split into key and value at the first occurrence of the split string.
// Leading and trailing spaces are trimmed from both key and value.
// Groups of key-value pairs are separated by empty lines. Each group is stored as a separate map in the resulting slice.
// Example input (split = ":"):
//
//	foo: bar
//	baz: qux
//
//	hello: world
//
//	Result: []map[string]string{
//	  {"foo": "bar", "baz": "qux"},
//	  {"hello": "world"},
//	}
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
