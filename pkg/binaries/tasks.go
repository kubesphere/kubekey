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
