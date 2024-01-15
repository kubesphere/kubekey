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

package connector

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"k8s.io/klog/v2"
	"k8s.io/utils/exec"
)

type localConnector struct {
	Cmd exec.Interface
}

func (c *localConnector) Init(ctx context.Context) error {
	return nil
}

func (c *localConnector) Close(ctx context.Context) error {
	return nil
}

func (c *localConnector) CopyFile(ctx context.Context, local []byte, remoteFile string, mode fs.FileMode) error {
	// create remote file
	if _, err := os.Stat(filepath.Dir(remoteFile)); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(remoteFile), mode); err != nil {
			klog.ErrorS(err, "Failed to create remote dir", "remote_file", remoteFile)
			return err
		}
	}
	rf, err := os.Create(remoteFile)
	if err != nil {
		klog.ErrorS(err, "Failed to create remote file", "remote_file", remoteFile)
		return err
	}
	if _, err := rf.Write(local); err != nil {
		klog.ErrorS(err, "Failed to write content to remote file", "remote_file", remoteFile)
		return err
	}
	return rf.Chmod(mode)
}

func (c *localConnector) FetchFile(ctx context.Context, remoteFile string, local io.Writer) error {
	var err error
	file, err := os.Open(remoteFile)
	if err != nil {
		klog.ErrorS(err, "Failed to read remote file failed", "remote_file", remoteFile)
		return err
	}
	if _, err := io.Copy(local, file); err != nil {
		klog.ErrorS(err, "Failed to copy remote file to local", "remote_file", remoteFile)
		return err
	}
	return nil
}

func (c *localConnector) ExecuteCommand(ctx context.Context, cmd string) ([]byte, error) {
	return c.Cmd.CommandContext(ctx, cmd).CombinedOutput()
}
