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

package network

import (
	"encoding/base64"
	"fmt"
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/plugins/network/calico"
	"github.com/kubesphere/kubekey/pkg/plugins/network/flannel"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/kubesphere/kubekey/pkg/util/ssh"
	"github.com/pkg/errors"
)

func DeployNetworkPlugin(mgr *manager.Manager) error {
	mgr.Logger.Infoln("Deploying network plugin ...")

	return mgr.RunTaskOnMasterNodes(deployNetworkPlugin, true)
}

func deployNetworkPlugin(mgr *manager.Manager, node *kubekeyapi.HostCfg, conn ssh.Connection) error {
	if mgr.Runner.Index == 0 {
		switch mgr.Cluster.Network.Plugin {
		case "calico":
			if err := deployCalico(mgr, node); err != nil {
				return err
			}
		case "flannel":
			if err := deployFlannel(mgr); err != nil {
				return err
			}
		case "macvlan":
			if err := deployMacvlan(mgr); err != nil {
				return err
			}
		default:
			return errors.New(fmt.Sprintf("This network plugin is not supported: %s", mgr.Cluster.Network.Plugin))
		}
	}
	return nil
}

func deployCalico(mgr *manager.Manager, node *kubekeyapi.HostCfg) error {
	calicoFile, err := calico.GenerateCalicoFiles(mgr.Cluster)
	if err != nil {
		return err
	}
	calicoFileBase64 := base64.StdEncoding.EncodeToString([]byte(calicoFile))
	_, err1 := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > /etc/kubernetes/calico.yaml\"", calicoFileBase64))
	if err1 != nil {
		return errors.Wrap(errors.WithStack(err1), "Failed to generate calico file")
	}

	_, err2 := mgr.Runner.RunCmd("/usr/local/bin/kubectl apply -f /etc/kubernetes/calico.yaml")
	if err2 != nil {
		return errors.Wrap(errors.WithStack(err2), "Failed to deploy calico")
	}
	return nil
}

func deployFlannel(mgr *manager.Manager) error {
	flannelFile, err := flannel.GenerateFlannelFiles(mgr.Cluster)
	if err != nil {
		return err
	}
	flannelFileBase64 := base64.StdEncoding.EncodeToString([]byte(flannelFile))
	_, err1 := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > /etc/kubernetes/flannel.yaml\"", flannelFileBase64))
	if err1 != nil {
		return errors.Wrap(errors.WithStack(err1), "Failed to generate flannel file")
	}

	_, err2 := mgr.Runner.RunCmd("/usr/local/bin/kubectl apply -f /etc/kubernetes/flannel.yaml")
	if err2 != nil {
		return errors.Wrap(errors.WithStack(err2), "Failed to deploy flannel")
	}
	return nil
}

func deployMacvlan(mgr *manager.Manager) error {
	return nil
}
