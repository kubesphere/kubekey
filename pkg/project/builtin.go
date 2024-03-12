//go:build builtin
// +build builtin

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

package project

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/kubesphere/kubekey/v4/builtin"
	kkcorev1 "github.com/kubesphere/kubekey/v4/pkg/apis/core/v1"
	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
)

func init() {
	builtinProjectFunc = func(pipeline kubekeyv1.Pipeline) (Project, error) {
		if pipeline.Spec.Playbook == "" {
			return nil, fmt.Errorf("playbook should not be empty")
		}
		if filepath.IsAbs(pipeline.Spec.Playbook) {
			return nil, fmt.Errorf("playbook should be relative path base on project.addr")
		}
		return &builtinProject{Pipeline: pipeline, FS: builtin.BuiltinPipeline, playbook: pipeline.Spec.Playbook}, nil
	}
}

type builtinProject struct {
	kubekeyv1.Pipeline

	fs.FS
	// playbook relpath base on projectDir
	playbook string
}

func (p builtinProject) getFilePath(path string, o GetFileOption) string {
	var find []string
	switch {
	case o.IsFile:
		if o.Role != "" {
			// find from project/roles/roleName
			find = append(find, filepath.Join(_const.ProjectRolesDir, o.Role, _const.ProjectRolesFilesDir, path))
			// find from pbPath dir like: current_playbook/roles/roleName
			find = append(find, filepath.Join(p.playbook, _const.ProjectRolesDir, o.Role, _const.ProjectRolesFilesDir, path))
		}
		find = append(find, filepath.Join(_const.ProjectRolesFilesDir, path))
	case o.IsTemplate:
		// find from project/roles/roleName
		if o.Role != "" {
			find = append(find, filepath.Join(_const.ProjectRolesDir, o.Role, _const.ProjectRolesTemplateDir, path))
			// find from pbPath dir like: current_playbook/roles/roleName
			find = append(find, filepath.Join(p.playbook, _const.ProjectRolesDir, o.Role, _const.ProjectRolesTemplateDir, path))
		}
		find = append(find, filepath.Join(_const.ProjectRolesTemplateDir, path))
	default:
		find = append(find, filepath.Join(path))
	}
	for _, s := range find {
		if _, err := fs.Stat(p.FS, s); err == nil {
			return s
		}
	}
	return ""
}

func (p builtinProject) Stat(path string, option GetFileOption) (os.FileInfo, error) {
	return fs.Stat(p.FS, p.getFilePath(path, option))
}

func (p builtinProject) WalkDir(path string, option GetFileOption, f fs.WalkDirFunc) error {
	return fs.WalkDir(p.FS, p.getFilePath(path, option), f)
}

func (p builtinProject) ReadFile(path string, option GetFileOption) ([]byte, error) {
	return fs.ReadFile(p.FS, p.getFilePath(path, option))
}

func (p builtinProject) MarshalPlaybook() (*kkcorev1.Playbook, error) {
	return marshalPlaybook(p.FS, p.playbook)
}
