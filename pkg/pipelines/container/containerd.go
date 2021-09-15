package container

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/files"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"github.com/pkg/errors"
	"path/filepath"
)

type SyncCrictlBinaries struct {
	common.KubeAction
}

func (s *SyncCrictlBinaries) Execute(runtime connector.Runtime) error {
	_, err := runtime.GetRunner().SudoCmd(
		fmt.Sprintf("if [ -d %s ]; then rm -rf %s ;fi && mkdir -p %s",
			common.TmpDir, common.TmpDir, common.TmpDir), false)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "reset tmp dir failed")
	}

	binariesMapObj, ok := s.PipelineCache.Get(common.KubeBinaries)
	if !ok {
		return errors.New("get KubeBinary by pipeline cache failed")
	}
	binariesMap := binariesMapObj.(map[string]files.KubeBinary)

	crictl, ok := binariesMap[common.Crictl]
	if !ok {
		return errors.New("get KubeBinary key crictl by pipeline cache failed")
	}
	dst := filepath.Join(common.TmpDir, crictl.Name)

	if err := runtime.GetRunner().SudoScp(crictl.Path, dst); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("sync crictl binaries failed"))
	}

	if _, err := runtime.GetRunner().SudoCmd(
		fmt.Sprintf("mkdir -p /usr/bin && tar -zxf %s -C /usr/bin ", dst),
		false); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("install crictl binaries failed"))
	}
	return nil
}

type EnableContainerd struct {
	common.KubeAction
}

func (e *EnableContainerd) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd(
		"systemctl daemon-reload && systemctl enable containerd && systemctl start containerd",
		false); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("enable and start containerd failed"))
	}
	return nil
}
