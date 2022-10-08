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
	"path/filepath"

	mapset "github.com/deckarep/golang-set"
	kubekeyapiv1alpha2 "github.com/kubesphere/kubekey/cmd/kk/apis/kubekey/v1alpha2"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/core/connector"
	"github.com/pkg/errors"
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
		if err := K8sFilesDownloadHTTP(d.KubeConf, runtime.GetWorkDir(), kubeVersion, arch, d.PipelineCache); err != nil {
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
		if err := K3sFilesDownloadHTTP(k.KubeConf, runtime.GetWorkDir(), kubeVersion, arch, k.PipelineCache); err != nil {
			return err
		}
	}
	return nil
}

type K8eDownload struct {
	common.KubeAction
}

func (k *K8eDownload) Execute(runtime connector.Runtime) error {
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
		if err := K8eFilesDownloadHTTP(k.KubeConf, runtime.GetWorkDir(), kubeVersion, arch, k.PipelineCache); err != nil {
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

	kubernetesSet := mapset.NewThreadUnsafeSet()
	for _, k := range manifest.KubernetesDistributions {
		kubernetesSet.Add(k)
	}

	kubernetesVersions := make([]string, 0, kubernetesSet.Cardinality())
	for _, k := range kubernetesSet.ToSlice() {
		k8s := k.(kubekeyapiv1alpha2.KubernetesDistribution)
		kubernetesVersions = append(kubernetesVersions, k8s.Version)
	}

	basePath := filepath.Join(runtime.GetWorkDir(), common.Artifact)
	for arch := range archMap {
		for _, version := range kubernetesVersions {
			if err := KubernetesArtifactBinariesDownload(a.Manifest, basePath, arch, version); err != nil {
				return err
			}
		}

		if err := RegistryBinariesDownload(a.Manifest, basePath, arch); err != nil {
			return err
		}
	}
	return nil
}

type K3sArtifactDownload struct {
	common.ArtifactAction
}

func (a *K3sArtifactDownload) Execute(runtime connector.Runtime) error {
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

	kubernetesSet := mapset.NewThreadUnsafeSet()
	for _, k := range manifest.KubernetesDistributions {
		kubernetesSet.Add(k)
	}

	kubernetesVersions := make([]string, 0, kubernetesSet.Cardinality())
	for _, k := range kubernetesSet.ToSlice() {
		k8s := k.(kubekeyapiv1alpha2.KubernetesDistribution)
		kubernetesVersions = append(kubernetesVersions, k8s.Version)
	}

	basePath := filepath.Join(runtime.GetWorkDir(), common.Artifact)
	for arch := range archMap {
		for _, version := range kubernetesVersions {
			if err := K3sArtifactBinariesDownload(a.Manifest, basePath, arch, version); err != nil {
				return err
			}
		}

		if err := RegistryBinariesDownload(a.Manifest, basePath, arch); err != nil {
			return err
		}
	}
	return nil
}

type K8eArtifactDownload struct {
	common.ArtifactAction
}

func (a *K8eArtifactDownload) Execute(runtime connector.Runtime) error {
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

	kubernetesSet := mapset.NewThreadUnsafeSet()
	for _, k := range manifest.KubernetesDistributions {
		kubernetesSet.Add(k)
	}

	kubernetesVersions := make([]string, 0, kubernetesSet.Cardinality())
	for _, k := range kubernetesSet.ToSlice() {
		k8s := k.(kubekeyapiv1alpha2.KubernetesDistribution)
		kubernetesVersions = append(kubernetesVersions, k8s.Version)
	}

	basePath := filepath.Join(runtime.GetWorkDir(), common.Artifact)
	for arch := range archMap {
		for _, version := range kubernetesVersions {
			if err := K8eArtifactBinariesDownload(a.Manifest, basePath, arch, version); err != nil {
				return err
			}
		}

		if err := RegistryBinariesDownload(a.Manifest, basePath, arch); err != nil {
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
	if err := RegistryPackageDownloadHTTP(k.KubeConf, runtime.GetWorkDir(), arch, k.PipelineCache); err != nil {
		return err
	}

	return nil
}

type CriDownload struct {
	common.KubeAction
}

func (d *CriDownload) Execute(runtime connector.Runtime) error {
	cfg := d.KubeConf.Cluster
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
		if err := CriDownloadHTTP(d.KubeConf, runtime.GetWorkDir(), arch, d.PipelineCache); err != nil {
			return err
		}
	}
	return nil
}
