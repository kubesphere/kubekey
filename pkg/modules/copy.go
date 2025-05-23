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
	"github.com/kubesphere/kubekey/v4/pkg/project"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

/*
The Copy module copies files or content from local to remote hosts.
This module allows users to transfer files or create files with specified content on remote hosts.

Configuration:
Users can specify either a source file or content to copy.

copy:
  src: /path/to/file    # optional: local file path to copy
  content: "text"       # optional: content to write to file
  dest: /remote/path    # required: destination path on remote host
  mode: 0644           # optional: file permissions (default: 0644)

Usage Examples in Playbook Tasks:
1. Copy local file:
   ```yaml
   - name: Copy configuration file
     copy:
       src: config.yaml
       dest: /etc/app/config.yaml
       mode: 0644
     register: copy_result
   ```

2. Create file with content:
   ```yaml
   - name: Create config file
     copy:
       content: |
         server: localhost
         port: 8080
       dest: /etc/app/config.yaml
     register: config_result
   ```

3. Copy directory:
   ```yaml
   - name: Copy application files
     copy:
       src: app/
       dest: /opt/app/
     register: app_files
   ```

Return Values:
- On success: Returns "Success" in stdout
- On failure: Returns error message in stderr
*/

type copyArgs struct {
	src     string
	content string
	dest    string
	mode    *uint32
}

func newCopyArgs(_ context.Context, raw runtime.RawExtension, vars map[string]any) (*copyArgs, error) {
	var err error
	ca := &copyArgs{}
	args := variable.Extension2Variables(raw)
	ca.src, _ = variable.StringVar(vars, args, "src")
	ca.content, _ = variable.StringVar(vars, args, "content")
	ca.dest, err = variable.StringVar(vars, args, "dest")
	if err != nil {
		return nil, errors.New("\"dest\" in args should be string")
	}
	mode, err := variable.IntVar(vars, args, "mode")
	if err != nil {
		klog.V(4).InfoS("get mode error", "error", err)
	} else {
		if *mode < 0 || *mode > math.MaxUint32 {
			return nil, errors.New("mode should be uint32")
		}
		ca.mode = ptr.To(uint32(*mode))
	}

	return ca, nil
}

// ModuleCopy handles the "copy" module, copying files or content to remote hosts
func ModuleCopy(ctx context.Context, options ExecOptions) (string, string) {
	// get host variable
	ha, err := options.getAllVariables()
	if err != nil {
		return "", err.Error()
	}

	ca, err := newCopyArgs(ctx, options.Args, ha)
	if err != nil {
		return "", err.Error()
	}

	// get connector
	conn, err := getConnector(ctx, options.Host, options.Variable)
	if err != nil {
		return "", fmt.Sprintf("get connector error: %v", err)
	}
	defer conn.Close(ctx)

	switch {
	case ca.src != "": // copy local file to remote
		return ca.copySrc(ctx, options, conn)
	case ca.content != "":
		return ca.copyContent(ctx, os.ModePerm, conn)
	default:
		return "", "either \"src\" or \"content\" must be provided."
	}
}

// copySrc copy src file to dest
func (ca copyArgs) copySrc(ctx context.Context, options ExecOptions, conn connector.Connector) (string, string) {
	if filepath.IsAbs(ca.src) { // if src is absolute path. find it in local path
		return ca.handleAbsolutePath(ctx, conn)
	}
	// if src is not absolute path. find file in project
	return ca.handleRelativePath(ctx, options, conn)
}

func (ca copyArgs) handleAbsolutePath(ctx context.Context, conn connector.Connector) (string, string) {
	fileInfo, err := os.Stat(ca.src)
	if err != nil {
		return "", fmt.Sprintf(" get src file %s in local path error: %v", ca.src, err)
	}

	if fileInfo.IsDir() { // src is dir
		if err := ca.absDir(ctx, conn); err != nil {
			return "", fmt.Sprintf("sync copy absolute dir error %s", err)
		}

		return StdoutSuccess, ""
	}

	// src is file
	data, err := os.ReadFile(ca.src)
	if err != nil {
		return "", fmt.Sprintf("read file error: %s", err)
	}
	if err := ca.readFile(ctx, data, fileInfo.Mode(), conn); err != nil {
		return "", fmt.Sprintf("sync copy absolute dir error %s", err)
	}

	return StdoutSuccess, ""
}

func (ca copyArgs) handleRelativePath(ctx context.Context, options ExecOptions, conn connector.Connector) (string, string) {
	pj, err := project.New(ctx, options.Playbook, false)
	if err != nil {
		return "", fmt.Sprintf("get project error: %v", err)
	}

	relPath := filepath.Join(options.Task.Annotations[kkcorev1alpha1.TaskAnnotationRelativePath], _const.ProjectRolesFilesDir, ca.src)
	fileInfo, err := pj.Stat(relPath)
	if err != nil {
		return "", fmt.Sprintf("get file %s from project error %v", ca.src, err)
	}

	if fileInfo.IsDir() {
		if err := ca.handleRelativeDir(ctx, pj, relPath, conn); err != nil {
			return "", fmt.Sprintf("sync copy relative dir error %s", err)
		}

		return StdoutSuccess, ""
	}

	// Handle single file
	data, err := pj.ReadFile(relPath)
	if err != nil {
		return "", fmt.Sprintf("read file error: %s", err)
	}
	if err := ca.readFile(ctx, data, fileInfo.Mode(), conn); err != nil {
		return "", fmt.Sprintf("sync copy relative dir error %s", err)
	}

	return StdoutSuccess, ""
}

func (ca copyArgs) handleRelativeDir(ctx context.Context, pj project.Project, relPath string, conn connector.Connector) error {
	return pj.WalkDir(relPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() { // only copy file
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return errors.Wrap(err, "failed to get file info")
		}

		mode := info.Mode()
		if ca.mode != nil {
			mode = os.FileMode(*ca.mode)
		}

		data, err := pj.ReadFile(path)
		if err != nil {
			return errors.Wrap(err, "failed to read file")
		}

		dest := ca.dest
		if strings.HasSuffix(ca.dest, "/") {
			rel, err := pj.Rel(relPath, path)
			if err != nil {
				return errors.Wrap(err, "failed to get relative file path")
			}
			dest = filepath.Join(ca.dest, rel)
		}

		return conn.PutFile(ctx, data, dest, mode)
	})
}

// copyContent convert content param and copy to dest
func (ca copyArgs) copyContent(ctx context.Context, mode fs.FileMode, conn connector.Connector) (string, string) {
	if strings.HasSuffix(ca.dest, "/") {
		return "", "\"content\" should copy to a file"
	}

	if ca.mode != nil {
		mode = os.FileMode(*ca.mode)
	}

	if err := conn.PutFile(ctx, []byte(ca.content), ca.dest, mode); err != nil {
		return "", fmt.Sprintf("copy file error: %v", err)
	}

	return StdoutSuccess, ""
}

// absFile when copy.src is absolute file, get file from os, and copy to remote.
func (ca copyArgs) readFile(ctx context.Context, data []byte, mode fs.FileMode, conn connector.Connector) error {
	dest := ca.dest
	if strings.HasSuffix(ca.dest, "/") {
		dest = filepath.Join(ca.dest, filepath.Base(ca.src))
	}

	if ca.mode != nil {
		mode = os.FileMode(*ca.mode)
	}

	return conn.PutFile(ctx, data, dest, mode)
}

// absDir when copy.src is absolute dir, get all files from os, and copy to remote.
func (ca copyArgs) absDir(ctx context.Context, conn connector.Connector) error {
	return filepath.WalkDir(ca.src, func(path string, d fs.DirEntry, err error) error {
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
		if ca.mode != nil {
			mode = os.FileMode(*ca.mode)
		}
		// read file
		data, err := os.ReadFile(path)
		if err != nil {
			return errors.Wrapf(err, "failed to read file %q", path)
		}
		// copy file to remote
		dest := ca.dest
		if strings.HasSuffix(ca.dest, "/") {
			rel, err := filepath.Rel(ca.src, path)
			if err != nil {
				return errors.Wrap(err, "failed to get relative filepath")
			}
			dest = filepath.Join(ca.dest, rel)
		}

		return conn.PutFile(ctx, data, dest, mode)
	})
}

func init() {
	utilruntime.Must(RegisterModule("copy", ModuleCopy))
}
