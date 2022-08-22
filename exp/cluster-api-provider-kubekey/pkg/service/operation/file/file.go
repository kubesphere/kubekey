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

// Type represents the type of file.
type Type string

var (
	// FileBinary represents a binary file.
	FileBinary = Type("fileBinary")
	// FileText represents a text file.
	FileText = Type("fileText")
	// FileTemplate represents a template file.
	FileTemplate = Type("fileTemplate")
)

// Params represents the parameters of a file.
type Params struct {
	SSHClient      ssh.Interface
	Name           string
	Type           Type
	LocalFullPath  string
	RemoteFullPath string
	RootFs         rootfs.Interface
}

// NewFile returns a new File object given a FileParams.
func NewFile(params Params) (*File, error) {
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

// File is an implementation of the File interface.
type File struct {
	sshClient      ssh.Interface
	name           string
	fileType       Type
	localFullPath  string
	remoteFullPath string
	rootFs         rootfs.Interface
}

// Name returns the name of the file.
func (s *File) Name() string {
	return s.name
}

// Type returns the type of the file.
func (s *File) Type() Type {
	return s.fileType
}

// SetLocalPath sets the local path of the file.
func (s *File) SetLocalPath(path string) {
	s.localFullPath = path
}

// SetRemotePath sets the remote path of the file.
func (s *File) SetRemotePath(path string) {
	s.remoteFullPath = path
}

// LocalPath returns the local path of the file.
func (s *File) LocalPath() string {
	return s.localFullPath
}

// RemotePath returns the remote path of the file.
func (s *File) RemotePath() string {
	return s.remoteFullPath
}

// LocalExist returns true if the file exists in the local path.
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

// RemoteExist returns true if the file exists in the remote path.
func (s *File) RemoteExist() bool {
	ok, err := s.sshClient.RemoteFileExist(s.RemotePath())
	if err != nil {
		return false
	}
	return ok
}

// Copy copies the file from the local path to the remote path.
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

// Fetch copies the file from the remote path to the local path.
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

// Chmod changes the mode of the file.
func (s *File) Chmod(option string) error {
	if !s.RemoteExist() {
		return errors.Errorf("remote file %s is not exist in the remote path %s", s.Name(), s.RemotePath())
	}

	_, err := s.sshClient.SudoCmdf("chmod %s %s", option, s.remoteFullPath)
	return err
}
