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
	"net"
	"os"

	"github.com/cockroachdb/errors"
	"k8s.io/klog/v2"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

// connectedType for connector
const (
	connectedSSH        = "ssh"
	connectedLocal      = "local"
	connectedKubernetes = "kubernetes"
	defaultSHELL        = "/bin/bash"
)

// Connector is the interface for connecting to a remote host
type Connector interface {
	// Init initializes the connection
	Init(ctx context.Context) error
	// Close closes the connection
	Close(ctx context.Context) error
	// PutFile copies a file from src to dst with mode.
	PutFile(ctx context.Context, src []byte, dst string, mode fs.FileMode) error
	// FetchFile copies a file from src to dst writer.
	FetchFile(ctx context.Context, src string, dst io.Writer) error
	// ExecuteCommand executes a command on the remote host
	ExecuteCommand(ctx context.Context, cmd string) ([]byte, error)
}

// NewConnector creates a new connector
// if set connector to "local", use local connector
// if set connector to "ssh", use ssh connector
// if set connector to "kubernetes", use kubernetes connector
// if connector is not set. when host is localhost, use local connector, else use ssh connector
// vars contains all inventory for host. It's best to define the connector info in inventory file.
func NewConnector(host string, v variable.Variable) (Connector, error) {
	vars, err := v.Get(variable.GetAllVariable(host))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get host %q variable", host)
	}
	connectorVars := make(map[string]any)
	if c1, ok := vars.(map[string]any)[_const.VariableConnector]; ok {
		if c2, ok := c1.(map[string]any); ok {
			connectorVars = c2
		}
	}

	connectedType, _ := variable.StringVar(nil, connectorVars, _const.VariableConnectorType)
	switch connectedType {
	case connectedLocal:
		return newLocalConnector(connectorVars), nil
	case connectedSSH:
		return newSSHConnector(host, connectorVars), nil
	case connectedKubernetes:
		workdir, err := v.Get(variable.GetWorkDir())
		if err != nil {
			return nil, errors.Wrap(err, "failed to get workdir from variable")
		}
		wd, ok := workdir.(string)
		if !ok {
			return nil, errors.New("workdir in variable should be string")
		}

		return newKubernetesConnector(host, wd, connectorVars)
	default:
		localHost, _ := os.Hostname()
		// get host in connector variable. if empty, set default host: host_name.
		hostParam, err := variable.StringVar(nil, connectorVars, _const.VariableConnectorHost)
		if err != nil {
			klog.V(4).Infof("connector host is empty use: %s", host)
			hostParam = host
		}
		if host == _const.VariableLocalHost || host == localHost || isLocalIP(hostParam) {
			return newLocalConnector(connectorVars), nil
		}

		return newSSHConnector(host, connectorVars), nil
	}
}

// GatherFacts get host info.
type GatherFacts interface {
	HostInfo(ctx context.Context) (map[string]any, error)
}

// isLocalIP check if given ipAddr is local network ip
func isLocalIP(ipAddr string) bool {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		klog.ErrorS(err, "get network address error")

		return false
	}

	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		default:
			klog.V(4).InfoS("unknown address type", "address", addr.String())

			continue
		}

		if ip.String() == ipAddr {
			return true
		}
	}

	return false
}
