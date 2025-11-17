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
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/cockroachdb/errors"
	kkcorev1alpha1 "github.com/kubesphere/kubekey/api/core/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"

	"github.com/kubesphere/kubekey/v4/pkg/connector"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/converter/tmpl"
	"github.com/kubesphere/kubekey/v4/pkg/project"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

/*
The Template module processes files using Go templates before copying them to remote hosts.
This module allows users to template files with variables before transferring them.

Configuration:
Users can specify source and destination paths:

template:
  src: /path/to/file    # required: local file path to template
  dest: /remote/path    # required: destination path on remote host
  mode: 0644           # optional: file permissions (default: 0644)

Usage Examples in Playbook Tasks:
1. Basic file templating:
   ```yaml
   - name: Template configuration file
     template:
       src: config.yaml.tmpl
       dest: /etc/app/config.yaml
       mode: 0644
     register: template_result
   ```

2. Template with variables:
   ```yaml
   - name: Template with variables
     template:
       src: app.conf.tmpl
       dest: /etc/app/app.conf
     register: app_config
   ```

Return Values:
- On success: Returns "Success" in stdout
- On failure: Returns error message in stderr
*/

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

// ModuleTemplate handles the "template" module, processing files with Go templates
func ModuleTemplate(ctx context.Context, options ExecOptions) (string, string, error) {
	ha, err := options.getAllVariables()
	if err != nil {
		return StdoutFailed, StderrGetHostVariable, err
	}

	ta, err := newTemplateArgs(ctx, options.Args, ha)
	if err != nil {
		return StdoutFailed, StderrParseArgument, err
	}

	conn, err := options.getConnector(ctx)
	if err != nil {
		return StdoutFailed, StderrGetConnector, err
	}
	defer conn.Close(ctx)

	if filepath.IsAbs(ta.src) {
		return handleAbsoluteTemplate(ctx, ta, conn, ha)
	}

	return handleRelativeTemplate(ctx, ta, conn, ha, options)
}

func handleAbsoluteTemplate(ctx context.Context, ta *templateArgs, conn connector.Connector, vars map[string]any) (string, string, error) {
	fileInfo, err := os.Stat(ta.src)
	if err != nil {
		return StdoutFailed, "failed to get src file in local path", err
	}

	if fileInfo.IsDir() {
		if err := ta.absDir(ctx, conn, vars); err != nil {
			return StdoutFailed, "failed to template absolute dir", err
		}

		return StdoutSuccess, "", nil
	}

	data, err := os.ReadFile(ta.src)
	if err != nil {
		return StdoutFailed, "failed to read file", err
	}
	if err := ta.readFile(ctx, string(data), fileInfo.Mode(), conn, vars); err != nil {
		return StdoutFailed, "failed to template file", err
	}

	return StdoutSuccess, "", nil
}

func handleRelativeTemplate(ctx context.Context, ta *templateArgs, conn connector.Connector, vars map[string]any, options ExecOptions) (string, string, error) {
	pj, err := project.New(ctx, options.Playbook, false)
	if err != nil {
		return StdoutFailed, "failed to get playbook", nil
	}

	relPath := filepath.Join(options.Task.Annotations[kkcorev1alpha1.TaskAnnotationRelativePath], _const.ProjectRolesTemplateDir, ta.src)
	fileInfo, err := pj.Stat(relPath)
	if err != nil {
		return StdoutFailed, "failed to stat relative path", err
	}

	if fileInfo.IsDir() {
		if err := handleRelativeDir(ctx, pj, relPath, ta, conn, vars); err != nil {
			return StdoutFailed, "failed to template relative dir", err
		}

		return StdoutSuccess, "", nil
	}

	data, err := pj.ReadFile(relPath)
	if err != nil {
		return StdoutFailed, "failed to read relative file", err
	}
	if err := ta.readFile(ctx, string(data), fileInfo.Mode(), conn, vars); err != nil {
		return StdoutFailed, "failed to template relative file", err
	}

	return StdoutSuccess, "", nil
}

func handleRelativeDir(ctx context.Context, pj project.Project, relPath string, ta *templateArgs, conn connector.Connector, vars map[string]any) error {
	return pj.WalkDir(relPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() { // only deal file
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return errors.Wrapf(err, "failed to get file %q info", path)
		}

		mode := info.Mode()
		if ta.mode != nil {
			mode = os.FileMode(*ta.mode)
		}

		data, err := pj.ReadFile(path)
		if err != nil {
			return errors.Wrapf(err, "failed to read file %q", path)
		}
		result, err := tmpl.Parse(vars, string(data))
		if err != nil {
			return errors.Wrapf(err, "failed to parse file %q", path)
		}

		dest := ta.dest
		if strings.HasSuffix(ta.dest, "/") {
			rel, err := pj.Rel(relPath, path)
			if err != nil {
				return errors.Wrap(err, "failed to get relative filepath")
			}
			dest = filepath.Join(ta.dest, rel)
		}
		tmpDest := filepath.Join("/tmp", dest)

		if err = conn.PutFile(ctx, result, tmpDest, mode); err != nil {
			return err
		}

		_, _, err = conn.ExecuteCommand(ctx, fmt.Sprintf("mkdir -p %s\nmv %s %s", filepath.Dir(dest), tmpDest, dest))

		return err
	})
}

// relFile when template.src is relative file, get file from project, parse it, and copy to remote.
func (ta templateArgs) readFile(ctx context.Context, data string, mode fs.FileMode, conn connector.Connector, vars map[string]any) error {
	result, err := tmpl.Parse(vars, data)
	if err != nil {
		return err
	}

	dest := ta.dest
	if strings.HasSuffix(ta.dest, "/") {
		dest = filepath.Join(ta.dest, filepath.Base(ta.src))
	}

	if ta.mode != nil {
		mode = os.FileMode(*ta.mode)
	}
	tmpDest := filepath.Join("/tmp", dest)

	if err = conn.PutFile(ctx, result, tmpDest, mode); err != nil {
		return err
	}

	_, _, err = conn.ExecuteCommand(ctx, fmt.Sprintf("mkdir -p %s\nmv %s %s", filepath.Dir(dest), tmpDest, dest))

	return err
}

// absDir when template.src is absolute dir, get all files by os, parse it, and copy to remote.
func (ta templateArgs) absDir(ctx context.Context, conn connector.Connector, vars map[string]any) error {
	if err := filepath.WalkDir(ta.src, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() { // only copy file
			return nil
		}
		if err != nil {
			return errors.WithStack(err)
		}

		// get file old mode
		info, err := d.Info()
		if err != nil {
			return errors.Wrapf(err, "failed to get file %q info", path)
		}
		mode := info.Mode()
		if ta.mode != nil {
			mode = os.FileMode(*ta.mode)
		}
		// read file
		data, err := os.ReadFile(path)
		if err != nil {
			return errors.Wrapf(err, "failed to read file %q", path)
		}
		result, err := tmpl.Parse(vars, string(data))
		if err != nil {
			return errors.Wrapf(err, "failed to parse file %q", path)
		}
		// copy file to remote
		dest := ta.dest
		if strings.HasSuffix(ta.dest, "/") {
			rel, err := filepath.Rel(ta.src, path)
			if err != nil {
				return errors.Wrap(err, "failed to get relative filepath")
			}
			dest = filepath.Join(ta.dest, rel)
		}

		tmpDest := filepath.Join("/tmp", dest)

		if err = conn.PutFile(ctx, result, tmpDest, mode); err != nil {
			return err
		}

		_, _, err = conn.ExecuteCommand(ctx, fmt.Sprintf("mkdir -p %s\nmv %s %s", filepath.Dir(dest), tmpDest, dest))

		return err
	}); err != nil {
		return errors.Wrapf(err, "failed to walk dir %q", ta.src)
	}

	return nil
}

func init() {
	utilruntime.Must(RegisterModule("template", ModuleTemplate))
}
