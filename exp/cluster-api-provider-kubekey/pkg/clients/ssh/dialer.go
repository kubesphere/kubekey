/*
 Copyright 2022 The KubeSphere Authors.

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
	"sync"
	"time"

	infrav1 "github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/api/v1beta1"
)

type Dialer struct {
	lock    sync.Mutex
	clients map[string]Interface
}

func NewDialer() *Dialer {
	return &Dialer{
		clients: make(map[string]Interface),
	}
}

func (d *Dialer) Ping(host string, auth *infrav1.Auth, retry int) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	var err error
	client := NewClient(host, auth)
	for i := 0; i < retry; i++ {
		err = client.Ping()
		if err == nil {
			break
		}
		time.Sleep(time.Duration(i) * time.Second)
	}
	return err
}

func (d *Dialer) Connect(host string, auth *infrav1.Auth) (Interface, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	client, ok := d.clients[host]
	if !ok {
		client = NewClient(host, auth)
		if err := client.Connect(); err != nil {
			return nil, err
		}
		d.clients[host] = client
	}

	return client, nil
}

func (d *Dialer) Close(host string) {
	client, ok := d.clients[host]
	if !ok {
		return
	}

	client.Close()

	c := client.(*Client)
	d.forgetClient(c)
}

func (d *Dialer) forgetClient(client *Client) {
	d.lock.Lock()
	defer d.lock.Unlock()

	for k := range d.clients {
		if d.clients[k] == client {
			delete(d.clients, k)
		}
	}
}
