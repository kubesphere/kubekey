/*
 Copyright 2022 The KubeSphere Authors.

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

package bootstrap

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/pkg/errors"

	"github.com/kubesphere/kubekey/pkg/service/operation/directory"
	"github.com/kubesphere/kubekey/pkg/service/operation/file"
	"github.com/kubesphere/kubekey/pkg/util/filesystem"
)

//go:embed templates
var f embed.FS

// AddUsers adds a kube user to the Linux.
func (s *Service) AddUsers() error {
	userService := s.getUserService("kube", "Kubernetes user")

	// todo: if need to create a etcd user
	return userService.Add()
}

// SetHostname sets the hostname of the machine.
func (s *Service) SetHostname() error {
	if _, err := s.sshClient.SudoCmdf(
		"hostnamectl set-hostname %s && sed -i '/^127.0.1.1/s/.*/127.0.1.1      %s/g' /etc/hosts",
		s.instanceScope.HostName(),
		s.instanceScope.HostName()); err != nil {
		return errors.Wrapf(err, "failed to set host name [%s]", s.instanceScope.HostName())
	}
	return nil
}

// CreateDirectory creates some common directories.
func (s *Service) CreateDirectory() error {
	makeDirs := []string{
		directory.BinDir,
		directory.KubeConfigDir,
		directory.KubeCertDir,
		directory.KubeManifestDir,
		directory.KubeScriptDir,
		directory.KubeletFlexvolumesPluginsDir,
		"/var/lib/calico",
		"/etc/cni/net.d",
		"/opt/cni/bin",
	}
	for _, dir := range makeDirs {
		dirService := s.getDirectoryService(dir, os.FileMode(filesystem.FileMode0755))
		if err := dirService.Make(); err != nil {
			return err
		}
	}

	chownDirs := []string{
		directory.KubeConfigDir,
		directory.KubeCertDir,
		directory.KubeManifestDir,
		directory.KubeScriptDir,
		"/usr/libexec/kubernetes",
		"/etc/cni",
		"/opt/cni",
		"/var/lib/calico",
	}
	for _, dir := range chownDirs {
		dirService := s.getDirectoryService(dir, os.FileMode(filesystem.FileMode0755))
		if err := dirService.Chown("kube"); err != nil {
			return err
		}
	}
	return nil
}

// ResetTmpDirectory resets the temporary "/tmp/kubekey" directory.
func (s *Service) ResetTmpDirectory() error {
	dirService := s.getDirectoryService(directory.TmpDir, os.FileMode(filesystem.FileMode0777))
	if err := dirService.Remove(); err != nil {
		return err
	}
	if err := dirService.Make(); err != nil {
		return err
	}
	return nil
}

// ExecInitScript executes the init script on the remote instance.
func (s *Service) ExecInitScript() error {
	var (
		hostsList []string
		lbHost    string
	)

	for _, host := range s.scope.AllInstancesInfo() {
		if host.Name != "" {
			hostsList = append(hostsList, fmt.Sprintf("%s  %s.%s %s",
				host.InternalAddress,
				host.Name,
				s.scope.KubernetesClusterName(),
				host.Name))
		}
	}
	hostsList = append(hostsList, lbHost)

	temp, err := template.ParseFS(f, "templates/initOS.sh")
	if err != nil {
		return err
	}

	svc, err := s.getTemplateService(
		temp,
		file.Data{
			"Hosts": hostsList,
		},
		filepath.Join(directory.KubeScriptDir, temp.Name()))
	if err != nil {
		return err
	}
	if err := svc.RenderToLocal(); err != nil {
		return err
	}
	if err := svc.Copy(true); err != nil {
		return err
	}
	if err := svc.Chmod("+x"); err != nil {
		return err
	}
	if _, err := s.sshClient.SudoCmd(svc.RemotePath()); err != nil {
		return err
	}
	return nil
}

// ResetNetwork resets the network configuration.
func (s *Service) ResetNetwork() error {
	networkResetCmds := []string{
		"iptables -F",
		"iptables -X",
		"iptables -F -t nat",
		"iptables -X -t nat",
		"ipvsadm -C",
		"ip link del kube-ipvs0",
		"ip link del nodelocaldns",
	}
	for _, cmd := range networkResetCmds {
		_, _ = s.sshClient.SudoCmd(cmd)
	}
	return nil
}

// RemoveFiles removes some directories and files that may have been created by the Kubernetes and other related components.
func (s *Service) RemoveFiles() error {
	removeDirs := []string{
		directory.KubeConfigDir,
		directory.KubeScriptDir,
		"/var/log/calico",
		"/etc/cni",
		"/var/log/pods/",
		"/var/lib/cni",
		"/var/lib/calico",
		"/var/lib/kubelet",
		"/var/lib/rook",
		"/run/calico",
		"/run/flannel",
		"/etc/flannel",
		"/var/openebs",
		"/etc/systemd/system/kubelet.service",
		"/etc/systemd/system/kubelet.service.d",
		"/tmp/kubekey",
		"/etc/kubekey",
		"/run/kubeadm/*",
	}
	for _, dir := range removeDirs {
		dirService := s.getDirectoryService(dir, 0)
		_ = dirService.Remove()
	}
	return nil
}

// DaemonReload reloads the systemd daemon and restart the containerd.
func (s *Service) DaemonReload() error {
	_, _ = s.sshClient.SudoCmd("systemctl daemon-reload && exit 0")
	_, _ = s.sshClient.SudoCmd("systemctl restart containerd")
	return nil
}

// UninstallK3s uninstalls the k3s.
func (s *Service) UninstallK3s() error {
	if s.instanceScope.IsControlPlane() {
		_, _ = s.sshClient.SudoCmd("/usr/local/bin/k3s-uninstall.sh")
	} else {
		_, _ = s.sshClient.SudoCmd("/usr/local/bin/k3s-agent-uninstall.sh")
	}
	return nil
}
