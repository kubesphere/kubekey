package connector

import (
	"sync"
	"time"
)

type Dialer struct {
	lock        sync.Mutex
	connections map[int]Connection
}

func NewDialer() *Dialer {
	return &Dialer{
		connections: make(map[int]Connection),
	}
}

func (d *Dialer) Connect(host Host) (Connection, error) {
	var err error

	d.lock.Lock()
	defer d.lock.Unlock()

	conn, ok := d.connections[host.GetIndex()]
	if !ok {
		opts := Cfg{
			Username:   host.GetUser(),
			Port:       host.GetPort(),
			Address:    host.GetAddress(),
			Password:   host.GetPassword(),
			PrivateKey: host.GetPrivateKey(),
			KeyFile:    host.GetPrivateKeyPath(),
			Timeout:    30 * time.Second,
		}
		conn, err = NewConnection(d, opts)
		if err != nil {
			return nil, err
		}
		d.connections[host.GetIndex()] = conn
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
