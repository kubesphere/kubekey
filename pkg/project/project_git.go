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

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"k8s.io/klog/v2"

	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
)

// gitProject from git
type gitProject struct {
	kubekeyv1.Pipeline

	localDir string
}

func (r gitProject) FS(ctx context.Context, update bool) (fs.FS, error) {
	if !update {
		return os.DirFS(r.localDir), nil
	}
	if err := r.init(ctx); err != nil {
		klog.ErrorS(err, "Init git project error", "project_addr", r.Pipeline.Spec.Project.Addr)
		return nil, err
	}
	return os.DirFS(r.localDir), nil
}

func (r gitProject) init(ctx context.Context) error {
	if _, err := os.Stat(r.localDir); err != nil {
		// git clone
		return r.gitClone(ctx)
	} else {
		// git pull
		return r.gitPull(ctx)
	}
}

func (r gitProject) gitClone(ctx context.Context) error {
	if _, err := git.PlainCloneContext(ctx, r.localDir, false, &git.CloneOptions{
		URL:             r.Pipeline.Spec.Project.Addr,
		Progress:        nil,
		ReferenceName:   plumbing.NewBranchReferenceName(r.Pipeline.Spec.Project.Branch),
		SingleBranch:    true,
		Auth:            &http.TokenAuth{r.Pipeline.Spec.Project.Token},
		InsecureSkipTLS: false,
	}); err != nil {
		klog.Errorf("clone project %s failed: %v", r.Pipeline.Spec.Project.Addr, err)
		return err
	}
	return nil
}

func (r gitProject) gitPull(ctx context.Context) error {
	open, err := git.PlainOpen(r.localDir)
	if err != nil {
		klog.ErrorS(err, "git open error", "local_dir", r.localDir)
		return err
	}
	wt, err := open.Worktree()
	if err != nil {
		klog.ErrorS(err, "git open worktree error", "local_dir", r.localDir)
		return err
	}
	if err := wt.PullContext(ctx, &git.PullOptions{
		RemoteURL:       r.Pipeline.Spec.Project.Addr,
		ReferenceName:   plumbing.NewBranchReferenceName(r.Pipeline.Spec.Project.Branch),
		SingleBranch:    true,
		Auth:            &http.TokenAuth{r.Pipeline.Spec.Project.Token},
		InsecureSkipTLS: false,
	}); err != nil && err != git.NoErrAlreadyUpToDate {
		klog.ErrorS(err, "git pull error", "local_dir", r.localDir)
		return err
	}

	return nil
}
