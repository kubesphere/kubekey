package initialization

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/config"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/kubesphere/kubekey/pkg/core/vars"
	"strings"
)

type NodePreCheck struct {
	action.BaseAction
}

func (n *NodePreCheck) Execute(runtime *config.Runtime, vars vars.Vars) error {
	var results = make(map[string]string)
	results["name"] = runtime.Runner.Host.Name
	for _, software := range baseSoftware {
		_, err := runtime.Runner.SudoCmd(fmt.Sprintf("which %s", software), false)
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
				dockerVersion, err := runtime.Runner.SudoCmd("docker version --format '{{.Server.Version}}'", false)
				if err != nil {
					results[software] = UnknownVersion
				} else {
					results[software] = dockerVersion
				}
			}
		}
	}

	output, err := runtime.Runner.Cmd("date +\"%Z %H:%M:%S\"", false)
	if err != nil {
		results["time"] = ""
	} else {
		results["time"] = strings.TrimSpace(output)
	}

	if res, ok := n.RootCache.Get("nodePreCheck"); ok {
		m := res.(map[string]map[string]string)
		m[runtime.Runner.Host.Name] = results
		n.RootCache.Set("nodePreCheck", m)
	} else {
		checkResults := make(map[string]map[string]string)
		checkResults[runtime.Runner.Host.Name] = results
		n.RootCache.Set("nodePreCheck", checkResults)
	}
	return nil
}
