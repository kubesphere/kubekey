package state

import (
	"fmt"
	"sync"

	"github.com/pkg/errors"

	kubekeyapi "github.com/pixiake/kubekey/apis/v1alpha1"
	"github.com/pixiake/kubekey/util/dialer/ssh"
	"github.com/pixiake/kubekey/util/runner"
)

// NodeTask is a task that is specifically tailored to run on a single node.
type NodeTask func(ctx *State, node *kubekeyapi.HostCfg, conn ssh.Connection) error

func (s *State) runTask(node *kubekeyapi.HostCfg, task NodeTask, prefixed bool) error {
	var (
		err  error
		conn ssh.Connection
	)

	// connect to the host (and do not close connection
	// because we want to re-use it for future task)
	conn, err = s.Connector.Connect(*node)
	if err != nil {
		return errors.Wrapf(err, "failed to connect to %s", node.Address)
	}

	prefix := ""
	if prefixed {
		prefix = fmt.Sprintf("[%s] ", node.Address)
	}

	s.Runner = &runner.Runner{
		Conn:    conn,
		Verbose: s.Verbose,
		//OS:      node.OS,
		Prefix: prefix,
	}

	return task(s, node, conn)
}

// RunTaskOnNodes runs the given task on the given selection of hosts.
func (s *State) RunTaskOnNodes(nodes []kubekeyapi.HostCfg, task NodeTask, parallel bool) error {
	var err error

	wg := sync.WaitGroup{}
	hasErrors := false

	for i := range nodes {
		ctx := s.Clone()
		ctx.Logger = ctx.Logger.WithField("node", nodes[i].Address)

		if parallel {
			wg.Add(1)
			go func(ctx *State, node *kubekeyapi.HostCfg) {
				err = ctx.runTask(node, task, parallel)
				if err != nil {
					ctx.Logger.Error(err)
					hasErrors = true
				}
				wg.Done()
			}(ctx, &nodes[i])
		} else {
			err = ctx.runTask(&nodes[i], task, parallel)
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
func (s *State) RunTaskOnAllNodes(task NodeTask, parallel bool) error {
	// It's not possible to concatenate host lists in this function.
	// Some of the task(determineOS, determineHostname) write to the state and sending a copy would break that.
	if err := s.RunTaskOnNodes(s.Cluster.Hosts, task, parallel); err != nil {
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
