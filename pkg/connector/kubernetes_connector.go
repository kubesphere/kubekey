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

	"github.com/cockroachdb/errors"
	"k8s.io/klog/v2"
	"k8s.io/utils/exec"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

const kubeconfigRelPath = ".kube/config"

var _ Connector = &kubernetesConnector{}

func newKubernetesConnector(host string, workdir string, connectorVars map[string]any) (*kubernetesConnector, error) {
	kubeconfig, err := variable.StringVar(nil, connectorVars, _const.VariableConnectorKubeconfig)
	if err != nil && host != _const.VariableLocalHost {
		return nil, errors.WithStack(err)
	}

	return &kubernetesConnector{
		workdir:     workdir,
		cmd:         exec.New(),
		clusterName: host,
		kubeconfig:  kubeconfig,
		shell:       defaultSHELL,
	}, nil
}

type kubernetesConnector struct {
	workdir     string
	homedir     string
	clusterName string
	kubeconfig  string
	// shell to execute command
	shell string
	cmd   exec.Interface
}

// Init connector, create home dir in local for each kubernetes.
func (c *kubernetesConnector) Init(_ context.Context) error {
	if c.clusterName == _const.VariableLocalHost && c.kubeconfig == "" {
		klog.V(4).InfoS("kubeconfig is not set, using local kubeconfig")
		// use default kubeconfig. skip
		return nil
	}
	// set home dir for each kubernetes
	c.homedir = filepath.Join(c.workdir, _const.KubernetesDir, c.clusterName)
	if _, err := os.Stat(c.homedir); err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrapf(err, "failed to stat local dir %q for cluster %q", c.homedir, c.clusterName)
		}
		// if dir is not exist, create it.
		if err := os.MkdirAll(c.homedir, os.ModePerm); err != nil {
			return errors.Wrapf(err, "failed to create local dir %q for cluster %q", c.homedir, c.clusterName)
		}
	}
	// create kubeconfig path in home dir
	kubeconfigPath := filepath.Join(c.homedir, kubeconfigRelPath)
	if _, err := os.Stat(kubeconfigPath); err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrapf(err, "failed to stat local path %q for cluster %q", kubeconfigPath, c.clusterName)
		}
		if err := os.MkdirAll(filepath.Dir(kubeconfigPath), os.ModePerm); err != nil {
			return errors.Wrapf(err, "failed to create local path %q for cluster %q", kubeconfigPath, c.clusterName)
		}
	}

	// write kubeconfig to home dir
	if err := os.WriteFile(kubeconfigPath, []byte(c.kubeconfig), os.ModePerm); err != nil {
		return errors.Wrapf(err, "failed to create kubeconfig file for cluster %q", c.clusterName)
	}
	// find command interpreter in env. default /bin/bash
	sl, ok := os.LookupEnv(_const.ENV_SHELL)
	if ok {
		c.shell = sl
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
	dst = filepath.Join(c.homedir, dst)
	if _, err := os.Stat(filepath.Dir(dst)); err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrapf(err, "failed to stat local dir %q for cluster %q", dst, c.clusterName)
		}
		if err := os.MkdirAll(filepath.Dir(dst), mode); err != nil {
			return errors.Wrapf(err, "failed to create local dir %q for cluster %q", dst, c.clusterName)
		}
	}
	if err := os.WriteFile(dst, src, mode); err != nil {
		return errors.Wrapf(err, "failed to write file %q for cluster %q", dst, c.clusterName)
	}

	return nil
}

// FetchFile copy src file to dst writer. src is the local filename, dst is the local writer.
func (c *kubernetesConnector) FetchFile(ctx context.Context, src string, dst io.Writer) error {
	// add "--kubeconfig" to src command
	klog.V(5).InfoS("exec local command", "cmd", src)
	command := c.cmd.CommandContext(ctx, c.shell, "-c", src)
	command.SetDir(c.homedir)
	command.SetEnv([]string{"KUBECONFIG=" + filepath.Join(c.homedir, kubeconfigRelPath)})
	command.SetStdout(dst)
	_, err := command.CombinedOutput()

	return err
}

// ExecuteCommand in a kubernetes cluster
func (c *kubernetesConnector) ExecuteCommand(ctx context.Context, cmd string) ([]byte, error) {
	// add "--kubeconfig" to src command
	klog.V(5).InfoS("exec local command", "cmd", cmd)
	command := c.cmd.CommandContext(ctx, c.shell, "-c", cmd)
	command.SetDir(c.homedir)
	command.SetEnv([]string{"KUBECONFIG=" + filepath.Join(c.homedir, kubeconfigRelPath)})

	return command.CombinedOutput()
}
