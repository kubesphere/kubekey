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
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cockroachdb/errors"
	"k8s.io/klog/v2"
	"k8s.io/utils/exec"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

var _ Connector = &localConnector{}
var _ GatherFacts = &localConnector{}

func newLocalConnector(connectorVars map[string]any) *localConnector {
	password, err := variable.StringVar(nil, connectorVars, _const.VariableConnectorPassword)
	if err != nil { // password is not necessary when execute with root user.
		klog.V(4).Info("Warning: Failed to obtain local connector password when executing command with sudo. Please ensure the 'kk' process is run by a root-privileged user.")
	}

	return &localConnector{Password: password, Cmd: exec.New(), shell: defaultSHELL}
}

type localConnector struct {
	Password string
	Cmd      exec.Interface
	// shell to execute command
	shell string
}

// Init initializes the local connector. This method does nothing for localConnector.
func (c *localConnector) Init(context.Context) error {
	// find command interpreter in env. default /bin/bash
	sl, ok := os.LookupEnv(_const.ENV_SHELL)
	if ok {
		c.shell = sl
	}

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
		if err := os.MkdirAll(filepath.Dir(dst), mode); err != nil {
			return errors.Wrapf(err, "failed to create local dir %q", dst)
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
func (c *localConnector) ExecuteCommand(ctx context.Context, cmd string) ([]byte, error) {
	klog.V(5).InfoS("exec local command", "cmd", cmd)

	// For consistency, use HERE document approach for all commands
	execCmd := fmt.Sprintf("%s << 'KUBEKEY_EOF'\n%s\nKUBEKEY_EOF\n", c.shell, cmd)

	// Check if command requires sudo (we don't always need sudo for local execution)
	if strings.Contains(cmd, "sudo ") || os.Geteuid() != 0 {
		// Command explicitly uses sudo or we're not running as root
		// First try running without sudo to see if it works
		regularCmd := c.Cmd.CommandContext(ctx, "bash", "-c", execCmd)
		regularOutput, regularErr := regularCmd.CombinedOutput()

		// If the command succeeds or doesn't indicate permission issues, return its results
		if regularErr == nil || !strings.Contains(string(regularOutput), "permission denied") {
			return regularOutput, regularErr
		}

		// Command needs sudo, prepare sudo command
		if c.Password != "" {
			// Use a temporary file to avoid showing password in process list
			pwFile, err := os.CreateTemp("", "kubekey-pw-*")
			if err != nil {
				return nil, errors.Wrap(err, "failed to create temp file for sudo password")
			}
			pwPath := pwFile.Name()
			defer os.Remove(pwPath)

			if _, err := pwFile.WriteString(c.Password); err != nil {
				pwFile.Close()
				return nil, errors.Wrap(err, "failed to write sudo password to temp file")
			}
			pwFile.Close()

			// Use sudo -S to read password from stdin
			sudoCmd := fmt.Sprintf("sudo -S %s", execCmd)
			command := c.Cmd.CommandContext(ctx, "bash", "-c", fmt.Sprintf("cat %s | %s", pwPath, sudoCmd))
			output, err := command.CombinedOutput()

			// Filter out password prompt from output
			output = bytes.Replace(output, []byte("Password:"), []byte(""), -1)
			output = bytes.Replace(output, []byte("[sudo] password for "+os.Getenv("USER")+":"), []byte(""), -1)

			return output, errors.Wrap(err, "command execution failed")
		} else {
			// Try sudo without password (relies on NOPASSWD in sudoers)
			sudoCmd := fmt.Sprintf("sudo %s", execCmd)
			command := c.Cmd.CommandContext(ctx, "bash", "-c", sudoCmd)
			output, err := command.CombinedOutput()
			return output, errors.Wrap(err, "command execution failed")
		}
	} else {
		// No need for sudo, run command directly
		command := c.Cmd.CommandContext(ctx, "bash", "-c", execCmd)
		output, err := command.CombinedOutput()
		return output, errors.Wrap(err, "command execution failed")
	}
}

// HostInfo gathers and returns host information for the local host.
func (c *localConnector) HostInfo(ctx context.Context) (map[string]any, error) {
	switch runtime.GOOS {
	case "linux":
		// os information
		osVars := make(map[string]any)
		var osRelease bytes.Buffer
		if err := c.FetchFile(ctx, "/etc/os-release", &osRelease); err != nil {
			return nil, err
		}
		osVars[_const.VariableOSRelease] = convertBytesToMap(osRelease.Bytes(), "=")
		kernel, err := c.ExecuteCommand(ctx, "uname -r")
		if err != nil {
			return nil, err
		}
		osVars[_const.VariableOSKernelVersion] = string(bytes.TrimSpace(kernel))
		hn, err := c.ExecuteCommand(ctx, "hostname")
		if err != nil {
			return nil, err
		}
		osVars[_const.VariableOSHostName] = string(bytes.TrimSpace(hn))
		arch, err := c.ExecuteCommand(ctx, "arch")
		if err != nil {
			return nil, err
		}
		osVars[_const.VariableOSArchitecture] = string(bytes.TrimSpace(arch))

		// process information
		procVars := make(map[string]any)
		var cpu bytes.Buffer
		if err := c.FetchFile(ctx, "/proc/cpuinfo", &cpu); err != nil {
			return nil, err
		}
		procVars[_const.VariableProcessCPU] = convertBytesToSlice(cpu.Bytes(), ":")
		var mem bytes.Buffer
		if err := c.FetchFile(ctx, "/proc/meminfo", &mem); err != nil {
			return nil, err
		}
		procVars[_const.VariableProcessMemory] = convertBytesToMap(mem.Bytes(), ":")

		return map[string]any{
			_const.VariableOS:      osVars,
			_const.VariableProcess: procVars,
		}, nil
	default:
		klog.V(4).ErrorS(nil, "Unsupported platform", "platform", runtime.GOOS)
	}

	return make(map[string]any), nil
}
