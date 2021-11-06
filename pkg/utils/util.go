package utils

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/pkg/errors"
)

func ResetTmpDir(runtime connector.Runtime) error {
	_, err := runtime.GetRunner().SudoCmd(fmt.Sprintf(
		"if [ -d %s ]; then rm -rf %s ;fi && mkdir -m 777 -p %s",
		common.TmpDir, common.TmpDir, common.TmpDir), false)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "reset tmp dir failed")
	}
	return nil
}
