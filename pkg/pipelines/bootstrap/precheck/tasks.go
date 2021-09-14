package precheck

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
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

	host := runtime.RemoteHost()
	if res, ok := host.GetCache().Get(common.NodePreCheck); ok {
		m := res.(map[string]string)
		m = results
		host.GetCache().Set(common.NodePreCheck, m)
	} else {
		host.GetCache().Set(common.NodePreCheck, results)
	}
	return nil
}
