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

const kubeconfigRelPath = ".kube/config"

var _ Connector = &kubernetesConnector{}

type kubernetesConnector struct {
	clusterName string
	kubeconfig  string
	homeDir     string
	Cmd         exec.Interface
}

// Init connector, create home dir in local for each kubernetes.
func (c *kubernetesConnector) Init(_ context.Context) error {
	if c.clusterName == _const.VariableLocalHost && c.kubeconfig == "" {
		klog.V(4).InfoS("kubeconfig is not set, using local kubeconfig")
		// use default kubeconfig. skip
		return nil
	}
	// set home dir for each kubernetes
	c.homeDir = filepath.Join(_const.GetWorkDir(), _const.KubernetesDir, c.clusterName)
	if _, err := os.Stat(c.homeDir); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(c.homeDir, os.ModePerm); err != nil {
			klog.V(4).ErrorS(err, "Failed to create local dir", "cluster", c.clusterName)
			// if dir is not exist, create it.
			return err
		}
	}
	// create kubeconfig path in home dir
	kubeconfigPath := filepath.Join(c.homeDir, kubeconfigRelPath)
	if _, err := os.Stat(kubeconfigPath); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(kubeconfigPath), os.ModePerm); err != nil {
			klog.V(4).ErrorS(err, "Failed to create local dir", "cluster", c.clusterName)

			return err
		}
	}
	// write kubeconfig to home dir
	if err := os.WriteFile(kubeconfigPath, []byte(c.kubeconfig), os.ModePerm); err != nil {
		klog.V(4).ErrorS(err, "Failed to create kubeconfig file", "cluster", c.clusterName)

		return err
	}

	return nil
}

// Close connector, do nothing
func (c *kubernetesConnector) Close(_ context.Context) error {
	return nil
}

// PutFile copy src file to dst file. src is the local filename, dst is the local filename.
// Typically, the configuration file for each cluster may be different,
// and it may be necessary to keep them in separate directories locally.
func (c *kubernetesConnector) PutFile(_ context.Context, src []byte, dst string, mode fs.FileMode) error {
	dst = filepath.Join(c.homeDir, dst)
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
	klog.V(5).InfoS("exec local command", "cmd", src)
	command := c.Cmd.CommandContext(ctx, "/bin/sh", "-c", src)
	command.SetDir(c.homeDir)
	command.SetEnv([]string{"KUBECONFIG=" + filepath.Join(c.homeDir, kubeconfigRelPath)})
	command.SetStdout(dst)
	_, err := command.CombinedOutput()

	return err
}

// ExecuteCommand in a kubernetes cluster
func (c *kubernetesConnector) ExecuteCommand(ctx context.Context, cmd string) ([]byte, error) {
	// add "--kubeconfig" to src command
	klog.V(5).InfoS("exec local command", "cmd", cmd)
	command := c.Cmd.CommandContext(ctx, "/bin/sh", "-c", cmd)
	command.SetDir(c.homeDir)
	command.SetEnv([]string{"KUBECONFIG=" + filepath.Join(c.homeDir, kubeconfigRelPath)})

	return command.CombinedOutput()
}
