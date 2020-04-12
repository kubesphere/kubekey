package ssh

import (
	"sync"
	"time"

	"github.com/pkg/errors"

	kubekeyapi "github.com/pixiake/kubekey/apis/v1alpha1"
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
func (dialer *Dialer) Tunnel(host kubekeyapi.HostConfig) (Tunneler, error) {
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
func (dialer *Dialer) Connect(host kubekeyapi.HostConfig) (Connection, error) {
	var err error

	dialer.lock.Lock()
	defer dialer.lock.Unlock()

	conn, found := dialer.connections[host.ID]
	if !found {
		opts := SSHCfg{
			Username:    host.SSHUsername,
			Port:        host.SSHPort,
			Hostname:    host.PublicAddress,
			KeyFile:     host.SSHPrivateKeyFile,
			AgentSocket: host.SSHAgentSocket,
			Timeout:     10 * time.Second,
			Bastion:     host.Bastion,
			BastionPort: host.BastionPort,
			BastionUser: host.BastionUser,
		}

		conn, err = NewConnection(opts)
		if err != nil {
			return nil, err
		}

		dialer.connections[host.ID] = conn
	}

	return conn, nil
}
