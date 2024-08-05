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

	"k8s.io/klog/v2"
	"k8s.io/utils/exec"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

// connectedType for connector
const (
	connectedDefault    = ""
	connectedSSH        = "ssh"
	connectedLocal      = "local"
	connectedKubernetes = "kubernetes"
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
func NewConnector(host string, connectorVars map[string]any) (Connector, error) {
	connectedType, _ := variable.StringVar(nil, connectorVars, _const.VariableConnectorType)
	switch connectedType {
	case connectedLocal:
		return &localConnector{Cmd: exec.New()}, nil
	case connectedSSH:
		// get host in connector variable. if empty, set default host: host_name.
		hostParam, err := variable.StringVar(nil, connectorVars, _const.VariableConnectorHost)
		if err != nil {
			klog.InfoS("get ssh port failed use default port 22", "error", err)
			hostParam = host
		}
		// get port in connector variable. if empty, set default port: 22.
		portParam, err := variable.IntVar(nil, connectorVars, _const.VariableConnectorPort)
		if err != nil {
			klog.V(4).Infof("connector port is empty use: %v", defaultSSHPort)
			portParam = defaultSSHPort
		}
		// get user in connector variable. if empty, set default user: root.
		userParam, err := variable.StringVar(nil, connectorVars, _const.VariableConnectorUser)
		if err != nil {
			klog.V(4).Infof("connector user is empty use: %s", defaultSSHUser)
			userParam = defaultSSHUser
		}
		// get password in connector variable. if empty, should connector by private key.
		passwdParam, err := variable.StringVar(nil, connectorVars, _const.VariableConnectorPassword)
		if err != nil {
			klog.V(4).InfoS("connector password is empty use public key")
		}
		// get private key path in connector variable. if empty, set default path: /root/.ssh/id_rsa.
		keyParam, err := variable.StringVar(nil, connectorVars, _const.VariableConnectorPrivateKey)
		if err != nil {
			klog.V(4).Infof("ssh public key is empty, use: %s", defaultSSHPrivateKey)
			keyParam = defaultSSHPrivateKey
		}
		return &sshConnector{
			Host:       hostParam,
			Port:       portParam,
			User:       userParam,
			Password:   passwdParam,
			PrivateKey: keyParam,
		}, nil
	case connectedKubernetes:
		kubeconfig, err := variable.StringVar(nil, connectorVars, _const.VariableConnectorKubeconfig)
		if err != nil && host != _const.VariableLocalHost {
			return nil, err
		}
		return &kubernetesConnector{Cmd: exec.New(), clusterName: host, kubeconfig: kubeconfig}, nil
	default:
		localHost, _ := os.Hostname()
		// get host in connector variable. if empty, set default host: host_name.
		hostParam, err := variable.StringVar(nil, connectorVars, _const.VariableConnectorHost)
		if err != nil {
			klog.V(4).Infof("connector host is empty use: %s", host)
			hostParam = host
		}
		if host == _const.VariableLocalHost || host == localHost || isLocalIP(hostParam) {
			return &localConnector{Cmd: exec.New()}, nil
		}
		// get port in connector variable. if empty, set default port: 22.
		portParam, err := variable.IntVar(nil, connectorVars, _const.VariableConnectorPort)
		if err != nil {
			klog.V(4).Infof("connector port is empty use: %v", defaultSSHPort)
			portParam = defaultSSHPort
		}
		// get user in connector variable. if empty, set default user: root.
		userParam, err := variable.StringVar(nil, connectorVars, _const.VariableConnectorUser)
		if err != nil {
			klog.V(4).Infof("connector user is empty use: %s", defaultSSHUser)
			userParam = defaultSSHUser
		}
		// get password in connector variable. if empty, should connector by private key.
		passwdParam, err := variable.StringVar(nil, connectorVars, _const.VariableConnectorPassword)
		if err != nil {
			klog.V(4).InfoS("connector password is empty use public key")
		}
		// get private key path in connector variable. if empty, set default path: /root/.ssh/id_rsa.
		keyParam, err := variable.StringVar(nil, connectorVars, _const.VariableConnectorPrivateKey)
		if err != nil {
			klog.V(4).Infof("ssh public key is empty, use: %s", defaultSSHPrivateKey)
			keyParam = defaultSSHPrivateKey
		}

		return &sshConnector{
			Host:       hostParam,
			Port:       portParam,
			User:       userParam,
			Password:   passwdParam,
			PrivateKey: keyParam,
		}, nil
	}
}

// GatherFacts get host info.
type GatherFacts interface {
	Info(ctx context.Context) (map[string]any, error)
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
