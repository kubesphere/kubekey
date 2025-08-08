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
	"os"
	"path/filepath"

	"github.com/cockroachdb/errors"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	kkprojectv1 "github.com/kubesphere/kubekey/api/project/v1"
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
	projectPath := GetProjectPath(playbook.Spec.Playbook)
	relPath, err := filepath.Rel(projectPath, playbook.Spec.Playbook)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get relative path for playbook %q", playbook.Spec.Playbook)
	}

	return &project{
		FS:           os.DirFS(projectPath),
		basePlaybook: relPath,
		Playbook:     &kkprojectv1.Playbook{},
		config:       playbook.Spec.Config.Value(),
	}, nil
}
