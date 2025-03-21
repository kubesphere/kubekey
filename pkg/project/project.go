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

package project

import (
	"context"
	"io/fs"
	"os"
	"strings"

	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	kkprojectv1 "github.com/kubesphere/kubekey/api/project/v1"
)

var builtinProjectFunc func(kkcorev1.Playbook) (Project, error)

// Project represent location of actual project.
// get project file should base on it
type Project interface {
	// MarshalPlaybook project file to playbook.
	MarshalPlaybook() (*kkprojectv1.Playbook, error)
	// Stat file or dir in project
	Stat(path string, option GetFileOption) (os.FileInfo, error)
	// WalkDir dir in project
	WalkDir(path string, option GetFileOption, f fs.WalkDirFunc) error
	// ReadFile file or dir in project
	ReadFile(path string, option GetFileOption) ([]byte, error)
	// Rel path file or dir in project
	Rel(root string, path string, option GetFileOption) (string, error)
}

// GetFileOption for file.
type GetFileOption struct {
	Role       string
	IsTemplate bool
	IsFile     bool
}

// New project.
// If project address is git format. newGitProject
// If playbook has BuiltinsProjectAnnotation. builtinProjectFunc
// Default newLocalProject
func New(ctx context.Context, playbook kkcorev1.Playbook, update bool) (Project, error) {
	if strings.HasPrefix(playbook.Spec.Project.Addr, "https://") ||
		strings.HasPrefix(playbook.Spec.Project.Addr, "http://") ||
		strings.HasPrefix(playbook.Spec.Project.Addr, "git@") {
		return newGitProject(ctx, playbook, update)
	}

	if _, ok := playbook.Annotations[kkcorev1.BuiltinsProjectAnnotation]; ok {
		return builtinProjectFunc(playbook)
	}

	return newLocalProject(playbook)
}
