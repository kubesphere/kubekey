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
	"io/ioutil"
	"net"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"k8s.io/klog/v2/klogr"

	infrav1 "github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/api/v1beta1"
	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/util/filesystem"
)

const (
	DefaultSSHPort = 22
	DefaultTimeout = 15
)

type Client struct {
	logr.Logger
	mu             sync.Mutex
	User           string
	Password       string
	Port           *int
	PrivateKey     string
	PrivateKeyPath string
	Timeout        *time.Duration
	host           string
	sshClient      *ssh.Client
	sftpClient     *sftp.Client
	fs             filesystem.Interface
}

func NewClient(host string, auth *infrav1.Auth, log *logr.Logger) Interface {
	if log == nil {
		l := klogr.New()
		log = &l
	}
	if auth.User == "" {
		auth.User = "root"
	}

	var port int
	port = DefaultSSHPort
	if auth.Port == nil {
		auth.Port = &port
	}

	var timeout time.Duration
	timeout = time.Duration(DefaultTimeout) * time.Second
	if auth.Timeout == nil {
		auth.Timeout = &timeout
	}

	return &Client{
		User:           auth.User,
		Password:       auth.Password,
		Port:           auth.Port,
		PrivateKey:     auth.PrivateKey,
		PrivateKeyPath: auth.PrivateKeyPath,
		Timeout:        auth.Timeout,
		host:           host,
		fs:             filesystem.NewFileSystem(),
		Logger:         *log,
	}
}

func (c *Client) Connect() error {
	authMethods, err := c.authMethod(c.Password, c.PrivateKey, c.PrivateKeyPath)
	if err != nil {
		return errors.Wrap(err, "The given SSH key could not be parsed")
	}

	sshConfig := &ssh.ClientConfig{
		User:            c.User,
		Timeout:         *c.Timeout,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	endpoint := net.JoinHostPort(c.host, strconv.Itoa(*c.Port))
	sshClient, err := ssh.Dial("tcp", endpoint, sshConfig)
	if err != nil {
		return errors.Wrapf(err, "could not establish connection to %s", endpoint)
	}

	c.sshClient = sshClient
	return nil
}

func (c *Client) ConnectSftpClient(opts ...sftp.ClientOption) error {
	sess1, err := c.sshClient.NewSession()
	if err != nil {
		return err
	}
	defer sess1.Close()

	cmd := `grep -oP "Subsystem\s+sftp\s+\K.*" /etc/ssh/sshd_config`
	buff, err := sess1.Output(cmd)
	if err != nil {
		return fmt.Errorf("cmd output errored %v", err)
	}

	sess2, err := c.sshClient.NewSession()
	if err != nil {
		return err
	}

	sftpServerPath := strings.ReplaceAll(string(buff), "\r", "")
	if match, _ := regexp.MatchString(`^sudo `, sftpServerPath); !match {
		sftpServerPath = "sudo" + " " + sftpServerPath
	}

	ok, err := sess2.SendRequest("exec", true, ssh.Marshal(struct{ Command string }{sftpServerPath}))
	if err == nil && !ok {
		return errors.New("ssh: exec request failed")
	}

	pw, err := sess2.StdinPipe()
	if err != nil {
		return err
	}
	pr, err := sess2.StdoutPipe()
	if err != nil {
		return err
	}

	sftpClient, err := sftp.NewClientPipe(pr, pw, opts...)
	if err != nil {
		return err
	}
	c.sftpClient = sftpClient
	return nil
}

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
		content, err := ioutil.ReadFile(filepath.Clean(privateKeyPath))
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

	c.Logger.V(2).Info(fmt.Sprintf("cmd: %s", cmd))

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return "", errors.Wrapf(err, "[%s] run command failed", c.host)
	}

	return string(output), nil
}

func (c *Client) Cmdf(cmd string, a ...any) (string, error) {
	return c.Cmd(fmt.Sprintf(cmd, a...))
}

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
	c.Logger.V(2).Info(fmt.Sprintf("cmd: %s", cmd))

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
			_, err = in.Write([]byte(c.Password + "\n"))
			if err != nil {
				break
			}
		}
	}

	err = session.Wait()
	if err != nil {
		return "", err
	}
	outStr := strings.TrimPrefix(string(output), fmt.Sprintf("[sudo] password for %s:", c.User))
	return strings.TrimSpace(outStr), nil
}

func (c *Client) SudoCmdf(cmd string, a ...any) (string, error) {
	return c.SudoCmd(fmt.Sprintf(cmd, a...))
}

func (c *Client) Copy(src, dst string) error {
	if err := c.Connect(); err != nil {
		return errors.Wrapf(err, "[%s] connect ssh client failed", c.host)
	}

	if err := c.ConnectSftpClient(); err != nil {
		return errors.Wrapf(err, "[%s] connect sftp client failed", c.host)
	}
	defer c.sshClient.Close()
	defer c.sftpClient.Close()

	f, err := os.Stat(src)
	if err != nil {
		return errors.Wrapf(err, "[%s] get file stat failed", c.host)
	}

	if f.IsDir() {
		return errors.Wrapf(err, "[%s] the source %s is not a file", c.host, src)
	}

	baseRemoteFilePath := filepath.Dir(dst)
	_, err = c.sftpClient.ReadDir(baseRemoteFilePath)
	if err != nil {
		if err = c.sftpClient.MkdirAll(baseRemoteFilePath); err != nil {
			return err
		}
	}

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
	srcMd5 = c.fs.MD5Sum(src)
	if exist, err := c.remoteFileExist(dst); err != nil {
		return err
	} else if exist {
		dstMd5 = c.remoteMd5Sum(dst)
		if srcMd5 == dstMd5 {
			c.Logger.V(2).Info(fmt.Sprintf("remote file %s md5 value is the same as local file, skip scp", dst))
			return nil
		}
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	// the dst file mod will be 0666
	dstFile, err := c.sftpClient.Create(dst)
	if err != nil {
		return err
	}
	fileStat, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("get file stat failed %v", err)
	}
	if err := dstFile.Chmod(fileStat.Mode()); err != nil {
		return fmt.Errorf("chmod remote file failed %v", err)
	}
	defer dstFile.Close()
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}
	dstMd5 = c.remoteMd5Sum(dst)
	if srcMd5 != dstMd5 {
		return fmt.Errorf("validate md5sum failed %s != %s", srcMd5, dstMd5)
	}
	return nil
}

func (c *Client) Fetch(local, remote string) error {
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
			logrus.Fatal("failed to close file")
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
			logrus.Fatal("failed to close file")
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

func (c *Client) Ping() error {
	if err := c.Connect(); err != nil {
		return errors.Wrapf(err, "[%s] connect failed", c.host)
	}
	defer c.Close()
	return nil
}

func (c *Client) Host() string {
	return c.host
}

func (c *Client) Fs() filesystem.Interface {
	return c.fs
}

func fileExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func SudoPrefix(cmd string) string {
	return fmt.Sprintf("sudo -E /bin/bash <<EOF\n%s\nEOF", cmd)
}
