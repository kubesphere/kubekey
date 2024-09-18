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

	"k8s.io/klog/v2"
	"k8s.io/utils/exec"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

var _ Connector = &localConnector{}
var _ GatherFacts = &localConnector{}

func newLocalConnector(connectorVars map[string]any) *localConnector {
	password, err := variable.StringVar(nil, connectorVars, _const.VariableConnectorPassword)
	if err != nil {
		klog.V(4).InfoS("get connector sudo password failed, execute command without sudo", "error", err)
	}

	return &localConnector{Password: password, Cmd: exec.New()}
}

type localConnector struct {
	Password string
	Cmd      exec.Interface
}

// Init connector. do nothing
func (c *localConnector) Init(context.Context) error {
	return nil
}

// Close connector. do nothing
func (c *localConnector) Close(context.Context) error {
	return nil
}

// PutFile copy src file to dst file. src is the local filename, dst is the local filename.
func (c *localConnector) PutFile(_ context.Context, src []byte, dst string, mode fs.FileMode) error {
	if _, err := os.Stat(filepath.Dir(dst)); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(dst), mode); err != nil {
			klog.V(4).ErrorS(err, "Failed to create local dir", "dst_file", dst)

			return err
		}
	}

	return os.WriteFile(dst, src, mode)
}

// FetchFile copy src file to dst writer. src is the local filename, dst is the local writer.
func (c *localConnector) FetchFile(_ context.Context, src string, dst io.Writer) error {
	var err error
	file, err := os.Open(src)
	if err != nil {
		klog.V(4).ErrorS(err, "Failed to read local file failed", "src_file", src)

		return err
	}

	if _, err := io.Copy(dst, file); err != nil {
		klog.V(4).ErrorS(err, "Failed to copy local file", "src_file", src)

		return err
	}

	return nil
}

// ExecuteCommand in local host
func (c *localConnector) ExecuteCommand(ctx context.Context, cmd string) ([]byte, error) {
	klog.V(5).InfoS("exec local command", "cmd", cmd)
	// find command interpreter in env. default /bin/bash

	command := c.Cmd.CommandContext(ctx, "sudo", "-E", localShell, "-c", "\"", cmd, "\"")
	if c.Password != "" {
		command.SetStdin(bytes.NewBufferString(c.Password + "\n"))
	}

	return command.CombinedOutput()
}

// HostInfo for GatherFacts
func (c *localConnector) HostInfo(ctx context.Context) (map[string]any, error) {
	switch runtime.GOOS {
	case "linux":
		// os information
		osVars := make(map[string]any)
		var osRelease bytes.Buffer
		if err := c.FetchFile(ctx, "/etc/os-release", &osRelease); err != nil {
			return nil, fmt.Errorf("failed to fetch os-release: %w", err)
		}
		osVars[_const.VariableOSRelease] = convertBytesToMap(osRelease.Bytes(), "=")
		kernel, err := c.ExecuteCommand(ctx, "uname -r")
		if err != nil {
			return nil, fmt.Errorf("get kernel version error: %w", err)
		}
		osVars[_const.VariableOSKernelVersion] = string(bytes.TrimSuffix(kernel, []byte("\n")))
		hn, err := c.ExecuteCommand(ctx, "hostname")
		if err != nil {
			return nil, fmt.Errorf("get hostname error: %w", err)
		}
		osVars[_const.VariableOSHostName] = string(bytes.TrimSuffix(hn, []byte("\n")))
		arch, err := c.ExecuteCommand(ctx, "arch")
		if err != nil {
			return nil, fmt.Errorf("get arch error: %w", err)
		}
		osVars[_const.VariableOSArchitecture] = string(bytes.TrimSuffix(arch, []byte("\n")))

		// process information
		procVars := make(map[string]any)
		var cpu bytes.Buffer
		if err := c.FetchFile(ctx, "/proc/cpuinfo", &cpu); err != nil {
			return nil, fmt.Errorf("get cpuinfo error: %w", err)
		}
		procVars[_const.VariableProcessCPU] = convertBytesToSlice(cpu.Bytes(), ":")
		var mem bytes.Buffer
		if err := c.FetchFile(ctx, "/proc/meminfo", &mem); err != nil {
			return nil, fmt.Errorf("get meminfo error: %w", err)
		}
		procVars[_const.VariableProcessMemory] = convertBytesToMap(mem.Bytes(), ":")

		return map[string]any{
			_const.VariableOS:      osVars,
			_const.VariableProcess: procVars,
		}, nil
	default:
		klog.V(4).ErrorS(nil, "Unsupported platform", "platform", runtime.GOOS)

		return make(map[string]any), nil
	}
}
