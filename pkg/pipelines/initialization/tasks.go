package initialization

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"github.com/pkg/errors"
	"strings"
)

type NodePreCheck struct {
	common.KubeAction
}

func (n *NodePreCheck) Execute(runtime connector.Runtime) error {
	var results = make(map[string]string)
	results["name"] = runtime.RemoteHost().GetName()
	for _, software := range baseSoftware {
		_, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("which %s", software), false)
		switch software {
		case showmount:
			software = nfs
		case rbd:
			software = ceph
		case glusterfs:
			software = glusterfs
		}
		if err != nil {
			results[software] = ""
			logger.Log.Debugf("exec cmd 'which %s' got err return: %v", software, err)
		} else {
			results[software] = "y"
			if software == docker {
				dockerVersion, err := runtime.GetRunner().SudoCmd("docker version --format '{{.Server.Version}}'", false)
				if err != nil {
					results[software] = UnknownVersion
				} else {
					results[software] = dockerVersion
				}
			}
		}
	}

	output, err := runtime.GetRunner().Cmd("date +\"%Z %H:%M:%S\"", false)
	if err != nil {
		results["time"] = ""
	} else {
		results["time"] = strings.TrimSpace(output)
	}

	if res, ok := n.RootCache.Get("nodePreCheck"); ok {
		m := res.(map[string]map[string]string)
		m[runtime.RemoteHost().GetName()] = results
		n.RootCache.Set("nodePreCheck", m)
	} else {
		checkResults := make(map[string]map[string]string)
		checkResults[runtime.RemoteHost().GetName()] = results
		n.RootCache.Set("nodePreCheck", checkResults)
	}
	return nil
}

type NodeConfigureOS struct {
	common.KubeAction
}

func (n *NodeConfigureOS) Execute(runtime connector.Runtime) error {

	host := runtime.RemoteHost()
	_ = addUsers(runtime, host)

	if err := createDirectories(runtime, host); err != nil {
		return err
	}

	tmpDir := "/tmp/kubekey"
	_, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("if [ -d %s ]; then rm -rf %s ;fi && mkdir -p %s", tmpDir, tmpDir, tmpDir), false)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to create tmp dir")
	}

	_, err1 := runtime.GetRunner().SudoCmd(fmt.Sprintf("hostnamectl set-hostname %s && sed -i '/^127.0.1.1/s/.*/127.0.1.1      %s/g' /etc/hosts", host.GetName(), host.GetName()), false)
	if err1 != nil {
		return errors.Wrap(errors.WithStack(err1), "Failed to override hostname")
	}

	return nil
}

func addUsers(runtime connector.Runtime, node connector.Host) error {
	if _, err := runtime.GetRunner().SudoCmd("useradd -M -c 'Kubernetes user' -s /sbin/nologin -r kube || :", false); err != nil {
		return err
	}

	if node.IsRole(common.Etcd) {
		if _, err := runtime.GetRunner().SudoCmd("useradd -M -c 'Etcd user' -s /sbin/nologin -r etcd || :", false); err != nil {
			return err
		}
	}

	return nil
}

func createDirectories(runtime connector.Runtime, node connector.Host) error {
	dirs := []string{
		common.BinDir,
		common.KubeConfigDir,
		common.KubeCertDir,
		common.KubeManifestDir,
		common.KubeScriptDir,
		common.KubeletFlexvolumesPluginsDir,
	}

	for _, dir := range dirs {
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("mkdir -p %s", dir), false); err != nil {
			return err
		}
		if dir == common.KubeletFlexvolumesPluginsDir {
			if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("chown kube -R %s", "/usr/libexec/kubernetes"), false); err != nil {
				return err
			}
		} else {
			if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("chown kube -R %s", dir), false); err != nil {
				return err
			}
		}
	}

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("mkdir -p %s && chown kube -R %s", "/etc/cni/net.d", "/etc/cni"), false); err != nil {
		return err
	}

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("mkdir -p %s && chown kube -R %s", "/opt/cni/bin", "/opt/cni"), false); err != nil {
		return err
	}

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("mkdir -p %s && chown kube -R %s", "/var/lib/calico", "/var/lib/calico"), false); err != nil {
		return err
	}

	if node.IsRole(common.Etcd) {
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("mkdir -p %s && chown etcd -R %s", "/var/lib/etcd", "/var/lib/etcd"), false); err != nil {
			return err
		}
	}

	return nil
}

type NodeExecScript struct {
	common.KubeAction
}

func (n *NodeExecScript) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("chmod +x %s/initOS.sh", "/tmp/kubekey"), false); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to chmod +x init os script")
	}

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("cp %s/initOS.sh %s && sudo %s/initOS.sh", "/tmp/kubekey", common.KubeScriptDir, common.KubeScriptDir), true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to configure operating system")
	}
	return nil
}
