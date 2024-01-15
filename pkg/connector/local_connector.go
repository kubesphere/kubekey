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
	if _, err := os.Stat(filepath.Dir(remoteFile)); err != nil {
		klog.Warningf("Failed to stat dir %s: %v create it", filepath.Dir(remoteFile), err)
		if err := os.MkdirAll(filepath.Dir(remoteFile), mode); err != nil {
			klog.Errorf("Failed to create dir %s: %v", filepath.Dir(remoteFile), err)
			return err
		}
	}
	rf, err := os.Create(remoteFile)
	if err != nil {
		klog.Errorf("Failed to create file %s: %v", remoteFile, err)
		return err
	}
	if _, err := rf.Write(local); err != nil {
		klog.Errorf("Failed to write file %s: %v", remoteFile, err)
		return err
	}
	return rf.Chmod(mode)
}

func (c *localConnector) FetchFile(ctx context.Context, remoteFile string, local io.Writer) error {
	var err error
	file, err := os.Open(remoteFile)
	if err != nil {
		klog.Errorf("Failed to read file %s: %v", remoteFile, err)
		return err
	}
	if _, err := io.Copy(local, file); err != nil {
		klog.Errorf("Failed to copy file %s: %v", remoteFile, err)
		return err
	}
	return nil
}

func (c *localConnector) ExecuteCommand(ctx context.Context, cmd string) ([]byte, error) {
	return c.Cmd.CommandContext(ctx, cmd).CombinedOutput()
}

func (c *localConnector) copyFile(sourcePath, destinationPath string) error {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}
