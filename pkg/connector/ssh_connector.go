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
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

const (
	defaultSSHPort = 22
	defaultSSHUser = "root"
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

func newSSHConnector(workdir, host string, hostVars map[string]any) *sshConnector {
	// get host in connector variable. if empty, set default host: host_name.
	hostParam, err := variable.StringVar(nil, hostVars, _const.VariableConnector, _const.VariableConnectorHost)
	if err != nil {
		klog.V(4).InfoS("get connector host failed use current hostname", "error", err)
		hostParam = host
	}
	// get port in connector variable. if empty, set default port: 22.
	portParam, err := variable.IntVar(nil, hostVars, _const.VariableConnector, _const.VariableConnectorPort)
	if err != nil {
		klog.V(4).Infof("connector port is empty use: %v", defaultSSHPort)
		portParam = ptr.To(defaultSSHPort)
	}
	// get user in connector variable. if empty, set default user: root.
	userParam, err := variable.StringVar(nil, hostVars, _const.VariableConnector, _const.VariableConnectorUser)
	if err != nil {
		klog.V(4).Infof("connector user is empty use: %s", defaultSSHUser)
		userParam = defaultSSHUser
	}
	// get password in connector variable. if empty, should connector by private key.
	passwdParam, err := variable.StringVar(nil, hostVars, _const.VariableConnector, _const.VariableConnectorPassword)
	if err != nil {
		klog.V(4).InfoS("connector password is empty use public key")
	}
	// get private key path in connector variable. if empty, set default path: /root/.ssh/id_rsa.
	keyParam, err := variable.StringVar(nil, hostVars, _const.VariableConnector, _const.VariableConnectorPrivateKey)
	if err != nil {
		klog.V(4).InfoS("ssh private key path is empty, using default", "path", defaultSSHPrivateKey)
		keyParam = defaultSSHPrivateKey
	}
	keycontentParam, err := variable.StringVar(nil, hostVars, _const.VariableConnector, _const.VariableConnectorPrivateKeyContent)
	if err != nil {
		klog.V(4).InfoS("ssh private key content is empty")
		// Leave keycontentParam as empty string - no default needed
	}
	cacheType, _ := variable.StringVar(nil, hostVars, _const.VariableGatherFactsCache)
	connector := &sshConnector{
		Host:              hostParam,
		Port:              *portParam,
		User:              userParam,
		Password:          passwdParam,
		PrivateKey:        keyParam,
		PrivateKeyContent: keycontentParam,
	}

	// Initialize the cacheGatherFact with a function that will call getHostInfoFromRemote
	connector.gatherFacts = newCacheGatherFact(_const.VariableLocalHost, cacheType, workdir, connector.getHostInfo)

	return connector
}

type sshConnector struct {
	Host              string
	Port              int
	User              string
	Password          string
	PrivateKey        string
	PrivateKeyContent string

	client *ssh.Client
	// shell to execute command
	shell string

	gatherFacts *cacheGatherFact

	mu sync.Mutex
}

// Init establishes SSH connection with the following authentication priority:
// - Password: Always included if set (independent)
// - Key auth (exclusive priority):
//   1. PrivateKeyContent - if set, use ONLY this
//   2. PrivateKey path - if set and content not set, use ONLY this
//   3. Default ~/.ssh/id_rsa - fallback if neither is set
func (c *sshConnector) Init(context.Context) error {
	if c.Host == "" {
		return errors.New("host is not set")
	}

	var auth []ssh.AuthMethod

	// Password: Independent, always add if provided
	if c.Password != "" {
		auth = append(auth, ssh.Password(c.Password))
	}

	// Key auth: EXCLUSIVE priority
	if c.PrivateKeyContent != "" {
		// Priority 1: Use ONLY PrivateKeyContent
		privateKey, err := ssh.ParsePrivateKey([]byte(c.PrivateKeyContent))
		if err != nil {
			return errors.Wrapf(err, "failed to parse private key content")
		}
		auth = append(auth, ssh.PublicKeys(privateKey))
		klog.V(4).InfoS("using private key content for authentication")
	} else if c.PrivateKey != "" {
		// Priority 2: Use ONLY PrivateKey path (if content not set)
		if _, err := os.Stat(c.PrivateKey); err != nil {
			return errors.Wrapf(err, "private key file not found: %s", c.PrivateKey)
		}
		key, err := os.ReadFile(c.PrivateKey)
		if err != nil {
			return errors.Wrapf(err, "failed to read private key %q", c.PrivateKey)
		}
		privateKey, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return errors.Wrapf(err, "failed to parse private key %q", c.PrivateKey)
		}
		auth = append(auth, ssh.PublicKeys(privateKey))
		klog.V(4).InfoS("using private key file for authentication", "path", c.PrivateKey)
	}
	// Note: If neither content nor path is set, c.PrivateKey already has default from newSSHConnector line 89

	// Validate we have at least one auth method
	if len(auth) == 0 {
		return errors.New("no authentication method available: provide password, private_key_content, or private_key")
	}

	sshClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", c.Host, c.Port), &ssh.ClientConfig{
		User:            c.User,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	})
	if err != nil {
		return errors.Wrapf(err, "failed to dial %s:%d ssh server", c.Host, c.Port)
	}
	c.client = sshClient

	// get shell from env
	session, err := sshClient.NewSession()
	if err != nil {
		return errors.Wrap(err, "failed to create session")
	}
	defer session.Close()

	output, err := session.CombinedOutput("echo $SHELL")
	if err != nil {
		return errors.Wrap(err, "failed to env command")
	}

	if strings.TrimSpace(string(output)) != "" {
		c.shell = strings.TrimSpace(string(output))
	} else {
		c.shell = "/bin/bash"
	}

	return nil
}

// Close connector
func (c *sshConnector) Close(context.Context) error {
	return c.client.Close()
}

func (c *sshConnector) session() (*ssh.Session, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client == nil {
		return nil, errors.New("connection closed")
	}

	sess, err := c.client.NewSession()
	if err != nil {
		return nil, err
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	err = sess.RequestPty("xterm", 100, 50, modes)
	if err != nil {
		return nil, err
	}

	return sess, nil
}

// PutFile to remote node. src is the file bytes. dst is the remote filename
func (c *sshConnector) PutFile(_ context.Context, src []byte, dst string, mode fs.FileMode) error {
	// create sftp client
	sftpClient, err := sftp.NewClient(c.client)
	if err != nil {
		return errors.Wrap(err, "failed to create sftp client")
	}
	defer sftpClient.Close()
	// create remote file
	if _, err := sftpClient.Stat(filepath.Dir(dst)); err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrapf(err, "failed to stat dir %q", dst)
		}
		if err := sftpClient.MkdirAll(filepath.Dir(dst)); err != nil {
			return errors.Wrapf(err, "failed to create remote dir %q", dst)
		}
	}

	rf, err := sftpClient.Create(dst)
	if err != nil {
		return errors.Wrapf(err, "failed to create remote file %q", dst)
	}
	defer rf.Close()
	if _, err = rf.Write(src); err != nil {
		return errors.Wrapf(err, "failed to write content to remote file %q", dst)
	}
	if err := rf.Chmod(mode); err != nil {
		return errors.Wrapf(err, "failed to chmod remote file %q", dst)
	}

	return nil
}

// FetchFile from remote node. src is the remote filename, dst is the local writer.
func (c *sshConnector) FetchFile(_ context.Context, src string, dst io.Writer) error {
	// create sftp client
	sftpClient, err := sftp.NewClient(c.client)
	if err != nil {
		return errors.Wrap(err, "failed to create sftp client")
	}
	defer sftpClient.Close()

	rf, err := sftpClient.Open(src)
	if err != nil {
		return errors.Wrapf(err, "failed to open remote file %q", src)
	}
	defer rf.Close()

	if _, err := io.Copy(dst, rf); err != nil {
		return errors.Wrapf(err, "failed to copy file %q", src)
	}

	return nil
}

// ExecuteCommand exec cmd with sudo
func (c *sshConnector) ExecuteCommand(_ context.Context, cmd string) ([]byte, []byte, error) {
	session, err := c.session()
	if err != nil {
		return nil, nil, err
	}
	defer session.Close()

	cmd = SudoPrefix(c.shell, cmd)

	in, err := session.StdinPipe()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get stdin pipe")
	}

	out, err := session.StdoutPipe()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get stdout pipe")
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get stderr pipe")
	}

	if err = session.Start(cmd); err != nil {
		return nil, nil, errors.Wrap(err, "failed to start session")
	}
	var (
		output []byte
		line   = ""
		r      = bufio.NewReader(out)
	)

	for {
		b, err := r.ReadByte()
		if err != nil {
			break
		}

		output = append(output, b)

		if b == byte('\n') {
			line = ""
			continue
		}

		line += string(b)

		if (strings.HasPrefix(line, "[sudo] password for ") || strings.HasPrefix(line, "Password")) && strings.HasSuffix(line, ": ") {
			_, err = in.Write([]byte(c.Password + "\n"))
			if err != nil {
				break
			}
		}
	}

	outStr := strings.TrimPrefix(string(output), fmt.Sprintf("[sudo] password for %s:", c.User))
	err = session.Wait()
	var stderrBuffer bytes.Buffer
	_, _ = io.Copy(&stderrBuffer, stderr)
	outStr = strings.TrimSpace(outStr)
	stderrData := stderrBuffer.Bytes()
	if err != nil {
		return []byte(outStr), nil, errors.Wrap(err, strings.TrimSpace(string(stderrData)))
	}
	return []byte(outStr), stderrData, nil
}

// SudoPrefix returns the prefix for sudo commands.
func SudoPrefix(shell, cmd string) string {
	return fmt.Sprintf("TERM=dumb; export LANG=C.UTF-8;sudo -E %s << 'KUBEKEY_EOF'\n%s\nKUBEKEY_EOF", shell, cmd)
}

// HostInfo from gatherFacts cache
func (c *sshConnector) HostInfo(ctx context.Context) (map[string]any, error) {
	return c.gatherFacts.HostInfo(ctx)
}

// getHostInfo from remote
func (c *sshConnector) getHostInfo(ctx context.Context) (map[string]any, error) {
	// os information
	osVars := make(map[string]any)
	var osRelease bytes.Buffer
	if err := c.FetchFile(ctx, "/etc/os-release", &osRelease); err != nil {
		return nil, err
	}
	osVars[_const.VariableOSRelease] = convertBytesToMap(osRelease.Bytes(), "=")
	kernel, kernelStderr, err := c.ExecuteCommand(ctx, "uname -r")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get kernel: %v, stderr: %q", err, string(kernelStderr))
	}
	osVars[_const.VariableOSKernelVersion] = string(bytes.TrimSpace(kernel))

	hn, hnStderr, err := c.ExecuteCommand(ctx, "hostname")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get hostname: %v, stderr: %q", err, string(hnStderr))
	}
	osVars[_const.VariableOSHostName] = string(bytes.TrimSpace(hn))

	arch, archStderr, err := c.ExecuteCommand(ctx, "arch")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get arch: %v, stderr: %q", err, string(archStderr))
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

	// persistence the hostInfo

	return map[string]any{
		_const.VariableOS:      osVars,
		_const.VariableProcess: procVars,
	}, nil
}
