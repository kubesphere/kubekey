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
	"io"
	"io/fs"
	"os"

	"k8s.io/utils/exec"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

// Connector is the interface for connecting to a remote host
type Connector interface {
	// Init initializes the connection
	Init(ctx context.Context) error
	// Close closes the connection
	Close(ctx context.Context) error
	// CopyFile copies a file from local to remote
	CopyFile(ctx context.Context, local []byte, remoteFile string, mode fs.FileMode) error
	// FetchFile copies a file from remote to local
	FetchFile(ctx context.Context, remoteFile string, local io.Writer) error
	// ExecuteCommand executes a command on the remote host
	ExecuteCommand(ctx context.Context, cmd string) ([]byte, error)
}

// NewConnector creates a new connector
func NewConnector(host string, vars variable.VariableData) Connector {
	switch vars["connector"] {
	case "local":
		return &localConnector{Cmd: exec.New()}
	case "ssh":
		if variable.StringVar(vars, "ssh_host") != nil {
			host = *variable.StringVar(vars, "ssh_host")
		}
		return &sshConnector{
			Host:     host,
			Port:     variable.IntVar(vars, "ssh_port"),
			User:     variable.StringVar(vars, "ssh_user"),
			Password: variable.StringVar(vars, "ssh_password"),
		}
	default:
		localHost, _ := os.Hostname()
		if host == _const.LocalHostName || localHost == host {
			return &localConnector{Cmd: exec.New()}
		}

		if variable.StringVar(vars, "ssh_host") != nil {
			host = *variable.StringVar(vars, "ssh_host")
		}
		return &sshConnector{
			Host:     host,
			Port:     variable.IntVar(vars, "ssh_port"),
			User:     variable.StringVar(vars, "ssh_user"),
			Password: variable.StringVar(vars, "ssh_password"),
		}
	}
}
