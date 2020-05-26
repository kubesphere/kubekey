/*
Copyright 2020 The KubeSphere Authors.

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
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"

	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
)

// Connector holds a map of Connections
type Dialer struct {
	lock        sync.Mutex
	connections map[int]Connection
}

// NewConnector constructor
func NewConnector() *Dialer {
	return &Dialer{
		connections: make(map[int]Connection),
	}
}

// Tunnel returns established SSH tunnel
func (dialer *Dialer) Tunnel(host kubekeyapi.HostCfg) (Tunneler, error) {
	conn, err := dialer.Connect(host)
	if err != nil {
		return nil, err
	}

	tunn, ok := conn.(Tunneler)
	if !ok {
		err = errors.New("Unable to assert Tunneler")
	}

	return tunn, err
}

// Connect to the node
func (dialer *Dialer) Connect(host kubekeyapi.HostCfg) (Connection, error) {
	var err error

	dialer.lock.Lock()
	defer dialer.lock.Unlock()

	found := false
	conn, found := dialer.connections[host.ID]
	if !found {
		port, _ := strconv.Atoi(host.Port)
		opts := SSHCfg{
			Username: host.User,
			Port:     port,
			Address:  host.Address,
			Password: host.Password,
			KeyFile:  host.PrivateKeyPath,
			Timeout:  10 * time.Second,
		}

		conn, err = NewConnection(opts)
		if err != nil {
			return nil, err
		}

		dialer.connections[host.ID] = conn
	}

	return conn, nil
}
