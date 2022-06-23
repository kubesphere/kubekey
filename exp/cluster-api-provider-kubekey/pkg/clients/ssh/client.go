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
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/povsister/scp"
	"golang.org/x/crypto/ssh"

	infrav1 "github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/api/v1beta1"
)

const (
	DefaultSSHPort = 22
	DefaultTimeout = 15
)

type Client struct {
	mu             sync.Mutex
	User           string
	Password       string
	Port           *int
	PrivateKey     string
	PrivateKeyPath string
	Timeout        *time.Duration
	host           string
	sshClient      *ssh.Client
	scpClient      *scp.Client
}

func NewClient(auth *infrav1.Auth) Interface {
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
	}
}

func (c *Client) Connect(host string) error {
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

	endpoint := net.JoinHostPort(host, strconv.Itoa(*c.Port))
	sshClient, err := ssh.Dial("tcp", endpoint, sshConfig)
	if err != nil {
		return errors.Wrapf(err, "could not establish connection to %s", endpoint)
	}
	scpClient, err := scp.NewClientFromExistingSSH(sshClient, &scp.ClientOption{})
	if err != nil {
		return errors.Wrapf(err, "coould not new scp client to : %v", endpoint)
	}

	c.host = host
	c.sshClient = sshClient
	c.scpClient = scpClient
	return nil
}

func (c *Client) Close() (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.sshClient == nil && c.scpClient == nil {
		return err
	}

	if c.sshClient != nil {
		err = c.sshClient.Close()
		c.sshClient = nil
	}
	if c.scpClient != nil {
		err = c.scpClient.Close()
		c.scpClient = nil
	}
	return err
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
	session, err := c.session()
	if err != nil {
		return "", errors.Wrapf(err, "[%s] create ssh session failed", c.host)
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return "", errors.Wrapf(err, "[%s] run command failed", c.host)
	}

	return string(output), nil
}

func (c *Client) SudoCmd(cmd string) (string, error) {
	session, err := c.session()
	if err != nil {
		return "", errors.Wrapf(err, "[%s] create ssh session failed", c.host)
	}
	defer session.Close()

	stdoutB := new(bytes.Buffer)
	session.Stdout = stdoutB
	in, _ := session.StdinPipe()

	go func(in io.Writer, output *bytes.Buffer) {
		for {
			if strings.Contains(string(output.Bytes()), "[sudo] password for ") {
				_, err = in.Write([]byte(c.Password + "\n"))
				if err != nil {
					break
				}
				break
			}
		}
	}(in, stdoutB)

	err = session.Run(SudoPrefix(cmd))
	if err != nil {
		return "", err
	}
	return stdoutB.String(), nil
}

func (c *Client) Copy(src, dst string) error {
	f, err := os.Stat(src)
	if err != nil {
		return errors.Wrapf(err, "[%s] get file stat failed", c.host)
	}

	if f.IsDir() {
		return errors.Wrapf(err, "[%s] the source %s is not a file", c.host, src)
	}

	if err := c.scpClient.CopyFileToRemote(src, dst, &scp.FileTransferOption{PreserveProp: true}); err != nil {
		return errors.Wrapf(err, "[%s] copy file failed", c.host)
	}
	return nil
}

func (c *Client) Fetch(local, remote string) error {
	ok, err := c.RemoteFileExist(remote)
	if err != nil {
		return errors.Wrapf(err, "[%s] check remote file failed", c.host)
	}
	if !ok {
		return errors.Errorf("[%s] remote file %s not exist", c.host, remote)
	}

	if err := c.scpClient.CopyFileFromRemote(remote, local, &scp.FileTransferOption{}); err != nil {
		return errors.Wrapf(err, "[%s] fetch file failed", c.host)
	}
	return nil
}

func (c *Client) RemoteFileExist(remote string) (bool, error) {
	remoteFileName := path.Base(remote)
	remoteFileDirName := path.Dir(remote)

	remoteFileCommand := fmt.Sprintf("ls -l %s/%s 2>/dev/null |wc -l", remoteFileDirName, remoteFileName)

	out, err := c.SudoCmd(remoteFileCommand)
	if err != nil {
		return false, err
	}
	count, err := strconv.Atoi(strings.TrimSpace(out))
	if err != nil {
		return false, err
	}
	return count != 0, nil
}

func (c *Client) Ping(host string) error {
	if err := c.Connect(host); err != nil {
		return errors.Wrapf(err, "[%s] connect failed", host)
	}

	if err := c.Close(); err != nil {
		return errors.Wrapf(err, "[%s] close connect failed", host)
	}
	return nil
}

func fileExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func SudoPrefix(cmd string) string {
	return fmt.Sprintf("sudo -E /bin/bash -c \"%s\"", cmd)
}
