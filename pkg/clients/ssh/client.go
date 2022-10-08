/*
 Copyright 2022 The KubeSphere Authors.

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

package ssh

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"k8s.io/klog/v2/klogr"

	infrav1 "github.com/kubesphere/kubekey/api/v1beta1"
	"github.com/kubesphere/kubekey/pkg/util/filesystem"
)

// Default values.
const (
	DefaultSSHPort = 22
	DefaultTimeout = 15
	ROOT           = "root"
)

// Client is a wrapper around the SSH client that provides a few helper.
type Client struct {
	logr.Logger
	mu             sync.Mutex
	user           string
	password       string
	port           *int
	privateKey     string
	privateKeyPath string
	timeout        *time.Duration
	host           string
	sshClient      *ssh.Client
	sftpClient     *sftp.Client
	fs             filesystem.Interface
}

// NewClient returns a new client given ssh information.
func NewClient(host string, auth infrav1.Auth, log *logr.Logger) Interface {
	if log == nil {
		l := klogr.New()
		log = &l
	}
	if auth.User == "" {
		auth.User = ROOT
	}

	port := DefaultSSHPort
	if auth.Port == nil {
		auth.Port = &port
	}

	timeout := time.Duration(DefaultTimeout) * time.Second
	if auth.Timeout == nil {
		auth.Timeout = &timeout
	}

	return &Client{
		user:           auth.User,
		password:       auth.Password,
		port:           auth.Port,
		privateKey:     auth.PrivateKey,
		privateKeyPath: auth.PrivateKeyPath,
		timeout:        auth.Timeout,
		host:           host,
		fs:             filesystem.NewFileSystem(),
		Logger:         *log,
	}
}

// Connect connects to the host using the provided ssh information.
func (c *Client) Connect() error {
	authMethods, err := c.authMethod(c.password, c.privateKey, c.privateKeyPath)
	if err != nil {
		return errors.Wrap(err, "The given SSH key could not be parsed")
	}

	sshConfig := &ssh.ClientConfig{
		User:    c.user,
		Timeout: *c.timeout,
		Auth:    authMethods,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	endpoint := net.JoinHostPort(c.host, strconv.Itoa(*c.port))
	sshClient, err := ssh.Dial("tcp", endpoint, sshConfig)
	if err != nil {
		return errors.Wrapf(err, "could not establish connection to %s", endpoint)
	}

	c.sshClient = sshClient
	return nil
}

// ConnectSftpClient connects to the host sftp client using the provided ssh information.
func (c *Client) ConnectSftpClient(opts ...sftp.ClientOption) error {
	var (
		sftpClient *sftp.Client
		err        error
	)

	sftpClient, err = sftp.NewClient(c.sshClient, opts...)
	c.sftpClient = sftpClient
	return err
}

// Close closes the underlying ssh and sftp connection.
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.sshClient == nil && c.sftpClient == nil {
		return
	}

	if c.sshClient != nil {
		_ = c.sshClient.Close()
		c.sshClient = nil
	}
	if c.sftpClient != nil {
		_ = c.sftpClient.Close()
		c.sftpClient = nil
	}
}

func (c *Client) authMethod(password, privateKey, privateKeyPath string) (auths []ssh.AuthMethod, err error) {
	if privateKey != "" || privateKeyPath != "" {
		am, err := c.privateKeyMethod(privateKey, privateKeyPath)
		if err != nil {
			return auths, err
		}
		auths = append(auths, am)
	}
	if password != "" {
		auths = append(auths, ssh.Password(password))
	}
	return auths, nil
}

func (c *Client) privateKeyMethod(privateKey, privateKeyPath string) (am ssh.AuthMethod, err error) {
	var signer ssh.Signer

	if fileExist(privateKeyPath) {
		content, err := os.ReadFile(filepath.Clean(privateKeyPath))
		if err != nil {
			return nil, err
		}
		if privateKey == "" {
			signer, err = ssh.ParsePrivateKey(content)
			if err != nil {
				return nil, err
			}
		} else {
			passphrase := []byte(privateKey)
			signer, err = ssh.ParsePrivateKeyWithPassphrase(content, passphrase)
			if err != nil {
				return nil, err
			}
		}
	} else {
		signer, err = ssh.ParsePrivateKey([]byte(privateKey))
		if err != nil {
			return nil, err
		}
	}

	return ssh.PublicKeys(signer), nil
}

func (c *Client) session() (*ssh.Session, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.sshClient == nil {
		return nil, errors.New("connection closed")
	}

	sess, err := c.sshClient.NewSession()
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

// Cmd executes a command on the remote host.
func (c *Client) Cmd(cmd string) (string, error) {
	if err := c.Connect(); err != nil {
		return "", errors.Wrapf(err, "[%s] connect ssh client failed", c.host)
	}
	session, err := c.session()
	if err != nil {
		return "", errors.Wrapf(err, "[%s] create ssh session failed", c.host)
	}
	defer session.Close()
	defer c.sshClient.Close()

	c.Logger.V(4).Info(fmt.Sprintf("cmd: %s", cmd))

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return "", errors.Wrapf(err, "[%s] run command failed", c.host)
	}

	return string(output), nil
}

// Cmdf execute a formatting command according to the format specifier.
func (c *Client) Cmdf(cmd string, a ...any) (string, error) {
	return c.Cmd(fmt.Sprintf(cmd, a...))
}

// SudoCmd executes a command on the remote host with sudo.
func (c *Client) SudoCmd(cmd string) (string, error) {
	if err := c.Connect(); err != nil {
		return "", errors.Wrapf(err, "[%s] connect ssh client failed", c.host)
	}
	defer c.sshClient.Close()
	return c.sudoCmd(cmd)
}

func (c *Client) sudoCmd(cmd string) (string, error) {
	session, err := c.session()
	if err != nil {
		return "", errors.Wrapf(err, "[%s] create ssh session failed", c.host)
	}
	defer session.Close()

	cmd = SudoPrefix(cmd)
	c.Logger.V(4).Info(fmt.Sprintf("cmd: %s", cmd))

	in, err := session.StdinPipe()
	if err != nil {
		return "", err
	}

	out, err := session.StdoutPipe()
	if err != nil {
		return "", err
	}

	if err := session.Start(cmd); err != nil {
		return "", err
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
			_, err = in.Write([]byte(c.password + "\n"))
			if err != nil {
				break
			}
		}
	}

	outStr := strings.TrimPrefix(string(output), fmt.Sprintf("[sudo] password for %s:", c.user))
	err = session.Wait()
	if err != nil {
		return strings.TrimSpace(outStr), errors.Wrap(err, strings.TrimSpace(outStr))
	}
	return strings.TrimSpace(outStr), nil
}

// SudoCmdf executes a formatting command on the remote host with sudo.
func (c *Client) SudoCmdf(cmd string, a ...any) (string, error) {
	return c.SudoCmd(fmt.Sprintf(cmd, a...))
}

// Copy copies a file to the remote host.
func (c *Client) Copy(src, dst string) error {
	if c.user == ROOT {
		return c.copy(src, dst)
	}
	return c.sudoCopy(src, dst)
}

func (c *Client) sudoCopy(src, dst string) error {
	// scp to tmp dir
	remoteTmp := filepath.Join("/tmp/kubekey", dst)
	if err := c.copy(src, remoteTmp); err != nil {
		return err
	}

	baseRemotePath := filepath.Dir(dst)
	if err := c.mkdirAll(baseRemotePath, ""); err != nil {
		return err
	}
	if _, err := c.SudoCmdf("mv -f %s %s", remoteTmp, dst); err != nil {
		return errors.Wrapf(err, "[%s] mv -f %s %s failed", c.host, remoteTmp, dst)
	}
	if _, err := c.SudoCmd("rm -rf /tmp/kubekey*"); err != nil {
		return errors.Wrapf(err, "[%s] rm -rf /tmp/kubekey* failed", c.host)
	}
	return nil
}

func (c *Client) copy(src, dst string) error {
	baseRemoteFilePath := filepath.Dir(dst)
	_ = c.mkdirAll(baseRemoteFilePath, "777")

	if err := c.Connect(); err != nil {
		return errors.Wrapf(err, "[%s] connect ssh client failed", c.host)
	}
	if err := c.ConnectSftpClient(); err != nil {
		return errors.Wrapf(err, "[%s] connect sftp client failed", c.host)
	}
	defer c.sshClient.Close()
	defer c.sftpClient.Close()

	if err := c.copyLocalFileToRemote(src, dst); err != nil {
		return errors.Wrapf(err, "[%s] copy file failed", c.host)
	}
	return nil
}

func (c *Client) copyLocalFileToRemote(src, dst string) error {
	// check remote file md5 first
	var (
		srcMd5, dstMd5 string
	)
	cleanSrc := filepath.Clean(src)
	srcMd5 = c.fs.MD5Sum(cleanSrc)
	if exist, err := c.remoteFileExist(dst); err != nil {
		return err
	} else if exist {
		dstMd5 = c.remoteMd5Sum(dst)
		if srcMd5 == dstMd5 {
			c.Logger.V(4).Info(fmt.Sprintf("remote file %s md5 value is the same as local file, skip scp", dst))
			return nil
		}
	}

	srcFile, err := os.Open(cleanSrc)
	if err != nil {
		return errors.Wrapf(err, "open local file %s failed", cleanSrc)
	}
	defer srcFile.Close()

	fileStat, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("get file stat failed %v", err)
	}
	if fileStat.IsDir() {
		return fmt.Errorf("the source %s is not a file", cleanSrc)
	}

	// the dst file mod will be 0666
	dstFile, err := c.sftpClient.Create(dst)
	if err != nil {
		return errors.Wrapf(err, "[%s] create remote file %s failed", c.host, dst)
	}
	if err := dstFile.Chmod(fileStat.Mode()); err != nil {
		return fmt.Errorf("chmod remote file failed %v", err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return errors.Wrapf(err, "[%s] io copy file %s to remote %s failed", c.host, cleanSrc, dst)
	}

	dstMd5 = c.remoteMd5Sum(dst)
	if srcMd5 != dstMd5 {
		return fmt.Errorf("validate md5sum failed %s != %s", srcMd5, dstMd5)
	}
	return nil
}

// Fetch fetches a file from the remote host.
func (c *Client) Fetch(local, remote string) error {
	if c.user == ROOT {
		return c.fetch(local, remote)
	}
	return c.sudoFetch(local, remote)
}

func (c *Client) sudoFetch(local, remote string) error {
	remoteTmp := filepath.Join("/tmp/kubekey", filepath.Base(remote))
	baseRemotePath := filepath.Dir(remoteTmp)
	if err := c.mkdirAll(baseRemotePath, "777"); err != nil {
		return err
	}
	if _, err := c.SudoCmdf("cp %s %s", remote, remoteTmp); err != nil {
		return errors.Wrapf(err, "[%s] cp %s %s failed", c.host, remote, remoteTmp)
	}

	if err := c.fetch(local, remoteTmp); err != nil {
		return err
	}
	if _, err := c.SudoCmd("rm -rf /tmp/kubekey*"); err != nil {
		return errors.Wrapf(err, "[%s] rm -rf /tmp/kubekey* failed", c.host)
	}
	return nil
}

func (c *Client) fetch(local, remote string) error {
	if err := c.Connect(); err != nil {
		return errors.Wrapf(err, "[%s] connect ssh client failed", c.host)
	}

	if err := c.ConnectSftpClient(); err != nil {
		return errors.Wrapf(err, "[%s] connect sftp client failed", c.host)
	}
	defer c.sshClient.Close()
	defer c.sftpClient.Close()

	ok, err := c.RemoteFileExist(remote)
	if err != nil {
		return errors.Wrapf(err, "[%s] check remote file failed", c.host)
	}
	if !ok {
		return errors.Errorf("[%s] remote file %s not exist", c.host, remote)
	}

	// open remote source file
	srcFile, err := c.sftpClient.Open(remote)
	if err != nil {
		return fmt.Errorf("open remote file failed %v, remote path: %s", err, remote)
	}
	defer func() {
		if err := srcFile.Close(); err != nil {
			c.Logger.Error(err, "failed to close file")
		}
	}()

	err = os.MkdirAll(filepath.Dir(local), os.ModePerm)
	if err != nil {
		return err
	}

	dstFile, err := os.Create(filepath.Clean(local))
	if err != nil {
		return fmt.Errorf("create local file failed %v", err)
	}
	defer func() {
		if err := dstFile.Close(); err != nil {
			c.Logger.Error(err, "failed to close file")
		}
	}()

	_, err = srcFile.WriteTo(dstFile)
	return err
}

func (c *Client) remoteMd5Sum(dst string) string {
	cmd := fmt.Sprintf("md5sum %s | cut -d\" \" -f1", dst)
	remoteMd5, err := c.sudoCmd(cmd)
	if err != nil {
		c.Logger.Error(err, fmt.Sprintf("sum remote file md5 failed, output: %s", remoteMd5))
		return ""
	}
	return remoteMd5
}

// RemoteFileExist checks if a file exists on the remote host.
func (c *Client) RemoteFileExist(remote string) (bool, error) {
	if err := c.Connect(); err != nil {
		return false, errors.Wrapf(err, "[%s] connect failed", c.host)
	}
	defer c.sshClient.Close()
	return c.remoteFileExist(remote)
}

func (c *Client) remoteFileExist(remote string) (bool, error) {
	remoteFileName := path.Base(remote)
	remoteFileDirName := path.Dir(remote)

	remoteFileCommand := fmt.Sprintf("ls -l %s/%s 2>/dev/null |wc -l", remoteFileDirName, remoteFileName)

	out, err := c.sudoCmd(remoteFileCommand)
	if err != nil {
		return false, err
	}
	count, err := strconv.Atoi(strings.TrimSpace(out))
	if err != nil {
		return false, err
	}
	return count != 0, nil
}

func (c *Client) mkdirAll(path, mode string) error {
	if mode == "" {
		mode = "775"
	}
	if _, err := c.SudoCmdf("mkdir -p -m %s %s", mode, path); err != nil {
		return errors.Wrapf(err, "[%s] mkdir -p -m %s %s failed", c.host, mode, path)
	}
	return nil
}

// Ping checks if the remote host is reachable.
func (c *Client) Ping() error {
	if err := c.Connect(); err != nil {
		return errors.Wrapf(err, "[%s] connect failed", c.host)
	}
	defer c.Close()
	return nil
}

// Host returns the host name of the ssh client.
func (c *Client) Host() string {
	return c.host
}

// Fs returns the filesystem of the ssh client.
func (c *Client) Fs() filesystem.Interface {
	return c.fs
}

func fileExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

// SudoPrefix returns the prefix for sudo commands.
func SudoPrefix(cmd string) string {
	return fmt.Sprintf("sudo -E /bin/bash <<EOF\n%s\nEOF", cmd)
}
