package ssh

import (
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/experiment/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/experiment/core/connector"
	"sync"
	"time"
)

type Dialer struct {
	lock        sync.Mutex
	connections map[int]connector.Connection
}

func NewDialer() *Dialer {
	return &Dialer{
		connections: make(map[int]connector.Connection),
	}
}

func (d *Dialer) Connect(host kubekeyapiv1alpha1.HostCfg) (connector.Connection, error) {
	var err error

	d.lock.Lock()
	defer d.lock.Unlock()

	conn, ok := d.connections[host.Index]
	if !ok {
		opts := Cfg{
			Username:   host.User,
			Port:       host.Port,
			Address:    host.Address,
			Password:   host.Password,
			PrivateKey: host.PrivateKey,
			KeyFile:    host.PrivateKeyPath,
			Timeout:    30 * time.Second,
		}
		conn, err = NewConnection(d, opts)
		if err != nil {
			return nil, err
		}
		d.connections[host.Index] = conn
	}

	return conn, nil
}

func (d *Dialer) forgetConnection(conn *connection) {
	d.lock.Lock()
	defer d.lock.Unlock()

	for k := range d.connections {
		if d.connections[k] == conn {
			delete(d.connections, k)
		}
	}
}
