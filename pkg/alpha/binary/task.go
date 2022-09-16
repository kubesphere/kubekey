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

type GetEtcdBinaryPath struct {
	common.KubeAction
}

func (g *GetEtcdBinaryPath) Execute(runtime connector.Runtime) error {
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
		if err := setEtcdBinaryPath(g.KubeConf, runtime.GetWorkDir(), arch, g.PipelineCache); err != nil {
			return err
		}
	}
	return nil
}

func setEtcdBinaryPath(kubeConf *common.KubeConf, path, arch string, pipelineCache *cache.Cache) error {
	binary := "etcd"
	binariesMap := make(map[string]*files.KubeBinary)
	kubeBinary := files.NewKubeBinary(binary, arch, kubekeyapiv1alpha2.DefaultEtcdVersion, path, kubeConf.Arg.DownloadCommand)
	binariesMap[kubeBinary.ID] = kubeBinary
	pipelineCache.Set(common.KubeBinaries+"-"+arch, binariesMap)
	return nil
}
