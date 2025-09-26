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
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

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
		klog.V(4).Infof("ssh public key is empty, use: %s", defaultSSHPrivateKey)
		keyParam = defaultSSHPrivateKey
	}
	keycontentParam, err := variable.StringVar(nil, hostVars, _const.VariableConnector, _const.VariableConnectorPrivateKeyContent)
	if err != nil {
		klog.V(4).Infof("ssh public key is empty, use: %s", defaultSSHPrivateKey)
		keyParam = defaultSSHPrivateKey
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
	if c.PrivateKeyContent != "" {
		privateKey, err := ssh.ParsePrivateKey([]byte(c.PrivateKeyContent))
		if err != nil {
			return errors.Wrapf(err, "failed to parse private key %q", c.PrivateKey)
		}
		auth = append(auth, ssh.PublicKeys(privateKey))
	}
	if _, err := os.Stat(c.PrivateKey); err == nil {
		key, err := os.ReadFile(c.PrivateKey)
		if err != nil {
			return errors.Wrapf(err, "failed to read private key %q", c.PrivateKey)
		}
		privateKey, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return errors.Wrapf(err, "failed to parse private key %q", c.PrivateKey)
		}
		auth = append(auth, ssh.PublicKeys(privateKey))
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

// ExecuteCommand in remote host
func (c *sshConnector) ExecuteCommand(_ context.Context, cmd string) ([]byte, []byte, error) {
	cmd = fmt.Sprintf("sudo -E %s << 'KUBEKEY_EOF'\n%s\nKUBEKEY_EOF\n", c.shell, cmd)
	klog.V(5).InfoS("exec ssh command", "cmd", cmd, "host", c.Host)
	// create ssh session
	session, err := c.client.NewSession()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create ssh session")
	}
	defer session.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	err = session.RequestPty("xterm", 100, 50, modes)
	if err != nil {
		return nil, nil, err
	}

	// get pipe from session
	stdin, err := session.StdinPipe()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get stdin pipe")
	}
	stdout, err := session.StdoutPipe()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get stdout pipe")
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get stderr pipe")
	}
	// Start the remote command
	if err := session.Start(cmd); err != nil {
		return nil, nil, errors.Wrap(err, "failed to start session")
	}
	if c.Password != "" {
		if _, err := stdin.Write([]byte(c.Password + "\n")); err != nil {
			return nil, nil, errors.Wrap(err, "failed to write password")
		}
	}
	if err := stdin.Close(); err != nil {
		return nil, nil, errors.Wrap(err, "failed to close stdin pipe")
	}

	// Create buffers to store stdout and stderr output
	var stdoutBuf, stderrBuf bytes.Buffer

	// When reading large amounts of data from stdout/stderr, the pipe buffer can fill up
	// and block the remote command from completing if we don't read from it continuously.
	// To prevent this deadlock scenario, we need to read stdout/stderr asynchronously
	// in separate goroutines while the command is running.
	// Create channels to signal when copying is complete
	stdoutDone := make(chan error, 1)
	stderrDone := make(chan error, 1)

	// Copy stdout and stderr concurrently to prevent pipe buffer from filling
	go func() {
		_, err := io.Copy(&stdoutBuf, stdout)
		stdoutDone <- err
	}()
	go func() {
		_, err := io.Copy(&stderrBuf, stderr)
		stderrDone <- err
	}()

	// Wait for command to complete
	err = session.Wait()

	// Wait for stdout and stderr copying to finish to ensure we've captured all output
	<-stdoutDone
	<-stderrDone

	outDataStr := strings.TrimSpace(CleanCommandOutput(stdoutBuf.String(), cmd))
	if (strings.HasPrefix(outDataStr, "[sudo] password for ") || strings.HasPrefix(outDataStr, "Password")) && strings.Contains(outDataStr, ":") {
		outDataSlice := strings.Split(outDataStr, ":")
		if len(outDataSlice) > 1 {
			outDataStr = strings.Join(outDataSlice[1:], ":")
		}
	}

	return []byte(strings.TrimSpace(outDataStr)), stderrBuf.Bytes(), errors.Wrap(err, "failed to execute ssh command")
}

// CleanOutput clean cmd output data
func CleanOutput(output string) string {
	if output == "" {
		return output
	}

	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	cleaned := ansiRegex.ReplaceAllString(output, "")

	decPrivateModeRegex := regexp.MustCompile(`\x1b\[\?[0-9;]*[hl]`)
	cleaned = decPrivateModeRegex.ReplaceAllString(cleaned, "")

	otherEscapeRegex := regexp.MustCompile(`\x1b[\(\)][BC0-9]|\x1b\][0-9][;:].*?\x07|\x1b[=>?]`)
	cleaned = otherEscapeRegex.ReplaceAllString(cleaned, "")

	titleRegex := regexp.MustCompile(`\x1b\]0;.*?\x07`)
	cleaned = titleRegex.ReplaceAllString(cleaned, "")

	cleaned = removeControlChars(cleaned)

	cleaned = removeTerminalNoise(cleaned)

	cleaned = normalizeWhitespace(cleaned)

	return strings.TrimSpace(cleaned)
}

// removeControlChars remove all control chars in cmd output
func removeControlChars(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' || (r >= 0x20 && r <= 0x7E) {
			return r
		}
		if r > 0x7E && r <= 0xFF && unicode.IsPrint(r) {
			return r
		}
		return -1
	}, s)
}

// removeTerminalNoise remove all terminal noise in cmd output
func removeTerminalNoise(s string) string {
	promptPatterns := []string{
		`^.*@.*:[/~].*[\$#]\s*`,        // user@host:path$
		`^\[\d+\]\s*`,                  // [0]
		`^\s*➜\s*`,                     // ➜
		`^\s*›\s*`,                     // ›
		`^\s*>\s*`,                     // >
		`^.*(Last login|Welcome to).*`, // login data
		`^.*(\x1b\]0;).*`,              // terminal title
	}

	for _, pattern := range promptPatterns {
		regex := regexp.MustCompile(pattern)
		s = regex.ReplaceAllString(s, "")
	}

	lines := strings.Split(s, "\n")
	var cleanLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !isLikelyCommandEcho(trimmed) {
			cleanLines = append(cleanLines, trimmed)
		}
	}

	return strings.Join(cleanLines, "\n")
}

// isLikelyCommandEcho check is cmd echo or not
func isLikelyCommandEcho(line string) bool {
	// 匹配常见的命令模式
	commandPatterns := []string{
		`^\s*(sudo|apt|docker|kubectl|systemctl|journalctl|cat|ls|cd|pwd|echo|printf)\s`,
		`^\s*[a-zA-Z0-9_\-]+\s+[a-zA-Z0-9_\-]`, // looks like cmd + args
	}

	for _, pattern := range commandPatterns {
		matched, _ := regexp.MatchString(pattern, line)
		if matched {
			return true
		}
	}

	return false
}

// normalizeWhitespace change cmd output white space to normal white space
func normalizeWhitespace(s string) string {
	s = strings.ReplaceAll(s, "\t", " ")

	spaceRegex := regexp.MustCompile(`\s+`)
	s = spaceRegex.ReplaceAllString(s, " ")

	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}

	var nonEmptyLines []string
	for _, line := range lines {
		if line != "" {
			nonEmptyLines = append(nonEmptyLines, line)
		}
	}

	return strings.Join(nonEmptyLines, "\n")
}

// CleanCommandOutput clean cmd output
func CleanCommandOutput(output, command string) string {
	cleaned := CleanOutput(output)

	if strings.Contains(command, "systemctl") {
		cleaned = cleanSystemctlOutput(cleaned)
	}

	if command != "" {
		commandRegex := regexp.MustCompile(regexp.QuoteMeta(command) + `\s*\n?`)
		cleaned = commandRegex.ReplaceAllString(cleaned, "")
	}

	return cleaned
}

// cleanSystemctlOutput clean system cmd output
func cleanSystemctlOutput(output string) string {
	statusValues := []string{
		"enabled", "disabled", "static", "masked", "linked", "indirect",
		"active", "inactive", "failed", "activating", "deactivating",
		"loaded", "not-found", "bad-setting", "error", "waiting",
	}

	lines := strings.Split(output, "\n")
	var cleanLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		isStatus := false
		for _, status := range statusValues {
			if strings.Contains(strings.ToLower(trimmed), status) {
				isStatus = true
				break
			}
		}

		if isStatus || isValidOutput(trimmed) {
			cleanLines = append(cleanLines, trimmed)
		}
	}

	return strings.Join(cleanLines, "\n")
}

// isValidOutput check is cmd output line valid
func isValidOutput(line string) bool {
	if line == "" {
		return false
	}

	if len(line) > 200 {
		return false
	}

	// Count non-printable characters (excluding spaces and tabs which are valid whitespace)
	// This detects lines dominated by control sequences, escape codes, or other terminal artifacts
	specialCharCount := 0
	for _, r := range line {
		if r != '\t' && (r != ' ' && !unicode.IsPrint(r)) {
			specialCharCount++
		}
	}

	// If more than 10% of characters are non-printable, this line is likely terminal noise
	// This heuristic helps distinguish between legitimate output and control sequence remnants
	return float64(specialCharCount)/float64(len(line)) <= 0.1
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
	if len(kernelStderr) > 0 {
		return nil, errors.Errorf("failed to get kernel, stderr: %q", string(kernelStderr))
	}
	osVars[_const.VariableOSKernelVersion] = string(bytes.TrimSpace(kernel))

	hn, hnStderr, err := c.ExecuteCommand(ctx, "hostname")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get hostname: %v, stderr: %q", err, string(hnStderr))
	}
	if len(hnStderr) > 0 {
		return nil, errors.Errorf("failed to get hostname, stderr: %q", string(hnStderr))
	}
	osVars[_const.VariableOSHostName] = string(bytes.TrimSpace(hn))

	arch, archStderr, err := c.ExecuteCommand(ctx, "arch")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get arch: %v, stderr: %q", err, string(archStderr))
	}
	if len(archStderr) > 0 {
		return nil, errors.Errorf("failed to get arch, stderr: %q", string(archStderr))
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
