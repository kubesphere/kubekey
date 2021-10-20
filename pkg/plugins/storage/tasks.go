package storage

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/pkg/errors"
	"path/filepath"
)

type DeployLocalVolume struct {
	common.KubeAction
}

func (d *DeployLocalVolume) Execute(runtime connector.Runtime) error {
	cmd := fmt.Sprintf("/usr/local/bin/kubectl apply -f %s", filepath.Join(common.KubeAddonsDir, "local-volume.yaml"))
	if _, err := runtime.GetRunner().SudoCmd(cmd, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "deploy local-volume.yaml failed")
	}
	return nil
}
