/*
 Copyright 2021 The KubeSphere Authors.

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
	"io"

	"github.com/kubesphere/kubekey/util/workflow/cache"
)

type Connection interface {
	Exec(cmd string, host Host) (stdout string, code int, err error)
	PExec(cmd string, stdin io.Reader, stdout io.Writer, stderr io.Writer, host Host) (code int, err error)
	Fetch(local, remote string, host Host) error
	ScpFile(local, remote string, host Host) error
	RemoteFileExist(remote string, host Host) (bool, error)
	MkDirAll(path string, mode string, host Host) error
	Close()
}

type Connector interface {
	Connect(host Host) (Connection, error)
	Close(host Host)
}

type ModuleRuntime interface {
	GetObjName() string
	SetObjName(name string)
	GenerateWorkDir() error
	GetHostWorkDir() string
	GetWorkDir() string
	GetIgnoreErr() bool
	GetAllHosts() []Host
	SetAllHosts([]Host)
	GetHostsByRole(role string) []Host
	DeleteHost(host Host)
	HostIsDeprecated(host Host) bool
	InitLogger() error
}

type Runtime interface {
	GetRunner() *Runner
	SetRunner(r *Runner)
	GetConnector() Connector
	SetConnector(c Connector)
	RemoteHost() Host
	Copy() Runtime
	ModuleRuntime
}

type Host interface {
	GetName() string
	SetName(name string)
	GetAddress() string
	SetAddress(str string)
	GetInternalAddress() string
	SetInternalAddress(str string)
	GetPort() int
	SetPort(port int)
	GetUser() string
	SetUser(u string)
	GetPassword() string
	SetPassword(password string)
	GetPrivateKey() string
	SetPrivateKey(privateKey string)
	GetPrivateKeyPath() string
	SetPrivateKeyPath(path string)
	GetArch() string
	SetArch(arch string)
	GetTimeout() int64
	SetTimeout(timeout int64)
	GetRoles() []string
	SetRoles(roles []string)
	IsRole(role string) bool
	GetCache() *cache.Cache
	SetCache(c *cache.Cache)
}
