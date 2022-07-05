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
	"fmt"
	"os"
	"path/filepath"

	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/service/bootstrap/templates"
	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/service/operation/directory"
	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/service/operation/file"
	"github.com/kubesphere/kubekey/pkg/common"
)

func (s *Service) AddUsers() error {
	userService := s.getUserService("kube", "Kubernetes user")

	// todo: if need to create a etcd user
	return userService.Add()
}

func (s *Service) CreateDirectory() error {
	makeDirs := []string{
		directory.BinDir,
		directory.KubeConfigDir,
		directory.KubeCertDir,
		directory.KubeManifestDir,
		directory.KubeScriptDir,
		common.KubeletFlexvolumesPluginsDir,
		"/var/lib/calico",
		"/etc/cni/net.d",
		"/opt/cni/bin",
	}
	for _, dir := range makeDirs {
		dirService := s.getDirectoryFactory(dir, os.ModeDir)
		if err := dirService.Make(); err != nil {
			return err
		}
	}

	chownDirs := []string{
		directory.BinDir,
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
		dirService := s.getDirectoryFactory(dir, os.ModeDir)
		if err := dirService.Chown("kube"); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) ResetTmpDirectory() error {
	dirService := s.getDirectoryFactory(directory.TmpDir, os.ModeDir)
	if err := dirService.Remove(); err != nil {
		return err
	}
	if err := dirService.Make(); err != nil {
		return err
	}
	return nil
}

func (s *Service) ExecInitScript() error {
	var (
		hostsList []string
		lbHost    string
	)

	if s.infraCluster.ControlPlaneEndpoint().Address != "" {
		lbHost = fmt.Sprintf("%s  %s", s.infraCluster.ControlPlaneEndpoint().Address, s.infraCluster.ControlPlaneEndpoint().Domain)
	}
	for _, host := range s.infraCluster.AllInstancesSpec() {
		if host.Name != "" {
			hostsList = append(hostsList, fmt.Sprintf("%s  %s.%s %s",
				host.InternalAddress,
				host.Name,
				s.infraCluster.KubernetesClusterName(),
				host.Name))
		}
	}
	hostsList = append(hostsList, lbHost)

	svc, err := s.getTemplateFactory(
		templates.InitOsScriptTmpl,
		file.Data{
			"Hosts": hostsList,
		},
		filepath.Join(directory.KubeScriptDir, templates.InitOsScriptTmpl.Name()))
	if err != nil {
		return err
	}
	if err := svc.RenderToLocal(); err != nil {
		return err
	}
	if err := svc.Copy(true); err != nil {
		return err
	}
	if err := svc.Chmod(os.ModeExclusive); err != nil {
		return err
	}
	if _, err := s.SSHClient.SudoCmd(svc.RemotePath()); err != nil {
		return err
	}
	return nil
}
