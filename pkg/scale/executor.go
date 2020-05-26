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

package scale

import (
	"fmt"
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	ssh "github.com/kubesphere/kubekey/pkg/util/ssh"
	log "github.com/sirupsen/logrus"
)

type Executor struct {
	cluster *kubekeyapi.ClusterSpec
	logger  *log.Logger
	Verbose bool
}

func NewExecutor(cluster *kubekeyapi.ClusterSpec, logger *log.Logger, verbose bool) *Executor {
	return &Executor{
		cluster: cluster,
		logger:  logger,
		Verbose: verbose,
	}
}

func (executor *Executor) Execute() error {
	mgr, err := executor.createManager()
	if err != nil {
		return err
	}
	return ExecTasks(mgr)
}

func (executor *Executor) createManager() (*manager.Manager, error) {
	mgr := &manager.Manager{}
	hostGroups := executor.cluster.GroupHosts()
	mgr.AllNodes = hostGroups.All
	mgr.EtcdNodes = hostGroups.Etcd
	mgr.MasterNodes = hostGroups.Master
	mgr.WorkerNodes = hostGroups.Worker
	mgr.K8sNodes = hostGroups.K8s
	mgr.Cluster = executor.cluster
	mgr.ClusterHosts = GenerateHosts(hostGroups, executor.cluster)
	mgr.Connector = ssh.NewConnector()
	mgr.Logger = executor.logger
	mgr.Verbose = executor.Verbose

	return mgr, nil
}

func GenerateHosts(hostGroups *kubekeyapi.HostGroups, cfg *kubekeyapi.ClusterSpec) []string {
	var lbHost string
	hostsList := []string{}

	if cfg.ControlPlaneEndpoint.Address != "" {
		lbHost = fmt.Sprintf("%s  %s", cfg.ControlPlaneEndpoint.Address, cfg.ControlPlaneEndpoint.Domain)
	} else {
		lbHost = fmt.Sprintf("%s  %s", hostGroups.Master[0].InternalAddress, cfg.ControlPlaneEndpoint.Domain)
	}

	for _, host := range cfg.Hosts {
		if host.Name != "" {
			hostsList = append(hostsList, fmt.Sprintf("%s  %s.%s %s", host.InternalAddress, host.Name, cfg.Kubernetes.ClusterName, host.Name))
		}
	}

	hostsList = append(hostsList, lbHost)
	return hostsList
}
