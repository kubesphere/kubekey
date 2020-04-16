package manager

import (
	"fmt"
	"github.com/pixiake/kubekey/util/ssh"
	"sync"
	"time"

	"github.com/pkg/errors"

	kubekeyapi "github.com/pixiake/kubekey/apis/v1alpha1"
	"github.com/pixiake/kubekey/util/runner"
)

const (
	DefaultCon = 10
	Timeout    = 600
)

// NodeTask is a task that is specifically tailored to run on a single node.
type NodeTask func(mgr *Manager, node *kubekeyapi.HostCfg, conn ssh.Connection) error

func (mgr *Manager) runTask(node *kubekeyapi.HostCfg, task NodeTask, prefixed bool) error {
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
		//OS:      node.OS,
		Prefix: prefix,
		Host:   node,
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
			go func(mgr *Manager, node *kubekeyapi.HostCfg, result chan string) {
				err = mgr.runTask(node, task, parallel)
				if err != nil {
					mgr.Logger.Error(err)
					hasErrors = true
				}
				result <- "done"
			}(mgrTask, &nodes[i], result)
		} else {
			err = mgrTask.runTask(&nodes[i], task, parallel)
			if err != nil {
				break
			}
		}
	}

	wg.Wait()

	if hasErrors {
		err = errors.New("at least one of the task has encountered an error")
	}

	return err
}

// RunTaskOnAllNodes runs the given task on all hosts.
func (mgr *Manager) RunTaskOnAllNodes(task NodeTask, parallel bool) error {
	// It's not possible to concatenate host lists in this function.
	// Some of the task(determineOS, determineHostname) write to the state and sending a copy would break that.
	if err := mgr.RunTaskOnNodes(mgr.Cluster.Hosts, task, parallel); err != nil {
		return err
	}
	//if s.Cluster.StaticWorkers != nil {
	//	return s.RunTaskOnNodes(s.Cluster.StaticWorkers, task, parallel)
	//}
	return nil
}

// RunTaskOnLeader runs the given task on the leader host.
//func (s *State) RunTaskOnLeader(task NodeTask) error {
//	leader, err := s.Cluster.Leader()
//	if err != nil {
//		return err
//	}
//
//	hosts := []kubekeyapi.HostConfig{
//		leader,
//	}
//
//	return s.RunTaskOnNodes(hosts, task, false)
//}

// RunTaskOnFollowers runs the given task on the follower hosts.
//func (s *State) RunTaskOnFollowers(task NodeTask, parallel bool) error {
//	return s.RunTaskOnNodes(s.Cluster.Followers(), task, parallel)
//}
//
//func (s *State) RunTaskOnStaticWorkers(task NodeTask, parallel bool) error {
//	return s.RunTaskOnNodes(s.Cluster.StaticWorkers, task, parallel)
//}
