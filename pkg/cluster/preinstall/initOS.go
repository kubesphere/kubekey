/*
Copyright 2020 The KubeSphere Authors.

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

package preinstall

import (
	"encoding/base64"
	"fmt"
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/cluster/preinstall/tmpl"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/kubesphere/kubekey/pkg/util/ssh"
	"github.com/pkg/errors"
)

func InitOS(mgr *manager.Manager) error {
	PrecheckConfirm(mgr)
	if err := Prepare(mgr); err != nil {
		return errors.Wrap(err, "Failed to load kube binaries")
	}

	mgr.Logger.Infoln("Configurating operating system ...")

	return mgr.RunTaskOnAllNodes(initOsOnNode, true)
}

func initOsOnNode(mgr *manager.Manager, node *kubekeyapi.HostCfg, conn ssh.Connection) error {
	tmpDir := "/tmp/kubekey"
	_, err := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"if [ -d %s ]; then rm -rf %s ;fi\" && mkdir -p %s", tmpDir, tmpDir, tmpDir))
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to configure operating system")
	}

	_, err1 := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"hostnamectl set-hostname %s && sed -i '/^127.0.1.1/s/.*/127.0.1.1      %s/g' /etc/hosts\"", node.Name, node.Name))
	if err1 != nil {
		return errors.Wrap(errors.WithStack(err1), "Failed to override hostname")
	}

	initOsScript, err2 := tmpl.InitOsScript(mgr)
	if err2 != nil {
		return err2
	}

	str := base64.StdEncoding.EncodeToString([]byte(initOsScript))
	_, err3 := mgr.Runner.RunCmd(fmt.Sprintf("echo %s | base64 -d > %s/initOS.sh && chmod +x %s/initOS.sh", str, tmpDir, tmpDir))
	if err3 != nil {
		return errors.Wrap(errors.WithStack(err3), "Failed to configure operating system")
	}

	_, err4 := mgr.Runner.RunCmd(fmt.Sprintf("sudo %s/initOS.sh", tmpDir))
	if err4 != nil {
		return errors.Wrap(errors.WithStack(err4), "Failed to configure operating system")
	}
	return nil
}
