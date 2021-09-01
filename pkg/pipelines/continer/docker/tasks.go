package docker

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"github.com/pkg/errors"
)

type InstallDocker struct {
	common.KubeAction
}

func (i *InstallDocker) Execute(runtime connector.Runtime) error {
	output, err := runtime.GetRunner().SudoCmd(
		"if [ -z $(which docker) ] || [ ! -e /var/run/docker.sock ]; "+
			"then curl https://kubernetes.pek3b.qingstor.com/tools/kubekey/docker-install.sh | sh && systemctl enable docker; "+
			"systemctl daemon-reload && systemctl restart docker; "+
			"fi", false)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to install docker: %s", output))
	}
	return nil
}

type ReloadDocker struct {
	common.KubeAction
}

func (r *ReloadDocker) Execute(runtime connector.Runtime) error {
	output, err := runtime.GetRunner().SudoCmd("systemctl daemon-reload && systemctl restart docker", false)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to reload docker: %s", output))
	}
	return nil
}
