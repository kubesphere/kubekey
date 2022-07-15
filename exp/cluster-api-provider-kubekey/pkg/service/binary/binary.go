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

package binary

import (
	"path/filepath"

	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/service/operation"
	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/service/operation/file"
)

func (s *Service) DownloadAll() error {
	kubeadm, err := s.getKubeadmService(s.SSHClient, s.instanceScope.KubernetesVersion(), s.instanceScope.Arch())
	if err != nil {
		return err
	}
	kubelet, err := s.getKubeletService(s.SSHClient, s.instanceScope.KubernetesVersion(), s.instanceScope.Arch())
	if err != nil {
		return err
	}
	kubecni, err := s.getKubecniService(s.SSHClient, file.KubecniDefaultVersion, s.instanceScope.Arch())
	if err != nil {
		return err
	}
	kubectl, err := s.getKubectlService(s.SSHClient, s.instanceScope.KubernetesVersion(), s.instanceScope.Arch())
	if err != nil {
		return err
	}

	binaries := []operation.Binary{
		kubeadm,
		kubelet,
		kubecni,
		kubectl,
	}

	for _, b := range binaries {
		skipGet := false
		if b.LocalExist() {
			if err := b.CompareChecksum(); err == nil {
				skipGet = true
			}
		}
		if !skipGet {
			if err := b.Get(); err != nil {
				return err
			}
			if err := b.CompareChecksum(); err != nil {
				return err
			}
		}
		if err := b.Copy(true); err != nil {
			return err
		}
	}

	if _, err := s.SSHClient.SudoCmdf("tar Cxzvf %s %s", filepath.Dir(kubecni.RemotePath()), kubecni.RemotePath()); err != nil {
		return err
	}

	return nil
}
