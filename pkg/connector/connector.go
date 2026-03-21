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
	"bytes"
	"context"
	"io"
	"io/fs"
	"net"
	"os"
	"strings"
	"text/template"

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
	connectedPrometheus = "prometheus"
)

// Connector is the interface for connecting to a remote host.
// It abstracts the operations required to interact with different types of hosts (e.g., SSH, local, Kubernetes, Prometheus).
// Implementations of this interface should provide mechanisms for initialization, cleanup, file transfer, and command execution.
type Connector interface {
	// Init initializes the connection.
	Init(ctx context.Context) error
	// Close closes the connection and releases any resources.
	Close(ctx context.Context) error
	// PutFile copies a file from src (as bytes) to dst (remote path) with the specified file mode.
	PutFile(ctx context.Context, src []byte, dst string, mode fs.FileMode) error
	// FetchFile copies a file from src (remote path) to dst (local writer).
	FetchFile(ctx context.Context, src string, dst io.Writer) error
	// ExecuteCommand executes a command on the remote host.
	// Returns stdout, stderr, and error (if any).
	ExecuteCommand(ctx context.Context, cmd string) ([]byte, []byte, error)
}

// NewConnector creates a new connector
// if set connector to "local", use local connector
// if set connector to "ssh", use ssh connector
// if set connector to "kubernetes", use kubernetes connector
// if set connector to "prometheus", use prometheus connector
// if connector is not set. when host is localhost, use local connector, else use ssh connector
// vars contains all inventory for host. It's best to define the connector info in inventory file.
func NewConnector(tpl *template.Template, host string, v variable.Variable) (Connector, error) {
	ha, err := v.Get(variable.GetAllVariable(host))
	if err != nil {
		return nil, err
	}
	vd, ok := ha.(map[string]any)
	if !ok {
		return nil, errors.Errorf("host: %s variable is not a map", host)
	}

	workdir, err := v.Get(variable.GetWorkDir())
	if err != nil {
		return nil, err
	}
	wd, ok := workdir.(string)
	if !ok {
		return nil, errors.New("workdir in variable should be string")
	}

	connectedType, _ := variable.StringVar(tpl, nil, vd, _const.VariableConnector, _const.VariableConnectorType)
	switch connectedType {
	case connectedLocal:
		return newLocalConnector(tpl, wd, vd), nil
	case connectedSSH:
		return newSSHConnector(tpl, wd, host, vd), nil
	case connectedKubernetes:
		return newKubernetesConnector(tpl, host, wd, vd)
	case connectedPrometheus:
		return newPrometheusConnector(tpl, vd), nil
	default:
		localHost, _ := os.Hostname()
		// get host in connector variable. if empty, set default host: host_name.
		hostParam, err := variable.StringVar(tpl, nil, vd, _const.VariableConnector, _const.VariableConnectorHost)
		if err != nil {
			klog.V(4).InfoS("connector host is empty, using provided host", "host", host)
			hostParam = host
		}
		if host == _const.VariableLocalHost || host == localHost || isLocalIP(hostParam) {
			return newLocalConnector(tpl, wd, vd), nil
		}

		return newSSHConnector(tpl, wd, host, vd), nil
	}
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

func parseHostInfo(c Connector, ctx context.Context) (map[string]any, error) {
	// os information
	osVars := make(map[string]any)
	var buf bytes.Buffer
	if err := c.FetchFile(ctx, "/etc/os-release", &buf); err != nil {
		return nil, err
	}
	osVars[_const.VariableOSRelease] = convertBytesToMap(buf.Bytes(), "=")
	buf.Reset()
	if err := c.FetchFile(ctx, "/proc/sys/kernel/hostname", &buf); err != nil {
		return nil, errors.Wrap(err, "failed to get hostname")
	}
	osVars[_const.VariableOSHostName] = string(bytes.TrimSpace(buf.Bytes()))
	buf.Reset()
	if err := c.FetchFile(ctx, "/proc/version", &buf); err != nil {
		return nil, err
	}
	versionParts := bytes.Split(buf.Bytes(), []byte(" "))
	if len(versionParts) < 3 {
		return nil, errors.New("failed to parse kernel version from /proc/version")
	}
	osVars[_const.VariableOSKernelVersion] = string(bytes.TrimSpace(versionParts[2]))
	matches := archRegex.FindStringSubmatch(buf.String())
	if len(matches) == 0 {
		return nil, errors.New("failed to get arch")
	}
	osVars[_const.VariableOSArchitecture] = strings.TrimSpace(matches[0])

	// process information
	procVars := make(map[string]any)
	buf.Reset()
	if err := c.FetchFile(ctx, "/proc/cpuinfo", &buf); err != nil {
		return nil, err
	}
	procVars[_const.VariableProcessCPU] = convertBytesToSlice(buf.Bytes(), ":")
	buf.Reset()
	if err := c.FetchFile(ctx, "/proc/meminfo", &buf); err != nil {
		return nil, err
	}
	procVars[_const.VariableProcessMemory] = convertBytesToMap(buf.Bytes(), ":")

	// persistence the hostInfo

	return map[string]any{
		_const.VariableOS:      osVars,
		_const.VariableProcess: procVars,
	}, nil
}
