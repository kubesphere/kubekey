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
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/klog/v2"

	kubekeyv1alpha1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/v4/pkg/converter/tmpl"
	"github.com/kubesphere/kubekey/v4/pkg/project"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

func ModuleTemplate(ctx context.Context, options ExecOptions) (string, string) {
	// get host variable
	ha, err := options.Variable.Get(variable.GetAllVariable(options.Host))
	if err != nil {
		klog.V(4).ErrorS(err, "failed to get host variable", "hostname", options.Host)
		return "", err.Error()
	}
	// check args
	args := variable.Extension2Variables(options.Args)
	srcParam, err := variable.StringVar(ha.(map[string]any), args, "src")
	if err != nil {
		return "", "\"src\" should be string"
	}
	destParam, err := variable.StringVar(ha.(map[string]any), args, "dest")
	if err != nil {
		return "", "\"dest\" should be string"
	}

	// get connector
	conn, err := getConnector(ctx, options.Host, ha.(map[string]any))
	if err != nil {
		return "", err.Error()
	}
	defer conn.Close(ctx)

	if filepath.IsAbs(srcParam) {
		fileInfo, err := os.Stat(srcParam)
		if err != nil {
			return "", fmt.Sprintf(" get src file %s in local path error: %v", srcParam, err)
		}

		if fileInfo.IsDir() { // src is dir
			if err := filepath.WalkDir(srcParam, func(path string, d fs.DirEntry, err error) error {
				if d.IsDir() { // only copy file
					return nil
				}
				if err != nil {
					return fmt.Errorf("walk dir %s error: %w", srcParam, err)
				}

				// get file old mode
				info, err := d.Info()
				if err != nil {
					return fmt.Errorf("get file info error: %w", err)
				}
				mode := info.Mode()
				if modeParam, err := variable.IntVar(ha.(map[string]any), args, "mode"); err == nil {
					mode = os.FileMode(modeParam)
				}
				// read file
				data, err := os.ReadFile(path)
				if err != nil {
					return fmt.Errorf("read file error: %w", err)
				}
				result, err := tmpl.ParseFile(ha.(map[string]any), data)
				if err != nil {
					return fmt.Errorf("parse file error: %w", err)
				}
				// copy file to remote
				if err := conn.CopyFile(ctx, []byte(result), path, mode); err != nil {
					return fmt.Errorf("copy file error: %w", err)
				}
				return nil
			}); err != nil {
				return "", fmt.Sprintf(" walk dir %s in local path error: %v", srcParam, err)
			}
		} else { // src is file
			data, err := os.ReadFile(srcParam)
			if err != nil {
				return "", fmt.Sprintf("read file error: %v", err)
			}
			result, err := tmpl.ParseFile(ha.(map[string]any), data)
			if err != nil {
				return "", fmt.Sprintf("parse file error: %v", err)
			}
			if strings.HasSuffix(destParam, "/") {
				destParam += filepath.Base(srcParam)
			}
			mode := fileInfo.Mode()
			if modeParam, err := variable.IntVar(ha.(map[string]any), args, "mode"); err == nil {
				mode = os.FileMode(modeParam)
			}
			if err := conn.CopyFile(ctx, []byte(result), destParam, mode); err != nil {
				return "", fmt.Sprintf("copy file error: %v", err)
			}
		}
	} else {
		pj, err := project.New(options.Pipeline, false)
		if err != nil {
			return "", fmt.Sprintf("get project error: %v", err)
		}
		fileInfo, err := pj.Stat(srcParam, project.GetFileOption{IsTemplate: true, Role: options.Task.Annotations[kubekeyv1alpha1.TaskAnnotationRole]})
		if err != nil {
			return "", fmt.Sprintf("get file %s from project error %v", srcParam, err)
		}

		if fileInfo.IsDir() {
			if err := pj.WalkDir(srcParam, project.GetFileOption{IsTemplate: true, Role: options.Task.Annotations[kubekeyv1alpha1.TaskAnnotationRole]}, func(path string, d fs.DirEntry, err error) error {
				if d.IsDir() { // only copy file
					return nil
				}
				if err != nil {
					return fmt.Errorf("walk dir %s error: %v", srcParam, err)
				}

				info, err := d.Info()
				if err != nil {
					return fmt.Errorf("get file info error: %v", err)
				}
				mode := info.Mode()
				if modeParam, err := variable.IntVar(ha.(map[string]any), args, "mode"); err == nil {
					mode = os.FileMode(modeParam)
				}
				data, err := pj.ReadFile(path, project.GetFileOption{Role: options.Task.Annotations[kubekeyv1alpha1.TaskAnnotationRole]})
				if err != nil {
					return fmt.Errorf("read file error: %v", err)
				}
				result, err := tmpl.ParseFile(ha.(map[string]any), data)
				if err != nil {
					return fmt.Errorf("parse file error: %v", err)
				}
				if err := conn.CopyFile(ctx, []byte(result), path, mode); err != nil {
					return fmt.Errorf("copy file error: %v", err)
				}
				return nil
			}); err != nil {
				return "", fmt.Sprintf("")
			}
		} else {
			data, err := pj.ReadFile(srcParam, project.GetFileOption{IsTemplate: true, Role: options.Task.Annotations[kubekeyv1alpha1.TaskAnnotationRole]})
			if err != nil {
				return "", fmt.Sprintf("read file error: %v", err)
			}
			result, err := tmpl.ParseFile(ha.(map[string]any), data)
			if err != nil {
				return "", fmt.Sprintf("parse file error: %v", err)
			}
			if strings.HasSuffix(destParam, "/") {
				destParam = destParam + filepath.Base(srcParam)
			}
			mode := fileInfo.Mode()
			if modeParam, err := variable.IntVar(ha.(map[string]any), args, "mode"); err == nil {
				mode = os.FileMode(modeParam)
			}
			if err := conn.CopyFile(ctx, []byte(result), destParam, mode); err != nil {
				return "", fmt.Sprintf("copy file error: %v", err)
			}
		}
	}
	return "success", ""
}
