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
	"strings"

	"k8s.io/klog/v2"

	kubekeyv1alpha1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/v4/pkg/connector"
	"github.com/kubesphere/kubekey/v4/pkg/converter/tmpl"
	"github.com/kubesphere/kubekey/v4/pkg/project"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

func ModuleCopy(ctx context.Context, options ExecOptions) (string, string) {
	// check args
	args := variable.Extension2Variables(options.Args)
	src := variable.StringVar(args, "src")
	content := variable.StringVar(args, "content")
	if src == nil && content == nil {
		return "", "\"src\" or \"content\" in args should be string"
	}
	dest := variable.StringVar(args, "dest")
	if dest == nil {
		return "", "\"dest\" in args should be string"
	}
	lv, err := options.Variable.Get(variable.LocationVars{
		HostName:    options.Host,
		LocationUID: string(options.Task.UID),
	})
	if err != nil {
		klog.V(4).ErrorS(err, "failed to get location vars")
		return "", err.Error()
	}
	destStr, err := tmpl.ParseString(lv.(variable.VariableData), *dest)
	if err != nil {
		klog.V(4).ErrorS(err, "template parse dest error")
		return "", err.Error()
	}

	var conn connector.Connector
	if v := ctx.Value("connector"); v != nil {
		conn = v.(connector.Connector)
	} else {
		// get connector
		ha, err := options.Variable.Get(variable.HostVars{HostName: options.Host})
		if err != nil {
			klog.V(4).ErrorS(err, "failed to get host vars")
			return "", err.Error()
		}
		conn = connector.NewConnector(options.Host, ha.(variable.VariableData))
	}
	if err := conn.Init(ctx); err != nil {
		klog.V(4).ErrorS(err, "failed to init connector")
		return "", err.Error()
	}
	defer conn.Close(ctx)

	if src != nil {
		// convert src
		srcStr, err := tmpl.ParseString(lv.(variable.VariableData), *src)
		if err != nil {
			klog.V(4).ErrorS(err, "template parse src error")
			return "", err.Error()
		}
		var baseFS fs.FS
		if filepath.IsAbs(srcStr) {
			baseFS = os.DirFS("/")
		} else {
			projectFs, err := project.New(project.Options{Pipeline: &options.Pipeline}).FS(ctx, false)
			if err != nil {
				klog.V(4).ErrorS(err, "failed to get project fs")
				return "", err.Error()
			}
			baseFS = projectFs
		}
		roleName := options.Task.Annotations[kubekeyv1alpha1.TaskAnnotationRole]
		flPath := project.GetFilesFromPlayBook(baseFS, options.Pipeline.Spec.Playbook, roleName, srcStr)
		fileInfo, err := fs.Stat(baseFS, flPath)
		if err != nil {
			klog.V(4).ErrorS(err, "failed to get src file in local")
			return "", err.Error()
		}
		if fileInfo.IsDir() {
			// src is dir
			if err := fs.WalkDir(baseFS, flPath, func(path string, info fs.DirEntry, err error) error {
				if err != nil {
					klog.V(4).ErrorS(err, "failed to walk dir")
					return err
				}
				rel, err := filepath.Rel(srcStr, path)
				if err != nil {
					klog.V(4).ErrorS(err, "failed to get relative path")
					return err
				}
				if info.IsDir() {
					return nil
				}
				fi, err := info.Info()
				if err != nil {
					klog.V(4).ErrorS(err, "failed to get file info")
					return err
				}
				mode := fi.Mode()
				if variable.IntVar(args, "mode") != nil {
					mode = os.FileMode(*variable.IntVar(args, "mode"))
				}
				data, err := fs.ReadFile(baseFS, rel)
				if err != nil {
					klog.V(4).ErrorS(err, "failed to read file")
					return err
				}
				if err := conn.CopyFile(ctx, data, filepath.Join(destStr, rel), mode); err != nil {
					klog.V(4).ErrorS(err, "failed to copy file", "src", srcStr, "dest", destStr)
					return err
				}
				return nil
			}); err != nil {
				klog.V(4).ErrorS(err, "failed to walk dir")
				return "", err.Error()
			}
		} else {
			// src is file
			data, err := fs.ReadFile(baseFS, flPath)
			if err != nil {
				klog.V(4).ErrorS(err, "failed to read file")
				return "", err.Error()
			}
			if strings.HasSuffix(destStr, "/") {
				destStr = destStr + filepath.Base(srcStr)
			}
			if err := conn.CopyFile(ctx, data, destStr, fileInfo.Mode()); err != nil {
				klog.V(4).ErrorS(err, "failed to copy file", "src", srcStr, "dest", destStr)
				return "", err.Error()
			}
			return "success", ""
		}
	} else if content != nil {
		if strings.HasSuffix(destStr, "/") {
			return "", "\"content\" should copy to a file"
		}
		mode := os.ModePerm
		if v := variable.IntVar(args, "mode"); v != nil {
			mode = os.FileMode(*v)
		}

		if err := conn.CopyFile(ctx, []byte(*content), destStr, mode); err != nil {
			return "", err.Error()
		}
	}
	return "success", ""
}
