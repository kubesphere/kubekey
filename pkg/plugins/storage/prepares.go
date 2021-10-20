package storage

import (
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/pkg/errors"
	"regexp"
)

type CheckDefaultStorageClass struct {
	common.KubePrepare
}

func (c *CheckDefaultStorageClass) PreCheck(runtime connector.Runtime) (bool, error) {
	output, err := runtime.GetRunner().SudoCmd(
		"/usr/local/bin/kubectl get sc --no-headers | grep '(default)' | wc -l", false)
	if err != nil {
		return false, errors.Wrap(errors.WithStack(err), "check default storageClass failed")
	}

	reg := regexp.MustCompile(`([\d])`)
	defaultStorageClassNum := reg.FindStringSubmatch(output)[0]
	if defaultStorageClassNum == "0" {
		return true, nil
	}
	host := runtime.RemoteHost()
	logger.Log.Messagef(host.GetName(), "Default storageClass in cluster is not unique!")
	return false, nil
}
