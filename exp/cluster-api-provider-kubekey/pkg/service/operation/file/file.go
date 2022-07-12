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

package file

import (
	"os"

	"github.com/pkg/errors"

	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/clients/ssh"
	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/rootfs"
)

type FileType string

var (
	FileBinary   = FileType("fileBinary")
	FileText     = FileType("fileText")
	FileTemplate = FileType("fileTemplate")
)

type FileParams struct {
	SSHClient      ssh.Interface
	Name           string
	Type           FileType
	LocalFullPath  string
	RemoteFullPath string
	RootFs         rootfs.Interface
}

func NewFile(params FileParams) (*File, error) {
	if params.SSHClient == nil {
		return nil, errors.New("ssh client is required when creating a File")
	}
	if params.RootFs == nil {
		return nil, errors.New("rootfs is required when creating a File")
	}
	if params.Type == "" {
		return nil, errors.New("file type is required when creating a File")
	}
	return &File{
		sshClient:      params.SSHClient,
		rootFs:         params.RootFs,
		name:           params.Name,
		fileType:       params.Type,
		localFullPath:  params.LocalFullPath,
		remoteFullPath: params.RemoteFullPath,
	}, nil
}

type File struct {
	sshClient      ssh.Interface
	name           string
	fileType       FileType
	localFullPath  string
	remoteFullPath string
	rootFs         rootfs.Interface
}

func (s *File) Name() string {
	return s.name
}

func (s *File) Type() FileType {
	return s.fileType
}

func (s *File) SetLocalPath(path string) {
	s.localFullPath = path
}

func (s *File) SetRemotePath(path string) {
	s.remoteFullPath = path
}

func (s *File) LocalPath() string {
	return s.localFullPath
}

func (s *File) RemotePath() string {
	return s.remoteFullPath
}

func (s *File) LocalExist() bool {
	_, err := os.Stat(s.LocalPath())
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		if os.IsNotExist(err) {
			return false
		}
		return false
	}
	return true
}

func (s *File) RemoteExist() bool {
	ok, err := s.sshClient.RemoteFileExist(s.RemotePath())
	if err != nil {
		return false
	}
	return ok
}

func (s *File) Copy(override bool) error {
	if !s.LocalExist() {
		return errors.Errorf("file %s is not exist in the local path %s", s.Name(), s.LocalPath())
	}

	if !override {
		if s.RemoteExist() {
			return nil
		}
	}
	return s.sshClient.Copy(s.LocalPath(), s.RemotePath())
}

func (s *File) Fetch(override bool) error {
	if !s.RemoteExist() {
		return errors.Errorf("remote file %s is not exist in the remote path %s", s.Name(), s.RemotePath())
	}

	if !override {
		if s.LocalExist() {
			return nil
		}
	}
	return s.sshClient.Fetch(s.LocalPath(), s.RemotePath())
}

func (s *File) Chmod(option string) error {
	if !s.RemoteExist() {
		return errors.Errorf("remote file %s is not exist in the remote path %s", s.Name(), s.RemotePath())
	}

	_, err := s.sshClient.SudoCmdf("chmod %s %s", option, s.remoteFullPath)
	return err
}
