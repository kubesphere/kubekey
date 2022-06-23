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
	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/util/filesystem"
)

type Interface interface {
	Connector
	Command
	Sftp
	LocalFileSystem
	Ping() error
	Host() string
}

type Connector interface {
	Connect() error
	Close()
}

type Command interface {
	Cmd(cmd string) (string, error)
	Cmdf(cmd string, a ...any) (string, error)
	SudoCmd(cmd string) (string, error)
	SudoCmdf(cmd string, a ...any) (string, error)
}

type Sftp interface {
	Copy(local, remote string) error
	Fetch(local, remote string) error
	RemoteFileExist(remote string) (bool, error)
}

type LocalFileSystem interface {
	Fs() filesystem.Interface
}
