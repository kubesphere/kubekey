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
	"io/fs"
	"os"
	"path/filepath"

	"github.com/cockroachdb/errors"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	kkprojectv1 "github.com/kubesphere/kubekey/api/project/v1"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
)

func newLocalProject(playbook kkcorev1.Playbook) (Project, error) {
	if !filepath.IsAbs(playbook.Spec.Playbook) {
		if playbook.Spec.Project.Addr == "" {
			wd, err := os.Getwd()
			if err != nil {
				return nil, err
			}
			playbook.Spec.Project.Addr = wd
		}
		playbook.Spec.Playbook = filepath.Join(playbook.Spec.Project.Addr, playbook.Spec.Playbook)
	}

	if _, err := os.Stat(playbook.Spec.Playbook); err != nil {
		return nil, errors.Wrapf(err, "cannot find playbook %q", playbook.Spec.Playbook)
	}

	if filepath.Base(filepath.Dir(playbook.Spec.Playbook)) != _const.ProjectPlaybooksDir {
		// the format of playbook is not correct
		return nil, errors.New("playbook should be projectDir/playbooks/playbookfile")
	}

	projectDir := filepath.Dir(filepath.Dir(playbook.Spec.Playbook))
	pb, err := filepath.Rel(projectDir, playbook.Spec.Playbook)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get rel path for playbook %q", playbook.Spec.Playbook)
	}

	return &localProject{Playbook: playbook, projectDir: projectDir, playbook: pb}, nil
}

type localProject struct {
	kkcorev1.Playbook

	projectDir string
	// playbook relpath base on projectDir
	playbook string
}

func (p localProject) getFilePath(path string, o GetFileOption) string {
	if filepath.IsAbs(path) {
		return path
	}
	var find []string
	switch {
	case o.IsFile:
		if o.Role != "" {
			// find from project/roles/roleName
			find = append(find, filepath.Join(p.projectDir, _const.ProjectRolesDir, o.Role, _const.ProjectRolesFilesDir, path))
			// find from pbPath dir like: current_playbook/roles/roleName
			find = append(find, filepath.Join(p.projectDir, p.playbook, _const.ProjectRolesDir, o.Role, _const.ProjectRolesFilesDir, path))
		}
		find = append(find, filepath.Join(p.projectDir, _const.ProjectRolesFilesDir, path))
	case o.IsTemplate:
		// find from project/roles/roleName
		if o.Role != "" {
			find = append(find, filepath.Join(p.projectDir, _const.ProjectRolesDir, o.Role, _const.ProjectRolesTemplateDir, path))
			// find from pbPath dir like: current_playbook/roles/roleName
			find = append(find, filepath.Join(p.projectDir, p.playbook, _const.ProjectRolesDir, o.Role, _const.ProjectRolesTemplateDir, path))
		}
		find = append(find, filepath.Join(p.projectDir, _const.ProjectRolesTemplateDir, path))
	default:
		find = append(find, filepath.Join(p.projectDir, path))
	}
	for _, s := range find {
		if _, err := os.Stat(s); err == nil {
			return s
		}
	}

	return ""
}

// MarshalPlaybook project file to playbook.
func (p localProject) MarshalPlaybook() (*kkprojectv1.Playbook, error) {
	return marshalPlaybook(os.DirFS(p.projectDir), p.playbook)
}

// Stat role/file/template file or dir in project
func (p localProject) Stat(path string, option GetFileOption) (os.FileInfo, error) {
	return os.Stat(p.getFilePath(path, option))
}

// WalkDir role/file/template dir in project
func (p localProject) WalkDir(path string, option GetFileOption, f fs.WalkDirFunc) error {
	return filepath.WalkDir(p.getFilePath(path, option), f)
}

// ReadFile role/file/template file or dir in project
func (p localProject) ReadFile(path string, option GetFileOption) ([]byte, error) {
	return os.ReadFile(p.getFilePath(path, option))
}

// Rel path for role/file/template file or dir in project
func (p localProject) Rel(root string, path string, option GetFileOption) (string, error) {
	return filepath.Rel(p.getFilePath(root, option), path)
}
