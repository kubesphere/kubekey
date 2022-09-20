package binary

import (
	"fmt"

	kubekeyapiv1alpha2 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha2"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/files"
	"github.com/pkg/errors"
)

type GetBinaryPath struct {
	common.KubeAction
	Binaries []string
}

func (g *GetBinaryPath) Execute(runtime connector.Runtime) error {
	cfg := g.KubeConf.Cluster

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
		if err := setBinaryPath(g.KubeConf, runtime.GetWorkDir(), arch, g.Binaries, g.PipelineCache); err != nil {
			return err
		}
	}
	return nil
}

func setBinaryPath(kubeConf *common.KubeConf, path, arch string, binaries []string, pipelineCache *cache.Cache) error {
	binariesMap := make(map[string]*files.KubeBinary)
	var kubeBinary *files.KubeBinary
	for _, binary := range binaries {
		switch binary {
		case "etcd":
			kubeBinary = files.NewKubeBinary(binary, arch, kubekeyapiv1alpha2.DefaultEtcdVersion, path, kubeConf.Arg.DownloadCommand)
		case "docker":
			kubeBinary = files.NewKubeBinary(binary, arch, kubekeyapiv1alpha2.DefaultDockerVersion, path, kubeConf.Arg.DownloadCommand)
		case "containerd":
			kubeBinary = files.NewKubeBinary(binary, arch, kubekeyapiv1alpha2.DefaultContainerdVersion, path, kubeConf.Arg.DownloadCommand)
		case "helm":
			kubeBinary = files.NewKubeBinary(binary, arch, kubekeyapiv1alpha2.DefaultHelmVersion, path, kubeConf.Arg.DownloadCommand)
		case "crictl":
			kubeBinary = files.NewKubeBinary("crictl", arch, kubekeyapiv1alpha2.DefaultCrictlVersion, path, kubeConf.Arg.DownloadCommand)
		case "runc":
			kubeBinary = files.NewKubeBinary("runc", arch, kubekeyapiv1alpha2.DefaultRuncVersion, path, kubeConf.Arg.DownloadCommand)
		default:
			return errors.New(fmt.Sprintf("Unsupported binary name: %s", binary))
		}
		binariesMap[kubeBinary.ID] = kubeBinary
	}
	pipelineCache.Set(common.KubeBinaries+"-"+arch, binariesMap)
	return nil
}
