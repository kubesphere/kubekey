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
	"path/filepath"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	kkprojectv1 "github.com/kubesphere/kubekey/api/project/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
)

func newGitProject(ctx context.Context, playbook kkcorev1.Playbook, update bool) (Project, error) {
	if playbook.Spec.Playbook == "" || playbook.Spec.Project.Addr == "" {
		return nil, errors.New("playbook and project.addr should not be empty")
	}

	if filepath.IsAbs(playbook.Spec.Playbook) {
		return nil, errors.New("playbook should be relative path base on project.addr")
	}

	// get project_dir from playbook
	projectDir, _, err := unstructured.NestedString(playbook.Spec.Config.Value(), _const.ProjectsDir)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get %q in config", _const.ProjectsDir)
	}

	// git clone to project dir
	if playbook.Spec.Project.Name == "" {
		playbook.Spec.Project.Name = strings.TrimSuffix(playbook.Spec.Project.Addr[strings.LastIndex(playbook.Spec.Project.Addr, "/")+1:], ".git")
	}

	p := &gitProject{
		Playbook:   playbook,
		projectDir: filepath.Join(projectDir, playbook.Spec.Project.Name),
		playbook:   playbook.Spec.Playbook,
	}

	if _, err := os.Stat(p.projectDir); os.IsNotExist(err) {
		// git clone
		if err := p.gitClone(ctx); err != nil {
			return nil, err
		}
	} else if update {
		// git pull
		if err := p.gitPull(ctx); err != nil {
			return nil, err
		}
	}

	return p, nil
}

// gitProject from git
type gitProject struct {
	kkcorev1.Playbook

	//location
	projectDir string
	// playbook relpath base on projectDir
	playbook string
}

func (p gitProject) getFilePath(path string, o GetFileOption) string {
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

func (p gitProject) gitClone(ctx context.Context) error {
	if _, err := git.PlainCloneContext(ctx, p.projectDir, false, &git.CloneOptions{
		URL:             p.Playbook.Spec.Project.Addr,
		Progress:        nil,
		ReferenceName:   plumbing.NewBranchReferenceName(p.Playbook.Spec.Project.Branch),
		SingleBranch:    true,
		Auth:            &http.TokenAuth{Token: p.Playbook.Spec.Project.Token},
		InsecureSkipTLS: false,
	}); err != nil {
		return errors.Wrapf(err, "failed to clone project %q", p.Playbook.Spec.Project.Addr)
	}

	return nil
}

func (p gitProject) gitPull(ctx context.Context) error {
	open, err := git.PlainOpen(p.projectDir)
	if err != nil {
		return errors.Wrapf(err, "failed to open git project %a", p.projectDir)
	}

	wt, err := open.Worktree()
	if err != nil {
		return errors.Wrapf(err, "failed to open git project %q worktree", p.projectDir)
	}

	if err := wt.PullContext(ctx, &git.PullOptions{
		RemoteURL:       p.Playbook.Spec.Project.Addr,
		ReferenceName:   plumbing.NewBranchReferenceName(p.Playbook.Spec.Project.Branch),
		SingleBranch:    true,
		Auth:            &http.TokenAuth{Token: p.Playbook.Spec.Project.Token},
		InsecureSkipTLS: false,
	}); err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return errors.Wrapf(err, "failed to pull git project %q", p.playbook)
	}

	return nil
}

// MarshalPlaybook project file to playbook.
func (p gitProject) MarshalPlaybook() (*kkprojectv1.Playbook, error) {
	return marshalPlaybook(os.DirFS(p.projectDir), p.Playbook.Spec.Playbook)
}

// Stat role/file/template file or dir in project
func (p gitProject) Stat(path string, option GetFileOption) (os.FileInfo, error) {
	return os.Stat(p.getFilePath(path, option))
}

// WalkDir role/file/template dir in project
func (p gitProject) WalkDir(path string, option GetFileOption, f fs.WalkDirFunc) error {
	return filepath.WalkDir(p.getFilePath(path, option), f)
}

// ReadFile role/file/template file or dir in project
func (p gitProject) ReadFile(path string, option GetFileOption) ([]byte, error) {
	return os.ReadFile(p.getFilePath(path, option))
}

// Rel path for role/file/template file or dir in project
func (p gitProject) Rel(root string, path string, option GetFileOption) (string, error) {
	return filepath.Rel(p.getFilePath(root, option), path)
}
