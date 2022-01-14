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
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/files"
	"github.com/pkg/errors"
	"os/exec"
)

// K3sFilesDownloadHTTP defines the kubernetes' binaries that need to be downloaded in advance and downloads them.
func K3sFilesDownloadHTTP(kubeConf *common.KubeConf, path, version, arch string, pipelineCache *cache.Cache) error {

	etcd := files.NewKubeBinary("etcd", arch, kubekeyapiv1alpha2.DefaultEtcdVersion, path, kubeConf.Arg.DownloadCommand)
	kubecni := files.NewKubeBinary("kubecni", arch, kubekeyapiv1alpha2.DefaultCniVersion, path, kubeConf.Arg.DownloadCommand)
	helm := files.NewKubeBinary("helm", arch, kubekeyapiv1alpha2.DefaultHelmVersion, path, kubeConf.Arg.DownloadCommand)
	k3s := files.NewKubeBinary("k3s", arch, version, path, kubeConf.Arg.DownloadCommand)

	binaries := []*files.KubeBinary{k3s, helm, kubecni, etcd}
	binariesMap := make(map[string]*files.KubeBinary)
	for _, binary := range binaries {
		if err := binary.CreateBaseDir(); err != nil {
			return errors.Wrapf(errors.WithStack(err), "create file %s base dir failed", binary.FileName)
		}

		logger.Log.Messagef(common.LocalHost, "downloading %s %s %s ...", arch, binary.ID, binary.Version)

		binariesMap[binary.ID] = binary
		if util.IsExist(binary.Path()) {
			// download it again if it's incorrect
			p := binary.Path()
			if err := binary.SHA256Check(); err != nil {
				_ = exec.Command("/bin/sh", "-c", fmt.Sprintf("rm -f %s", p)).Run()
			} else {
				continue
			}
		}

		if err := binary.Download(); err != nil {
			return fmt.Errorf("Failed to download %s binary: %s error: %w ", binary.ID, binary.GetCmd(), err)
		}
	}

	pipelineCache.Set(common.KubeBinaries+"-"+arch, binariesMap)
	return nil
}
