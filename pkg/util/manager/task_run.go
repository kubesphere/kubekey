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

package manager

import (
	"github.com/kubesphere/kubekey/pkg/util/ssh"
	"sync"
	"time"

	"github.com/pkg/errors"

	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/util/runner"
)

const (
	DefaultCon = 10
	Timeout    = 1800
)

// NodeTask is a task that is specifically tailored to run on a single node.
type NodeTask func(mgr *Manager, node *kubekeyapi.HostCfg, conn ssh.Connection) error

func (mgr *Manager) runTask(node *kubekeyapi.HostCfg, task NodeTask, index int) error {
	var (
		err  error
		conn ssh.Connection
	)

	conn, err = mgr.Connector.Connect(*node)
	if err != nil {
		return errors.Wrapf(err, "Failed to connect to %s", node.Address)
	}

	mgr.Runner = &runner.Runner{
		Conn:  conn,
		Debug: mgr.Debug,
		Host:  node,
		Index: index,
	}

	return task(mgr, node, conn)
}

func (mgr *Manager) RunTaskOnNodes(nodes []kubekeyapi.HostCfg, task NodeTask, parallel bool) error {
	var err error
	hasErrors := false

	wg := &sync.WaitGroup{}
	result := make(chan string)
	ccons := make(chan struct{}, DefaultCon)
	defer close(result)
	defer close(ccons)
	hostNum := len(nodes)

	if parallel {
		go func(result chan string, ccons chan struct{}) {
			for i := 0; i < hostNum; i++ {
				select {
				case <-result:
				case <-time.After(time.Second * Timeout):
					mgr.Logger.Fatalf("Execute task timeout, Timeout=%ds", Timeout)
				}
				wg.Done()
				<-ccons
			}
		}(result, ccons)
	}

	for i := range nodes {
		mgrTask := mgr.Copy()
		mgrTask.Logger = mgrTask.Logger.WithField("node", nodes[i].Address)

		if parallel {
			ccons <- struct{}{}
			wg.Add(1)
			go func(mgr *Manager, node *kubekeyapi.HostCfg, result chan string, index int) {
				err = mgr.runTask(node, task, index)
				if err != nil {
					mgr.Logger.Error(err)
					hasErrors = true
				}
				result <- "done"
			}(mgrTask, &nodes[i], result, i)
		} else {
			err = mgrTask.runTask(&nodes[i], task, i)
			if err != nil {
				break
			}
		}
	}

	wg.Wait()

	if hasErrors {
		err = errors.New("interrupted by error")
	}

	return err
}

func (mgr *Manager) RunTaskOnAllNodes(task NodeTask, parallel bool) error {
	if err := mgr.RunTaskOnNodes(mgr.AllNodes, task, parallel); err != nil {
		return err
	}
	return nil
}

func (mgr *Manager) RunTaskOnEtcdNodes(task NodeTask, parallel bool) error {
	if err := mgr.RunTaskOnNodes(mgr.EtcdNodes, task, parallel); err != nil {
		return err
	}
	return nil
}

func (mgr *Manager) RunTaskOnMasterNodes(task NodeTask, parallel bool) error {
	if err := mgr.RunTaskOnNodes(mgr.MasterNodes, task, parallel); err != nil {
		return err
	}
	return nil
}

func (mgr *Manager) RunTaskOnWorkerNodes(task NodeTask, parallel bool) error {
	if err := mgr.RunTaskOnNodes(mgr.WorkerNodes, task, parallel); err != nil {
		return err
	}
	return nil
}

func (mgr *Manager) RunTaskOnK8sNodes(task NodeTask, parallel bool) error {
	if err := mgr.RunTaskOnNodes(mgr.K8sNodes, task, parallel); err != nil {
		return err
	}
	return nil
}
