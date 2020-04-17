package preinstall

import (
	"encoding/base64"
	"fmt"
	kubekeyapi "github.com/pixiake/kubekey/apis/v1alpha1"
	"github.com/pixiake/kubekey/cluster/preinstall/tmpl"
	"github.com/pixiake/kubekey/util/manager"
	"github.com/pixiake/kubekey/util/ssh"
	"github.com/pkg/errors"
)

func InitOS(mgr *manager.Manager) error {
	mgr.Logger.Infoln("Initialize operating system")

	return mgr.RunTaskOnAllNodes(initOsOnNode, false)
}

func initOsOnNode(mgr *manager.Manager, node *kubekeyapi.HostCfg, conn ssh.Connection) error {
	tmpDir := "/tmp/kubekey"
	_, err := mgr.Runner.RunCmd(fmt.Sprintf("mkdir -p %s", tmpDir))
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to init operating system")
	}

	initOsScript, err1 := tmpl.InitOsScript(mgr.Cluster)
	if err1 != nil {
		return err1
	}

	str := base64.StdEncoding.EncodeToString([]byte(initOsScript))
	_, err2 := mgr.Runner.RunCmd(fmt.Sprintf("echo %s | base64 -d > %s/initOS.sh && chmod +x %s/initOS.sh", str, tmpDir, tmpDir))
	if err2 != nil {
		return errors.Wrap(errors.WithStack(err2), "failed to init operating system")
	}

	_, err3 := mgr.Runner.RunCmd(fmt.Sprintf("sudo %s/initOS.sh", tmpDir))
	if err3 != nil {
		return errors.Wrap(errors.WithStack(err3), "failed to init operating system")
	}
	return nil
}
