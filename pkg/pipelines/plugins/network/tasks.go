package network

import (
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"github.com/pkg/errors"
)

type DeployNetworkPlugin struct {
	common.KubeAction
}

func (d *DeployNetworkPlugin) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd(
		"/usr/local/bin/kubectl apply -f /etc/kubernetes/network-plugin.yaml --force", true); err != nil {
		return errors.Wrap(errors.WithStack(err), "deploy network plugin failed")
	}
	return nil
}
