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

var builtinProjectFunc func(kkcorev1.Pipeline) (Project, error)

// Project represent location of actual project.
// get project file should base on it
type Project interface {
	MarshalPlaybook() (*kkprojectv1.Playbook, error)
	Stat(path string, option GetFileOption) (os.FileInfo, error)
	WalkDir(path string, option GetFileOption, f fs.WalkDirFunc) error
	ReadFile(path string, option GetFileOption) ([]byte, error)
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
// If pipeline has BuiltinsProjectAnnotation. builtinProjectFunc
// Default newLocalProject
func New(ctx context.Context, pipeline kkcorev1.Pipeline, update bool) (Project, error) {
	if strings.HasPrefix(pipeline.Spec.Project.Addr, "https://") ||
		strings.HasPrefix(pipeline.Spec.Project.Addr, "http://") ||
		strings.HasPrefix(pipeline.Spec.Project.Addr, "git@") {
		return newGitProject(ctx, pipeline, update)
	}

	if _, ok := pipeline.Annotations[kkcorev1.BuiltinsProjectAnnotation]; ok {
		return builtinProjectFunc(pipeline)
	}

	return newLocalProject(pipeline)
}
