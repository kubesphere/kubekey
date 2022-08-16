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

package operation

import (
	"time"

	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/service/operation/file"
)

type File interface {
	Name() string
	Type() file.FileType
	LocalPath() string
	RemotePath() string
	LocalExist() bool
	RemoteExist() bool
	Copy(override bool) error
	Fetch(override bool) error
	Chmod(option string) error
}

type Binary interface {
	File
	ID() string
	Arch() string
	Version() string
	Url() string
	SetZone(zone string)
	Get(timeout time.Duration) error
	CompareChecksum() error
}

type Template interface {
	File
	RenderToLocal() error
}

type User interface {
	Add() error
}

type Directory interface {
	Make() error
	Chown(user string) error
	Remove() error
}

type Repository interface {
	Update() error
	Install(pkg ...string) error
}
