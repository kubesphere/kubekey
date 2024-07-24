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
func NewConnector(host string, vars map[string]any) (Connector, error) {
	switch vars["connector"] {
	case "local":
		return &localConnector{Cmd: exec.New()}, nil
	case "ssh":
		addressParam, err := variable.StringVar(nil, vars, "address")
		if err != nil {
			return nil, err
		}

		portParam, err := variable.IntVar(nil, vars, "port")
		if err != nil { // default port 22
			klog.InfoS("get ssh port failed use default port 22", "error", err)
			portParam = 22
		}

		userParam, err := variable.StringVar(nil, vars, "user")
		if err != nil {
			return nil, err
		}

		passParam, err := variable.StringVar(nil, vars, "password")
		if err != nil {
			return nil, err
		}
		return &sshConnector{
			Address:  addressParam,
			Port:     portParam,
			User:     userParam,
			Password: passParam,
		}, nil
	case "kubernetes":
		kubeconfig, err := variable.StringVar(nil, vars, "kubeconfig")
		if err != nil && host != _const.LocalHostName {
			return nil, err
		}
		return &kubernetesConnector{Cmd: exec.New(), clusterName: host, kubeconfig: kubeconfig}, nil
	default:
		localHost, _ := os.Hostname()
		hostParam, err := variable.StringVar(nil, vars, "address")
		if err != nil {
			klog.V(4).Infof("ssh_address is empty use: %s", host)
			hostParam = host
		}
		if host == _const.LocalHostName || localHost == host || isLocalIP(hostParam) {
			return &localConnector{Cmd: exec.New()}, nil
		}

		portParam, err := variable.IntVar(nil, vars, "port")
		if err != nil {
			klog.V(4).Infof("ssh_port is empty use: %v", defaultSSHPort)
			portParam = defaultSSHPort
		}

		userParam, err := variable.StringVar(nil, vars, "user")
		if err != nil {
			klog.V(4).Infof("ssh_user is empty use: %s", defaultSSHUser)
			userParam = defaultSSHUser
		}

		passParam, err := variable.StringVar(nil, vars, "password")
		if err != nil {
			klog.V(4).InfoS("ssh_password is empty use public key")
		}

		priParam, err := variable.StringVar(nil, vars, "key")
		if err != nil {
			klog.V(4).Infof("ssh public key is empty, use: %s", defaultSSHPrivateKey)
			priParam = defaultSSHPrivateKey
		}

		return &sshConnector{
			Address:    hostParam,
			Port:       portParam,
			User:       userParam,
			Password:   passParam,
			PrivateKey: priParam,
		}, nil
	}
}

// GatherFacts get host info.
type GatherFacts interface {
	Info(ctx context.Context) (map[string]any, error)
}

func isLocalIP(ipAddr string) bool {
	interfaces, err := net.Interfaces()
	if err != nil {
		klog.V(4).ErrorS(err, "get net interfaces error")
		return false
	}
	for _, i := range interfaces {
		addrs, err := i.Addrs()
		if err != nil {
			klog.V(4).ErrorS(err, "get address for net interface error", "interface", i.Name)
			continue
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
	}
	return false
}
