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
		err = errors.New("unable to assert Tunneler")
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
			Address:  host.SSHAddress,
			Password: host.Password,
			KeyFile:  host.SSHKeyPath,
			//AgentSocket: host.SSHAgentSocket,
			Timeout: 10 * time.Second,
			//Bastion:     host.Bastion,
			//BastionPort: host.BastionPort,
			//BastionUser: host.BastionUser,
		}

		conn, err = NewConnection(opts)
		if err != nil {
			return nil, err
		}

		dialer.connections[host.ID] = conn
	}

	return conn, nil
}
