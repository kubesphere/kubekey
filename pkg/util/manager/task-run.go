package manager

import (
	"fmt"
	"github.com/pixiake/kubekey/pkg/util/ssh"
	"sync"
	"time"

	"github.com/pkg/errors"

	kubekeyapi "github.com/pixiake/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/pixiake/kubekey/pkg/util/runner"
)

const (
	DefaultCon = 10
	Timeout    = 600
)

// NodeTask is a task that is specifically tailored to run on a single node.
type NodeTask func(mgr *Manager, node *kubekeyapi.HostCfg, conn ssh.Connection) error

func (mgr *Manager) runTask(node *kubekeyapi.HostCfg, task NodeTask, prefixed bool, index int) error {
	var (
		err  error
		conn ssh.Connection
	)
	// connect to the host (and do not close connection
	// because we want to re-use it for future task)
	conn, err = mgr.Connector.Connect(*node)
	if err != nil {
		return errors.Wrapf(err, "failed to connect to %s", node.SSHAddress)
	}

	prefix := ""
	if prefixed {
		prefix = fmt.Sprintf("[%s] ", node.HostName)
	}

	mgr.Runner = &runner.Runner{
		Conn:    conn,
		Verbose: mgr.Verbose,
		Prefix:  prefix,
		Host:    node,
		Index:   index,
	}

	return task(mgr, node, conn)
}

// RunTaskOnNodes runs the given task on the given selection of hosts.
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
					fmt.Sprintf("getSSHClient error,SSH-Read-TimeOut,Timeout=%ds", Timeout)
				}
				wg.Done()
				<-ccons
			}
		}(result, ccons)
	}

	for i := range nodes {
		mgrTask := mgr.Clone()
		mgrTask.Logger = mgrTask.Logger.WithField("node", nodes[i].SSHAddress)

		if parallel {
			ccons <- struct{}{}
			wg.Add(1)
			go func(mgr *Manager, node *kubekeyapi.HostCfg, result chan string, index int) {
				err = mgr.runTask(node, task, parallel, index)
				if err != nil {
					mgr.Logger.Error(err)
					hasErrors = true
				}
				result <- "done"
			}(mgrTask, &nodes[i], result, i)
		} else {
			err = mgrTask.runTask(&nodes[i], task, parallel, i)
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
	if err := mgr.RunTaskOnNodes(mgr.AllNodes.Hosts, task, parallel); err != nil {
		return err
	}
	return nil
}

func (mgr *Manager) RunTaskOnEtcdNodes(task NodeTask, parallel bool) error {
	if err := mgr.RunTaskOnNodes(mgr.EtcdNodes.Hosts, task, parallel); err != nil {
		return err
	}
	return nil
}

func (mgr *Manager) RunTaskOnMasterNodes(task NodeTask, parallel bool) error {
	if err := mgr.RunTaskOnNodes(mgr.MasterNodes.Hosts, task, parallel); err != nil {
		return err
	}
	return nil
}

func (mgr *Manager) RunTaskOnWorkerNodes(task NodeTask, parallel bool) error {
	if err := mgr.RunTaskOnNodes(mgr.WorkerNodes.Hosts, task, parallel); err != nil {
		return err
	}
	return nil
}

func (mgr *Manager) RunTaskOnK8sNodes(task NodeTask, parallel bool) error {
	if err := mgr.RunTaskOnNodes(mgr.K8sNodes.Hosts, task, parallel); err != nil {
		return err
	}
	return nil
}

func (mgr *Manager) RunTaskOnClientNode(task NodeTask, parallel bool) error {
	if err := mgr.RunTaskOnNodes(mgr.ClientNode.Hosts, task, parallel); err != nil {
		return err
	}
	return nil
}
