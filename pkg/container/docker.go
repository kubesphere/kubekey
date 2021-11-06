package container

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/files"
	"github.com/kubesphere/kubekey/pkg/utils"
	"github.com/pkg/errors"
	"path/filepath"
)

type SyncDockerBinaries struct {
	common.KubeAction
}

func (s *SyncDockerBinaries) Execute(runtime connector.Runtime) error {
	if err := utils.ResetTmpDir(runtime); err != nil {
		return err
	}

	binariesMapObj, ok := s.PipelineCache.Get(common.KubeBinaries)
	if !ok {
		return errors.New("get KubeBinary by pipeline cache failed")
	}
	binariesMap := binariesMapObj.(map[string]files.KubeBinary)

	docker, ok := binariesMap[common.Docker]
	if !ok {
		return errors.New("get KubeBinary key docker by pipeline cache failed")
	}
	dst := filepath.Join(common.TmpDir, docker.Name)

	if err := runtime.GetRunner().Scp(docker.Path, dst); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("sync docker binaries failed"))
	}

	if _, err := runtime.GetRunner().SudoCmd(
		fmt.Sprintf("mkdir -p /usr/bin && tar -zxf %s && mv docker/* /usr/bin && rm -rf docker", dst),
		false); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("install container runtime docker binaries failed"))
	}
	return nil
}

type EnableDocker struct {
	common.KubeAction
}

func (e *EnableDocker) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd(
		"systemctl daemon-reload && systemctl enable docker && systemctl start docker",
		false); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("enable and start docker failed"))
	}
	return nil
}
