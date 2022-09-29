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

// Package operation define the remote instance operations interface.
package operation

import (
	"net/url"
	"time"

	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/service/operation/file"
	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/service/operation/file/checksum"
)

// File interface defines the operations for normal file which needed to be copied to remote.
type File interface {
	Name() string
	Type() file.Type
	LocalPath() string
	RemotePath() string
	LocalExist() bool
	RemoteExist() bool
	Copy(override bool) error
	Fetch(override bool) error
	Chmod(option string) error
}

// Binary interface defines the operations for Kubernetes needed binaries which usually needed to be copied to remote.
type Binary interface {
	File
	ID() string
	Arch() string
	Version() string
	URL() *url.URL
	SetURL(url string)
	SetHost(host string)
	SetPath(path string)
	SetZone(zone string)
	AppendChecksum(c checksum.Interface)
	Get(timeout time.Duration) error
	CompareChecksum() error
}

// Template interface defines the operations for Kubernetes needed template files (systemd files, config files .e.g)
// which usually needed to be copied to remote.
type Template interface {
	File
	RenderToLocal() error
}

// User interface defines the operations for remote instance Linux user.
type User interface {
	Add() error
}

// Directory interface defines the operations for remote instance Linux directory.
type Directory interface {
	Make() error
	Chown(user string) error
	Remove() error
}

// Repository interface defines the operations for remote instance Linux repository.
type Repository interface {
	Update() error
	Install(pkg ...string) error
	Add(path string) error
}
