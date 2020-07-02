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

package kubernetes

import (
	"encoding/base64"
	"fmt"
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/cluster/kubernetes/tmpl"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/kubesphere/kubekey/pkg/util/ssh"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strings"
)

func SyncKubeBinaries(mgr *manager.Manager) error {
	mgr.Logger.Infoln("Synchronizing kube binaries")
	return mgr.RunTaskOnK8sNodes(syncKubeBinaries, true)
}

func syncKubeBinaries(mgr *manager.Manager, node *kubekeyapi.HostCfg, _ ssh.Connection) error {
	if !strings.Contains(clusterStatus["clusterInfo"], node.Name) && !strings.Contains(clusterStatus["clusterInfo"], node.InternalAddress) {

		currentDir, err1 := filepath.Abs(filepath.Dir(os.Args[0]))
		if err1 != nil {
			return errors.Wrap(err1, "Failed to get current dir")
		}

		filesDir := fmt.Sprintf("%s/%s/%s/%s", currentDir, kubekeyapi.DefaultPreDir, mgr.Cluster.Kubernetes.Version, node.Arch)

		kubeadm := fmt.Sprintf("kubeadm")
		kubelet := fmt.Sprintf("kubelet")
		kubectl := fmt.Sprintf("kubectl")
		helm := fmt.Sprintf("helm")
		kubecni := fmt.Sprintf("cni-plugins-linux-%s-%s.tgz", node.Arch, kubekeyapi.DefaultCniVersion)
		binaryList := []string{kubeadm, kubelet, kubectl, helm, kubecni}

		var cmdlist []string

		for _, binary := range binaryList {
			err2 := mgr.Runner.ScpFile(fmt.Sprintf("%s/%s", filesDir, binary), fmt.Sprintf("%s/%s", "/tmp/kubekey", binary))
			if err2 != nil {
				return errors.Wrap(errors.WithStack(err2), fmt.Sprintf("Failed to sync binaries"))
			}

			if strings.Contains(binary, "cni-plugins-linux") {
				cmdlist = append(cmdlist, fmt.Sprintf("mkdir -p /opt/cni/bin && tar -zxf %s/%s -C /opt/cni/bin", "/tmp/kubekey", binary))
			} else {
				cmdlist = append(cmdlist, fmt.Sprintf("cp -f /tmp/kubekey/%s /usr/local/bin/%s && chmod +x /usr/local/bin/%s", binary, strings.Split(binary, "-")[0], strings.Split(binary, "-")[0]))
			}
		}
		cmd := strings.Join(cmdlist, " && ")
		_, err3 := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"%s\"", cmd), 2, false)
		if err3 != nil {
			return errors.Wrap(errors.WithStack(err3), fmt.Sprintf("Failed to create kubelet link"))
		}

		if err := setKubelet(mgr, node); err != nil {
			return err
		}
	}
	return nil
}

func setKubelet(mgr *manager.Manager, node *kubekeyapi.HostCfg) error {

	kubeletService, err1 := tmpl.GenerateKubeletService()
	if err1 != nil {
		return err1
	}
	kubeletServiceBase64 := base64.StdEncoding.EncodeToString([]byte(kubeletService))
	_, err2 := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > /etc/systemd/system/kubelet.service && systemctl enable kubelet && ln -snf /usr/local/bin/kubelet /usr/bin/kubelet\"", kubeletServiceBase64), 5, false)
	if err2 != nil {
		return errors.Wrap(errors.WithStack(err2), "Failed to generate kubelet service")
	}

	kubeletEnv, err3 := tmpl.GenerateKubeletEnv(node)
	if err3 != nil {
		return err3
	}
	kubeletEnvBase64 := base64.StdEncoding.EncodeToString([]byte(kubeletEnv))
	_, err4 := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"mkdir -p /etc/systemd/system/kubelet.service.d && echo %s | base64 -d > /etc/systemd/system/kubelet.service.d/10-kubeadm.conf\"", kubeletEnvBase64), 2, false)
	if err4 != nil {
		return errors.Wrap(errors.WithStack(err2), "Failed to generate kubelet env")
	}

	return nil
}
