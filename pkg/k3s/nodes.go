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

package k3s

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/k3s/config"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/pkg/errors"
)

// InstallKubeBinaries is used to install kubernetes' binaries to os' PATH.
func InstallKubeBinaries(mgr *manager.Manager, node *kubekeyapiv1alpha1.HostCfg) error {
	if !ExistNode(node) {
		if err := SyncKubeBinaries(mgr, node); err != nil {
			return err
		}

		if err := SetK3s(mgr); err != nil {
			return err
		}
	}
	return nil
}

// ExistNode is used determine if the node already exists.
func ExistNode(node *kubekeyapiv1alpha1.HostCfg) bool {
	var version bool
	_, name := allNodesInfo[node.Name]
	if name && allNodesInfo[node.Name] != "" {
		version = true
	}
	_, ip := allNodesInfo[node.InternalAddress]
	return version || ip
}

// SyncKubeBinaries is used to sync kubernetes' binaries to each node.
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

	k3s := "k3s"
	helm := "helm"
	kubecni := fmt.Sprintf("cni-plugins-linux-%s-%s.tgz", node.Arch, kubekeyapiv1alpha1.DefaultCniVersion)
	binaryList := []string{k3s, helm, kubecni}

	var cmdlist []string
	for _, binary := range binaryList {
		if err := mgr.Runner.ScpFile(fmt.Sprintf("%s/%s", filesDir, binary), fmt.Sprintf("%s/%s", "/tmp/kubekey", binary)); err != nil {
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to sync binaries"))
		}

		if strings.Contains(binary, "cni-plugins-linux") {
			cmdlist = append(cmdlist, fmt.Sprintf("mkdir -p /opt/cni/bin && tar -zxf %s/%s -C /opt/cni/bin", "/tmp/kubekey", binary))
		}
	}

	cmd := strings.Join(cmdlist, " && ")
	if _, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"%s\"", cmd), 2, false); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to install kube cni"))
	}

	return nil
}

// SetK3s is used to configure the kubelet's startup parameters.
func SetK3s(mgr *manager.Manager) error {

	if _, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"%s\"", "cp -f /tmp/kubekey/k3s /usr/local/bin/k3s && chmod +x /usr/local/bin/k3s"), 2, false); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to create kubelet link"))
	}

	k3sService, err1 := config.GenerateK3sService()
	if err1 != nil {
		return err1
	}
	k3sServiceBase64 := base64.StdEncoding.EncodeToString([]byte(k3sService))
	if _, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > /etc/systemd/system/k3s.service\"", k3sServiceBase64), 5, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to generate kubelet service")
	}

	binaries := []string{"kubectl", "crictl", "ctr"}
	createLinkCmds := []string{}
	for _, binary := range binaries {
		createLinkCmds = append(createLinkCmds, fmt.Sprintf("ln -snf /usr/local/bin/k3s /usr/local/bin/%s", binary))
	}

	if _, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"%s\"", strings.Join(createLinkCmds, " && ")), 5, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to create link")
	}

	if _, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"%s\"", "cp -f /tmp/kubekey/helm /usr/local/bin/helm && chmod +x /usr/local/bin/helm"), 2, false); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to install helm"))
	}

	tmpDir := "/tmp/kubekey"
	killallScript, err := config.GenerateK3sKillallScript()
	if err != nil {
		return err
	}

	if _, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("echo %s | base64 -d > %s/k3s-killall.sh && chmod +x %s/k3s-killall.sh", base64.StdEncoding.EncodeToString([]byte(killallScript)), tmpDir, tmpDir), 1, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to generate k3s-killall script")
	}

	if _, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo cp %s/k3s-killall.sh %s", tmpDir, "/usr/local/bin"), 1, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to sync k3s-killall.sh")
	}

	uninstallScript, err := config.GenerateK3sUninstallScript()
	if err != nil {
		return err
	}
	if _, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("echo %s | base64 -d > %s/k3s-uninstall.sh && chmod +x %s/k3s-uninstall.sh", base64.StdEncoding.EncodeToString([]byte(uninstallScript)), tmpDir, tmpDir), 1, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to generate k3s-uninstall script")
	}

	if _, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo cp %s/k3s-uninstall.sh %s", tmpDir, "/usr/local/bin"), 1, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to sync k3s-uninstall")
	}

	return nil
}
