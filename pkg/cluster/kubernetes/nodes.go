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
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/api/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/cluster/kubernetes/tmpl"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strings"
)

func InstallKubeBinaries(mgr *manager.Manager) error {
	mgr.Logger.Infoln("Installing kube binaries")
	return mgr.RunTaskOnK8sNodes(installKubeBinaries, true)
}

func installKubeBinaries(mgr *manager.Manager, node *kubekeyapiv1alpha1.HostCfg) error {
	if !ExistNodeName(node.Name) {
		if err := SyncKubeBinaries(mgr, node); err != nil {
			return err
		}

		if err := SetKubelet(mgr, node); err != nil {
			return err
		}
	}
	return nil
}

func ExistNodeName(nodename string) bool {
	_, ok := allNodesName[nodename]
	return ok
}

func SyncKubeBinaries(mgr *manager.Manager, node *kubekeyapiv1alpha1.HostCfg) error {

	tmpDir := "/tmp/kubekey"
	_, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"if [ -d %s ]; then rm -rf %s ;fi\" && mkdir -p %s", tmpDir, tmpDir, tmpDir), 1, false)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to create tmp dir")
	}

	currentDir, err1 := filepath.Abs(filepath.Dir(os.Args[0]))
	if err1 != nil {
		return errors.Wrap(err1, "Failed to get current dir")
	}

	filesDir := fmt.Sprintf("%s/%s/%s/%s", currentDir, kubekeyapiv1alpha1.DefaultPreDir, mgr.Cluster.Kubernetes.Version, node.Arch)

	kubeadm := "kubeadm"
	kubelet := "kubelet"
	kubectl := "kubectl"
	helm := "helm"
	kubecni := fmt.Sprintf("cni-plugins-linux-%s-%s.tgz", node.Arch, kubekeyapiv1alpha1.DefaultCniVersion)
	binaryList := []string{kubeadm, kubelet, kubectl, helm, kubecni}

	var cmdlist []string

	for _, binary := range binaryList {
		if err := mgr.Runner.ScpFile(fmt.Sprintf("%s/%s", filesDir, binary), fmt.Sprintf("%s/%s", "/tmp/kubekey", binary)); err != nil {
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to sync binaries"))
		}

		if strings.Contains(binary, "cni-plugins-linux") {
			cmdlist = append(cmdlist, fmt.Sprintf("mkdir -p /opt/cni/bin && tar -zxf %s/%s -C /opt/cni/bin", "/tmp/kubekey", binary))
		} else if strings.Contains(binary, "kubelet") {
			continue
		} else {
			cmdlist = append(cmdlist, fmt.Sprintf("cp -f /tmp/kubekey/%s /usr/local/bin/%s && chmod +x /usr/local/bin/%s", binary, binary, binary))
		}
	}
	cmd := strings.Join(cmdlist, " && ")
	if _, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"%s\"", cmd), 2, false); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to create kubelet link"))
	}

	return nil
}

func SetKubelet(mgr *manager.Manager, node *kubekeyapiv1alpha1.HostCfg) error {

	if _, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"%s\"", "cp -f /tmp/kubekey/kubelet /usr/local/bin/kubelet && chmod +x /usr/local/bin/kubelet"), 2, false); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to create kubelet link"))
	}

	kubeletService, err1 := tmpl.GenerateKubeletService()
	if err1 != nil {
		return err1
	}
	kubeletServiceBase64 := base64.StdEncoding.EncodeToString([]byte(kubeletService))
	if _, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > /etc/systemd/system/kubelet.service\"", kubeletServiceBase64), 5, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to generate kubelet service")
	}

	if _, err := mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"systemctl enable kubelet && ln -snf /usr/local/bin/kubelet /usr/bin/kubelet\"", 5, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to enable kubelet service")
	}

	kubeletEnv, err3 := tmpl.GenerateKubeletEnv(node)
	if err3 != nil {
		return err3
	}
	kubeletEnvBase64 := base64.StdEncoding.EncodeToString([]byte(kubeletEnv))
	if _, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"mkdir -p /etc/systemd/system/kubelet.service.d && echo %s | base64 -d > /etc/systemd/system/kubelet.service.d/10-kubeadm.conf\"", kubeletEnvBase64), 2, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to generate kubelet env")
	}

	return nil
}
