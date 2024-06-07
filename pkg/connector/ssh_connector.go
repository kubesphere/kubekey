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
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"k8s.io/klog/v2"
)

type sshConnector struct {
	Host     string
	Port     int
	User     string
	Password string
	client   *ssh.Client
}

func (c *sshConnector) Init(ctx context.Context) error {
	if c.Host == "" {
		return fmt.Errorf("host is not set")
	}
	var auth []ssh.AuthMethod
	if c.Password != "" {
		auth = []ssh.AuthMethod{
			ssh.Password(c.Password),
		}
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

	return nil
}

func (c *sshConnector) Close(ctx context.Context) error {
	return c.client.Close()
}

// PutFile to remote node. src is the file bytes. dst is the remote filename
func (c *sshConnector) PutFile(ctx context.Context, src []byte, dst string, mode fs.FileMode) error {
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
func (c *sshConnector) FetchFile(ctx context.Context, src string, dst io.Writer) error {
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

func (c *sshConnector) ExecuteCommand(ctx context.Context, cmd string) ([]byte, error) {
	klog.V(4).InfoS("exec ssh command", "cmd", cmd, "host", c.Host)
	// create ssh session
	session, err := c.client.NewSession()
	if err != nil {
		klog.V(4).ErrorS(err, "Failed to create ssh session")
		return nil, err
	}
	defer session.Close()

	return session.CombinedOutput(cmd)
}
