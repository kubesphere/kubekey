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
	"path/filepath"
	"strings"

	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
)

type Project interface {
	FS(ctx context.Context, update bool) (fs.FS, error)
}

type Options struct {
	*kubekeyv1.Pipeline
}

func New(o Options) Project {
	if strings.HasPrefix(o.Pipeline.Spec.Project.Addr, "https://") ||
		strings.HasPrefix(o.Pipeline.Spec.Project.Addr, "http://") ||
		strings.HasPrefix(o.Pipeline.Spec.Project.Addr, "git@") {
		// git clone to project dir
		if o.Pipeline.Spec.Project.Name == "" {
			o.Pipeline.Spec.Project.Name = strings.TrimSuffix(o.Pipeline.Spec.Project.Addr[strings.LastIndex(o.Pipeline.Spec.Project.Addr, "/")+1:], ".git")
		}
		return &gitProject{Pipeline: *o.Pipeline, localDir: filepath.Join(_const.GetWorkDir(), _const.ProjectDir, o.Spec.Project.Name)}
	}
	return &localProject{Pipeline: *o.Pipeline}

}
