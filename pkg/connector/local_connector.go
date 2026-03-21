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
	"bytes"
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"text/template"

	"github.com/cockroachdb/errors"
	"k8s.io/klog/v2"
	"k8s.io/utils/exec"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

var _ Connector = &localConnector{}
var _ GatherFacts = &localConnector{}

func newLocalConnector(tpl *template.Template, workdir string, hostVars map[string]any) *localConnector {
	password, err := variable.StringVar(tpl, nil, hostVars, _const.VariableConnector, _const.VariableConnectorPassword)
	if err != nil { // password is not necessary when execute with root user.
		klog.V(4).Info("Warning: Failed to obtain local connector password when executing command with sudo. Please ensure the 'kk' process is run by a root-privileged user.")
	}
	cacheType, _ := variable.StringVar(tpl, nil, hostVars, _const.VariableGatherFactsCache)
	connector := &localConnector{
		Password: password,
		Cmd:      exec.New(),
	}
	// Initialize the cacheGatherFact with a function that will call getHostInfoFromRemote
	connector.gatherFacts = newCacheGatherFact(_const.VariableLocalHost, cacheType, workdir, connector.getHostInfo)

	return connector
}

type localConnector struct {
	Password string
	Cmd      exec.Interface
	// shell to execute command
	shell string

	gatherFacts *cacheGatherFact
}

// Init initializes the local connector. This method does nothing for localConnector.
func (c *localConnector) Init(context.Context) error {
	// find command interpreter in env. default /bin/bash
	c.shell = _const.Getenv(_const.Shell)

	return nil
}

// Close closes the local connector. This method does nothing for localConnector.
func (c *localConnector) Close(context.Context) error {
	return nil
}

// PutFile copies the src file to the dst file. src is the local filename, dst is the local filename.
func (c *localConnector) PutFile(_ context.Context, src []byte, dst string, mode fs.FileMode) error {
	if _, err := os.Stat(filepath.Dir(dst)); err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrapf(err, "failed to stat local dir %q", dst)
		}
		if err := os.MkdirAll(filepath.Dir(dst), _const.PermDirPublic); err != nil {
			return errors.Wrapf(err, "failed to create local dir of path %q", dst)
		}
	}
	if err := os.WriteFile(dst, src, mode); err != nil {
		return errors.Wrapf(err, "failed to write file %q", dst)
	}

	return nil
}

// FetchFile copies the src file to the dst writer. src is the local filename, dst is the local writer.
func (c *localConnector) FetchFile(_ context.Context, src string, dst io.Writer) error {
	file, err := os.Open(src)
	if err != nil {
		return errors.Wrapf(err, "failed to open local file %q", src)
	}
	if _, err := io.Copy(dst, file); err != nil {
		return errors.Wrapf(err, "failed to copy local file %q", src)
	}

	return nil
}

// ExecuteCommand executes a command on the local host.
func (c *localConnector) ExecuteCommand(ctx context.Context, cmd string) ([]byte, []byte, error) {
	klog.V(5).InfoS("exec local command", "cmd", cmd)
	// in
	command := c.Cmd.CommandContext(ctx, "sudo", "-SE", c.shell, "-c", cmd)
	if c.Password != "" {
		command.SetStdin(bytes.NewBufferString(c.Password + "\n"))
	}
	// out
	var stdoutBuf, stderrBuf bytes.Buffer
	command.SetStdout(&stdoutBuf)
	command.SetStderr(&stderrBuf)
	err := command.Run()
	stdout := stdoutBuf.Bytes()
	stderr := stderrBuf.Bytes()
	if c.Password != "" {
		// Filter out the "Password:" prompt from the output
		stdout = bytes.ReplaceAll(stdout, []byte("Password:"), []byte(""))
		stderr = bytes.ReplaceAll(stderr, []byte("Password:"), []byte(""))
	}

	return stdout, stderr, err
}

// HostInfo from gatherFacts cache
func (c *localConnector) HostInfo(ctx context.Context) (map[string]any, error) {
	return c.gatherFacts.HostInfo(ctx)
}

// getHostInfo from remote
func (c *localConnector) getHostInfo(ctx context.Context) (map[string]any, error) {
	switch runtime.GOOS {
	case "linux":
		return parseHostInfo(c, ctx)
	default:
		klog.V(4).ErrorS(nil, "Unsupported platform", "platform", runtime.GOOS)
	}

	return make(map[string]any), nil
}
