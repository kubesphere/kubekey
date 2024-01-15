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
	"path/filepath"
	"strconv"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
)

type sshConnector struct {
	Host     string
	Port     *int
	User     *string
	Password *string
	client   *ssh.Client
}

func (c *sshConnector) Init(ctx context.Context) error {
	if c.Host == "" {
		return fmt.Errorf("host is not set")
	}
	if c.Port == nil {
		c.Port = pointer.Int(22)
	}
	var auth []ssh.AuthMethod
	if c.Password != nil {
		auth = []ssh.AuthMethod{
			ssh.Password(*c.Password),
		}
	}
	sshClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", c.Host, strconv.Itoa(*c.Port)), &ssh.ClientConfig{
		User:            pointer.StringDeref(c.User, ""),
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		return err
	}
	c.client = sshClient

	return nil
}

func (c *sshConnector) Close(ctx context.Context) error {
	return c.client.Close()
}

func (c *sshConnector) CopyFile(ctx context.Context, src []byte, remoteFile string, mode fs.FileMode) error {
	// create sftp client
	sftpClient, err := sftp.NewClient(c.client)
	if err != nil {
		klog.Errorf("Failed to create sftp client: %v", err)
		return err
	}
	defer sftpClient.Close()
	// create remote file
	if _, err := sftpClient.Stat(filepath.Dir(remoteFile)); err != nil {
		klog.Warningf("Failed to stat dir %s: %v create it", filepath.Dir(remoteFile), err)
		if err := sftpClient.MkdirAll(filepath.Dir(remoteFile)); err != nil {
			klog.Errorf("Failed to create dir %s: %v", filepath.Dir(remoteFile), err)
			return err
		}
	}
	rf, err := sftpClient.Create(remoteFile)
	if err != nil {
		klog.Errorf("Failed to create file %s: %v", remoteFile, err)
		return err
	}
	defer rf.Close()

	if _, err = rf.Write(src); err != nil {
		klog.Errorf("Failed to write file %s: %v", remoteFile, err)
		return err
	}
	return rf.Chmod(mode)
}

func (c *sshConnector) FetchFile(ctx context.Context, remoteFile string, local io.Writer) error {
	// create sftp client
	sftpClient, err := sftp.NewClient(c.client)
	if err != nil {
		klog.Errorf("Failed to create sftp client: %v", err)
		return err
	}
	defer sftpClient.Close()
	rf, err := sftpClient.Open(remoteFile)
	if err != nil {
		klog.Errorf("Failed to open file %s: %v", remoteFile, err)
		return err
	}
	defer rf.Close()
	if _, err := io.Copy(local, rf); err != nil {
		klog.Errorf("Failed to copy file %s: %v", remoteFile, err)
		return err
	}
	return nil
}

func (c *sshConnector) ExecuteCommand(ctx context.Context, cmd string) ([]byte, error) {
	// create ssh session
	session, err := c.client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	return session.CombinedOutput(cmd)
}
