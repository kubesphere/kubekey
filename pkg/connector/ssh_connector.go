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
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

const (
	defaultSSHPort  = 22
	defaultSSHUser  = "root"
	defaultSSHSHELL = "/bin/bash"
)

var defaultSSHPrivateKey string

func init() {
	if currentUser, err := user.Current(); err == nil {
		defaultSSHPrivateKey = filepath.Join(currentUser.HomeDir, ".ssh/id_rsa")
	} else {
		defaultSSHPrivateKey = filepath.Join(defaultSSHUser, ".ssh/id_rsa")
	}
}

var _ Connector = &sshConnector{}
var _ GatherFacts = &sshConnector{}

func newSSHConnector(host string, connectorVars map[string]any) *sshConnector {
	// get host in connector variable. if empty, set default host: host_name.
	hostParam, err := variable.StringVar(nil, connectorVars, _const.VariableConnectorHost)
	if err != nil {
		klog.V(4).InfoS("get connector host failed use current hostname", "error", err)
		hostParam = host
	}
	// get port in connector variable. if empty, set default port: 22.
	portParam, err := variable.IntVar(nil, connectorVars, _const.VariableConnectorPort)
	if err != nil {
		klog.V(4).Infof("connector port is empty use: %v", defaultSSHPort)
		portParam = ptr.To(defaultSSHPort)
	}
	// get user in connector variable. if empty, set default user: root.
	userParam, err := variable.StringVar(nil, connectorVars, _const.VariableConnectorUser)
	if err != nil {
		klog.V(4).Infof("connector user is empty use: %s", defaultSSHUser)
		userParam = defaultSSHUser
	}
	// get password in connector variable. if empty, should connector by private key.
	passwdParam, err := variable.StringVar(nil, connectorVars, _const.VariableConnectorPassword)
	if err != nil {
		klog.V(4).InfoS("connector password is empty use public key")
	}
	// get private key path in connector variable. if empty, set default path: /root/.ssh/id_rsa.
	keyParam, err := variable.StringVar(nil, connectorVars, _const.VariableConnectorPrivateKey)
	if err != nil {
		klog.V(4).Infof("ssh public key is empty, use: %s", defaultSSHPrivateKey)
		keyParam = defaultSSHPrivateKey
	}

	return &sshConnector{
		Host:       hostParam,
		Port:       *portParam,
		User:       userParam,
		Password:   passwdParam,
		PrivateKey: keyParam,
		shell:      defaultSSHSHELL,
	}
}

type sshConnector struct {
	Host       string
	Port       int
	User       string
	Password   string
	PrivateKey string

	client *ssh.Client
	// shell to execute command
	shell string
}

// Init connector, get ssh.Client
func (c *sshConnector) Init(context.Context) error {
	if c.Host == "" {
		return errors.New("host is not set")
	}

	var auth []ssh.AuthMethod
	if c.Password != "" {
		auth = append(auth, ssh.Password(c.Password))
	}
	if _, err := os.Stat(c.PrivateKey); err == nil {
		key, err := os.ReadFile(c.PrivateKey)
		if err != nil {
			return fmt.Errorf("read private key error: %w", err)
		}
		privateKey, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return fmt.Errorf("parse private key error: %w", err)
		}
		auth = append(auth, ssh.PublicKeys(privateKey))
	}

	sshClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:%v", c.Host, c.Port), &ssh.ClientConfig{
		User:            c.User,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	})
	if err != nil {
		klog.V(4).ErrorS(err, "Dial ssh server failed", "host", c.Host, "port", c.Port)

		return err
	}
	c.client = sshClient

	// get shell from env
	session, err := sshClient.NewSession()
	if err != nil {
		return fmt.Errorf("create session error: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput("echo $SHELL")
	if err != nil {
		return fmt.Errorf("env command error: %w", err)
	}

	if strings.TrimSuffix(string(output), "\n") != "" {
		c.shell = strings.TrimSuffix(string(output), "\n")
	}

	return nil
}

// Close connector
func (c *sshConnector) Close(context.Context) error {
	return c.client.Close()
}

// PutFile to remote node. src is the file bytes. dst is the remote filename
func (c *sshConnector) PutFile(_ context.Context, src []byte, dst string, mode fs.FileMode) error {
	// create sftp client
	sftpClient, err := sftp.NewClient(c.client)
	if err != nil {
		klog.V(4).ErrorS(err, "Failed to create sftp client")

		return err
	}
	defer sftpClient.Close()
	// create remote file
	if _, err := sftpClient.Stat(filepath.Dir(dst)); err != nil && os.IsNotExist(err) {
		if err := sftpClient.MkdirAll(filepath.Dir(dst)); err != nil {
			klog.V(4).ErrorS(err, "Failed to create remote dir", "remote_file", dst)

			return err
		}
	}

	rf, err := sftpClient.Create(dst)
	if err != nil {
		klog.V(4).ErrorS(err, "Failed to  create remote file", "remote_file", dst)

		return err
	}
	defer rf.Close()

	if _, err = rf.Write(src); err != nil {
		klog.V(4).ErrorS(err, "Failed to write content to remote file", "remote_file", dst)

		return err
	}

	return rf.Chmod(mode)
}

// FetchFile from remote node. src is the remote filename, dst is the local writer.
func (c *sshConnector) FetchFile(_ context.Context, src string, dst io.Writer) error {
	// create sftp client
	sftpClient, err := sftp.NewClient(c.client)
	if err != nil {
		klog.V(4).ErrorS(err, "Failed to create sftp client", "remote_file", src)

		return err
	}
	defer sftpClient.Close()

	rf, err := sftpClient.Open(src)
	if err != nil {
		klog.V(4).ErrorS(err, "Failed to open file", "remote_file", src)

		return err
	}
	defer rf.Close()

	if _, err := io.Copy(dst, rf); err != nil {
		klog.V(4).ErrorS(err, "Failed to copy file", "remote_file", src)

		return err
	}

	return nil
}

// ExecuteCommand in remote host
func (c *sshConnector) ExecuteCommand(_ context.Context, cmd string) ([]byte, error) {
	klog.V(5).InfoS("exec ssh command", "cmd", cmd, "host", c.Host)
	// create ssh session
	session, err := c.client.NewSession()
	if err != nil {
		klog.V(4).ErrorS(err, "Failed to create ssh session")

		return nil, err
	}
	defer session.Close()

	cmd = fmt.Sprintf("sudo -E %s -c \"%q\"", c.shell, cmd)
	// get pipe from session
	stdin, _ := session.StdinPipe()
	stdout, _ := session.StdoutPipe()
	stderr, _ := session.StderrPipe()
	// Request a pseudo-terminal (required for sudo password input)
	if err := session.RequestPty("xterm", 80, 40, ssh.TerminalModes{}); err != nil {
		return nil, err
	}
	// Start the remote command
	if err := session.Start(cmd); err != nil {
		return nil, err
	}
	if c.Password != "" {
		// Write sudo password to the standard input
		if _, err := io.WriteString(stdin, c.Password+"\n"); err != nil {
			return nil, err
		}
	}
	// Read the command output
	output := make([]byte, 0)
	stdoutData, _ := io.ReadAll(stdout)
	stderrData, _ := io.ReadAll(stderr)
	output = append(output, stdoutData...)
	output = append(output, stderrData...)
	// Wait for the command to complete
	if err := session.Wait(); err != nil {
		return nil, err
	}

	return output, nil
}

// HostInfo for GatherFacts
func (c *sshConnector) HostInfo(ctx context.Context) (map[string]any, error) {
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
}
