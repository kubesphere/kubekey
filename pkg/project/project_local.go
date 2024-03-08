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

	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
	"github.com/kubesphere/kubekey/v4/project"
)

type localProject struct {
	kubekeyv1.Pipeline

	fs fs.FS
}

func (r localProject) FS(ctx context.Context, update bool) (fs.FS, error) {
	if _, ok := r.Pipeline.Annotations[kubekeyv1.BuiltinsProjectAnnotation]; ok {
		return project.InternalPipeline, nil
	}
	if filepath.IsAbs(r.Pipeline.Spec.Playbook) {
		return os.DirFS("/"), nil
	}

	if r.fs != nil {
		return r.fs, nil
	}

	if r.Pipeline.Spec.Project.Addr != "" {
		return os.DirFS(r.Pipeline.Spec.Project.Addr), nil
	}

	return nil, fmt.Errorf("cannot get filesystem from absolute project %s", r.Pipeline.Spec.Project.Addr)
}
