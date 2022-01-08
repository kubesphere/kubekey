/*
 Copyright 2021 The KubeSphere Authors.

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

package binaries

import (
	"fmt"
	kubekeyapiv1alpha2 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha2"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/pkg/errors"
	"path/filepath"
)

type Download struct {
	common.KubeAction
}

func (d *Download) Execute(runtime connector.Runtime) error {
	cfg := d.KubeConf.Cluster

	var kubeVersion string
	if cfg.Kubernetes.Version == "" {
		kubeVersion = kubekeyapiv1alpha2.DefaultKubeVersion
	} else {
		kubeVersion = cfg.Kubernetes.Version
	}

	archMap := make(map[string]bool)
	for _, host := range cfg.Hosts {
		switch host.Arch {
		case "amd64":
			archMap["amd64"] = true
		case "arm64":
			archMap["arm64"] = true
		default:
			return errors.New(fmt.Sprintf("Unsupported architecture: %s", host.Arch))
		}
	}

	for arch := range archMap {
		binariesDir := filepath.Join(runtime.GetWorkDir(), kubeVersion, arch)
		if err := util.CreateDir(binariesDir); err != nil {
			return errors.Wrap(err, "Failed to create download target dir")
		}

		if err := K8sFilesDownloadHTTP(d.KubeConf, binariesDir, kubeVersion, arch, d.PipelineCache); err != nil {
			return err
		}
	}
	return nil
}

type K3sDownload struct {
	common.KubeAction
}

func (k *K3sDownload) Execute(runtime connector.Runtime) error {
	cfg := k.KubeConf.Cluster

	var kubeVersion string
	if cfg.Kubernetes.Version == "" {
		kubeVersion = kubekeyapiv1alpha2.DefaultKubeVersion
	} else {
		kubeVersion = cfg.Kubernetes.Version
	}

	archMap := make(map[string]bool)
	for _, host := range cfg.Hosts {
		switch host.Arch {
		case "amd64":
			archMap["amd64"] = true
		case "arm64":
			archMap["arm64"] = true
		default:
			return errors.New(fmt.Sprintf("Unsupported architecture: %s", host.Arch))
		}
	}

	for arch := range archMap {
		binariesDir := filepath.Join(runtime.GetWorkDir(), kubeVersion, arch)
		if err := util.CreateDir(binariesDir); err != nil {
			return errors.Wrap(err, "Failed to create download target dir")
		}

		if err := K3sFilesDownloadHTTP(k.KubeConf, binariesDir, kubeVersion, arch, k.PipelineCache); err != nil {
			return err
		}
	}
	return nil
}

type ArtifactDownload struct {
	common.ArtifactAction
}

func (a *ArtifactDownload) Execute(runtime connector.Runtime) error {
	manifest := a.Manifest.Spec

	archMap := make(map[string]bool)
	for _, arch := range manifest.Arches {
		switch arch {
		case "amd64":
			archMap["amd64"] = true
		case "arm64":
			archMap["arm64"] = true
		default:
			return errors.New(fmt.Sprintf("Unsupported architecture: %s", arch))
		}
	}

	for arch := range archMap {
		binariesDir := filepath.Join(runtime.GetWorkDir(), common.Artifact, manifest.KubernetesDistribution.Version, arch)
		if err := util.CreateDir(binariesDir); err != nil {
			return errors.Wrap(err, "Failed to create download target dir")
		}

		if err := KubernetesArtifactBinariesDownload(a.Manifest, binariesDir, arch); err != nil {
			return err
		}
	}
	return nil
}

type RegistryPackageDownload struct {
	common.KubeAction
}

func (k *RegistryPackageDownload) Execute(runtime connector.Runtime) error {
	arch := runtime.GetHostsByRole(common.Registry)[0].GetArch()

	packageDir := filepath.Join(runtime.GetWorkDir(), "registry", arch)
	if err := util.CreateDir(packageDir); err != nil {
		return errors.Wrap(err, "Failed to create download target dir")
	}
	if err := RegistryPackageDownloadHTTP(k.KubeConf, packageDir, arch, k.PipelineCache); err != nil {
		return err
	}

	return nil
}
