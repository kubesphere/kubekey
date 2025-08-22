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
	"github.com/kubesphere/kubekey/v4/pkg/utils"
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
	projectDir = filepath.Join(projectDir, playbook.Spec.Project.Name)
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		// git clone
		if err := gitClone(ctx, projectDir, playbook.Spec.Project); err != nil {
			return nil, err
		}
	} else if update {
		// git pull
		if err := gitPull(ctx, projectDir, playbook.Spec.Project); err != nil {
			return nil, err
		}
	}

	return &project{
		FS:            os.DirFS(filepath.Join(projectDir, playbook.Spec.Project.Name)),
		basePlaybook:  playbook.Spec.Playbook,
		Playbook:      &kkprojectv1.Playbook{},
		config:        playbook.Spec.Config.Value(),
		playbookGraph: utils.NewGraph(),
	}, nil
}

func gitClone(ctx context.Context, localDir string, project kkcorev1.PlaybookProject) error {
	if _, err := git.PlainCloneContext(ctx, localDir, false, &git.CloneOptions{
		URL:             project.Addr,
		Progress:        nil,
		ReferenceName:   plumbing.NewBranchReferenceName(project.Branch),
		SingleBranch:    true,
		Auth:            &http.TokenAuth{Token: project.Token},
		InsecureSkipTLS: false,
	}); err != nil {
		return errors.Wrapf(err, "failed to clone project %q", project.Addr)
	}

	return nil
}

func gitPull(ctx context.Context, localDir string, project kkcorev1.PlaybookProject) error {
	open, err := git.PlainOpen(localDir)
	if err != nil {
		return errors.Wrapf(err, "failed to open git project %a", localDir)
	}

	wt, err := open.Worktree()
	if err != nil {
		return errors.Wrapf(err, "failed to open git project %q worktree", localDir)
	}

	if err := wt.PullContext(ctx, &git.PullOptions{
		RemoteURL:       project.Addr,
		ReferenceName:   plumbing.NewBranchReferenceName(project.Branch),
		SingleBranch:    true,
		Auth:            &http.TokenAuth{Token: project.Token},
		InsecureSkipTLS: false,
	}); err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return errors.Wrapf(err, "failed to pull git project %q", project.Addr)
	}

	return nil
}
