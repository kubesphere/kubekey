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

	"k8s.io/klog/v2"
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
	// PutFile copies a file from local to remote
	PutFile(ctx context.Context, local []byte, remoteFile string, mode fs.FileMode) error
	// FetchFile copies a file from remote to local
	FetchFile(ctx context.Context, remoteFile string, local io.Writer) error
	// ExecuteCommand executes a command on the remote host
	ExecuteCommand(ctx context.Context, cmd string) ([]byte, error)
}

// NewConnector creates a new connector
// if set connector to local, use local connector
// if set connector to ssh, use ssh connector
// if connector is not set. when host is localhost, use local connector, else use ssh connector
// vars contains all inventory for host. It's best to define the connector info in inventory file.
func NewConnector(host string, vars map[string]any) (Connector, error) {
	switch vars["connector"] {
	case "local":
		return &localConnector{Cmd: exec.New()}, nil
	case "ssh":
		hostParam, err := variable.StringVar(nil, vars, "ssh_host")
		if err != nil {
			return nil, err
		}

		portParam, err := variable.IntVar(nil, vars, "ssh_port")
		if err != nil { // default port 22
			klog.InfoS("get ssh port failed use default port 22", "error", err)
			portParam = 22
		}

		userParam, err := variable.StringVar(nil, vars, "ssh_user")
		if err != nil {
			return nil, err
		}

		passParam, err := variable.StringVar(nil, vars, "ssh_password")
		if err != nil {
			return nil, err
		}
		return &sshConnector{
			Host:     hostParam,
			Port:     portParam,
			User:     userParam,
			Password: passParam,
		}, nil
	default:
		localHost, _ := os.Hostname()
		if host == _const.LocalHostName || localHost == host {
			return &localConnector{Cmd: exec.New()}, nil
		}

		hostParam, err := variable.StringVar(nil, vars, "ssh_host")
		if err != nil {
			return nil, err
		}

		portParam, err := variable.IntVar(nil, vars, "ssh_port")
		if err != nil {
			return nil, err
		}

		userParam, err := variable.StringVar(nil, vars, "ssh_user")
		if err != nil {
			return nil, err
		}

		passParam, err := variable.StringVar(nil, vars, "ssh_password")
		if err != nil {
			return nil, err
		}
		return &sshConnector{
			Host:     hostParam,
			Port:     portParam,
			User:     userParam,
			Password: passParam,
		}, nil
	}
}
