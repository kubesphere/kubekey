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
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"k8s.io/klog/v2"

	kkcorev1 "github.com/kubesphere/kubekey/v4/pkg/apis/core/v1"
	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
)

func newGitProject(pipeline kubekeyv1.Pipeline, update bool) (Project, error) {
	if pipeline.Spec.Playbook == "" || pipeline.Spec.Project.Addr == "" {
		return nil, fmt.Errorf("playbook and project.addr should not be empty")
	}
	if filepath.IsAbs(pipeline.Spec.Playbook) {
		return nil, fmt.Errorf("playbook should be relative path base on project.addr")
	}

	// git clone to project dir
	if pipeline.Spec.Project.Name == "" {
		pipeline.Spec.Project.Name = strings.TrimSuffix(pipeline.Spec.Project.Addr[strings.LastIndex(pipeline.Spec.Project.Addr, "/")+1:], ".git")
	}
	p := &gitProject{
		Pipeline:   pipeline,
		projectDir: filepath.Join(_const.GetWorkDir(), _const.ProjectDir, pipeline.Spec.Project.Name),
		playbook:   pipeline.Spec.Playbook,
	}
	if _, err := os.Stat(p.projectDir); os.IsNotExist(err) {
		// git clone
		if err := p.gitClone(context.Background()); err != nil {
			return nil, fmt.Errorf("clone git project error: %w", err)
		}
	} else if update {
		// git pull
		if err := p.gitPull(context.Background()); err != nil {
			return nil, fmt.Errorf("pull git project error: %w", err)
		}
	}
	return p, nil
}

// gitProject from git
type gitProject struct {
	kubekeyv1.Pipeline

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

func (p gitProject) Stat(path string, option GetFileOption) (os.FileInfo, error) {
	return os.Stat(p.getFilePath(path, option))
}

func (p gitProject) WalkDir(path string, option GetFileOption, f fs.WalkDirFunc) error {
	return filepath.WalkDir(p.getFilePath(path, option), f)
}

func (p gitProject) ReadFile(path string, option GetFileOption) ([]byte, error) {
	return os.ReadFile(p.getFilePath(path, option))
}

func (p gitProject) MarshalPlaybook() (*kkcorev1.Playbook, error) {
	return marshalPlaybook(os.DirFS(p.projectDir), p.Pipeline.Spec.Playbook)
}

func (p gitProject) gitClone(ctx context.Context) error {
	if _, err := git.PlainCloneContext(ctx, p.projectDir, false, &git.CloneOptions{
		URL:             p.Pipeline.Spec.Project.Addr,
		Progress:        nil,
		ReferenceName:   plumbing.NewBranchReferenceName(p.Pipeline.Spec.Project.Branch),
		SingleBranch:    true,
		Auth:            &http.TokenAuth{p.Pipeline.Spec.Project.Token},
		InsecureSkipTLS: false,
	}); err != nil {
		klog.Errorf("clone project %s failed: %v", p.Pipeline.Spec.Project.Addr, err)
		return err
	}
	return nil
}

func (p gitProject) gitPull(ctx context.Context) error {
	open, err := git.PlainOpen(p.projectDir)
	if err != nil {
		klog.V(4).ErrorS(err, "git open error", "local_dir", p.projectDir)
		return err
	}
	wt, err := open.Worktree()
	if err != nil {
		klog.V(4).ErrorS(err, "git open worktree error", "local_dir", p.projectDir)
		return err
	}
	if err := wt.PullContext(ctx, &git.PullOptions{
		RemoteURL:       p.Pipeline.Spec.Project.Addr,
		ReferenceName:   plumbing.NewBranchReferenceName(p.Pipeline.Spec.Project.Branch),
		SingleBranch:    true,
		Auth:            &http.TokenAuth{p.Pipeline.Spec.Project.Token},
		InsecureSkipTLS: false,
	}); err != nil && err != git.NoErrAlreadyUpToDate {
		klog.V(4).ErrorS(err, "git pull error", "local_dir", p.projectDir)
		return err
	}

	return nil
}
