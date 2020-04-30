package preinstall

import (
	"encoding/base64"
	"fmt"
	kubekeyapi "github.com/pixiake/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/pixiake/kubekey/pkg/cluster/preinstall/tmpl"
	"github.com/pixiake/kubekey/pkg/util/manager"
	"github.com/pixiake/kubekey/pkg/util/ssh"
	"github.com/pkg/errors"
)

func InitOS(mgr *manager.Manager) error {
	mgr.Logger.Infoln("Initialize operating system")

	return mgr.RunTaskOnAllNodes(initOsOnNode, false)
}

func initOsOnNode(mgr *manager.Manager, node *kubekeyapi.HostCfg, conn ssh.Connection) error {
	tmpDir := "/tmp/kubekey"
	_, err := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"if [ -d %s ]; then rm -rf %s ;fi\" && mkdir -p %s", tmpDir, tmpDir, tmpDir))
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to init operating system")
	}

	_, err1 := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"hostnamectl set-hostname %s && sed -i '/^127.0.1.1/s/.*/127.0.1.1      %s/g' /etc/hosts\"", node.HostName, node.HostName))
	if err1 != nil {
		return errors.Wrap(errors.WithStack(err1), "failed to override hostname")
	}

	initOsScript, err2 := tmpl.InitOsScript(mgr.Cluster)
	if err2 != nil {
		return err2
	}

	str := base64.StdEncoding.EncodeToString([]byte(initOsScript))
	_, err3 := mgr.Runner.RunCmd(fmt.Sprintf("echo %s | base64 -d > %s/initOS.sh && chmod +x %s/initOS.sh", str, tmpDir, tmpDir))
	if err3 != nil {
		return errors.Wrap(errors.WithStack(err3), "failed to init operating system")
	}

	_, err4 := mgr.Runner.RunCmd(fmt.Sprintf("sudo %s/initOS.sh", tmpDir))
	if err4 != nil {
		return errors.Wrap(errors.WithStack(err4), "failed to init operating system")
	}
	return nil
}
