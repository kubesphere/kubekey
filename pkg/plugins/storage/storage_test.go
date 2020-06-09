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
package storage

import (
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/kubesphere/kubekey/pkg/util/executor"
	"testing"
)

func TestDeployNeonSANCSI(t *testing.T) {
	spec := &kubekeyapi.ClusterSpec{
		Hosts: []kubekeyapi.HostCfg{
			{
				Name:            "master",
				Address:         "192.168.0.17",
				InternalAddress: "192.168.0.17",
				PrivateKeyPath:  "/root/.ssh/id_rsa",
			},
			{
				Name:            "node1",
				Address:         "192.168.0.18",
				InternalAddress: "192.168.0.18",
				PrivateKeyPath:  "/root/.ssh/id_rsa",
			},
			{
				Name:            "node2",
				Address:         "192.168.0.19",
				InternalAddress: "192.168.0.19",
				PrivateKeyPath:  "/root/.ssh/id_rsa",
			},
		},
		RoleGroups: kubekeyapi.RoleGroups{
			Etcd:   []string{"master"},
			Master: []string{"master"},
			Worker: []string{"node1", "node2"},
		},
		ControlPlaneEndpoint: kubekeyapi.ControlPlaneEndpoint{},
		Kubernetes:           kubekeyapi.Kubernetes{},
		Network:              kubekeyapi.NetworkConfig{},
		Registry:             kubekeyapi.RegistryConfig{},
		Storage: kubekeyapi.Storage{
			NeonsanCSI: kubekeyapi.NeonsanCSI{
				Enable:  true,
				Pool:    "kube",
				Replica: 1,
				FsType:  "ext4",
			},
		},
		KubeSphere: kubekeyapi.KubeSphere{},
	}
	logger := util.InitLogger(true)
	exec := executor.NewExecutor(spec, logger, true)
	manager, err := exec.CreateManager()
	if err != nil {
		t.Error(err)
		return
	}
	err = DeployStoragePlugins(manager)
	if err != nil {
		t.Error(err)
		return
	}

}
