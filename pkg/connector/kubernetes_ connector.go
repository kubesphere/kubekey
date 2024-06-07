/*
Copyright 2024 The KubeSphere Authors.

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

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
)

type kubernetesConnector struct {
	clusterName    string
	kubeconfig     string
	kubeconfigPath string
	Cmd            exec.Interface
}

func (c *kubernetesConnector) Init(ctx context.Context) error {
	c.kubeconfigPath = filepath.Join(_const.GetWorkDir(), "kubernetes", c.clusterName, "kubeconfig")
	if _, err := os.Stat(c.kubeconfigPath); err == nil || !os.IsNotExist(err) {
		klog.V(4).InfoS("kubeconfig file already exists", "cluster", c.clusterName)
		return err
	}
	// store kubeconfig to work dir
	if _, err := os.Stat(filepath.Dir(c.kubeconfigPath)); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(c.kubeconfigPath), 0644); err != nil {
			klog.V(4).ErrorS(err, "Failed to create local dir", "cluster", c.clusterName)
			return err
		}
	}
	if err := os.WriteFile(c.kubeconfigPath, []byte(c.kubeconfig), 0644); err != nil {
		klog.V(4).ErrorS(err, "Failed to create kubeconfig file", "cluster", c.clusterName)
		return err
	}
	return nil
}

func (c *kubernetesConnector) Close(ctx context.Context) error {
	return nil
}

// PutFile copy src file to dst file. src is the local filename, dst is the local filename.
// Typically, the configuration file for each cluster may be different,
// and it may be necessary to keep them in separate directories locally.
func (c *kubernetesConnector) PutFile(ctx context.Context, src []byte, dst string, mode fs.FileMode) error {
	if _, err := os.Stat(filepath.Dir(dst)); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(dst), mode); err != nil {
			klog.V(4).ErrorS(err, "Failed to create local dir", "dst_file", dst)
			return err
		}
	}
	return os.WriteFile(dst, src, mode)
}

// FetchFile copy src file to dst writer. src is the local filename, dst is the local writer.
func (c *kubernetesConnector) FetchFile(ctx context.Context, src string, dst io.Writer) error {
	// add "--kubeconfig" to src command
	command := c.Cmd.CommandContext(ctx, "/bin/sh", "-c", src)
	command.SetEnv([]string{"KUBECONFIG=" + c.kubeconfigPath})
	command.SetStdout(dst)
	_, err := command.CombinedOutput()
	return err
}

func (c *kubernetesConnector) ExecuteCommand(ctx context.Context, cmd string) ([]byte, error) {
	// add "--kubeconfig" to src command
	klog.V(4).InfoS("exec local command", "cmd", cmd)
	command := c.Cmd.CommandContext(ctx, "/bin/sh", "-c", cmd)
	command.SetEnv([]string{"KUBECONFIG=" + c.kubeconfigPath})
	return command.CombinedOutput()
}
