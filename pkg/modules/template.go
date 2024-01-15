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

package modules

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"

	"k8s.io/klog/v2"

	kubekeyv1alpha1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/v4/pkg/connector"
	"github.com/kubesphere/kubekey/v4/pkg/converter/tmpl"
	"github.com/kubesphere/kubekey/v4/pkg/project"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

func ModuleTemplate(ctx context.Context, options ExecOptions) (string, string) {
	// check args
	args := variable.Extension2Variables(options.Args)
	src := variable.StringVar(args, "src")
	if src == nil {
		return "", "\"src\" should be string"
	}
	dest := variable.StringVar(args, "dest")
	if dest == nil {
		return "", "\"dest\" should be string"
	}

	lv, err := options.Variable.Get(variable.LocationVars{
		HostName:    options.Host,
		LocationUID: string(options.Task.UID),
	})
	if err != nil {
		klog.ErrorS(err, "failed to get location vars")
		return "", err.Error()
	}
	srcStr, err := tmpl.ParseString(lv.(variable.VariableData), *src)
	if err != nil {
		klog.ErrorS(err, "template parse src error", "input", *src)
		return "", err.Error()
	}
	destStr, err := tmpl.ParseString(lv.(variable.VariableData), *dest)
	if err != nil {
		klog.ErrorS(err, "template parse dest error", "input", *dest)
		return "", err.Error()
	}

	var baseFS fs.FS
	if filepath.IsAbs(srcStr) {
		baseFS = os.DirFS("/")
	} else {
		projectFs, err := project.New(project.Options{Pipeline: &options.Pipeline}).FS(ctx, false)
		if err != nil {
			klog.ErrorS(err, "failed to get project fs")
			return "", err.Error()
		}
		baseFS = projectFs
	}
	roleName := options.Task.Annotations[kubekeyv1alpha1.TaskAnnotationRole]
	flPath := project.GetTemplatesFromPlayBook(baseFS, options.Pipeline.Spec.Playbook, roleName, srcStr)
	if _, err := fs.Stat(baseFS, flPath); err != nil {
		klog.ErrorS(err, "find src error")
		return "", err.Error()
	}

	var conn connector.Connector
	if v := ctx.Value("connector"); v != nil {
		conn = v.(connector.Connector)
	} else {
		// get connector
		ha, err := options.Variable.Get(variable.HostVars{HostName: options.Host})
		if err != nil {
			klog.ErrorS(err, "failed to get host vars")
			return "", err.Error()
		}
		conn = connector.NewConnector(options.Host, ha.(variable.VariableData))
	}
	if err := conn.Init(ctx); err != nil {
		klog.ErrorS(err, "failed to init connector")
		return "", err.Error()
	}
	defer conn.Close(ctx)

	// find src file
	lg, err := options.Variable.Get(variable.LocationVars{
		HostName:    options.Host,
		LocationUID: string(options.Task.UID),
	})
	if err != nil {
		klog.ErrorS(err, "failed to get location vars")
		return "", err.Error()
	}

	data, err := fs.ReadFile(baseFS, flPath)
	if err != nil {
		klog.ErrorS(err, "failed to read src file", "file_path", flPath)
		return "", err.Error()
	}
	result, err := tmpl.ParseFile(lg.(variable.VariableData), data)
	if err != nil {
		klog.ErrorS(err, "failed to parse file", "file_path", flPath)
		return "", err.Error()
	}

	// copy file
	mode := fs.ModePerm
	if v := variable.IntVar(args, "mode"); v != nil {
		mode = fs.FileMode(*v)
	}
	if err := conn.CopyFile(ctx, []byte(result), destStr, mode); err != nil {
		klog.ErrorS(err, "failed to copy file", "src", flPath, "dest", destStr)
		return "", err.Error()
	}
	return "success", ""
}
