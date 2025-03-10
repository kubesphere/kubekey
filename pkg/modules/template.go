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
	"math"
	"os"
	"path/filepath"
	"strings"

	kkcorev1alpha1 "github.com/kubesphere/kubekey/api/core/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubesphere/kubekey/v4/pkg/connector"
	"github.com/kubesphere/kubekey/v4/pkg/converter/tmpl"
	"github.com/kubesphere/kubekey/v4/pkg/project"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

type templateArgs struct {
	src  string
	dest string
	mode *uint32
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

	mode, err := variable.IntVar(vars, args, "mode")
	if err != nil {
		klog.V(4).InfoS("get mode error", "error", err)
	} else {
		if *mode < 0 || *mode > math.MaxUint32 {
			return nil, errors.New("mode should be uint32")
		}
		ta.mode = ptr.To(uint32(*mode))
	}

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
	conn, err := getConnector(ctx, options.Host, options.Variable)
	if err != nil {
		return "", err.Error()
	}
	defer conn.Close(ctx)

	dealAbsoluteFilePath := func() (string, string) {
		fileInfo, err := os.Stat(ta.src)
		if err != nil {
			return "", fmt.Sprintf(" get src file %s in local path error: %v", ta.src, err)
		}

		if fileInfo.IsDir() { // src is dir
			if err := ta.absDir(ctx, conn, ha); err != nil {
				return "", fmt.Sprintf("sync template absolute dir error %s", err)
			}
		} else { // src is file
			data, err := os.ReadFile(ta.src)
			if err != nil {
				return "", fmt.Sprintf("read file error: %s", err)
			}
			if err := ta.readFile(ctx, string(data), fileInfo.Mode(), conn, ha); err != nil {
				return "", fmt.Sprintf("sync template absolute file error %s", err)
			}
		}

		return StdoutSuccess, ""
	}
	dealRelativeFilePath := func() (string, string) {
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
			data, err := pj.ReadFile(ta.src, project.GetFileOption{IsTemplate: true, Role: options.Task.Annotations[kkcorev1alpha1.TaskAnnotationRole]})
			if err != nil {
				return "", fmt.Sprintf("read file error: %s", err)
			}
			if err := ta.readFile(ctx, string(data), fileInfo.Mode(), conn, ha); err != nil {
				return "", fmt.Sprintf("sync template relative dir error: %s", err)
			}
		}

		return StdoutSuccess, ""
	}
	if filepath.IsAbs(ta.src) {
		return dealAbsoluteFilePath()
	}

	return dealRelativeFilePath()
}

// relFile when template.src is relative file, get file from project, parse it, and copy to remote.
func (ta templateArgs) readFile(ctx context.Context, data string, mode fs.FileMode, conn connector.Connector, vars map[string]any) any {
	result, err := tmpl.Parse(vars, data)
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

	if err := conn.PutFile(ctx, result, dest, mode); err != nil {
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
		result, err := tmpl.Parse(vars, string(data))
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

		if err := conn.PutFile(ctx, result, dest, mode); err != nil {
			return fmt.Errorf("copy file error: %w", err)
		}

		return nil
	}); err != nil {
		return err
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
		result, err := tmpl.Parse(vars, string(data))
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

		if err := conn.PutFile(ctx, result, dest, mode); err != nil {
			return fmt.Errorf("copy file error: %w", err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}
