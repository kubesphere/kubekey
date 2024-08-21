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

// SourceType how to store variable
type SourceType int

const (
	// MemorySource store variable in memory
	MemorySource SourceType = iota
	// FileSource store variable in file
	FileSource SourceType = iota
)

// Source is the source from which config is loaded.
type Source interface {
	Read() (map[string][]byte, error)
	Write(data []byte, filename string) error
}
