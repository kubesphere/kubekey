package container

import (
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"strings"
)

type DockerExist struct {
	common.KubePrepare
	Not bool
}

func (d *DockerExist) PreCheck(runtime connector.Runtime) (bool, error) {
	output, err := runtime.GetRunner().SudoCmd("if [ -z $(which docker) ] || [ ! -e /var/run/docker.sock ]; "+
		"then echo 'not exist'; "+
		"fi", false)
	if err != nil {
		return false, err
	}
	if strings.Contains(output, "not exist") {
		return d.Not, nil
	} else {
		return !d.Not, nil
	}
}

type CrictlExist struct {
	common.KubePrepare
	Not bool
}

func (c *CrictlExist) PreCheck(runtime connector.Runtime) (bool, error) {
	output, err := runtime.GetRunner().SudoCmd(
		"if [ -z $(which crictl) ]; "+
			"then echo 'not exist'; "+
			"fi", false)
	if err != nil {
		return false, err
	}
	if strings.Contains(output, "not exist") {
		return c.Not, nil
	} else {
		return !c.Not, nil
	}
}

type ContainerdExist struct {
	common.KubePrepare
	Not bool
}

func (c *ContainerdExist) PreCheck(runtime connector.Runtime) (bool, error) {
	output, err := runtime.GetRunner().SudoCmd(
		"if [ -z $(which containerd) ] || [ ! -e /run/containerd/containerd.sock ]; "+
			"then echo 'not exist'; "+
			"fi", false)
	if err != nil {
		return false, err
	}
	if strings.Contains(output, "not exist") {
		return c.Not, nil
	} else {
		return !c.Not, nil
	}
}
