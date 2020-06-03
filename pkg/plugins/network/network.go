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
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/kubesphere/kubekey/pkg/util/ssh"
	"github.com/pkg/errors"
	"os/exec"
	"strings"
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
	var calicoFileBase64 string
	if util.IsExist(fmt.Sprintf("%s/calico.yaml", mgr.WorkDir)) {
		output, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("cat %s/calico.yaml | base64 --wrap=0", mgr.WorkDir)).CombinedOutput()
		if err != nil {
			fmt.Println(string(output))
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to read custom calico manifests: %s/calico.yaml", mgr.WorkDir))
		}
		calicoFileBase64 = strings.TrimSpace(string(output))
	} else {
		calicoFile, err := calico.GenerateCalicoFiles(mgr.Cluster)
		if err != nil {
			return err
		}
		calicoFileBase64 = base64.StdEncoding.EncodeToString([]byte(calicoFile))
	}

	_, err1 := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > /etc/kubernetes/calico.yaml\"", calicoFileBase64))
	if err1 != nil {
		return errors.Wrap(errors.WithStack(err1), "Failed to generate calico file")
	}

	_, err2 := mgr.Runner.RunCmdOutput("/usr/local/bin/kubectl apply -f /etc/kubernetes/calico.yaml")
	if err2 != nil {
		return errors.Wrap(errors.WithStack(err2), "Failed to deploy calico")
	}
	return nil
}

func deployFlannel(mgr *manager.Manager) error {
	var flannelFileBase64 string
	if util.IsExist(fmt.Sprintf("%s/flannel.yaml", mgr.WorkDir)) {
		output, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("cat %s/flannel.yaml | base64 --wrap=0", mgr.WorkDir)).CombinedOutput()
		if err != nil {
			fmt.Println(string(output))
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to read custom flannel manifests: %s/flannel.yaml", mgr.WorkDir))
		}
		flannelFileBase64 = strings.TrimSpace(string(output))
	} else {
		flannelFile, err := flannel.GenerateFlannelFiles(mgr.Cluster)
		if err != nil {
			return err
		}
		flannelFileBase64 = base64.StdEncoding.EncodeToString([]byte(flannelFile))
	}

	_, err1 := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > /etc/kubernetes/flannel.yaml\"", flannelFileBase64))
	if err1 != nil {
		return errors.Wrap(errors.WithStack(err1), "Failed to generate flannel file")
	}

	_, err2 := mgr.Runner.RunCmdOutput("/usr/local/bin/kubectl apply -f /etc/kubernetes/flannel.yaml")
	if err2 != nil {
		return errors.Wrap(errors.WithStack(err2), "Failed to deploy flannel")
	}
	return nil
}

func deployMacvlan(mgr *manager.Manager) error {
	return nil
}
