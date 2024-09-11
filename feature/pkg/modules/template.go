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
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubesphere/kubekey/v4/pkg/connector"

	kkcorev1alpha1 "github.com/kubesphere/kubekey/v4/pkg/apis/core/v1alpha1"
	"github.com/kubesphere/kubekey/v4/pkg/converter/tmpl"
	"github.com/kubesphere/kubekey/v4/pkg/project"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

type templateArgs struct {
	src  string
	dest string
	mode *int
}

func newTemplateArgs(_ context.Context, raw runtime.RawExtension, vars map[string]any) (*templateArgs, error) {
	var err error
	// check args
	ta := &templateArgs{}
	args := variable.Extension2Variables(raw)

	ta.src, err = variable.StringVar(vars, args, "src")
	if err != nil {
		klog.V(4).ErrorS(err, "\"src\" should be string")

		return nil, errors.New("\"src\" should be string")
	}

	ta.dest, err = variable.StringVar(vars, args, "dest")
	if err != nil {
		return nil, errors.New("\"dest\" should be string")
	}

	ta.mode, _ = variable.IntVar(vars, args, "mode")

	return ta, nil
}

// ModuleTemplate deal "template" module
func ModuleTemplate(ctx context.Context, options ExecOptions) (string, string) {
	// get host variable
	ha, err := options.getAllVariables()
	if err != nil {
		return "", err.Error()
	}

	ta, err := newTemplateArgs(ctx, options.Args, ha)
	if err != nil {
		klog.V(4).ErrorS(err, "get template args error", "task", ctrlclient.ObjectKeyFromObject(&options.Task))

		return "", err.Error()
	}

	// get connector
	conn, err := getConnector(ctx, options.Host, ha)
	if err != nil {
		return "", err.Error()
	}
	defer conn.Close(ctx)

	if filepath.IsAbs(ta.src) {
		fileInfo, err := os.Stat(ta.src)
		if err != nil {
			return "", fmt.Sprintf(" get src file %s in local path error: %v", ta.src, err)
		}

		if fileInfo.IsDir() { // src is dir
			if err := ta.absDir(ctx, conn, ha); err != nil {
				return "", fmt.Sprintf("sync template absolute dir error %s", err)
			}
		} else { // src is file
			if err := ta.absFile(ctx, fileInfo.Mode(), conn, ha); err != nil {
				return "", fmt.Sprintf("sync template absolute file error %s", err)
			}
		}
	} else {
		pj, err := project.New(ctx, options.Pipeline, false)
		if err != nil {
			return "", fmt.Sprintf("get project error: %v", err)
		}

		fileInfo, err := pj.Stat(ta.src, project.GetFileOption{IsTemplate: true, Role: options.Task.Annotations[kkcorev1alpha1.TaskAnnotationRole]})
		if err != nil {
			return "", fmt.Sprintf("get file %s from project error: %v", ta.src, err)
		}

		if fileInfo.IsDir() {
			if err := ta.relDir(ctx, pj, options.Task.Annotations[kkcorev1alpha1.TaskAnnotationRole], conn, ha); err != nil {
				return "", fmt.Sprintf("sync template relative dir error: %s", err)
			}
		} else {
			if err := ta.relFile(ctx, pj, options.Task.Annotations[kkcorev1alpha1.TaskAnnotationRole], fileInfo.Mode(), conn, ha); err != nil {
				return "", fmt.Sprintf("sync template relative dir error: %s", err)
			}
		}
	}

	return StdoutSuccess, ""
}

// relFile when template.src is relative file, get file from project, parse it, and copy to remote.
func (ta templateArgs) relFile(ctx context.Context, pj project.Project, role string, mode fs.FileMode, conn connector.Connector, vars map[string]any) any {
	data, err := pj.ReadFile(ta.src, project.GetFileOption{IsTemplate: true, Role: role})
	if err != nil {
		return fmt.Errorf("read file error: %w", err)
	}

	result, err := tmpl.ParseString(vars, string(data))
	if err != nil {
		return fmt.Errorf("parse file error: %w", err)
	}

	dest := ta.dest
	if strings.HasSuffix(ta.dest, "/") {
		dest = filepath.Join(ta.dest, filepath.Base(ta.src))
	}

	if ta.mode != nil {
		mode = os.FileMode(*ta.mode)
	}

	if err := conn.PutFile(ctx, []byte(result), dest, mode); err != nil {
		return fmt.Errorf("copy file error: %w", err)
	}

	return nil
}

// relDir when template.src is relative dir, get all files from project, parse it, and copy to remote.
func (ta templateArgs) relDir(ctx context.Context, pj project.Project, role string, conn connector.Connector, vars map[string]any) error {
	if err := pj.WalkDir(ta.src, project.GetFileOption{IsTemplate: true, Role: role}, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() { // only copy file
			return nil
		}
		if err != nil {
			return fmt.Errorf("walk dir %s error: %w", ta.src, err)
		}

		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("get file info error: %w", err)
		}

		mode := info.Mode()
		if ta.mode != nil {
			mode = os.FileMode(*ta.mode)
		}

		data, err := pj.ReadFile(path, project.GetFileOption{IsTemplate: true, Role: role})
		if err != nil {
			return fmt.Errorf("read file error: %w", err)
		}
		result, err := tmpl.ParseString(vars, string(data))
		if err != nil {
			return fmt.Errorf("parse file error: %w", err)
		}

		dest := ta.dest
		if strings.HasSuffix(ta.dest, "/") {
			rel, err := pj.Rel(ta.src, path, project.GetFileOption{IsTemplate: true, Role: role})
			if err != nil {
				return fmt.Errorf("get relative file path error: %w", err)
			}
			dest = filepath.Join(ta.dest, rel)
		}

		if err := conn.PutFile(ctx, []byte(result), dest, mode); err != nil {
			return fmt.Errorf("copy file error: %w", err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

// absFile when template.src is absolute file, get file by os, parse it, and copy to remote.
func (ta templateArgs) absFile(ctx context.Context, mode fs.FileMode, conn connector.Connector, vars map[string]any) error {
	data, err := os.ReadFile(ta.src)
	if err != nil {
		return fmt.Errorf("read file error: %w", err)
	}

	result, err := tmpl.ParseString(vars, string(data))
	if err != nil {
		return fmt.Errorf("parse file error: %w", err)
	}

	dest := ta.dest
	if strings.HasSuffix(ta.dest, "/") {
		dest = filepath.Join(ta.dest, filepath.Base(ta.src))
	}

	if ta.mode != nil {
		mode = os.FileMode(*ta.mode)
	}

	if err := conn.PutFile(ctx, []byte(result), dest, mode); err != nil {
		return fmt.Errorf("copy file error: %w", err)
	}

	return nil
}

// absDir when template.src is absolute dir, get all files by os, parse it, and copy to remote.
func (ta templateArgs) absDir(ctx context.Context, conn connector.Connector, vars map[string]any) error {
	if err := filepath.WalkDir(ta.src, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() { // only copy file
			return nil
		}
		if err != nil {
			return fmt.Errorf("walk dir %s error: %w", ta.src, err)
		}

		// get file old mode
		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("get file info error: %w", err)
		}
		mode := info.Mode()
		if ta.mode != nil {
			mode = os.FileMode(*ta.mode)
		}
		// read file
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read file error: %w", err)
		}
		result, err := tmpl.ParseString(vars, string(data))
		if err != nil {
			return fmt.Errorf("parse file error: %w", err)
		}
		// copy file to remote
		dest := ta.dest
		if strings.HasSuffix(ta.dest, "/") {
			rel, err := filepath.Rel(ta.src, path)
			if err != nil {
				return fmt.Errorf("get relative file path error: %w", err)
			}
			dest = filepath.Join(ta.dest, rel)
		}

		if err := conn.PutFile(ctx, []byte(result), dest, mode); err != nil {
			return fmt.Errorf("copy file error: %w", err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}
